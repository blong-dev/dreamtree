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

// netVerdict integrates a window into a SIGNED outcome magnitude M_O_net =
// V_pool − R_pool, where each pool is the paper-shape aggregate of its
// direction's reports (both 1×, capped at M_cap = cap_mult × S_issuance). The
// 2× negative asymmetry is NOT here — it is applied only to the contributor
// (spec: the R-update asymmetry), so the window's verdict is symmetric and a
// false accusation is neutralized by an equal defense (1:1), not 2:1.
func (k Keeper) netVerdict(p reputation.Params, pe reputation.PendingEvent) math.LegacyDec {
	capM := d(p.OutcomeCapMult).Mul(pe.TargetSIssuance)
	if !capM.IsPositive() {
		capM = pe.BaseMagnitude
	}
	var vPool, rPool math.LegacyDec
	if pe.OutcomeRefutes {
		rPool = paperShapeAdd(pe.BaseMagnitude, pe.Corroboration, capM)
		vPool = pe.Refutation // opposite direction = defenses = validations
	} else {
		vPool = paperShapeAdd(pe.BaseMagnitude, pe.Corroboration, capM)
		rPool = pe.Refutation // opposite direction = contests = refutations
	}
	return vPool.Sub(rPool)
}

// applyFloored adds a reputation contribution, capping a negative delta so the
// recipient's standing cannot go below 0. Reputation is [0, ∞): 0 is the floor,
// there is no negative "debt", and recovery is from 0 — a bounded crowd (paper-
// shape) plus this floor make character assassination unrecoverable-hole-proof.
func (k Keeper) applyFloored(ctx context.Context, addr, domain string, delta math.LegacyDec, bucket reputation.RateBucket, source uint64) error {
	if delta.IsNegative() {
		cur := k.StandingOf(ctx, addr, domain)
		if delta.Neg().GT(cur) {
			delta = cur.Neg() // cap: drive to exactly 0, never below
		}
	}
	if delta.IsZero() {
		return nil
	}
	return k.addContribution(ctx, addr, domain, delta, bucket, source)
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
		net := k.netVerdict(p, pe) // signed M_O_net = V_pool − R_pool
		if pe.CounterTargetPending != 0 {
			// Reversal: negate the overturned outcome's contributions + 2× penalty
			// to its reporter (target_attestor). Uses the verdict magnitude.
			return k.reverse(ctx, p, pe, net.Abs())
		}
		// Apply to the contributor (target's author), durable. Validated (net≥0)
		// → +net; refuted (net<0) → neg×net (bad work hits the author 2× harder).
		// The 2× lives ONLY here (spec's R-update asymmetry) — never inside the
		// window integration, so defenses count 1× and can't be double-penalized.
		contribDelta := net
		if net.IsNegative() {
			contribDelta = net.Mul(d(p.NegAsymmetry))
		}
		if err := k.applyFloored(ctx, pe.TargetAttestor, pe.TargetDomain, contribDelta,
			reputation.RateBucket_RATE_BUCKET_DURABLE_25Y, pe.SourceAttId); err != nil {
			return err
		}
		// Propagate to co-attestors and the contributor's endorsers (liability),
		// signed by the verdict, at 1× (no author's 2×), each floored at 0.
		return k.propagate(ctx, p, pe, net)
	}
	return nil
}

// propagate applies an outcome to the captured prop targets (co-attestors +
// endorsers). Co-attestor weight = coattestor_weight × specificity; endorser
// weight = endorse_inherit. net is the signed verdict (sign already carries
// validated(+)/refuted(−)); each move is floored at 0 (no debt for anyone).
func (k Keeper) propagate(ctx context.Context, p reputation.Params, pe reputation.PendingEvent, net math.LegacyDec) error {
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
		if err := k.applyFloored(ctx, pt.Address, pt.Domain, net.Mul(weight),
			reputation.RateBucket_RATE_BUCKET_DURABLE_25Y, pe.SourceAttId); err != nil {
			return err
		}
	}
	return nil
}

// reverse un-applies a settled outcome's contributions and penalizes its
// reporter. mag is the countering verdict's magnitude (net.Abs()).
func (k Keeper) reverse(ctx context.Context, p reputation.Params, pe reputation.PendingEvent, mag math.LegacyDec) error {
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
		// Floored negation (Z2 fix, DT-21 comb): if the beneficiary's standing
		// has meanwhile decayed or been spent, the claw-back caps at driving
		// them to 0 — never a stored net-negative. Same floor every other
		// R-move already honors; the remainder is simply unrecoverable, which
		// is what "no debt" means.
		if err := k.applyFloored(ctx, c.Signer, c.Domain, c.Magnitude.Neg(), c.RateBucket, pe.SourceAttId); err != nil {
			return err
		}
	}
	// 2× penalty to the original reporter (the countered outcome's author),
	// floored at 0 — a false accuser is driven to 0, not into debt.
	if mag.IsPositive() && pe.TargetAttestor != "" {
		penalty := mag.Mul(d(p.NegAsymmetry)).Neg()
		if err := k.applyFloored(ctx, pe.TargetAttestor, pe.TargetDomain, penalty,
			reputation.RateBucket_RATE_BUCKET_DURABLE_25Y, pe.SourceAttId); err != nil {
			return err
		}
	}
	return nil
}
