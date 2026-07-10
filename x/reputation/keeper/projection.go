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

// relevance attenuates a contribution earned in domain ki toward query domain k
// by taxonomy distance — longest common prefix depth over the 5-level path.
// Same node 1.0; then 0.70 / 0.40 / 0.15 / 0.03; cross-class 0. (spec table)
func relevance(k, ki string) float64 {
	if k == ki {
		return 1.0
	}
	a, b := strings.Split(k, "/"), strings.Split(ki, "/")
	d := 0
	for d < len(a) && d < len(b) && a[d] == b[d] && a[d] != "" {
		d++
	}
	switch d {
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
	default: // 5 shared but not equal string (shouldn't happen) → treat as near-full
		return 0.85
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
// point S: R for R<=S, else S + k·log(1+(R-S)/S).
func effectiveR(raw, S, kDamp float64) float64 {
	if raw <= S {
		return raw
	}
	return S + kDamp*math.Log(1+(raw-S)/S)
}

// ReputationRaw returns the pre-saturation raw R and the contribution count.
func (k Keeper) reputationRaw(ctx context.Context, signer, domain string, now int64) (float64, float64, uint32, error) {
	p, err := k.Params.Get(ctx)
	if err != nil {
		return 0, 0, 0, err
	}
	obsolescence, saturation := k.domainShaping(ctx, p, domain)
	sum := 0.0
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
		sum += mag * rel * decay
		count++
		return false, nil
	})
	if err != nil {
		return 0, 0, 0, err
	}
	raw := pf(p.BaselineKyc) + sum
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
