package keeper

import (
	"context"
	"math"
	"strconv"

	"cosmossdk.io/collections"

	"github.com/blong-dev/dreamtree/x/attest"
)

// The read-time projection: strength, refutation, and work value. NONE of this
// is consensus state — it is a reading derived from the stored attestation log
// (data-model.md), so it uses exact float math freely. R(signer,domain) is a
// flat baseline stand-in until x/reputation lands.

const secondsPerYear = 31557600.0 // 365.25 days

type paramsF struct {
	baselineKyc  float64
	obsolescence float64
	sMax         float64
	weight       map[attest.ProofType]float64
	lambda       map[attest.ProofType]float64
}

func f(s string) float64 { v, _ := strconv.ParseFloat(s, 64); return v }

func loadParamsF(p attest.Params) paramsF {
	return paramsF{
		baselineKyc:  f(p.BaselineKyc),
		obsolescence: f(p.ObsolescenceDefault),
		sMax:         f(p.SMax),
		weight: map[attest.ProofType]float64{
			attest.ProofType_PROOF_TYPE_ORIGIN:      float64(p.WeightOrigin) / 10000,
			attest.ProofType_PROOF_TYPE_RIGOR:       float64(p.WeightRigor) / 10000,
			attest.ProofType_PROOF_TYPE_USE:         float64(p.WeightUse) / 10000,
			attest.ProofType_PROOF_TYPE_REPLICATION: float64(p.WeightReplication) / 10000,
			attest.ProofType_PROOF_TYPE_OUTCOME:     float64(p.WeightOutcome) / 10000,
		},
		lambda: map[attest.ProofType]float64{
			attest.ProofType_PROOF_TYPE_ORIGIN:      f(p.LambdaOrigin),
			attest.ProofType_PROOF_TYPE_REPLICATION: f(p.LambdaReplication),
			attest.ProofType_PROOF_TYPE_RIGOR:       f(p.LambdaRigor),
			attest.ProofType_PROOF_TYPE_USE:         f(p.LambdaUse),
			attest.ProofType_PROOF_TYPE_OUTCOME:     0, // outcomes don't age as work-value inputs
		},
	}
}

func specificityFactor(bps uint32) float64 {
	if bps == 0 {
		return 1.0 // unset = fully specific
	}
	return float64(bps) / 10000
}

func (pf paramsF) decay(a attest.Attestation, now int64) float64 {
	lam := pf.lambda[a.ProofType] * pf.obsolescence
	if lam <= 0 {
		return 1.0 // Origin / outcome: permanent
	}
	years := float64(now-a.IssuedAt) / secondsPerYear
	if years < 0 {
		years = 0
	}
	return math.Exp(-lam * years)
}

// reputationOf resolves R(signer,domain) through the reputation seam, falling
// back to baseline_kyc when x/reputation is not wired.
func (k Keeper) reputationOf(ctx context.Context, pf paramsF, signer, domain string) float64 {
	if k.rep != nil {
		return k.rep.ReputationOf(ctx, signer, domain)
	}
	return pf.baselineKyc
}

// rawStrengthR is S without the refutation term, given a resolved R — used both
// directly and as the weight a refuting outcome carries (breaks recursion).
func (pf paramsF) rawStrengthR(a attest.Attestation, now int64, r float64) float64 {
	return r *
		specificityFactor(a.SpecificityBps) *
		pf.weight[a.ProofType] *
		pf.decay(a, now)
}

// refutedFraction aggregates REFUTED (and half-weighted PARTIAL) outcomes
// targeting id, paper-shape: 1 - Π(1 - share_i), share_i = rawStrength/sMax.
func (k Keeper) refutedFraction(ctx context.Context, id uint64, pf paramsF, now int64) (float64, error) {
	prod := 1.0
	rng := collections.NewPrefixedPairRange[uint64, uint64](id)
	err := k.TargetIndex.Walk(ctx, rng, func(key collections.Pair[uint64, uint64]) (bool, error) {
		out, err := k.Attestations.Get(ctx, key.K2())
		if err != nil {
			return false, err
		}
		if out.ProofType != attest.ProofType_PROOF_TYPE_OUTCOME {
			return false, nil
		}
		var w float64
		switch out.OutcomeKind {
		case attest.OutcomeKind_OUTCOME_KIND_REFUTED:
			w = 1.0
		case attest.OutcomeKind_OUTCOME_KIND_PARTIAL:
			w = 0.5
		default:
			return false, nil // VALIDATED does not reduce work strength
		}
		rOut := k.reputationOf(ctx, pf, out.Attestor, out.Domain)
		share := w * pf.rawStrengthR(out, now, rOut) / pf.sMax
		if share > 1 {
			share = 1
		}
		prod *= (1 - share)
		return false, nil
	})
	if err != nil {
		return 0, err
	}
	return 1 - prod, nil
}

// strength is the full S(att,t) including the refutation term.
func (k Keeper) strength(ctx context.Context, a attest.Attestation, pf paramsF, now int64) (float64, float64, error) {
	rf, err := k.refutedFraction(ctx, a.Id, pf, now)
	if err != nil {
		return 0, 0, err
	}
	r := k.reputationOf(ctx, pf, a.Attestor, a.Domain)
	return pf.rawStrengthR(a, now, r) * (1 - rf), rf, nil
}

// workValue is the paper-shape aggregation over all non-outcome attestations on
// a subject: V = 1 - Π(1 - S_i/S_max). demand_signal = 1.0 at v0.
func (k Keeper) workValue(ctx context.Context, subject string, pf paramsF, now int64) (float64, uint32, error) {
	prod := 1.0
	var count uint32
	rng := collections.NewPrefixedPairRange[string, uint64](subject)
	err := k.SubjectIndex.Walk(ctx, rng, func(key collections.Pair[string, uint64]) (bool, error) {
		a, err := k.Attestations.Get(ctx, key.K2())
		if err != nil {
			return false, err
		}
		if a.ProofType == attest.ProofType_PROOF_TYPE_OUTCOME || a.ProofType == attest.ProofType_PROOF_TYPE_ENDORSEMENT {
			return false, nil // outcomes/endorsements move reputation, not work value
		}
		s, _, err := k.strength(ctx, a, pf, now)
		if err != nil {
			return false, err
		}
		share := s / pf.sMax
		if share > 1 {
			share = 1
		}
		prod *= (1 - share)
		count++
		return false, nil
	})
	if err != nil {
		return 0, 0, err
	}
	return 1 - prod, count, nil
}
