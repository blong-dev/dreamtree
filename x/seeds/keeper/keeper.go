package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/blong-dev/dreamtree/x/seeds"
)

// Keeper manages the seeds module state.
type Keeper struct {
	cdc          codec.BinaryCodec
	addressCodec address.Codec

	// authority is the address allowed to run MsgUpdateParams (usually x/gov).
	authority string

	Schema collections.Schema
	Params collections.Item[seeds.Params]
	// Seeds maps a global id -> the anchored commitment.
	Seeds collections.Map[uint64, seeds.Seed]
	// Seq assigns the monotonic commitment id.
	Seq collections.Sequence
	// SubjectIndex is a (subject, id) key set for by-subject lookups.
	SubjectIndex collections.KeySet[collections.Pair[string, uint64]]
}

// NewKeeper creates a new seeds Keeper.
func NewKeeper(cdc codec.BinaryCodec, addressCodec address.Codec, storeService storetypes.KVStoreService, authority string) Keeper {
	if _, err := addressCodec.StringToBytes(authority); err != nil {
		panic(fmt.Errorf("invalid authority address: %w", err))
	}

	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		cdc:          cdc,
		addressCodec: addressCodec,
		authority:    authority,
		Params:       collections.NewItem(sb, seeds.ParamsKey, "params", codec.CollValue[seeds.Params](cdc)),
		Seeds:        collections.NewMap(sb, seeds.SeedsKey, "seeds", collections.Uint64Key, codec.CollValue[seeds.Seed](cdc)),
		Seq:          collections.NewSequence(sb, seeds.SeqKey, "seq"),
		SubjectIndex: collections.NewKeySet(sb, seeds.SubjectIndexKey, "subject_index", collections.PairKeyCodec(collections.StringKey, collections.Uint64Key)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string { return k.authority }
