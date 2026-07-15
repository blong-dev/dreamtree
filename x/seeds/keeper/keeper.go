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
//
// Storage model (leaf model — docs/specs/seed-atom-conformance.md): the stored
// unit is the Batch; a batch registers new_count leaf-seeds
// [first_seed_id, first_seed_id+new_count) under one Merkle root. Individual
// Seed objects are synthesized on read (see resolve.go). O(1) state per batch
// regardless of leaf count.
type Keeper struct {
	cdc          codec.BinaryCodec
	addressCodec address.Codec

	// authority is the address allowed to run MsgUpdateParams (usually x/gov).
	authority string

	Schema collections.Schema
	Params collections.Item[seeds.Params]
	// Batches maps batch_id -> the anchored batch (the stored unit).
	Batches collections.Map[uint64, seeds.Batch]
	// BatchSeq assigns the monotonic batch id.
	BatchSeq collections.Sequence
	// RangeIndex maps first_seed_id -> batch_id (ordered; a leaf id resolves
	// to the greatest first_seed_id <= id, then bounds-checks new_count).
	RangeIndex collections.Map[uint64, uint64]
	// Seq assigns the monotonic leaf-seed id (advanced by new_count per batch).
	Seq collections.Sequence
	// SubjectIndex is a (subject, first_seed_id) key set — one entry per batch.
	SubjectIndex collections.KeySet[collections.Pair[string, uint64]]

	// photons is the ingestion mint seam; nil when x/photons is absent.
	photons seeds.PhotonHooks
}

// SetPhotonHooks wires the ingestion mint seam (once, at app assembly).
func (k *Keeper) SetPhotonHooks(h seeds.PhotonHooks) { k.photons = h }

// Photons returns the wired ingestion seam (nil when x/photons is absent).
func (k Keeper) Photons() seeds.PhotonHooks { return k.photons }

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
		Batches:      collections.NewMap(sb, seeds.BatchesKey, "batches", collections.Uint64Key, codec.CollValue[seeds.Batch](cdc)),
		BatchSeq:     collections.NewSequence(sb, seeds.BatchSeqKey, "batch_seq"),
		RangeIndex:   collections.NewMap(sb, seeds.RangeIndexKey, "range_index", collections.Uint64Key, collections.Uint64Value),
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
