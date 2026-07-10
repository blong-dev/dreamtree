package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/blong-dev/dreamtree/x/reputation"
)

// proofTypeToBucket maps an x/attest ProofType enum value to the decay bucket an
// unvalidated "bet" contribution decays at. (Enum values mirror x/attest;
// documented coupling — see docs/specs/x-reputation-design.md §6.)
func proofTypeToBucket(proofType int32) (reputation.RateBucket, bool) {
	switch proofType {
	case 1: // ORIGIN — authorship never becomes false
		return reputation.RateBucket_RATE_BUCKET_PERMANENT, true
	case 2: // RIGOR
		return reputation.RateBucket_RATE_BUCKET_RIGOR, true
	case 3: // USE
		return reputation.RateBucket_RATE_BUCKET_USE, true
	case 4: // REPLICATION
		return reputation.RateBucket_RATE_BUCKET_REPLICATION, true
	default: // OUTCOME (5) / unspecified — no direct bet contribution in P1
		return reputation.RateBucket_RATE_BUCKET_UNSPECIFIED, false
	}
}

// OnAttestation is the hook x/attest calls when an attestation is recorded. P1:
// making an attestation is an unvalidated bet — a small contribution to the
// attestor's R in that domain, at the proof-type decay bucket. Magnitude is
// attest_bet_scale × specificity, all fixed-point (consensus-safe). The big,
// durable moves come from validated outcomes (P3).
func (k Keeper) OnAttestation(ctx context.Context, signer, domain string, proofType int32, specificityBps uint32, sourceAttID uint64) error {
	bucket, ok := proofTypeToBucket(proofType)
	if !ok {
		return nil
	}
	p, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}
	scale, err := math.LegacyNewDecFromStr(p.AttestBetScale)
	if err != nil {
		return err
	}
	bps := specificityBps
	if bps == 0 {
		bps = 10000 // unset = fully specific
	}
	spec := math.LegacyNewDec(int64(bps)).QuoInt64(10000)
	magnitude := scale.Mul(spec)
	return k.addContribution(ctx, signer, domain, magnitude, bucket, sourceAttID)
}

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
