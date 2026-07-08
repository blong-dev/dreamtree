package seeds

// Default parameter bounds. Commitments are digests/Merkle roots, so these are
// deliberately small — bodies never enter consensus.
const (
	DefaultMaxCommitmentBytes uint32 = 512
	DefaultMaxSourceRefBytes  uint32 = 512
)

// DefaultParams returns default module parameters.
func DefaultParams() Params {
	return Params{
		MaxCommitmentBytes: DefaultMaxCommitmentBytes,
		MaxSourceRefBytes:  DefaultMaxSourceRefBytes,
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

// Validate does a sanity check on the params.
func (p Params) Validate() error {
	return nil
}
