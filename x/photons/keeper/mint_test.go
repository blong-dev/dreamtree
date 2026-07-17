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
