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
	for _, s := range data.Seeds {
		if err := k.Seeds.Set(ctx, s.Id, s); err != nil {
			return err
		}
		if s.Subject != "" {
			if err := k.SubjectIndex.Set(ctx, collections.Join(s.Subject, s.Id)); err != nil {
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

// ExportGenesis exports module state to genesis.
func (k *Keeper) ExportGenesis(ctx context.Context) (*seeds.GenesisState, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	var list []seeds.Seed
	if err := k.Seeds.Walk(ctx, nil, func(_ uint64, value seeds.Seed) (bool, error) {
		list = append(list, value)
		return false, nil
	}); err != nil {
		return nil, err
	}
	next, err := k.Seq.Peek(ctx)
	if err != nil {
		return nil, err
	}
	return &seeds.GenesisState{Params: params, Seeds: list, NextId: next}, nil
}
