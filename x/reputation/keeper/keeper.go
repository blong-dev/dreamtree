package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/blong-dev/dreamtree/x/reputation"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	addressCodec address.Codec
	authority    string

	Schema        collections.Schema
	Params        collections.Item[reputation.Params]
	Contributions collections.Map[uint64, reputation.Contribution]
	Seq           collections.Sequence
	SignerIndex   collections.KeySet[collections.Pair[string, uint64]]
	DomainConfigs collections.Map[string, reputation.DomainConfig]
}

func NewKeeper(cdc codec.BinaryCodec, addressCodec address.Codec, storeService storetypes.KVStoreService, authority string) Keeper {
	if _, err := addressCodec.StringToBytes(authority); err != nil {
		panic(fmt.Errorf("invalid authority address: %w", err))
	}
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		cdc:           cdc,
		addressCodec:  addressCodec,
		authority:     authority,
		Params:        collections.NewItem(sb, reputation.ParamsKey, "params", codec.CollValue[reputation.Params](cdc)),
		Contributions: collections.NewMap(sb, reputation.ContributionsKey, "contributions", collections.Uint64Key, codec.CollValue[reputation.Contribution](cdc)),
		Seq:           collections.NewSequence(sb, reputation.SeqKey, "seq"),
		SignerIndex:   collections.NewKeySet(sb, reputation.SignerIndexKey, "signer_index", collections.PairKeyCodec(collections.StringKey, collections.Uint64Key)),
		DomainConfigs: collections.NewMap(sb, reputation.DomainConfigKey, "domain_configs", collections.StringKey, codec.CollValue[reputation.DomainConfig](cdc)),
	}
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

func (k Keeper) GetAuthority() string { return k.authority }
