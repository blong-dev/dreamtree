package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/blong-dev/dreamtree/x/photons"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	addressCodec address.Codec
	authority    string
	bank         photons.BankKeeper

	Schema collections.Schema
	Params collections.Item[photons.Params]
	Minted collections.Sequence // increments once per ingestion mint; = data-seed count
	// Per-block mint-ceiling accumulator (trust-layer W3). MintHeight is the
	// block height BlockMinted is counting for; when the height advances the
	// accumulator resets lazily (a stale height reads as 0). Both are absent
	// until the first mint after this code ships — no migration needed.
	MintHeight  collections.Item[int64]
	BlockMinted collections.Item[uint64]
}

func NewKeeper(cdc codec.BinaryCodec, addressCodec address.Codec, storeService storetypes.KVStoreService, authority string, bank photons.BankKeeper) Keeper {
	if _, err := addressCodec.StringToBytes(authority); err != nil {
		panic(fmt.Errorf("invalid authority address: %w", err))
	}
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		cdc:          cdc,
		addressCodec: addressCodec,
		authority:    authority,
		bank:         bank,
		Params:       collections.NewItem(sb, photons.ParamsKey, "params", codec.CollValue[photons.Params](cdc)),
		Minted:       collections.NewSequence(sb, photons.MintedKey, "minted"),
		MintHeight:   collections.NewItem(sb, photons.MintHeightKey, "mint_height", collections.Int64Value),
		BlockMinted:  collections.NewItem(sb, photons.BlockMintedKey, "block_minted", collections.Uint64Value),
	}
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

func (k Keeper) GetAuthority() string { return k.authority }
