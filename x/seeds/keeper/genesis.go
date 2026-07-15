package keeper

import (
	"context"

	"cosmossdk.io/collections"

	"github.com/blong-dev/dreamtree/x/seeds"
)

// InitGenesis initializes module state from genesis.
func (k *Keeper) InitGenesis(ctx context.Context, data *seeds.GenesisState) error {
	if err := k.Params.Set(ctx, data.Params); err != nil {
		return err
	}
	for _, b := range data.Batches {
		if err := k.Batches.Set(ctx, b.Id, b); err != nil {
			return err
		}
		if b.NewCount > 0 {
			if err := k.RangeIndex.Set(ctx, b.FirstSeedId, b.Id); err != nil {
				return err
			}
		}
		if b.Subject != "" {
			if err := k.SubjectIndex.Set(ctx, collections.Join(b.Subject, b.Id)); err != nil {
				return err
			}
		}
	}
	next := data.NextId
	if next == 0 {
		next = 1
	}
	if err := k.Seq.Set(ctx, next); err != nil {
		return err
	}
	nextBatch := data.NextBatchId
	if nextBatch == 0 {
		nextBatch = 1
	}
	return k.BatchSeq.Set(ctx, nextBatch)
}

// ExportGenesis exports module state to genesis.
func (k *Keeper) ExportGenesis(ctx context.Context) (*seeds.GenesisState, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	var list []seeds.Batch
	if err := k.Batches.Walk(ctx, nil, func(_ uint64, value seeds.Batch) (bool, error) {
		list = append(list, value)
		return false, nil
	}); err != nil {
		return nil, err
	}
	next, err := k.Seq.Peek(ctx)
	if err != nil {
		return nil, err
	}
	nextBatch, err := k.BatchSeq.Peek(ctx)
	if err != nil {
		return nil, err
	}
	return &seeds.GenesisState{Params: params, Batches: list, NextId: next, NextBatchId: nextBatch}, nil
}
