package keeper_test

import (
	"context"
	"testing"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/stretchr/testify/require"

	"github.com/blong-dev/dreamtree/x/licenses"
	"github.com/blong-dev/dreamtree/x/licenses/keeper"
	licensesmodule "github.com/blong-dev/dreamtree/x/licenses/module"
)

// mockBank — balances by address string; records transfers.
type mockBank struct {
	balances  map[string]math.Int
	transfers []string // "from>to:amt" for assertion-by-eye on failures
}

// Balances are keyed by RAW address bytes — the global sdk bech32 config need
// not match the test codec's "dream" prefix.
func (m *mockBank) GetBalance(_ context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	b, ok := m.balances[string(addr)]
	if !ok {
		b = math.ZeroInt()
	}
	return sdk.NewCoin(denom, b)
}

func (m *mockBank) SendCoins(_ context.Context, from, to sdk.AccAddress, amt sdk.Coins) error {
	a := amt.AmountOf(licenses.PhotonDenom)
	m.balances[string(from)] = m.balances[string(from)].Sub(a)
	if _, ok := m.balances[string(to)]; !ok {
		m.balances[string(to)] = math.ZeroInt()
	}
	m.balances[string(to)] = m.balances[string(to)].Add(a)
	m.transfers = append(m.transfers, string(from)+">"+string(to)+":"+a.String())
	return nil
}

// mockSeeds — fixed (dataType, producer) per id.
type mockSeeds struct{ producers map[uint64]string }

func (m *mockSeeds) SeedInfo(_ context.Context, id uint64) (string, string, bool) {
	p, ok := m.producers[id]
	if !ok {
		return "", "", false
	}
	return "dt.record", p, true
}

type fixture struct {
	ctx         sdk.Context
	k           keeper.Keeper
	msg         licenses.MsgServer
	bank        *mockBank
	authority   string
	buyer       string
	producerKey string // raw-bytes balance key for the producer
}

func setup(t *testing.T) *fixture {
	t.Helper()
	key := storetypes.NewKVStoreKey(licenses.ModuleName)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(licensesmodule.AppModule{})
	addrCodec := addresscodec.NewBech32Codec("dream")

	mk := func(seed byte) (string, []byte) {
		b := make([]byte, 20)
		b[0] = seed
		s, err := addrCodec.BytesToString(b)
		require.NoError(t, err)
		return s, b
	}
	authority, _ := mk(0)
	buyer, buyerRaw := mk(1)
	producer, producerRaw := mk(2)

	bank := &mockBank{balances: map[string]math.Int{string(buyerRaw): math.NewInt(100 * licenses.UphotonPerPhoton)}}
	seedsMock := &mockSeeds{producers: map[uint64]string{1: producer, 2: producer, 3: producer}}
	k := keeper.NewKeeper(encCfg.Codec, addrCodec, runtime.NewKVStoreService(key), authority, bank, seedsMock)
	// No treasury → no toll/tax; the constant-price arithmetic stands alone.
	p := licenses.DefaultParams()
	p.TreasuryRecipient = ""
	require.NoError(t, k.Params.Set(testCtx.Ctx, p))
	return &fixture{ctx: testCtx.Ctx, k: k, msg: keeper.NewMsgServerImpl(k), bank: bank, authority: authority, buyer: buyer, producerKey: string(producerRaw)}
}

// Constant pricing (upgrade-1 R4): a 3-seed swath at access_duration_days=1
// costs exactly 3 photons — no price table involved.
func TestPurchaseConstantPrice(t *testing.T) {
	f := setup(t)
	resp, err := f.k.Purchase(f.ctx, f.buyer, []uint64{1, 2, 3})
	require.NoError(t, err)
	require.Equal(t, uint32(3), resp.SeedsPurchased)
	require.Equal(t, uint64(3*licenses.UphotonPerPhoton), resp.TotalPaid)
	require.Equal(t, math.NewInt(3*licenses.UphotonPerPhoton), f.bank.balances[f.producerKey])

	// Unresolvable seeds are skipped, priced seeds still clear.
	resp, err = f.k.Purchase(f.ctx, f.buyer, []uint64{1, 99})
	require.NoError(t, err)
	require.Equal(t, uint32(1), resp.SeedsPurchased)
	require.Equal(t, uint64(licenses.UphotonPerPhoton), resp.TotalPaid)
}

// SetTypePrice is retired (there is no price to set).
func TestSetTypePriceRetired(t *testing.T) {
	f := setup(t)
	_, err := f.msg.SetTypePrice(f.ctx, &licenses.MsgSetTypePrice{Authority: f.authority, DataType: "dt.record", Price: 5})
	require.ErrorIs(t, err, licenses.ErrRetired)
}
