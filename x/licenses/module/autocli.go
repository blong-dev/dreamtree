package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"github.com/blong-dev/dreamtree/x/licenses"
)

func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: licenses.Query_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{RpcMethod: "TypePrice", Use: "price [data-type]", Short: "The market rate N_a for a data type", PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "data_type"}}},
				{RpcMethod: "Access", Use: "access [buyer] [seed-id]", Short: "Whether a buyer holds access to a seed", PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "buyer"}, {ProtoField: "seed_id"}}},
				{RpcMethod: "Params", Use: "params", Short: "Marketplace parameters"},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: licenses.Msg_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{RpcMethod: "Purchase", Use: "purchase [seed-ids...]", Short: "Buy time-bound access to a swath of seeds", PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "seed_ids", Varargs: true}}},
				{RpcMethod: "SetTypePrice", Skip: true},
				{RpcMethod: "UpdateParams", Skip: true},
			},
		},
	}
}
