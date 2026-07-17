package licenses

import "cosmossdk.io/collections"

const (
	ModuleName  = "licenses"
	PhotonDenom = "uphoton" // the base denom sales settle in (1 photon = 10^6 uphoton)
	// UphotonPerPhoton — display-to-base factor. Since upgrade-1 R4 the access
	// price is the protocol CONSTANT 1 photon per seed per day (owner
	// 2026-07-16): price-per-seed = access_duration_days × this, in uphoton.
	UphotonPerPhoton = 1_000_000
)

var (
	ParamsKey      = collections.NewPrefix(0)
	TypePriceKey   = collections.NewPrefix(1) // data_type -> N_a
	AccessGrantKey = collections.NewPrefix(2) // (buyer, seed_id) -> AccessGrant
)
