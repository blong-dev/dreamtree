package photons

import "cosmossdk.io/collections"

const (
	ModuleName = "photons"
	// PhotonDenom is the base currency denom. One photon mints per data-seed
	// (photons = seeds). Integer/non-divisible at base — market prices N_a are
	// whole photons per the spec's marketplace.
	PhotonDenom = "photon"
)

var (
	ParamsKey = collections.NewPrefix(0)
	MintedKey = collections.NewPrefix(1)
)
