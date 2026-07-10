package keeper

import (
	"context"

	"cosmossdk.io/collections"

	"github.com/blong-dev/dreamtree/x/reputation"
)

func (k *Keeper) InitGenesis(ctx context.Context, data *reputation.GenesisState) error {
	if err := k.Params.Set(ctx, data.Params); err != nil {
		return err
	}
	for _, cfg := range data.DomainConfigs {
		if err := k.DomainConfigs.Set(ctx, cfg.Path, cfg); err != nil {
			return err
		}
	}
	for _, c := range data.Contributions {
		if err := k.Contributions.Set(ctx, c.Id, c); err != nil {
			return err
		}
		if err := k.SignerIndex.Set(ctx, collections.Join(c.Signer, c.Id)); err != nil {
			return err
		}
	}
	for _, pe := range data.PendingEvents {
		if err := k.Pending.Set(ctx, pe.Id, pe); err != nil {
			return err
		}
		if err := k.CloseTimeIndex.Set(ctx, collections.Join(pe.CloseTime, pe.Id)); err != nil {
			return err
		}
		if pe.Kind == reputation.PendingKind_PENDING_KIND_OUTCOME && pe.CounterTargetPending == 0 {
			if err := k.PendingByTarget.Set(ctx, pe.TargetAttId, pe.Id); err != nil {
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
	nextPending := data.NextPendingId
	if nextPending == 0 {
		nextPending = 1
	}
	return k.PendingSeq.Set(ctx, nextPending)
}

func (k *Keeper) ExportGenesis(ctx context.Context) (*reputation.GenesisState, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	var contribs []reputation.Contribution
	if err := k.Contributions.Walk(ctx, nil, func(_ uint64, v reputation.Contribution) (bool, error) {
		contribs = append(contribs, v)
		return false, nil
	}); err != nil {
		return nil, err
	}
	var cfgs []reputation.DomainConfig
	if err := k.DomainConfigs.Walk(ctx, nil, func(_ string, v reputation.DomainConfig) (bool, error) {
		cfgs = append(cfgs, v)
		return false, nil
	}); err != nil {
		return nil, err
	}
	next, err := k.Seq.Peek(ctx)
	if err != nil {
		return nil, err
	}
	var pending []reputation.PendingEvent
	if err := k.Pending.Walk(ctx, nil, func(_ uint64, v reputation.PendingEvent) (bool, error) {
		pending = append(pending, v)
		return false, nil
	}); err != nil {
		return nil, err
	}
	nextPending, err := k.PendingSeq.Peek(ctx)
	if err != nil {
		return nil, err
	}
	return &reputation.GenesisState{Params: params, Contributions: contribs, DomainConfigs: cfgs, NextId: next, PendingEvents: pending, NextPendingId: nextPending}, nil
}
