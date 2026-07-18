package keeper_test

import (
	"context"
	"testing"

	storetypes "cosmossdk.io/store/types"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/stretchr/testify/require"

	"github.com/blong-dev/dreamtree/x/photons"
	"github.com/blong-dev/dreamtree/x/photons/keeper"
	photonsmodule "github.com/blong-dev/dreamtree/x/photons/module"
)

type mockBank struct{ minted sdk.Coins }

func (m *mockBank) MintCoins(_ context.Context, _ string, amt sdk.Coins) error {
	m.minted = m.minted.Add(amt...)
	return nil
}
func (m *mockBank) SendCoinsFromModuleToAccount(_ context.Context, _ string, _ sdk.AccAddress, _ sdk.Coins) error {
	return nil
}

// ALL kinds mint (upgrade-1 R3): the mintable_kinds whitelist gates nothing —
// every atom is an observation, and the peg is photons = distinct atoms.
func TestAllKindsMint(t *testing.T) {
	key := storetypes.NewKVStoreKey(photons.ModuleName)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(photonsmodule.AppModule{})
	addrCodec := addresscodec.NewBech32Codec("dream")
	authority, err := addrCodec.BytesToString(make([]byte, 20))
	require.NoError(t, err)

	bank := &mockBank{}
	k := keeper.NewKeeper(encCfg.Codec, addrCodec, runtime.NewKVStoreService(key), authority, bank)
	require.NoError(t, k.Params.Set(testCtx.Ctx, photons.DefaultParams()))

	// A kind OUTSIDE the legacy whitelist mints all the same.
	require.NoError(t, k.OnRecordBatch(testCtx.Ctx, "some.novel.kind", 7))
	require.Equal(t, int64(7*photons.UphotonPerPhoton), bank.minted.AmountOf(photons.PhotonDenom).Int64())

	minted, err := k.Minted.Peek(testCtx.Ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(7), minted)

	// Convergence still never re-mints: new_count 0 is a no-op.
	require.NoError(t, k.OnRecordBatch(testCtx.Ctx, "record", 0))
	minted, err = k.Minted.Peek(testCtx.Ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(7), minted)
}

// TestMintCeiling exercises the per-block mint ceiling (trust-layer W3):
// accumulation within a block, rejection past the ceiling with NO mint, and a
// lazy reset when the block advances.
func TestMintCeiling(t *testing.T) {
	key := storetypes.NewKVStoreKey(photons.ModuleName)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(photonsmodule.AppModule{})
	addrCodec := addresscodec.NewBech32Codec("dream")
	authority, err := addrCodec.BytesToString(make([]byte, 20))
	require.NoError(t, err)

	bank := &mockBank{}
	k := keeper.NewKeeper(encCfg.Codec, addrCodec, runtime.NewKVStoreService(key), authority, bank)
	require.NoError(t, k.Params.Set(testCtx.Ctx, photons.DefaultParams()))

	uphoton := func(photonCount uint64) int64 { return int64(photonCount * photons.UphotonPerPhoton) }

	// Block 100: mint half the ceiling — OK, accumulator = 50k.
	ctx100 := testCtx.Ctx.WithBlockHeight(100)
	require.NoError(t, k.OnRecordBatch(ctx100, "record", 50_000))
	require.Equal(t, uphoton(50_000), bank.minted.AmountOf(photons.PhotonDenom).Int64())

	// Block 100: a batch that would push past the ceiling (50k+60k > 100k) is
	// REJECTED, and nothing is minted (bank + Minted seq unchanged).
	err = k.OnRecordBatch(ctx100, "record", 60_000)
	require.ErrorIs(t, err, photons.ErrMintCeilingExceeded)
	require.Equal(t, uphoton(50_000), bank.minted.AmountOf(photons.PhotonDenom).Int64())
	minted, err := k.Minted.Peek(ctx100)
	require.NoError(t, err)
	require.Equal(t, uint64(50_000), minted)

	// Block 100: exactly hitting the ceiling (50k+50k == 100k) is allowed.
	require.NoError(t, k.OnRecordBatch(ctx100, "record", 50_000))
	require.Equal(t, uphoton(100_000), bank.minted.AmountOf(photons.PhotonDenom).Int64())

	// Block 100: any further mint (even 1) now exceeds the ceiling.
	require.ErrorIs(t, k.OnRecordBatch(ctx100, "record", 1), photons.ErrMintCeilingExceeded)

	// Block 101: the accumulator resets — a near-ceiling batch mints fine.
	ctx101 := testCtx.Ctx.WithBlockHeight(101)
	require.NoError(t, k.OnRecordBatch(ctx101, "record", 90_000))
	require.Equal(t, uphoton(190_000), bank.minted.AmountOf(photons.PhotonDenom).Int64())

	// The soft-warn event fired (90k >= 25k warn threshold) in block 101.
	warned := false
	for _, e := range ctx101.EventManager().Events() {
		if e.Type == "photon_mint_ceiling_warn" {
			warned = true
		}
	}
	require.True(t, warned, "expected photon_mint_ceiling_warn event past the soft threshold")
}

// TestHonestBatchWellUnderCeiling: a realistic batch (observed max ~732) is
// nowhere near the ceiling, so honest ingestion is never rejected.
func TestHonestBatchWellUnderCeiling(t *testing.T) {
	key := storetypes.NewKVStoreKey(photons.ModuleName)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(photonsmodule.AppModule{})
	addrCodec := addresscodec.NewBech32Codec("dream")
	authority, err := addrCodec.BytesToString(make([]byte, 20))
	require.NoError(t, err)
	k := keeper.NewKeeper(encCfg.Codec, addrCodec, runtime.NewKVStoreService(key), authority, &mockBank{})
	require.NoError(t, k.Params.Set(testCtx.Ctx, photons.DefaultParams()))

	ctx := testCtx.Ctx.WithBlockHeight(1)
	// 100 max-size batches in one block still fits comfortably.
	for i := 0; i < 100; i++ {
		require.NoError(t, k.OnRecordBatch(ctx, "record", 732))
	}
}
