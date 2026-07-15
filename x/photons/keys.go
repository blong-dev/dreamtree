package photons

import "cosmossdk.io/collections"

const (
	ModuleName = "photons"
	// PhotonDenom is the base denom: micro-photon. One photon = 10^6 uphoton;
	// the peg counts photons (photons = seeds = distinct atoms). Sub-photon
	// granularity exists for staking power, fees, and marketplace splits — the
	// meter identity lives at the photon (display) level.
	PhotonDenom = "uphoton"
	// UphotonPerPhoton is the base-unit factor (display exponent 6).
	UphotonPerPhoton = 1_000_000
	// DisplayDenom is the human unit registered in bank denom metadata.
	DisplayDenom = "photon"
)

var (
	ParamsKey = collections.NewPrefix(0)
	MintedKey = collections.NewPrefix(1)
)
