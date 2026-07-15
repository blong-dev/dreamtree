package keeper_test

import (
	"context"
	"testing"

	storetypes "cosmossdk.io/store/types"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/stretchr/testify/require"

	"github.com/blong-dev/dreamtree/x/seeds"
	"github.com/blong-dev/dreamtree/x/seeds/keeper"
	seedsmodule "github.com/blong-dev/dreamtree/x/seeds/module"

	"github.com/cosmos/cosmos-sdk/runtime"
)

// mintCall records one photon-seam notification.
type mintCall struct {
	kind     string
	newCount uint32
}

// mockPhotons records OnRecordBatch calls (the x/photons stand-in).
type mockPhotons struct{ calls []mintCall }

func (m *mockPhotons) OnRecordBatch(_ context.Context, kind string, newCount uint32) error {
	m.calls = append(m.calls, mintCall{kind, newCount})
	return nil
}

type fixture struct {
	ctx     sdk.Context
	k       keeper.Keeper
	msg     seeds.MsgServer
	photons *mockPhotons
	addr    string
}

func setup(t *testing.T) *fixture {
	t.Helper()
	key := storetypes.NewKVStoreKey(seeds.ModuleName)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(seedsmodule.AppModule{})
	addrCodec := addresscodec.NewBech32Codec("dream")

	authority, err := addrCodec.BytesToString(make([]byte, 20))
	require.NoError(t, err)
	committerBytes := make([]byte, 20)
	committerBytes[0] = 1
	committer, err := addrCodec.BytesToString(committerBytes)
	require.NoError(t, err)

	k := keeper.NewKeeper(encCfg.Codec, addrCodec, runtime.NewKVStoreService(key), authority)
	ph := &mockPhotons{}
	k.SetPhotonHooks(ph)
	require.NoError(t, k.InitGenesis(testCtx.Ctx, seeds.NewGenesisState()))

	return &fixture{
		ctx:     testCtx.Ctx,
		k:       k,
		msg:     keeper.NewMsgServerImpl(k),
		photons: ph,
		addr:    committer,
	}
}

const root = "aa11bb22cc33dd44ee55ff66aa11bb22cc33dd44ee55ff66aa11bb22cc33dd44"

func TestCommitBatchAllocatesLeafRange(t *testing.T) {
	f := setup(t)
	resp, err := f.msg.CommitBatch(f.ctx, &seeds.MsgCommitBatch{
		Committer: f.addr, Subject: "did:web:x", MerkleRoot: root,
		LeafCount: 5, NewCount: 5, Kind: "record", DataType: "dt.reading@1",
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), resp.FirstId)

	// Every leaf id resolves; the one past the range does not.
	for id := uint64(1); id <= 5; id++ {
		b, ok, err := f.k.BatchOf(f.ctx, id)
		require.NoError(t, err)
		require.True(t, ok, "leaf %d must resolve", id)
		s := keeper.SynthesizeSeed(b, id)
		require.Equal(t, root, s.Commitment)
		require.Equal(t, uint32(id-1), s.LeafIndex)
		require.Equal(t, "record", s.Kind)
	}
	_, ok, err := f.k.BatchOf(f.ctx, 6)
	require.NoError(t, err)
	require.False(t, ok, "unallocated id must not resolve")

	// The marketplace seam sees per-leaf pricing facts.
	dataType, producer, found := f.k.SeedInfo(f.ctx, 3)
	require.True(t, found)
	require.Equal(t, "dt.reading@1", dataType)
	require.Equal(t, f.addr, producer)

	// Photons = seeds: the seam was told 5.
	require.Equal(t, []mintCall{{"record", 5}}, f.photons.calls)
}

func TestConvergedAtomsDoNotRemint(t *testing.T) {
	f := setup(t)
	// 10 leaves under the root, only 3 new (7 re-observed): 3 ids, 3 photons.
	resp, err := f.msg.CommitBatch(f.ctx, &seeds.MsgCommitBatch{
		Committer: f.addr, MerkleRoot: root,
		LeafCount: 10, NewCount: 3, Kind: "record",
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), resp.FirstId)

	_, ok, err := f.k.BatchOf(f.ctx, 3)
	require.NoError(t, err)
	require.True(t, ok)
	_, ok, err = f.k.BatchOf(f.ctx, 4)
	require.NoError(t, err)
	require.False(t, ok, "converged leaves must not allocate ids")

	require.Equal(t, []mintCall{{"record", 3}}, f.photons.calls)
}

func TestMultiBatchResolution(t *testing.T) {
	f := setup(t)
	_, err := f.msg.CommitBatch(f.ctx, &seeds.MsgCommitBatch{
		Committer: f.addr, MerkleRoot: root, LeafCount: 5, NewCount: 5, Kind: "record",
	})
	require.NoError(t, err)
	respB, err := f.msg.CommitBatch(f.ctx, &seeds.MsgCommitBatch{
		Committer: f.addr, MerkleRoot: root, LeafCount: 3, NewCount: 3, Kind: "kg_claim",
	})
	require.NoError(t, err)
	require.Equal(t, uint64(6), respB.FirstId)

	a, ok, _ := f.k.BatchOf(f.ctx, 5)
	require.True(t, ok)
	require.Equal(t, "record", a.Kind)
	b, ok, _ := f.k.BatchOf(f.ctx, 6)
	require.True(t, ok)
	require.Equal(t, "kg_claim", b.Kind)
	_, ok, _ = f.k.BatchOf(f.ctx, 9)
	require.False(t, ok)
}

func TestCommitSeedIsBatchOfOne(t *testing.T) {
	f := setup(t)
	resp, err := f.msg.CommitSeed(f.ctx, &seeds.MsgCommitSeed{
		Committer: f.addr, Commitment: root, Kind: "record", Subject: "did:web:y",
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), resp.Id)

	b, ok, err := f.k.BatchOf(f.ctx, 1)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, uint32(1), b.NewCount)
	require.Equal(t, uint32(1), b.LeafCount)
	require.Equal(t, root, b.MerkleRoot)
	require.Equal(t, []mintCall{{"record", 1}}, f.photons.calls)
}

func TestBatchValidation(t *testing.T) {
	f := setup(t)
	// leaf_count = 0
	_, err := f.msg.CommitBatch(f.ctx, &seeds.MsgCommitBatch{
		Committer: f.addr, MerkleRoot: root, LeafCount: 0, NewCount: 0, Kind: "record",
	})
	require.ErrorIs(t, err, seeds.ErrBadCounts)
	// new_count > leaf_count
	_, err = f.msg.CommitBatch(f.ctx, &seeds.MsgCommitBatch{
		Committer: f.addr, MerkleRoot: root, LeafCount: 2, NewCount: 3, Kind: "record",
	})
	require.ErrorIs(t, err, seeds.ErrBadCounts)
	// aggregate kinds are retired — kind names the leaf
	_, err = f.msg.CommitBatch(f.ctx, &seeds.MsgCommitBatch{
		Committer: f.addr, MerkleRoot: root, LeafCount: 1, NewCount: 1, Kind: "reflow.batch_root",
	})
	require.ErrorIs(t, err, seeds.ErrRetiredKind)
	// non-hex root
	_, err = f.msg.CommitBatch(f.ctx, &seeds.MsgCommitBatch{
		Committer: f.addr, MerkleRoot: "not-hex!", LeafCount: 1, NewCount: 1, Kind: "record",
	})
	require.ErrorIs(t, err, seeds.ErrCommitmentNotHex)
	// supply-griefing guard: new_count above the per-batch cap
	over := seeds.DefaultMaxBatchNewCount + 1
	_, err = f.msg.CommitBatch(f.ctx, &seeds.MsgCommitBatch{
		Committer: f.addr, MerkleRoot: root, LeafCount: over, NewCount: over, Kind: "record",
	})
	require.ErrorIs(t, err, seeds.ErrBadCounts)
	// nothing allocated, nothing minted
	require.Empty(t, f.photons.calls)
	_, ok, _ := f.k.BatchOf(f.ctx, 1)
	require.False(t, ok)
}

func TestPureConvergenceBatch(t *testing.T) {
	f := setup(t)
	// D9 re-fetch shape: all 14 leaves re-observed, nothing new. Anchors the
	// root as provenance; allocates no ids; mints nothing.
	resp, err := f.msg.CommitBatch(f.ctx, &seeds.MsgCommitBatch{
		Committer: f.addr, Subject: "did:web:x", MerkleRoot: root,
		LeafCount: 14, NewCount: 0, Kind: "record",
	})
	require.NoError(t, err)
	require.Equal(t, uint64(0), resp.FirstId)
	require.Equal(t, uint64(1), resp.BatchId)
	require.Empty(t, f.photons.calls, "convergence must not mint")

	// The id sequence is untouched: the next real batch starts at 1.
	next, err := f.msg.CommitBatch(f.ctx, &seeds.MsgCommitBatch{
		Committer: f.addr, MerkleRoot: root, LeafCount: 2, NewCount: 2, Kind: "record",
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), next.FirstId)
	require.Equal(t, []mintCall{{"record", 2}}, f.photons.calls)
}

func TestGenesisRoundtripAndValidate(t *testing.T) {
	f := setup(t)
	for i := 0; i < 3; i++ {
		_, err := f.msg.CommitBatch(f.ctx, &seeds.MsgCommitBatch{
			Committer: f.addr, Subject: "did:web:z", MerkleRoot: root,
			LeafCount: 4, NewCount: 2, Kind: "record",
		})
		require.NoError(t, err)
	}
	exported, err := f.k.ExportGenesis(f.ctx)
	require.NoError(t, err)
	require.Len(t, exported.Batches, 3)
	require.Equal(t, uint64(7), exported.NextId)      // 3 × 2 new + 1
	require.Equal(t, uint64(4), exported.NextBatchId) // 3 batches + 1
	require.NoError(t, exported.Validate())

	// Re-import into a fresh keeper: same resolution behavior.
	g := setup(t)
	require.NoError(t, g.k.InitGenesis(g.ctx, exported))
	b, ok, err := g.k.BatchOf(g.ctx, 4)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, uint64(3), b.FirstSeedId)

	// Overlapping ranges must be rejected.
	bad := *exported
	bad.Batches = append([]seeds.Batch{}, exported.Batches...)
	bad.Batches[1].FirstSeedId = bad.Batches[0].FirstSeedId + 1 // overlaps batch 0 (new_count 2)
	require.Error(t, bad.Validate())

	// Sequence-reissue guards: sequences that don't clear the carried batches
	// would silently overwrite anchored state on the next commit.
	badSeq := *exported
	badSeq.Batches = exported.Batches
	badSeq.NextId = 0 // defaults to 1 — inside batch 0's range
	require.Error(t, badSeq.Validate(), "next_id must clear every carried range")
	badBatchSeq := *exported
	badBatchSeq.Batches = exported.Batches
	badBatchSeq.NextBatchId = exported.Batches[len(exported.Batches)-1].Id // == max id, would reissue
	require.Error(t, badBatchSeq.Validate(), "next_batch_id must clear every carried batch id")

	// uint64 wraparound in a range must be rejected, not wrapped past.
	wrap := *exported
	wrap.Batches = append([]seeds.Batch{}, exported.Batches...)
	wrap.Batches[0].FirstSeedId = ^uint64(0) - 1 // first + new_count(2) overflows
	require.Error(t, wrap.Validate())
}
