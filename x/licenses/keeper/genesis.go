package keeper

import (
	"context"

	"cosmossdk.io/collections"

	"github.com/blong-dev/dreamtree/x/licenses"
)

func (k *Keeper) InitGenesis(ctx context.Context, data *licenses.GenesisState) error {
	if err := k.Params.Set(ctx, data.Params); err != nil {
		return err
	}
	for _, tp := range data.TypePrices {
		if err := k.TypePrices.Set(ctx, tp.DataType, tp.Price); err != nil {
			return err
		}
	}
	for _, g := range data.AccessGrants {
		if err := k.AccessGrants.Set(ctx, collections.Join(g.Buyer, g.SeedId), g); err != nil {
			return err
		}
	}
	return nil
}

func (k *Keeper) ExportGenesis(ctx context.Context) (*licenses.GenesisState, error) {
	p, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	gs := &licenses.GenesisState{Params: p}
	if err := k.TypePrices.Walk(ctx, nil, func(dt string, price uint64) (bool, error) {
		gs.TypePrices = append(gs.TypePrices, licenses.TypePrice{DataType: dt, Price: price})
		return false, nil
	}); err != nil {
		return nil, err
	}
	if err := k.AccessGrants.Walk(ctx, nil, func(_ collections.Pair[string, uint64], g licenses.AccessGrant) (bool, error) {
		gs.AccessGrants = append(gs.AccessGrants, g)
		return false, nil
	}); err != nil {
		return nil, err
	}
	return gs, nil
}
