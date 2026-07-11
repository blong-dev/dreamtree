package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/blong-dev/dreamtree/x/licenses"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	addressCodec address.Codec
	authority    string
	bank         licenses.BankKeeper
	seeds        licenses.SeedReader

	Schema       collections.Schema
	Params       collections.Item[licenses.Params]
	TypePrices   collections.Map[string, uint64]
	AccessGrants collections.Map[collections.Pair[string, uint64], licenses.AccessGrant]
}

func NewKeeper(cdc codec.BinaryCodec, addressCodec address.Codec, storeService storetypes.KVStoreService, authority string, bank licenses.BankKeeper, seeds licenses.SeedReader) Keeper {
	if _, err := addressCodec.StringToBytes(authority); err != nil {
		panic(fmt.Errorf("invalid authority address: %w", err))
	}
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		cdc:          cdc,
		addressCodec: addressCodec,
		authority:    authority,
		bank:         bank,
		seeds:        seeds,
		Params:       collections.NewItem(sb, licenses.ParamsKey, "params", codec.CollValue[licenses.Params](cdc)),
		TypePrices:   collections.NewMap(sb, licenses.TypePriceKey, "type_prices", collections.StringKey, collections.Uint64Value),
		AccessGrants: collections.NewMap(sb, licenses.AccessGrantKey, "access_grants", collections.PairKeyCodec(collections.StringKey, collections.Uint64Key), codec.CollValue[licenses.AccessGrant](cdc)),
	}
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

func (k Keeper) GetAuthority() string { return k.authority }
