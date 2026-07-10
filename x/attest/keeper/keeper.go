package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/blong-dev/dreamtree/x/attest"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	addressCodec address.Codec
	authority    string

	Schema        collections.Schema
	Params        collections.Item[attest.Params]
	Attestations  collections.Map[uint64, attest.Attestation]
	Seq           collections.Sequence
	SubjectIndex  collections.KeySet[collections.Pair[string, uint64]]
	AttestorIndex collections.KeySet[collections.Pair[string, uint64]]
	TargetIndex   collections.KeySet[collections.Pair[uint64, uint64]]

	// rep is the reputation seam; nil when x/reputation is absent (falls back
	// to baseline_kyc for R, and a no-op OnAttestation).
	rep attest.ReputationKeeper
}

// SetReputationKeeper wires the reputation seam (called once at app assembly).
func (k *Keeper) SetReputationKeeper(rep attest.ReputationKeeper) { k.rep = rep }

// Rep returns the wired reputation seam (nil when x/reputation is absent).
func (k Keeper) Rep() attest.ReputationKeeper { return k.rep }

func NewKeeper(cdc codec.BinaryCodec, addressCodec address.Codec, storeService storetypes.KVStoreService, authority string) Keeper {
	if _, err := addressCodec.StringToBytes(authority); err != nil {
		panic(fmt.Errorf("invalid authority address: %w", err))
	}
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		cdc:           cdc,
		addressCodec:  addressCodec,
		authority:     authority,
		Params:        collections.NewItem(sb, attest.ParamsKey, "params", codec.CollValue[attest.Params](cdc)),
		Attestations:  collections.NewMap(sb, attest.AttestationsKey, "attestations", collections.Uint64Key, codec.CollValue[attest.Attestation](cdc)),
		Seq:           collections.NewSequence(sb, attest.SeqKey, "seq"),
		SubjectIndex:  collections.NewKeySet(sb, attest.SubjectIndexKey, "subject_index", collections.PairKeyCodec(collections.StringKey, collections.Uint64Key)),
		AttestorIndex: collections.NewKeySet(sb, attest.AttestorIndexKey, "attestor_index", collections.PairKeyCodec(collections.StringKey, collections.Uint64Key)),
		TargetIndex:   collections.NewKeySet(sb, attest.TargetIndexKey, "target_index", collections.PairKeyCodec(collections.Uint64Key, collections.Uint64Key)),
	}
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

func (k Keeper) GetAuthority() string { return k.authority }
