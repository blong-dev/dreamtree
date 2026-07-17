package attest

import (
	"fmt"
	"strconv"
)

// Default parameters — the stand-ins from parameters.md (v0.7.0). Weights in
// basis points; rates/reputation as decimal strings parsed only in the
// read-time projection.
func DefaultParams() Params {
	return Params{
		// type_weight — Origin/Rigor/Replication carry full weight; Use is
		// lighter (a citation is weaker signal than authorship/review); Outcome
		// is full (it moves reputation, not work value, but is weighted here for
		// completeness).
		WeightOrigin:      10000,
		WeightRigor:       10000,
		WeightUse:         5000,
		WeightReplication: 10000,
		WeightOutcome:     10000,

		LambdaOrigin:      "0.0",
		LambdaReplication: "0.015",
		LambdaRigor:       "0.04",
		LambdaUse:         "0.08",

		ObsolescenceDefault: "1.0",
		BaselineKyc:         "1.0",
		SMax:                "10.0",

		// creation-credit-forward (backtest M2): promoted from a compile-time
		// const; a citation from a top-value work counts up to (1 + this)x.
		CitationUpliftLambda: "1.0",
	}
}

func parsePos(name, v string) (float64, error) {
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", name, err)
	}
	if f < 0 {
		return 0, fmt.Errorf("%s must be >= 0", name)
	}
	return f, nil
}

// Validate checks the decimal params parse and are non-negative, and that the
// basis-point weights are within range.
func (p Params) Validate() error {
	for _, w := range []uint32{p.WeightOrigin, p.WeightRigor, p.WeightUse, p.WeightReplication, p.WeightOutcome} {
		if w > 10000 {
			return fmt.Errorf("type weight %d exceeds 10000 bps", w)
		}
	}
	for name, v := range map[string]string{
		"lambda_origin": p.LambdaOrigin, "lambda_replication": p.LambdaReplication,
		"lambda_rigor": p.LambdaRigor, "lambda_use": p.LambdaUse,
		"obsolescence_default": p.ObsolescenceDefault, "baseline_kyc": p.BaselineKyc, "s_max": p.SMax,
	} {
		if _, err := parsePos(name, v); err != nil {
			return err
		}
	}
	if sMax, _ := strconv.ParseFloat(p.SMax, 64); sMax <= 0 {
		return fmt.Errorf("s_max must be > 0")
	}
	// citation_uplift_lambda: empty tolerated (pre-upgrade state; readers fall
	// back to the pre-M2 const 1.0), else must parse >= 0.
	if p.CitationUpliftLambda != "" {
		if _, err := parsePos("citation_uplift_lambda", p.CitationUpliftLambda); err != nil {
			return err
		}
	}
	return nil
}
