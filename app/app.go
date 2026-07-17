package app

import (
	_ "embed"
	"io"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/core/appconfig"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	clienthelpers "cosmossdk.io/client/v2/helpers"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	attestkeeper "github.com/blong-dev/dreamtree/x/attest/keeper"
	licenseskeeper "github.com/blong-dev/dreamtree/x/licenses/keeper"
	reputationkeeper "github.com/blong-dev/dreamtree/x/reputation/keeper"

	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	_ "cosmossdk.io/api/cosmos/tx/config/v1"          // import for side-effects
	_ "cosmossdk.io/x/upgrade"                        // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/auth"           // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/bank"           // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/consensus"      // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/distribution"   // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/gov"            // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/staking"        // import for side-effects

	_ "github.com/blong-dev/dreamtree/x/attest/module"     // import for side-effects
	_ "github.com/blong-dev/dreamtree/x/licenses/module"   // import for side-effects
	_ "github.com/blong-dev/dreamtree/x/photons/module"    // import for side-effects
	_ "github.com/blong-dev/dreamtree/x/reputation/module" // import for side-effects
	_ "github.com/blong-dev/dreamtree/x/seeds/module"      // import for side-effects

	// pulsar (protobuf-v2) descriptors — registered for client/v2 autocli
	// (enum flags + typed query rendering).
	_ "github.com/blong-dev/dreamtree/api/dreamtree/attest/v1"
	_ "github.com/blong-dev/dreamtree/api/dreamtree/licenses/v1"
	_ "github.com/blong-dev/dreamtree/api/dreamtree/photons/v1"
	_ "github.com/blong-dev/dreamtree/api/dreamtree/reputation/v1"
	_ "github.com/blong-dev/dreamtree/api/dreamtree/seeds/v1"
)

// DefaultNodeHome default home directories for the application daemon
var DefaultNodeHome string

//go:embed app.yaml
var AppConfigYAML []byte

var (
	_ runtime.AppI            = (*DreamtreeApp)(nil)
	_ servertypes.Application = (*DreamtreeApp)(nil)
)

// DreamtreeApp extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type DreamtreeApp struct {
	*runtime.App
	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry codectypes.InterfaceRegistry

	// keepers
	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	GovKeeper             *govkeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	ConsensusParamsKeeper consensuskeeper.Keeper
	LicensesKeeper        licenseskeeper.Keeper
	ReputationKeeper      reputationkeeper.Keeper
	AttestKeeper          attestkeeper.Keeper

	// simulation manager
	sm *module.SimulationManager
}

func init() {
	var err error
	clienthelpers.EnvPrefix = "MINI"
	DefaultNodeHome, err = clienthelpers.GetNodeHomeDirectory(".dreamtreed")
	if err != nil {
		panic(err)
	}
}

// AppConfig returns the default app config.
func AppConfig() depinject.Config {
	return depinject.Configs(
		appconfig.LoadYAML(AppConfigYAML),
		depinject.Supply(
			&appv1alpha1.Config{}, // hack until https://github.com/cosmos/cosmos-sdk/pull/21042
			// supply custom module basics
			map[string]module.AppModuleBasic{
				genutiltypes.ModuleName: genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
			},
		),
	)
}

// NewDreamtreeApp returns a reference to an initialized DreamtreeApp.
func NewDreamtreeApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) (*DreamtreeApp, error) {
	var (
		app        = &DreamtreeApp{}
		appBuilder *runtime.AppBuilder
	)

	if err := depinject.Inject(
		depinject.Configs(
			AppConfig(),
			depinject.Supply(
				logger,
				appOpts,
			),
		),
		&appBuilder,
		&app.appCodec,
		&app.legacyAmino,
		&app.txConfig,
		&app.interfaceRegistry,
		&app.AccountKeeper,
		&app.BankKeeper,
		&app.StakingKeeper,
		&app.DistrKeeper,
		&app.GovKeeper,
		&app.UpgradeKeeper,
		&app.ConsensusParamsKeeper,
		&app.LicensesKeeper,
		&app.ReputationKeeper,
		&app.AttestKeeper,
	); err != nil {
		return nil, err
	}

	app.App = appBuilder.Build(db, traceStore, baseAppOptions...)

	app.RegisterUpgradeHandlers()

	// register streaming services
	if err := app.RegisterStreamingServices(appOpts, app.kvStoreKeys()); err != nil {
		return nil, err
	}

	/****  Module Options ****/

	// create the simulation manager and define the order of the modules for deterministic simulations
	// NOTE: this is not required apps that don't use the simulator for fuzz testing transactions
	app.sm = module.NewSimulationManagerFromAppModules(app.ModuleManager.Modules, make(map[string]module.AppModuleSimulation, 0))
	app.sm.RegisterStoreDecoders()

	if err := app.Load(loadLatest); err != nil {
		return nil, err
	}

	return app, nil
}

// LegacyAmino returns DreamtreeApp's amino codec.
func (app *DreamtreeApp) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// GetKey returns the KVStoreKey for the provided store key.
func (app *DreamtreeApp) GetKey(storeKey string) *storetypes.KVStoreKey {
	sk := app.UnsafeFindStoreKey(storeKey)
	kvStoreKey, ok := sk.(*storetypes.KVStoreKey)
	if !ok {
		return nil
	}
	return kvStoreKey
}

func (app *DreamtreeApp) kvStoreKeys() map[string]*storetypes.KVStoreKey {
	keys := make(map[string]*storetypes.KVStoreKey)
	for _, k := range app.GetStoreKeys() {
		if kv, ok := k.(*storetypes.KVStoreKey); ok {
			keys[kv.Name()] = kv
		}
	}

	return keys
}

// SimulationManager implements the SimulationApp interface
func (app *DreamtreeApp) SimulationManager() *module.SimulationManager {
	return app.sm
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *DreamtreeApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	app.App.RegisterAPIRoutes(apiSvr, apiConfig)
	// register swagger API in app.go so that other applications can override easily
	if err := server.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}
}
