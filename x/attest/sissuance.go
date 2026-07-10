package attest

import "cosmossdk.io/math"

// typeWeightDec returns the fixed-point type_weight for a proof type (bps/10000).
func (p Params) typeWeightDec(pt ProofType) math.LegacyDec {
	var bps uint32
	switch pt {
	case ProofType_PROOF_TYPE_ORIGIN:
		bps = p.WeightOrigin
	case ProofType_PROOF_TYPE_RIGOR:
		bps = p.WeightRigor
	case ProofType_PROOF_TYPE_USE:
		bps = p.WeightUse
	case ProofType_PROOF_TYPE_REPLICATION:
		bps = p.WeightReplication
	case ProofType_PROOF_TYPE_OUTCOME:
		bps = p.WeightOutcome
	case ProofType_PROOF_TYPE_ENDORSEMENT:
		return math.LegacyOneDec() // endorsements carry full weight (no param)
	}
	return math.LegacyNewDec(int64(bps)).QuoInt64(10000)
}

// specDec returns the fixed-point specificity factor (bps/10000; 0 = full).
func specDec(bps uint32) math.LegacyDec {
	if bps == 0 {
		bps = 10000
	}
	return math.LegacyNewDec(int64(bps)).QuoInt64(10000)
}

// SIssuance computes the rational strength-at-issuance for an attestation given
// the attestor's standing: standing × specificity × type_weight.
func (p Params) SIssuance(standing math.LegacyDec, pt ProofType, specBps uint32) math.LegacyDec {
	return standing.Mul(specDec(specBps)).Mul(p.typeWeightDec(pt))
}
