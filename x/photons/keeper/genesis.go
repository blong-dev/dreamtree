package keeper

import (
	"context"

	"github.com/blong-dev/dreamtree/x/photons"
)

func (k *Keeper) InitGenesis(ctx context.Context, data *photons.GenesisState) error {
	if err := k.Params.Set(ctx, data.Params); err != nil {
		return err
	}
	return k.Minted.Set(ctx, data.Minted)
}

func (k *Keeper) ExportGenesis(ctx context.Context) (*photons.GenesisState, error) {
	p, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	n, err := k.Minted.Peek(ctx)
	if err != nil {
		return nil, err
	}
	return &photons.GenesisState{Params: p, Minted: n}, nil
}
