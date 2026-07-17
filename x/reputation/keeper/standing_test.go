package keeper_test

import (
	"testing"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/stretchr/testify/require"

	"github.com/blong-dev/dreamtree/x/reputation"
	"github.com/blong-dev/dreamtree/x/reputation/keeper"
	reputationmodule "github.com/blong-dev/dreamtree/x/reputation/module"
)

type fixture struct {
	ctx       sdk.Context
	k         keeper.Keeper
	msg       reputation.MsgServer
	authority string
	addr      string // a plain (non-authority) address
	nextID    uint64
}

func setup(t *testing.T) *fixture {
	t.Helper()
	key := storetypes.NewKVStoreKey(reputation.ModuleName)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(reputationmodule.AppModule{})
	addrCodec := addresscodec.NewBech32Codec("dream")

	authority, err := addrCodec.BytesToString(make([]byte, 20))
	require.NoError(t, err)
	b := make([]byte, 20)
	b[0] = 1
	addr, err := addrCodec.BytesToString(b)
	require.NoError(t, err)

	k := keeper.NewKeeper(encCfg.Codec, addrCodec, runtime.NewKVStoreService(key), authority)
	require.NoError(t, k.Params.Set(testCtx.Ctx, reputation.DefaultParams()))
	return &fixture{
		ctx:       testCtx.Ctx,
		k:         k,
		msg:       keeper.NewMsgServerImpl(k),
		authority: authority,
		addr:      addr,
		nextID:    1,
	}
}

// contribute settles a contribution directly into the store (the walk reads
// id order = settlement order).
func (f *fixture) contribute(t *testing.T, signer, domain, magnitude string, bucket reputation.RateBucket) {
	t.Helper()
	c := reputation.Contribution{
		Id:         f.nextID,
		Signer:     signer,
		Domain:     domain,
		Magnitude:  math.LegacyMustNewDecFromStr(magnitude),
		RateBucket: bucket,
	}
	require.NoError(t, f.k.Contributions.Set(f.ctx, c.Id, c))
	require.NoError(t, f.k.SignerIndex.Set(f.ctx, collections.Join(signer, c.Id)))
	f.nextID++
}

// Standing starts at ZERO (upgrade-1 R2): an unknown address has no baseline;
// a governed MsgSetVerified grant confers it; revocation removes it.
func TestStandingStartsAtZero(t *testing.T) {
	f := setup(t)
	dom := "science/biology"

	require.True(t, f.k.StandingOf(f.ctx, f.addr, dom).IsZero())

	// Non-authority may not grant.
	_, err := f.msg.SetVerified(f.ctx, &reputation.MsgSetVerified{Authority: f.addr, Address: f.addr, Verified: true})
	require.Error(t, err)

	// Governed grant confers the baseline floor.
	_, err = f.msg.SetVerified(f.ctx, &reputation.MsgSetVerified{Authority: f.authority, Address: f.addr, Verified: true})
	require.NoError(t, err)
	require.Equal(t, "1.000000000000000000", f.k.StandingOf(f.ctx, f.addr, dom).String())

	// Revocation removes it.
	_, err = f.msg.SetVerified(f.ctx, &reputation.MsgSetVerified{Authority: f.authority, Address: f.addr, Verified: false})
	require.NoError(t, err)
	require.True(t, f.k.StandingOf(f.ctx, f.addr, dom).IsZero())
}

// The prep-school shape is dead (upgrade-1 R5): forty identical endorsements
// aggregate paper-shape to at most e_cap_mult × the strongest single one —
// they do NOT stack linearly to the saturation knee.
func TestEndorsementBreadthBounded(t *testing.T) {
	f := setup(t)
	dom := "science/biology"

	// Two equal endorsements (e = 0.25 each): cap = 2 × 0.25 = 0.5,
	// E = 0.5·(1 − (1−0.5)²) = 0.375 — exact in fixed-point.
	for i := 0; i < 2; i++ {
		f.contribute(t, f.addr, dom, "0.25", reputation.RateBucket_RATE_BUCKET_ENDORSEMENT)
	}
	require.Equal(t, "0.375000000000000000", f.k.StandingOf(f.ctx, f.addr, dom).String())

	// Thirty-eight more (the enforced-recs crowd): E approaches the cap 0.5
	// asymptotically. Linear stacking would have given 10.0.
	for i := 0; i < 38; i++ {
		f.contribute(t, f.addr, dom, "0.25", reputation.RateBucket_RATE_BUCKET_ENDORSEMENT)
	}
	got := f.k.StandingOf(f.ctx, f.addr, dom)
	require.True(t, got.LTE(math.LegacyMustNewDecFromStr("0.5")), "crowd breached the cap: %s", got)
	require.True(t, got.GT(math.LegacyMustNewDecFromStr("0.49")), "fold collapsed: %s", got)
}

// A reversal negation (negative ENDORSEMENT entry) subtracts from the folded
// total, floored at zero — reversed endorsements cannot leave a debt.
func TestEndorsementReversalSubtracts(t *testing.T) {
	f := setup(t)
	dom := "science/biology"
	f.contribute(t, f.addr, dom, "0.25", reputation.RateBucket_RATE_BUCKET_ENDORSEMENT)
	f.contribute(t, f.addr, dom, "0.25", reputation.RateBucket_RATE_BUCKET_ENDORSEMENT)
	f.contribute(t, f.addr, dom, "-0.25", reputation.RateBucket_RATE_BUCKET_ENDORSEMENT)
	// fold(0.25, 0.25) = 0.375; − 0.25 = 0.125
	require.Equal(t, "0.125000000000000000", f.k.StandingOf(f.ctx, f.addr, dom).String())

	f.contribute(t, f.addr, dom, "-1.0", reputation.RateBucket_RATE_BUCKET_ENDORSEMENT)
	require.True(t, f.k.StandingOf(f.ctx, f.addr, dom).IsZero())
}

// Work contributions still walk with the Z2 running floor — from a ZERO base
// for the unverified: a deep negative excursion is forgiven where it happened
// and later work counts in full.
func TestRunningFloorFromZeroBase(t *testing.T) {
	f := setup(t)
	dom := "science/biology"
	f.contribute(t, f.addr, dom, "-5.0", reputation.RateBucket_RATE_BUCKET_USE)
	f.contribute(t, f.addr, dom, "1.0", reputation.RateBucket_RATE_BUCKET_USE)
	require.Equal(t, "1.000000000000000000", f.k.StandingOf(f.ctx, f.addr, dom).String())
}
