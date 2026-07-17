// Package projection is the attest read-time projection — strength, refutation,
// work value — decoupled from the store (backtest M1, docs/specs/
// measurement-backtest.md Part B). NONE of this is consensus state: it is a
// reading derived from the attestation log, so it uses exact float math freely
// and must never be wired into a msg handler or Begin/EndBlock path.
//
// The same functions run in two worlds: the keeper supplies store-backed
// readers for live queries; the backtest harness supplies snapshot-backed ones
// over an exported log. Same math, same numbers, two data sources.
package projection

import (
	"math"
	"strconv"

	"github.com/blong-dev/dreamtree/x/attest"
)

const secondsPerYear = 31557600.0 // 365.25 days

// LogReader is the projection's view of the attestation log.
type LogReader interface {
	GetAttestation(id uint64) (attest.Attestation, error)
	WalkSubject(subject string, fn func(attest.Attestation) (stop bool, err error)) error
	WalkTarget(targetID uint64, fn func(attest.Attestation) (stop bool, err error)) error
}

// ReputationResolver resolves R(signer, domain). The keeper backs this with the
// x/reputation seam; the backtest holds it at baseline (see spec Limitations).
type ReputationResolver interface {
	ReputationOf(signer, domain string) float64
}

// ParamsF is the float view of attest.Params the projection computes with.
type ParamsF struct {
	BaselineKyc    float64
	Obsolescence   float64
	SMax           float64
	CitationUplift float64
	Weight         map[attest.ProofType]float64
	Lambda         map[attest.ProofType]float64
}

func f(s string) float64 { v, _ := strconv.ParseFloat(s, 64); return v }

// LoadParamsF converts stored params to the float view. citation_uplift_lambda
// falls back to the pre-M2 compile-time value (1.0) when the stored params
// predate the field.
func LoadParamsF(p attest.Params) ParamsF {
	uplift := 1.0
	if p.CitationUpliftLambda != "" {
		uplift = f(p.CitationUpliftLambda)
	}
	return ParamsF{
		BaselineKyc:    f(p.BaselineKyc),
		Obsolescence:   f(p.ObsolescenceDefault),
		SMax:           f(p.SMax),
		CitationUplift: uplift,
		Weight: map[attest.ProofType]float64{
			attest.ProofType_PROOF_TYPE_ORIGIN:      float64(p.WeightOrigin) / 10000,
			attest.ProofType_PROOF_TYPE_RIGOR:       float64(p.WeightRigor) / 10000,
			attest.ProofType_PROOF_TYPE_USE:         float64(p.WeightUse) / 10000,
			attest.ProofType_PROOF_TYPE_REPLICATION: float64(p.WeightReplication) / 10000,
			attest.ProofType_PROOF_TYPE_OUTCOME:     float64(p.WeightOutcome) / 10000,
		},
		Lambda: map[attest.ProofType]float64{
			attest.ProofType_PROOF_TYPE_ORIGIN:      f(p.LambdaOrigin),
			attest.ProofType_PROOF_TYPE_REPLICATION: f(p.LambdaReplication),
			attest.ProofType_PROOF_TYPE_RIGOR:       f(p.LambdaRigor),
			attest.ProofType_PROOF_TYPE_USE:         f(p.LambdaUse),
			attest.ProofType_PROOF_TYPE_OUTCOME:     0, // outcomes don't age as work-value inputs
		},
	}
}

// SpecificityFactor maps bps to (0,1]; unset = fully specific.
func SpecificityFactor(bps uint32) float64 {
	if bps == 0 {
		return 1.0
	}
	return float64(bps) / 10000
}

// Decay is exp(-λ·years) for the attestation's proof type; permanent types = 1.
func (pf ParamsF) Decay(a attest.Attestation, now int64) float64 {
	lam := pf.Lambda[a.ProofType] * pf.Obsolescence
	if lam <= 0 {
		return 1.0
	}
	years := float64(now-a.IssuedAt) / secondsPerYear
	if years < 0 {
		years = 0
	}
	return math.Exp(-lam * years)
}

// RawStrengthR is S without the refutation term, given a resolved R — used both
// directly and as the weight a refuting outcome carries (breaks recursion).
func (pf ParamsF) RawStrengthR(a attest.Attestation, now int64, r float64) float64 {
	return r *
		SpecificityFactor(a.SpecificityBps) *
		pf.Weight[a.ProofType] *
		pf.Decay(a, now)
}

// Projector computes the readings against any log + reputation source. It
// carries no SDK context — callers close their context into the readers.
type Projector struct {
	Params ParamsF
	Log    LogReader
	Rep    ReputationResolver
}

// RefutedFraction aggregates REFUTED (and half-weighted PARTIAL) outcomes
// targeting id, paper-shape: 1 - Π(1 - share_i), share_i = rawStrength/sMax.
func (p Projector) RefutedFraction(id uint64, now int64) (float64, error) {
	prod := 1.0
	err := p.Log.WalkTarget(id, func(out attest.Attestation) (bool, error) {
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
		share := w * p.Params.RawStrengthR(out, now, p.Rep.ReputationOf(out.Attestor, out.Domain)) / p.Params.SMax
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

// Strength is the full S(att,t) including the refutation term; also returns
// the refuted fraction for decomposed reporting.
func (p Projector) Strength(a attest.Attestation, now int64) (float64, float64, error) {
	rf, err := p.RefutedFraction(a.Id, now)
	if err != nil {
		return 0, 0, err
	}
	r := p.Rep.ReputationOf(a.Attestor, a.Domain)
	return p.Params.RawStrengthR(a, now, r) * (1 - rf), rf, nil
}

// WorkValueBase is V without citation uplift — a work's intrinsic value from
// its own attestations. Non-recursive, so it is safe as the success weight the
// uplift multiplies by (breaks any citation-graph recursion at one hop).
func (p Projector) WorkValueBase(subject string, now int64) (float64, uint32, error) {
	prod := 1.0
	var count uint32
	err := p.Log.WalkSubject(subject, func(a attest.Attestation) (bool, error) {
		if a.ProofType == attest.ProofType_PROOF_TYPE_OUTCOME || a.ProofType == attest.ProofType_PROOF_TYPE_ENDORSEMENT {
			return false, nil // outcomes/endorsements move reputation, not work value
		}
		s, _, err := p.Strength(a, now)
		if err != nil {
			return false, err
		}
		share := s / p.Params.SMax
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

// WorkValue is WorkValueBase plus creation-credit-forward: a USE attestation on
// `subject` records an edge (used_by = the new work B) → subject (the prior
// work A). A's citation contribution is scaled by (1 + λ·V_base(B)), so A's
// value rises with the value of the works built on it. One hop off V_base (no
// recursion), deterministic, signal-only.
func (p Projector) WorkValue(subject string, now int64) (float64, uint32, error) {
	prod := 1.0
	var count uint32
	err := p.Log.WalkSubject(subject, func(a attest.Attestation) (bool, error) {
		if a.ProofType == attest.ProofType_PROOF_TYPE_OUTCOME || a.ProofType == attest.ProofType_PROOF_TYPE_ENDORSEMENT {
			return false, nil
		}
		s, _, err := p.Strength(a, now)
		if err != nil {
			return false, err
		}
		share := s / p.Params.SMax
		if a.ProofType == attest.ProofType_PROOF_TYPE_USE && a.UsedBy != "" {
			vB, _, err := p.WorkValueBase(a.UsedBy, now)
			if err != nil {
				return false, err
			}
			share *= 1 + p.Params.CitationUplift*vB
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
