package licenses

import "cosmossdk.io/collections"

const (
	ModuleName  = "licenses"
	PhotonDenom = "uphoton" // the base denom sales settle in (1 photon = 10^6 uphoton; type prices N_a are in uphoton)
)

var (
	ParamsKey      = collections.NewPrefix(0)
	TypePriceKey   = collections.NewPrefix(1) // data_type -> N_a
	AccessGrantKey = collections.NewPrefix(2) // (buyer, seed_id) -> AccessGrant
)
