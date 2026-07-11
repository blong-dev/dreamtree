package app_test

import (
	"testing"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/stretchr/testify/require"

	"github.com/blong-dev/dreamtree/app"
	_ "github.com/blong-dev/dreamtree/app/params" // init(): set the "dream" bech32 prefix
	attestkeeper "github.com/blong-dev/dreamtree/x/attest/keeper"
	licenseskeeper "github.com/blong-dev/dreamtree/x/licenses/keeper"
	photonskeeper "github.com/blong-dev/dreamtree/x/photons/keeper"
	repkeeper "github.com/blong-dev/dreamtree/x/reputation/keeper"
	seedskeeper "github.com/blong-dev/dreamtree/x/seeds/keeper"
)

type emptyAppOpts struct{}

func (emptyAppOpts) Get(string) interface{} { return nil }

// TestModuleSeamsBind asserts the cross-module seams actually wire at assembly.
// They are optional depinject inputs bound by implicit single-implementer
// resolution — if that ever silently resolves to nil, reputation would stop
// accruing and photons would stop minting while the chain still runs. This test
// is the guard against that regression (the pre-launch audit's top ask).
func TestModuleSeamsBind(t *testing.T) {
	var (
		attestK   attestkeeper.Keeper
		seedsK    seedskeeper.Keeper
		repK      repkeeper.Keeper
		photonsK  photonskeeper.Keeper
		licensesK licenseskeeper.Keeper
	)
	err := depinject.Inject(
		depinject.Configs(
			app.AppConfig(),
			depinject.Supply(log.NewNopLogger(), servertypes.AppOptions(emptyAppOpts{})),
		),
		&attestK, &seedsK, &repK, &photonsK, &licensesK,
	)
	require.NoError(t, err, "app must assemble")

	// The optional seams must be non-nil (else the dependent logic is dead).
	require.NotNil(t, attestK.Rep(), "attest→reputation seam unbound: reputation would be silently dead")
	require.NotNil(t, seedsK.Photons(), "seeds→photons seam unbound: photon minting would be silently dead")

	// licensesK extracting at all proves its REQUIRED bank + seed-reader inputs
	// bound (Inject would error otherwise) — the marketplace can read seeds and
	// move photons.
}
