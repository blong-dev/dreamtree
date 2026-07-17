package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	"github.com/blong-dev/dreamtree/x/reputation"
)

// The RATIONAL view of reputation, for consensus math (S_issuance, cred). Uses
// fixed-point LegacyDec and the discrete relevance table — NO decay, NO
// saturation, NO float. This is the "standing" the determinism rule says
// consensus reads instead of the float R (docs/specs/x-reputation-design.md §1).

// relevanceDec is the fixed-point relevance (same table as the float version).
func relevanceDec(k, ki string) math.LegacyDec {
	switch relevanceDepth(k, ki) {
	case 0:
		return math.LegacyZeroDec()
	case 1:
		return math.LegacyMustNewDecFromStr("0.03")
	case 2:
		return math.LegacyMustNewDecFromStr("0.15")
	case 3:
		return math.LegacyMustNewDecFromStr("0.40")
	case 4:
		return math.LegacyMustNewDecFromStr("0.70")
	default:
		return math.LegacyOneDec()
	}
}

// eCapMultDec returns the ratified endorsement breadth-cap multiple, falling
// back to the canonical default when the stored param predates upgrade-1
// (empty string) or is out of range (< 1 would make the cap tighter than the
// strongest single endorsement — nonsensical).
func eCapMultDec(p reputation.Params) math.LegacyDec {
	m, err := math.LegacyNewDecFromStr(p.ECapMult)
	if err != nil || m.LT(math.LegacyOneDec()) {
		return math.LegacyMustNewDecFromStr(reputation.DefaultECapMult)
	}
	return m
}

// endorseFoldDec aggregates positive ENDORSEMENT-bucket contributions
// paper-shape (ratified 2026-07-17): E_total = E_cap·[1−Π(1−eᵢ/E_cap)] with
// E_cap = e_cap_mult × max(eᵢ). Reversal negations (negative ENDORSEMENT
// entries) subtract linearly from the folded total, floored at zero — under
// the fold an endorsement's marginal effect was ≤ its magnitude, so linear
// subtraction over-punishes reversals slightly; conservative by design.
func endorseFoldDec(p reputation.Params, pos []math.LegacyDec, neg math.LegacyDec) math.LegacyDec {
	total := math.LegacyZeroDec()
	if len(pos) > 0 {
		max := pos[0]
		for _, e := range pos[1:] {
			if e.GT(max) {
				max = e
			}
		}
		cap := eCapMultDec(p).Mul(max)
		if cap.IsPositive() {
			one := math.LegacyOneDec()
			prod := one
			for _, e := range pos {
				frac := e.Quo(cap)
				if frac.GT(one) {
					frac = one
				}
				prod = prod.Mul(one.Sub(frac))
			}
			total = cap.Mul(one.Sub(prod))
		}
	}
	total = total.Sub(neg)
	if total.IsNegative() {
		total = math.LegacyZeroDec()
	}
	return total
}

// StandingOf = base + Σ magnitude · relevance(domain, kᵢ) + E_total, fixed-point,
// undecayed. base = baseline_kyc ONLY for members of the verified set (upgrade-1
// R2: standing starts at ZERO; verification grants the floor). ENDORSEMENT-bucket
// contributions do not sum inline — they fold paper-shape after the walk (the
// spec's cold-start additive-term shape). The seam x/attest reads for
// S_issuance; also the basis for cred.
func (k Keeper) StandingOf(ctx context.Context, signer, domain string) math.LegacyDec {
	p, err := k.Params.Get(ctx)
	if err != nil {
		return math.LegacyZeroDec()
	}
	sum := math.LegacyZeroDec()
	if ok, _ := k.Verified.Has(ctx, signer); ok {
		sum = d(p.BaselineKyc)
	}
	var endorsePos []math.LegacyDec
	endorseNeg := math.LegacyZeroDec()
	rng := collections.NewPrefixedPairRange[string, uint64](signer)
	_ = k.SignerIndex.Walk(ctx, rng, func(key collections.Pair[string, uint64]) (bool, error) {
		c, err := k.Contributions.Get(ctx, key.K2())
		if err != nil {
			return false, err
		}
		rel := relevanceDec(domain, c.Domain)
		if rel.IsZero() {
			return false, nil
		}
		e := c.Magnitude.Mul(rel)
		if c.RateBucket == reputation.RateBucket_RATE_BUCKET_ENDORSEMENT {
			if e.IsPositive() {
				endorsePos = append(endorsePos, e)
			} else {
				endorseNeg = endorseNeg.Add(e.Neg())
			}
			return false, nil
		}
		sum = sum.Add(e)
		// Running floor (Z2, spec §floor-is-zero): the walk is contribution-id
		// order = settlement order, so flooring each step makes "recovery is
		// genuinely from zero" literal — a hole deeper than 0 is forgiven at
		// the moment it happened, and later positive work counts in full,
		// instead of first refilling an invisible debt.
		if sum.IsNegative() {
			sum = math.LegacyZeroDec()
		}
		return false, nil
	})
	return sum.Add(endorseFoldDec(p, endorsePos, endorseNeg))
}

// credOf is the reporter's rational standing: zero for an unknown signer,
// baseline for a verified-but-unaccumulated one, higher with track record. At
// baseline √cred=1 ⟹ M_O≈S (the spec's β=1 calibration). (2-hop recursion is P4.)
func (k Keeper) credOf(ctx context.Context, reporter, domain string) math.LegacyDec {
	return k.StandingOf(ctx, reporter, domain)
}
