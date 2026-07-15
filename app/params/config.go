package params

import (
	"cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// CoinUnit is the base denom: micro-photon (1 photon = 10^6 uphoton).
	// Photons = seeds = distinct atoms; minted per NEW leaf-seed at ingestion.
	CoinUnit = "uphoton"
	// BondDenom is the photon base unit: the photon IS the chain's native asset
	// per the protocol spec (seed-atom conformance, 2026-07-15). Validators bond
	// genesis-corpus photons; dtvp (the 2026-07-10 staking expedient) is retired.
	// With the SDK's default PowerReduction (10^6), voting power = whole photons.
	BondDenom = CoinUnit

	DefaultBondDenom = BondDenom

	// Bech32PrefixAccAddr defines the Bech32 prefix of an account's address.
	Bech32PrefixAccAddr = "dream"
)

var (
	// Bech32PrefixAccPub defines the Bech32 prefix of an account's public key.
	Bech32PrefixAccPub = Bech32PrefixAccAddr + "pub"
	// Bech32PrefixValAddr defines the Bech32 prefix of a validator's operator address.
	Bech32PrefixValAddr = Bech32PrefixAccAddr + "valoper"
	// Bech32PrefixValPub defines the Bech32 prefix of a validator's operator public key.
	Bech32PrefixValPub = Bech32PrefixAccAddr + "valoperpub"
	// Bech32PrefixConsAddr defines the Bech32 prefix of a consensus node address.
	Bech32PrefixConsAddr = Bech32PrefixAccAddr + "valcons"
	// Bech32PrefixConsPub defines the Bech32 prefix of a consensus node public key.
	Bech32PrefixConsPub = Bech32PrefixAccAddr + "valconspub"
)

func init() {
	SetAddressPrefixes()
	RegisterDenoms()
}

func RegisterDenoms() {
	// Register the photon display unit and its micro base unit (single-token
	// chain — the two-denoms-at-factor-1.0 normalizer gotcha that bit the dtvp
	// era no longer applies: these register at distinct factors).
	if err := sdk.RegisterDenom("photon", math.LegacyOneDec()); err != nil {
		panic(err)
	}
	if err := sdk.RegisterDenom(CoinUnit, math.LegacyNewDecWithPrec(1, 6)); err != nil {
		panic(err)
	}
}

func SetAddressPrefixes() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(Bech32PrefixAccAddr, Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(Bech32PrefixValAddr, Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(Bech32PrefixConsAddr, Bech32PrefixConsPub)

	// This is copied from the cosmos sdk v0.43.0-beta1
	// source: https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-beta1/types/address.go#L141
	config.SetAddressVerifier(func(bytes []byte) error {
		if len(bytes) == 0 {
			return errors.Wrap(sdkerrors.ErrUnknownAddress, "addresses cannot be empty")
		}

		if len(bytes) > address.MaxAddrLen {
			return errors.Wrapf(sdkerrors.ErrUnknownAddress, "address max length is %d, got %d", address.MaxAddrLen, len(bytes))
		}

		if len(bytes) != 20 && len(bytes) != 32 {
			return errors.Wrapf(sdkerrors.ErrUnknownAddress, "address length must be 20 or 32 bytes, got %d", len(bytes))
		}

		return nil
	})
}
