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

// Work value V is the paper-shape aggregation over a subject's non-outcome
// attestations: V = 1 - Π(1 - S_i/S_max). demand_signal = 1.0 at v0.
//
// citationUpliftLambda controls creation-credit-forward: how much a maximally-
// valuable citing work amplifies a USE citation's contribution to the source it
// builds on. 1.0 ⇒ a citation from a top-value work counts up to 2×. This is a
// value SIGNAL only — it lives in the read-projection (float, off consensus),
// never mints or splits photons. TODO: promote to a governable attest param.
const citationUpliftLambda = 1.0

// workValueBase is V without citation uplift — a work's intrinsic value from its
// own attestations. Non-recursive, so it is safe to read as the success weight
// the uplift multiplies by (breaks any citation-graph recursion at one hop).
func (k Keeper) workValueBase(ctx context.Context, subject string, pf paramsF, now int64) (float64, uint32, error) {
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

// workValue is workValueBase plus creation-credit-forward: a USE attestation on
// `subject` records an edge (used_by = the new work B) → subject (the prior work
// A). A's citation contribution is scaled by (1 + λ·V_base(B)), so A's value
// rises with the value of the works built on it. One hop off `V_base` (no
// recursion), deterministic, signal-only. Returns count of contributing attns.
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
			return false, nil
		}
		s, _, err := k.strength(ctx, a, pf, now)
		if err != nil {
			return false, err
		}
		share := s / pf.sMax
		// Creation-credit-forward: uplift a citation by the citing work's value.
		if a.ProofType == attest.ProofType_PROOF_TYPE_USE && a.UsedBy != "" {
			vB, _, err := k.workValueBase(ctx, a.UsedBy, pf, now)
			if err != nil {
				return false, err
			}
			share *= 1 + citationUpliftLambda*vB
		}
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
