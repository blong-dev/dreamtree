package attest

import (
	"context"

	"cosmossdk.io/math"
)

// ReputationKeeper is the seam x/attest reads R through and notifies of new
// attestations. Defined here (the consumer) and implemented by x/reputation;
// x/attest never imports x/reputation. If unwired (module absent), x/attest
// falls back to baseline_kyc and OnAttestation is a no-op — so it runs alone.
//
// The math.LegacyDec import keeps this ready for later phases that pass
// fixed-point magnitudes across the seam; P1 uses only the two methods below.
type ReputationKeeper interface {
	// ReputationOf returns R(signer,domain,t) — a read-time float projection.
	ReputationOf(ctx context.Context, signer, domain string) float64
	// OnAttestation records the reputation effect of a newly stored attestation.
	OnAttestation(ctx context.Context, signer, domain string, proofType int32, specificityBps uint32, sourceAttID uint64) error
}

var _ = math.LegacyDec{}
