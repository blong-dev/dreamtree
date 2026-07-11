package licenses

import "cosmossdk.io/collections"

const (
	ModuleName  = "licenses"
	PhotonDenom = "photon" // the currency sales settle in
)

var (
	ParamsKey      = collections.NewPrefix(0)
	TypePriceKey   = collections.NewPrefix(1) // data_type -> N_a
	AccessGrantKey = collections.NewPrefix(2) // (buyer, seed_id) -> AccessGrant
)
