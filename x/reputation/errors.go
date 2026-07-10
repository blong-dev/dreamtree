package reputation

import "cosmossdk.io/errors"

var (
	ErrEmptySigner = errors.Register(ModuleName, 2, "signer must not be empty")
	ErrEmptyDomain = errors.Register(ModuleName, 3, "domain must not be empty")
	ErrBadParams   = errors.Register(ModuleName, 4, "invalid params")
)
