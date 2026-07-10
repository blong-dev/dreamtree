package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"github.com/blong-dev/dreamtree/x/photons"
)

func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: photons.Query_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{RpcMethod: "Supply", Use: "supply", Short: "Photons minted at ingestion (= data-seed count; the peg)"},
				{RpcMethod: "Params", Use: "params", Short: "Photon module parameters"},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:           photons.Msg_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{{RpcMethod: "UpdateParams", Skip: true}},
		},
	}
}
