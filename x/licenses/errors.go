package licenses

import "cosmossdk.io/errors"

var (
	ErrNoPricedSeeds = errors.Register(ModuleName, 2, "no priced seeds in the swath")
	ErrInsufficient  = errors.Register(ModuleName, 3, "buyer has insufficient photons for the swath")
	ErrBadToll       = errors.Register(ModuleName, 4, "marketplace_toll must be in [0,1]")
	ErrEmptyDataType = errors.Register(ModuleName, 5, "data_type must not be empty")
)
