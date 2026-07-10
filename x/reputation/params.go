package reputation

import (
	"fmt"
	"strconv"
)

// Default parameters — parameters.md stand-ins. Decimal strings are read-side
// (decay/saturation) inputs parsed in the projection.
func DefaultParams() Params {
	return Params{
		BaselineKyc:          "1.0",
		DampeningK:           "5.0",
		SaturationStandard:   "10.0",
		ObsolescenceStandard: "1.0",
		LambdaPermanent:      "0.0",
		LambdaDurable:        "0.0277", // ln2/25 ≈ 25yr half-life
		LambdaRigor:          "0.04",
		LambdaUse:            "0.08",
		LambdaReplication:    "0.015",
		LambdaEndorsement:    "0.08",
		AttestBetScale:       "0.1", // an unvalidated attestation is a small bet
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
		"attest_bet_scale": p.AttestBetScale,
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
