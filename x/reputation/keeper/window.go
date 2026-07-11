package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/blong-dev/dreamtree/x/reputation"
)

// Review windows + settlement. An event is enqueued as a PendingEvent with a
// close_time = now + τ(M); EndBlock drains matured events and settles them into
// Contributions. All arithmetic here is LegacyDec (fixed-point) — consensus math,
// no transcendental. τ uses ApproxSqrt (deterministic Newton iteration).

const secondsPerDay = 86400

func d(s string) math.LegacyDec { return math.LegacyMustNewDecFromStr(s) }

// tauSeconds = review_window_base(days) × sqrt(M / threshold) × 86400.
func tauSeconds(p reputation.Params, m math.LegacyDec) int64 {
	if !m.IsPositive() {
		return 0
	}
	base := d(p.ReviewWindowBase)
	threshold := d(p.ReviewWindowThreshold)
	if !threshold.IsPositive() {
		threshold = math.LegacyOneDec()
	}
	root, err := m.Quo(threshold).ApproxSqrt()
	if err != nil {
		return 0
	}
	tauDays := base.Mul(root)
	return tauDays.MulInt64(secondsPerDay).TruncateInt64()
}

// paperShapeAdd aggregates x into agg with diminishing returns, capped:
// agg' = cap·[1 − (1 − agg/cap)(1 − x/cap)]. Sybil-resistant by construction.
func paperShapeAdd(agg, x, cap math.LegacyDec) math.LegacyDec {
	if !cap.IsPositive() {
		return agg.Add(x)
	}
	one := math.LegacyOneDec()
	fa := one.Sub(agg.Quo(cap))
	fx := one.Sub(x.Quo(cap))
	if fx.IsNegative() {
		fx = math.LegacyZeroDec()
	}
	return cap.Mul(one.Sub(fa.Mul(fx)))
}

// enqueue assigns a pending id + close_time and stores the event with its indexes.
func (k Keeper) enqueue(ctx context.Context, pe reputation.PendingEvent) (uint64, error) {
	p, err := k.Params.Get(ctx)
	if err != nil {
		return 0, err
	}
	id, err := k.PendingSeq.Next(ctx)
	if err != nil {
		return 0, err
	}
	now := sdk.UnwrapSDKContext(ctx).BlockTime().Unix()
	pe.Id = id
	pe.OpenedAt = now
	pe.CloseTime = now + tauSeconds(p, pe.BaseMagnitude)
	if err := k.Pending.Set(ctx, id, pe); err != nil {
		return 0, err
	}
	if err := k.CloseTimeIndex.Set(ctx, collections.Join(pe.CloseTime, id)); err != nil {
		return 0, err
	}
	// A non-reversal outcome opens the review window for its target.
	if pe.Kind == reputation.PendingKind_PENDING_KIND_OUTCOME && pe.CounterTargetPending == 0 {
		if err := k.PendingByTarget.Set(ctx, pe.TargetAttId, id); err != nil {
			return 0, err
		}
	}
	return id, nil
}

// EndBlock settles every pending event whose window has closed.
func (k Keeper) EndBlock(ctx context.Context) error {
	now := sdk.UnwrapSDKContext(ctx).BlockTime().Unix()
	// close_time in [0, now] — matured. Ordered by (close_time, id): deterministic.
	rng := new(collections.Range[collections.Pair[int64, uint64]]).
		EndInclusive(collections.Join(now, ^uint64(0)))
	var matured []collections.Pair[int64, uint64]
	if err := k.CloseTimeIndex.Walk(ctx, rng, func(key collections.Pair[int64, uint64]) (bool, error) {
		matured = append(matured, key)
		return false, nil
	}); err != nil {
		return err
	}
	for _, key := range matured {
		id := key.K2()
		pe, err := k.Pending.Get(ctx, id)
		if err != nil {
			continue
		}
		if err := k.settle(ctx, pe); err != nil {
			return err
		}
		_ = k.Pending.Remove(ctx, id)
		_ = k.CloseTimeIndex.Remove(ctx, key)
		if pe.Kind == reputation.PendingKind_PENDING_KIND_OUTCOME && pe.CounterTargetPending == 0 {
			_ = k.PendingByTarget.Remove(ctx, pe.TargetAttId)
		}
	}
	return nil
}

// claimStrength integrates a window: paper-shape(base, corroboration) − neg×refutation, ≥ 0.
func (k Keeper) claimStrength(p reputation.Params, pe reputation.PendingEvent) math.LegacyDec {
	// The paper-shape ceiling is M_cap = cap_mult × S_issuance (spec's hard 5×
	// bound on total outcome magnitude — corroboration can't breach it).
	capM := d(p.OutcomeCapMult).Mul(pe.TargetSIssuance)
	if !capM.IsPositive() {
		capM = pe.BaseMagnitude
	}
	support := paperShapeAdd(pe.BaseMagnitude, pe.Corroboration, capM)
	net := support.Sub(d(p.NegAsymmetry).Mul(pe.Refutation))
	if net.IsNegative() {
		return math.LegacyZeroDec()
	}
	return net
}

func (k Keeper) settle(ctx context.Context, pe reputation.PendingEvent) error {
	p, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}
	switch pe.Kind {
	case reputation.PendingKind_PENDING_KIND_BET:
		// A bet settles at its face magnitude (no accumulators in practice).
		return k.addContribution(ctx, pe.Signer, pe.Domain, pe.BaseMagnitude, pe.RateBucket, pe.SourceAttId)

	case reputation.PendingKind_PENDING_KIND_OUTCOME:
		strength := k.claimStrength(p, pe)
		if pe.CounterTargetPending != 0 {
			// Reversal: negate the overturned outcome's contributions + 2× penalty
			// to its reporter (target_attestor).
			return k.reverse(ctx, p, pe, strength)
		}
		if !strength.IsPositive() {
			return nil // refuted into nothing
		}
		// Apply to the contributor (target's author), durable. Validated → +;
		// refuted → −neg× (bad work hits the author harder).
		delta := strength
		if pe.OutcomeRefutes {
			delta = strength.Mul(d(p.NegAsymmetry)).Neg()
		}
		if err := k.addContribution(ctx, pe.TargetAttestor, pe.TargetDomain, delta,
			reputation.RateBucket_RATE_BUCKET_DURABLE_25Y, pe.SourceAttId); err != nil {
			return err
		}
		// Propagate to co-attestors and the contributor's endorsers (liability).
		// Their move follows the outcome's sign (no extra 2× — that's the
		// author's penalty), scaled by the kind's weight.
		return k.propagate(ctx, p, pe, strength)
	}
	return nil
}

// propagate applies an outcome to the captured prop targets (co-attestors +
// endorsers). Co-attestor weight = coattestor_weight × specificity; endorser
// weight = endorse_inherit. Sign follows validated(+)/refuted(−).
func (k Keeper) propagate(ctx context.Context, p reputation.Params, pe reputation.PendingEvent, strength math.LegacyDec) error {
	for _, pt := range pe.PropTargets {
		var weight math.LegacyDec
		switch pt.Kind {
		case reputation.PropKind_PROP_KIND_COATTESTOR:
			weight = d(p.CoattestorWeight).Mul(pt.BaseFactor)
		case reputation.PropKind_PROP_KIND_ENDORSER:
			weight = d(p.EndorseInherit).Mul(pt.BaseFactor)
		default:
			continue
		}
		delta := strength.Mul(weight)
		if pe.OutcomeRefutes {
			delta = delta.Neg()
		}
		if delta.IsZero() {
			continue
		}
		if err := k.addContribution(ctx, pt.Address, pt.Domain, delta,
			reputation.RateBucket_RATE_BUCKET_DURABLE_25Y, pe.SourceAttId); err != nil {
			return err
		}
	}
	return nil
}

// reverse un-applies a settled outcome's contributions and penalizes its reporter.
func (k Keeper) reverse(ctx context.Context, p reputation.Params, pe reputation.PendingEvent, strength math.LegacyDec) error {
	overturned := pe.TargetAttId // the outcome attestation being countered
	// Idempotent: an outcome can only be overturned once. Parallel counter-outcomes
	// on the same outcome must not double-negate its contributions.
	if has, _ := k.Reversed.Has(ctx, overturned); has {
		return nil
	}
	if err := k.Reversed.Set(ctx, overturned); err != nil {
		return err
	}
	rng := collections.NewPrefixedPairRange[uint64, uint64](overturned)
	var toNegate []uint64
	if err := k.SourceIndex.Walk(ctx, rng, func(key collections.Pair[uint64, uint64]) (bool, error) {
		toNegate = append(toNegate, key.K2())
		return false, nil
	}); err != nil {
		return err
	}
	for _, cid := range toNegate {
		c, err := k.Contributions.Get(ctx, cid)
		if err != nil {
			continue
		}
		if err := k.addContribution(ctx, c.Signer, c.Domain, c.Magnitude.Neg(), c.RateBucket, pe.SourceAttId); err != nil {
			return err
		}
	}
	// 2× penalty to the original reporter (the countered outcome's author).
	if strength.IsPositive() && pe.TargetAttestor != "" {
		penalty := strength.Mul(d(p.NegAsymmetry)).Neg()
		if err := k.addContribution(ctx, pe.TargetAttestor, pe.TargetDomain, penalty,
			reputation.RateBucket_RATE_BUCKET_DURABLE_25Y, pe.SourceAttId); err != nil {
			return err
		}
	}
	return nil
}
