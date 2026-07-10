package attest

import (
	"context"

	"cosmossdk.io/math"
)

// ReputationKeeper is the seam x/attest reads reputation through and notifies of
// events. Defined here (the consumer); implemented by x/reputation. x/attest
// never imports x/reputation. When unwired (module absent), x/attest falls back
// to baseline everywhere and the hooks are no-ops — so it runs alone.
type ReputationKeeper interface {
	// ReputationOf — R(signer,domain,t), the read-time float projection (S/V).
	ReputationOf(ctx context.Context, signer, domain string) float64
	// StandingOf — the rational (fixed-point) reputation view, for S_issuance.
	StandingOf(ctx context.Context, signer, domain string) math.LegacyDec
	// OnAttestation — a new attestation's unvalidated bet.
	OnAttestation(ctx context.Context, signer, domain string, proofType int32, specificityBps uint32, sourceAttID uint64) error
	// OnOutcome — an OUTCOME attestation (validate/refute of targetAttID).
	OnOutcome(ctx context.Context, reporter string, refutes bool, targetAttID uint64, targetAttestor, targetDomain string, targetSIssuance math.LegacyDec, targetIsOutcome bool, sourceAttID uint64) error
	// OnEndorsement — A (endorser) vouches for B (endorsed) in a domain.
	OnEndorsement(ctx context.Context, endorser, endorsed, domain string, sourceAttID uint64) error
}
