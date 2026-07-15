package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"github.com/blong-dev/dreamtree/x/seeds"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: seeds.Query_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod:      "Seed",
					Use:            "seed [id]",
					Short:          "Get a single leaf-seed by id (synthesized from its batch)",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				{
					RpcMethod:      "Batch",
					Use:            "batch [id]",
					Short:          "Get a stored anchoring batch by batch id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				{
					RpcMethod: "Seeds",
					Use:       "seeds",
					Short:     "List anchored batches (one entry per batch; use 'seed [id]' for a leaf)",
				},
				{
					RpcMethod:      "SeedsBySubject",
					Use:            "seeds-by-subject [subject]",
					Short:          "List commitments for a subject (e.g. a wallet DID)",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "subject"}},
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Get the seeds module parameters",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: seeds.Msg_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "CommitSeed",
					Use:       "commit-seed [commitment] [kind]",
					Short:     "Anchor a single commitment on-chain (stored as a batch of one); set --subject and --source-ref as needed",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "commitment"},
						{ProtoField: "kind"},
					},
				},
				{
					RpcMethod: "CommitBatch",
					Use:       "commit-batch [merkle-root] [leaf-count] [new-count] [kind]",
					Short:     "Anchor N leaf-seeds under one Merkle root (seed = atom); new-count excludes converged re-observations",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "merkle_root"},
						{ProtoField: "leaf_count"},
						{ProtoField: "new_count"},
						{ProtoField: "kind"},
					},
				},
				{
					RpcMethod: "UpdateParams",
					Skip:      true,
				},
			},
		},
	}
}
