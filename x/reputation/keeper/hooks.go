package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/blong-dev/dreamtree/x/reputation"
)

// The x/attest → reputation hooks. Events don't move R directly — they enqueue
// PendingEvents that settle at their review-window close (see window.go).

func proofTypeToBucket(proofType int32) (reputation.RateBucket, bool) {
	switch proofType {
	case 1: // ORIGIN
		return reputation.RateBucket_RATE_BUCKET_PERMANENT, true
	case 2: // RIGOR
		return reputation.RateBucket_RATE_BUCKET_RIGOR, true
	case 3: // USE
		return reputation.RateBucket_RATE_BUCKET_USE, true
	case 4: // REPLICATION
		return reputation.RateBucket_RATE_BUCKET_REPLICATION, true
	default: // OUTCOME / unspecified — no direct bet
		return reputation.RateBucket_RATE_BUCKET_UNSPECIFIED, false
	}
}

// OnAttestation enqueues an unvalidated "bet" for a newly recorded attestation:
// attest_bet_scale × specificity, at the proof-type bucket, through a review
// window (trivial magnitude → τ≈0 → settles next block).
func (k Keeper) OnAttestation(ctx context.Context, signer, domain string, proofType int32, specificityBps uint32, sourceAttID uint64) error {
	bucket, ok := proofTypeToBucket(proofType)
	if !ok {
		return nil
	}
	p, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}
	magnitude := d(p.AttestBetScale).Mul(specDec(specificityBps))
	_, err = k.enqueue(ctx, reputation.PendingEvent{
		Kind:          reputation.PendingKind_PENDING_KIND_BET,
		Signer:        signer,
		Domain:        domain,
		RateBucket:    bucket,
		BaseMagnitude: magnitude,
		Corroboration: math.LegacyZeroDec(),
		Refutation:    math.LegacyZeroDec(),
		SourceAttId:   sourceAttID,
	})
	return err
}

// OnOutcome handles an OUTCOME attestation. Non-reversal outcomes open (or
// corroborate/refute) a review window on the target attestation; at settlement
// the integrated M_O propagates to the contributor. A reversal (target is
// itself an outcome) negates the overturned outcome + penalizes its reporter.
func (k Keeper) OnOutcome(ctx context.Context, reporter string, refutes bool, targetAttID uint64, targetAttestor, targetDomain string, targetSIssuance math.LegacyDec, targetIsOutcome bool, propTargets []reputation.PropTarget, sourceAttID uint64) error {
	p, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}
	// M_O = min(M_cap, β · S_issuance · √cred(reporter)).  M_cap = cap_mult · S_issuance.
	cred := k.credOf(ctx, reporter, targetDomain)
	sqrtCred, err := cred.ApproxSqrt()
	if err != nil {
		return err
	}
	mO := d(p.OutcomeBeta).Mul(targetSIssuance).Mul(sqrtCred)
	capM := d(p.OutcomeCapMult).Mul(targetSIssuance)
	if mO.GT(capM) {
		mO = capM
	}
	if !mO.IsPositive() {
		return nil
	}

	if targetIsOutcome {
		// Reversal: enqueue a counter-outcome that overturns targetAttID.
		_, err = k.enqueue(ctx, reputation.PendingEvent{
			Kind:                 reputation.PendingKind_PENDING_KIND_OUTCOME,
			TargetAttId:          targetAttID,
			TargetAttestor:       targetAttestor, // the overturned outcome's reporter
			TargetDomain:         targetDomain,
			TargetSIssuance:      targetSIssuance,
			OutcomeRefutes:       refutes,
			CounterTargetPending: targetAttID, // non-zero ⇒ reversal
			BaseMagnitude:        mO,
			Corroboration:        math.LegacyZeroDec(),
			Refutation:           math.LegacyZeroDec(),
			SourceAttId:          sourceAttID,
		})
		return err
	}

	// Is a window already open on this target? Then corroborate / refute it.
	if openID, err := k.PendingByTarget.Get(ctx, targetAttID); err == nil {
		pe, err := k.Pending.Get(ctx, openID)
		if err == nil {
			capPool := pe.BaseMagnitude.MulInt64(4)
			if refutes == pe.OutcomeRefutes {
				pe.Corroboration = paperShapeAdd(pe.Corroboration, mO, capPool)
			} else {
				pe.Refutation = paperShapeAdd(pe.Refutation, mO, capPool)
			}
			return k.Pending.Set(ctx, openID, pe)
		}
	}

	// Otherwise open a fresh window.
	_, err = k.enqueue(ctx, reputation.PendingEvent{
		Kind:            reputation.PendingKind_PENDING_KIND_OUTCOME,
		TargetAttId:     targetAttID,
		TargetAttestor:  targetAttestor,
		TargetDomain:    targetDomain,
		TargetSIssuance: targetSIssuance,
		OutcomeRefutes:  refutes,
		BaseMagnitude:   mO,
		Corroboration:   math.LegacyZeroDec(),
		Refutation:      math.LegacyZeroDec(),
		PropTargets:     propTargets,
		SourceAttId:     sourceAttID,
	})
	return err
}

// OnEndorsement credits the endorsed B with endorse_inherit × the endorser's
// rational standing, in the ENDORSEMENT bucket, through a review window. Because
// StandingOf already sums endorsement contributions, geometric multi-hop
// inheritance and the 2-hop cred recursion emerge automatically — and terminate
// (contributions are snapshots, not runtime recursion). Self-endorsement is
// rejected upstream, so a shell cannot bootstrap itself.
func (k Keeper) OnEndorsement(ctx context.Context, endorser, endorsed, domain string, sourceAttID uint64) error {
	p, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}
	magnitude := d(p.EndorseInherit).Mul(k.StandingOf(ctx, endorser, domain))
	if !magnitude.IsPositive() {
		return nil
	}
	// Settles identically to a bet: credit `endorsed`, ENDORSEMENT bucket.
	_, err = k.enqueue(ctx, reputation.PendingEvent{
		Kind:          reputation.PendingKind_PENDING_KIND_BET,
		Signer:        endorsed,
		Domain:        domain,
		RateBucket:    reputation.RateBucket_RATE_BUCKET_ENDORSEMENT,
		BaseMagnitude: magnitude,
		Corroboration: math.LegacyZeroDec(),
		Refutation:    math.LegacyZeroDec(),
		SourceAttId:   sourceAttID,
	})
	return err
}

func specDec(bps uint32) math.LegacyDec {
	if bps == 0 {
		bps = 10000
	}
	return math.LegacyNewDec(int64(bps)).QuoInt64(10000)
}

// addContribution writes a settled contribution + its indexes (signer, source).
func (k Keeper) addContribution(ctx context.Context, signer, domain string, magnitude math.LegacyDec, bucket reputation.RateBucket, sourceAttID uint64) error {
	id, err := k.Seq.Next(ctx)
	if err != nil {
		return err
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	c := reputation.Contribution{
		Id:          id,
		Signer:      signer,
		Domain:      domain,
		Magnitude:   magnitude,
		RateBucket:  bucket,
		SettledAt:   sdkCtx.BlockTime().Unix(),
		SourceAttId: sourceAttID,
		Height:      sdkCtx.BlockHeight(),
	}
	if err := k.Contributions.Set(ctx, id, c); err != nil {
		return err
	}
	if err := k.SignerIndex.Set(ctx, collections.Join(signer, id)); err != nil {
		return err
	}
	if sourceAttID != 0 {
		if err := k.SourceIndex.Set(ctx, collections.Join(sourceAttID, id)); err != nil {
			return err
		}
	}
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"reputation_contribution",
		sdk.NewAttribute("id", fmt.Sprintf("%d", id)),
		sdk.NewAttribute("signer", signer),
		sdk.NewAttribute("domain", domain),
		sdk.NewAttribute("magnitude", magnitude.String()),
		sdk.NewAttribute("bucket", bucket.String()),
	))
	return nil
}
