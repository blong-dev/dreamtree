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
	ErrEmptyKind = errors.Register(ModuleName, 5, "kind must not be empty")
)
