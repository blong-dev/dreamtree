package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
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

// StandingOf = baseline_kyc + Σ magnitude · relevance(domain, kᵢ), fixed-point,
// undecayed. The seam x/attest reads for S_issuance; also the basis for cred.
func (k Keeper) StandingOf(ctx context.Context, signer, domain string) math.LegacyDec {
	p, err := k.Params.Get(ctx)
	if err != nil {
		return math.LegacyOneDec()
	}
	sum := d(p.BaselineKyc)
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
		sum = sum.Add(c.Magnitude.Mul(rel))
		return false, nil
	})
	if sum.IsNegative() {
		return math.LegacyZeroDec()
	}
	return sum
}

// credOf is a basic continuous cred: 1 + standing, so √cred grows with a
// reporter's track record. (2-hop recursion is P4.)
func (k Keeper) credOf(ctx context.Context, reporter, domain string) math.LegacyDec {
	return math.LegacyOneDec().Add(k.StandingOf(ctx, reporter, domain))
}
