package seeds

// Default parameter bounds. Commitments are digests/Merkle roots, so these are
// deliberately small — bodies never enter consensus.
const (
	DefaultMaxCommitmentBytes uint32 = 512
	DefaultMaxSourceRefBytes  uint32 = 512
	// DefaultMaxBatchNewCount caps a single batch's committer-asserted
	// new_count (each new leaf mints a photon — an uncapped claim is a
	// single-tx supply-griefing vector). Sized generously above the observed
	// per-cycle rate (~hundreds of atoms) while bounding one tx's mint to
	// well under a day of honest corpus growth. Governance-tunable.
	DefaultMaxBatchNewCount uint32 = 1_000_000
)

// DefaultParams returns default module parameters.
func DefaultParams() Params {
	return Params{
		MaxCommitmentBytes: DefaultMaxCommitmentBytes,
		MaxSourceRefBytes:  DefaultMaxSourceRefBytes,
		MaxBatchNewCount:   DefaultMaxBatchNewCount,
	}
}

// MaxCommitment returns the effective commitment bound (falling back to default).
func (p Params) MaxCommitment() uint32 {
	if p.MaxCommitmentBytes == 0 {
		return DefaultMaxCommitmentBytes
	}
	return p.MaxCommitmentBytes
}

// MaxSourceRef returns the effective source_ref bound (falling back to default).
func (p Params) MaxSourceRef() uint32 {
	if p.MaxSourceRefBytes == 0 {
		return DefaultMaxSourceRefBytes
	}
	return p.MaxSourceRefBytes
}

// MaxBatchNew returns the effective per-batch new_count cap (falling back to
// default).
func (p Params) MaxBatchNew() uint32 {
	if p.MaxBatchNewCount == 0 {
		return DefaultMaxBatchNewCount
	}
	return p.MaxBatchNewCount
}

// Validate does a sanity check on the params.
func (p Params) Validate() error {
	return nil
}
