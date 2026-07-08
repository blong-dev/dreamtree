package keeper

import (
	"context"

	"cosmossdk.io/collections"

	"github.com/blong-dev/dreamtree/x/attest"
)

func (k *Keeper) InitGenesis(ctx context.Context, data *attest.GenesisState) error {
	if err := k.Params.Set(ctx, data.Params); err != nil {
		return err
	}
	for _, a := range data.Attestations {
		if err := k.Attestations.Set(ctx, a.Id, a); err != nil {
			return err
		}
		if err := k.SubjectIndex.Set(ctx, collections.Join(a.Subject, a.Id)); err != nil {
			return err
		}
		if err := k.AttestorIndex.Set(ctx, collections.Join(a.Attestor, a.Id)); err != nil {
			return err
		}
		if a.ProofType == attest.ProofType_PROOF_TYPE_OUTCOME && a.TargetId != 0 {
			if err := k.TargetIndex.Set(ctx, collections.Join(a.TargetId, a.Id)); err != nil {
				return err
			}
		}
	}
	next := data.NextId
	if next == 0 {
		next = 1
	}
	return k.Seq.Set(ctx, next)
}

func (k *Keeper) ExportGenesis(ctx context.Context) (*attest.GenesisState, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	var list []attest.Attestation
	if err := k.Attestations.Walk(ctx, nil, func(_ uint64, v attest.Attestation) (bool, error) {
		list = append(list, v)
		return false, nil
	}); err != nil {
		return nil, err
	}
	next, err := k.Seq.Peek(ctx)
	if err != nil {
		return nil, err
	}
	return &attest.GenesisState{Params: params, Attestations: list, NextId: next}, nil
}
