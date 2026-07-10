package module

import (
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"

	"github.com/cosmos/cosmos-sdk/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/blong-dev/dreamtree/x/reputation"
	"github.com/blong-dev/dreamtree/x/reputation/keeper"
)

var _ appmodule.AppModule = AppModule{}

func (am AppModule) IsOnePerModuleType() {}
func (am AppModule) IsAppModule()        {}

func init() {
	appconfig.Register(
		&reputation.Module{},
		appconfig.Provide(ProvideModule),
	)
}

type ModuleInputs struct {
	depinject.In

	Cdc          codec.Codec
	StoreService store.KVStoreService
	AddressCodec address.Codec

	Config *reputation.Module
}

type ModuleOutputs struct {
	depinject.Out

	Module appmodule.AppModule
	Keeper keeper.Keeper
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	authority := authtypes.NewModuleAddress("gov")
	if in.Config.Authority != "" {
		authority = authtypes.NewModuleAddressOrBech32Address(in.Config.Authority)
	}
	k := keeper.NewKeeper(in.Cdc, in.AddressCodec, in.StoreService, authority.String())
	m := NewAppModule(in.Cdc, k)
	return ModuleOutputs{Module: m, Keeper: k}
}
