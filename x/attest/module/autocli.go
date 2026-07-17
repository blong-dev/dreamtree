package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"github.com/blong-dev/dreamtree/x/attest"
)

func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: attest.Query_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{RpcMethod: "Attestation", Use: "attestation [id]", Short: "Get an attestation by id", PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}}},
				{RpcMethod: "AttestationsBySubject", Use: "by-subject [subject]", Short: "List attestations on a work/subject", PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "subject"}}},
				{RpcMethod: "AttestationsByAttestor", Use: "by-attestor [attestor]", Short: "List attestations by a signer", PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "attestor"}}},
				{RpcMethod: "Strength", Use: "strength [id]", Short: "Compute S(att,t) for an attestation (projection)", PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}}},
				{RpcMethod: "WorkValue", Use: "work-value [subject]", Short: "Compute V(w,t) for a subject (projection)", PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "subject"}}},
				{RpcMethod: "StrengthAt", Use: "strength-at [id]", Short: "S(att,t) under override params and/or an as-of clock (the dial; flags --as-of, gRPC for params_override)", PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}}},
				{RpcMethod: "WorkValueAt", Use: "work-value-at [subject]", Short: "V(w,t) under override params and/or an as-of clock (the dial)", PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "subject"}}},
				{RpcMethod: "Params", Use: "params", Short: "Get the attest module parameters"},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: attest.Msg_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod:      "Attest",
					Use:            "attest [subject] [proof-type]",
					Short:          "Record an attestation (proof-type: 1=origin 2=rigor 3=use 4=replication 5=outcome). Use --domain --specificity-bps; for outcome --outcome-kind --target-id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "subject"}, {ProtoField: "proof_type"}},
				},
				{RpcMethod: "UpdateParams", Skip: true},
			},
		},
	}
}
