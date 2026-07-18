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
	// Per-block mint-ceiling accumulator (trust-layer W3). MintHeightKey holds
	// the block height the accumulator is for; BlockMintedKey holds the photons
	// minted so far in that block. Both default to zero when absent, so they
	// are safe to introduce on a running chain with no migration.
	MintHeightKey  = collections.NewPrefix(2)
	BlockMintedKey = collections.NewPrefix(3)
)
