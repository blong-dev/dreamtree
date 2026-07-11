package reputation

import (
	"fmt"
	"strconv"
)

// Default parameters — parameters.md stand-ins. Decimal strings are read-side
// (decay/saturation) inputs parsed in the projection.
func DefaultParams() Params {
	return Params{
		BaselineKyc:           "1.0",
		DampeningK:            "5.0",
		SaturationStandard:    "10.0",
		ObsolescenceStandard:  "1.0",
		LambdaPermanent:       "0.0",
		LambdaDurable:         "0.0277", // ln2/25 ≈ 25yr half-life
		LambdaRigor:           "0.04",
		LambdaUse:             "0.08",
		LambdaReplication:     "0.015",
		LambdaEndorsement:     "0.08",
		AttestBetScale:        "0.1", // an unvalidated attestation is a small bet
		NegAsymmetry:          "2.0",
		OutcomeBeta:           "1.0",
		OutcomeCapMult:        "5.0",
		ReviewWindowBase:      "1.0", // days
		ReviewWindowThreshold: "4.0", // tuned up: √ has a fat tail near 0, so a higher threshold shortens trivial-event windows toward "instant" (bets ~hours). √ can't fully match log's trivial-instant/large-weeks spread — revisit the curve or route bets around the window if the spread matters.
		CoattestorWeight:      "0.25",
		EndorseInherit:        "0.25",
	}
}

func mustPos(name, v string) error {
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	if f < 0 {
		return fmt.Errorf("%s must be >= 0", name)
	}
	return nil
}

func (p Params) Validate() error {
	for name, v := range map[string]string{
		"baseline_kyc": p.BaselineKyc, "dampening_k": p.DampeningK,
		"saturation_standard": p.SaturationStandard, "obsolescence_standard": p.ObsolescenceStandard,
		"lambda_permanent": p.LambdaPermanent, "lambda_durable": p.LambdaDurable,
		"lambda_rigor": p.LambdaRigor, "lambda_use": p.LambdaUse,
		"lambda_replication": p.LambdaReplication, "lambda_endorsement": p.LambdaEndorsement,
		"attest_bet_scale": p.AttestBetScale, "neg_asymmetry": p.NegAsymmetry,
		"outcome_beta": p.OutcomeBeta, "outcome_cap_mult": p.OutcomeCapMult,
		"review_window_base": p.ReviewWindowBase, "review_window_threshold": p.ReviewWindowThreshold,
		"coattestor_weight": p.CoattestorWeight, "endorse_inherit": p.EndorseInherit,
	} {
		if err := mustPos(name, v); err != nil {
			return err
		}
	}
	if s, _ := strconv.ParseFloat(p.SaturationStandard, 64); s <= 0 {
		return fmt.Errorf("saturation_standard must be > 0")
	}
	return nil
}
