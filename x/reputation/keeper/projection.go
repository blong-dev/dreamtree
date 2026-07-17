package keeper

import (
	"context"
	"math"
	"strconv"
	"strings"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/blong-dev/dreamtree/x/reputation"
)

// The read-time projection of R(j,k,t). Float math, single-node reading — never
// consensus state (see docs/specs/x-reputation-design.md §1, §8). Consensus
// stores integrated Contributions (LegacyDec magnitudes); the exp decay and log
// saturation live here and here only.

const secondsPerYear = 31557600.0 // 365.25 days

func pf(s string) float64 { v, _ := strconv.ParseFloat(s, 64); return v }

// relevanceDepth is the longest common prefix depth of two 5-level taxonomy
// paths — the shared basis for both the float and fixed-point relevance.
func relevanceDepth(k, ki string) int {
	if k == ki {
		return 5
	}
	a, b := strings.Split(k, "/"), strings.Split(ki, "/")
	d := 0
	for d < len(a) && d < len(b) && a[d] == b[d] && a[d] != "" {
		d++
	}
	return d
}

// relevance attenuates a contribution earned in domain ki toward query domain k.
// Same node 1.0; then 0.70 / 0.40 / 0.15 / 0.03; cross-class 0. (spec table)
func relevance(k, ki string) float64 {
	switch relevanceDepth(k, ki) {
	case 0:
		return 0.0
	case 1:
		return 0.03
	case 2:
		return 0.15
	case 3:
		return 0.40
	case 4:
		return 0.70
	default:
		return 1.0
	}
}

func (k Keeper) lambda(p reputation.Params, bucket reputation.RateBucket) float64 {
	switch bucket {
	case reputation.RateBucket_RATE_BUCKET_PERMANENT:
		return pf(p.LambdaPermanent)
	case reputation.RateBucket_RATE_BUCKET_DURABLE_25Y:
		return pf(p.LambdaDurable)
	case reputation.RateBucket_RATE_BUCKET_RIGOR:
		return pf(p.LambdaRigor)
	case reputation.RateBucket_RATE_BUCKET_USE:
		return pf(p.LambdaUse)
	case reputation.RateBucket_RATE_BUCKET_REPLICATION:
		return pf(p.LambdaReplication)
	case reputation.RateBucket_RATE_BUCKET_ENDORSEMENT:
		return pf(p.LambdaEndorsement)
	}
	return pf(p.LambdaUse)
}

// obsolescence + saturation for a domain (defaults when no DomainConfig set).
func (k Keeper) domainShaping(ctx context.Context, p reputation.Params, domain string) (obsolescence, saturation float64) {
	obsolescence, saturation = pf(p.ObsolescenceStandard), pf(p.SaturationStandard)
	if cfg, err := k.DomainConfigs.Get(ctx, domain); err == nil {
		if v := pf(cfg.ObsolescenceMultiplier.String()); v > 0 {
			obsolescence = v
		}
		if v := pf(cfg.SaturationPoint.String()); v > 0 {
			saturation = v
		}
	}
	return
}

// effectiveR applies the two-piece linear+log dampening past the saturation
// point S: R for R<=S, else S + k·log(1+(R-S)/S). Floored at 0 — a domain R can
// be destroyed to zero, and per spec §floor-is-zero recovery is genuinely from
// zero: no contribution debt survives (running floor in reputationRaw /
// StandingOf; floored negations in window.go's reverse path).
func effectiveR(raw, S, kDamp float64) float64 {
	if raw <= 0 {
		return 0
	}
	if raw <= S {
		return raw
	}
	return S + kDamp*math.Log(1+(raw-S)/S)
}

// endorseFold is the float mirror of standing.go's endorseFoldDec: positive
// ENDORSEMENT contributions (already relevance- and decay-attenuated) aggregate
// paper-shape with E_cap = e_cap_mult × max(eᵢ); reversal negations subtract
// linearly from the folded total, floored at zero.
func endorseFold(p reputation.Params, pos []float64, neg float64) float64 {
	total := 0.0
	if len(pos) > 0 {
		max := pos[0]
		for _, e := range pos[1:] {
			if e > max {
				max = e
			}
		}
		mult := pf(p.ECapMult)
		if mult < 1 {
			mult, _ = strconv.ParseFloat(reputation.DefaultECapMult, 64)
		}
		if cap := mult * max; cap > 0 {
			prod := 1.0
			for _, e := range pos {
				frac := e / cap
				if frac > 1 {
					frac = 1
				}
				prod *= 1 - frac
			}
			total = cap * (1 - prod)
		}
	}
	total -= neg
	if total < 0 {
		total = 0
	}
	return total
}

// ReputationRaw returns the pre-saturation raw R and the contribution count.
// The baseline term applies ONLY to verified-set members (upgrade-1 R2:
// standing starts at zero); ENDORSEMENT-bucket contributions fold paper-shape
// after the walk instead of summing inline (upgrade-1 R5).
func (k Keeper) reputationRaw(ctx context.Context, signer, domain string, now int64) (float64, float64, uint32, error) {
	p, err := k.Params.Get(ctx)
	if err != nil {
		return 0, 0, 0, err
	}
	obsolescence, saturation := k.domainShaping(ctx, p, domain)
	base := 0.0
	if ok, _ := k.Verified.Has(ctx, signer); ok {
		base = pf(p.BaselineKyc)
	}
	sum := 0.0
	var endorsePos []float64
	endorseNeg := 0.0
	var count uint32
	rng := collections.NewPrefixedPairRange[string, uint64](signer)
	err = k.SignerIndex.Walk(ctx, rng, func(key collections.Pair[string, uint64]) (bool, error) {
		c, err := k.Contributions.Get(ctx, key.K2())
		if err != nil {
			return false, err
		}
		rel := relevance(domain, c.Domain)
		if rel == 0 {
			return false, nil
		}
		lam := k.lambda(p, c.RateBucket) * obsolescence
		decay := 1.0
		if lam > 0 {
			years := float64(now-c.SettledAt) / secondsPerYear
			if years < 0 {
				years = 0
			}
			decay = math.Exp(-lam * years)
		}
		mag, _ := strconv.ParseFloat(c.Magnitude.String(), 64)
		count++
		v := mag * rel * decay
		if c.RateBucket == reputation.RateBucket_RATE_BUCKET_ENDORSEMENT {
			if v > 0 {
				endorsePos = append(endorsePos, v)
			} else {
				endorseNeg += -v
			}
			return false, nil
		}
		sum += v
		// Running floor (Z2, mirrors standing.go): id order = settlement
		// order; a below-zero excursion is forgiven where it happened, so
		// later work recovers from zero, not from a hole. (Read-time float
		// projection — the rational consensus path applies the same rule.)
		if base+sum < 0 {
			sum = -base
		}
		return false, nil
	})
	if err != nil {
		return 0, 0, 0, err
	}
	raw := base + sum + endorseFold(p, endorsePos, endorseNeg)
	return raw, saturation, count, nil
}

// ReputationOf is the ReputationSource seam x/attest reads through.
func (k Keeper) ReputationOf(ctx context.Context, signer, domain string) float64 {
	now := sdk.UnwrapSDKContext(ctx).BlockTime().Unix()
	raw, saturation, _, err := k.reputationRaw(ctx, signer, domain, now)
	if err != nil {
		return 0
	}
	p, _ := k.Params.Get(ctx)
	return effectiveR(raw, saturation, pf(p.DampeningK))
}
