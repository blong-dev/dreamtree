package seeds

import "cosmossdk.io/errors"

var (
	// ErrEmptyCommitment is returned when a commit carries no digest.
	ErrEmptyCommitment = errors.Register(ModuleName, 2, "commitment must not be empty")
	// ErrCommitmentTooLong is returned when the digest exceeds the param bound.
	ErrCommitmentTooLong = errors.Register(ModuleName, 3, "commitment exceeds max length")
	// ErrSourceRefTooLong is returned when the source ref exceeds the param bound.
	ErrSourceRefTooLong = errors.Register(ModuleName, 4, "source_ref exceeds max length")
	// ErrEmptyKind is returned when a commit carries no kind label.
	ErrEmptyKind        = errors.Register(ModuleName, 5, "kind must not be empty")
	ErrCommitmentNotHex = errors.Register(ModuleName, 6, "commitment must be hex (a digest or Merkle root, not a body)")
	// ErrBadCounts is returned when a batch's counts are invalid
	// (need 0 < new_count <= leaf_count).
	ErrBadCounts = errors.Register(ModuleName, 7, "batch counts invalid: need 0 < new_count <= leaf_count")
	// ErrRetiredKind is returned for aggregate kind labels — the aggregate is
	// no longer a seed kind; kind names the LEAF (seed = atom,
	// docs/specs/seed-atom-conformance.md).
	ErrRetiredKind = errors.Register(ModuleName, 8, "kind names the leaf; batch_root-style aggregate kinds are retired (seed = atom)")
	// ErrSeedNotFound is returned when a leaf-seed id resolves to no batch.
	ErrSeedNotFound = errors.Register(ModuleName, 9, "seed id not found")
)
