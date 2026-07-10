package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"github.com/blong-dev/dreamtree/x/reputation"
)

func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: reputation.Query_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{RpcMethod: "Reputation", Use: "reputation [signer] [domain]", Short: "R(signer,domain,t) — the read projection", PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "signer"}, {ProtoField: "domain"}}},
				{RpcMethod: "Contributions", Use: "contributions [signer]", Short: "A signer's reputation contributions", PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "signer"}}},
				{RpcMethod: "DomainConfig", Use: "domain-config [path]", Short: "A domain's saturation/obsolescence tiers", PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "path"}}},
				{RpcMethod: "PendingEvents", Use: "pending", Short: "Reputation events currently in their review window"},
				{RpcMethod: "Params", Use: "params", Short: "Reputation module parameters"},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:           reputation.Msg_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{{RpcMethod: "UpdateParams", Skip: true}, {RpcMethod: "SetDomainConfig", Skip: true}},
		},
	}
}
