package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/blong-dev/dreamtree/x/photons"
)

// OnRecordBatch is the x/seeds ingestion hook. For a data-contribution kind it
// mints EXACTLY ONE photon per NEW leaf-seed (photons = seeds = distinct atoms;
// converged re-observations do not re-mint) and routes them to the storer-reward
// recipient (dt-as-first-storer today; per-storer distribution when a storer set
// exists). This is the ONLY minting event in the protocol.
func (k Keeper) OnRecordBatch(ctx context.Context, kind string, newCount uint32) error {
	if newCount == 0 {
		return nil
	}
	p, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}
	// Per-block mint ceiling (trust-layer W3): bound the total photons a single
	// block can mint, so a leaked ANCHORD_TOKEN can't mint an unbounded amount
	// in a burst. Honest ingestion is orders of magnitude under the ceiling, so
	// this never trips for real traffic; it caps a flood and each hit is an
	// on-chain rejection. Accumulator resets lazily when the block advances.
	h := sdk.UnwrapSDKContext(ctx).BlockHeight()
	blockMinted := uint64(0)
	if mh, err := k.MintHeight.Get(ctx); err == nil && mh == h {
		if bm, err := k.BlockMinted.Get(ctx); err == nil {
			blockMinted = bm
		}
	} else if err != nil && !errors.IsOf(err, collections.ErrNotFound) {
		return err
	}
	newBlockTotal := blockMinted + uint64(newCount)
	if newBlockTotal > photons.MaxMintPerBlock {
		return errors.Wrapf(photons.ErrMintCeilingExceeded,
			"block %d: %d already minted + %d requested > %d ceiling",
			h, blockMinted, newCount, photons.MaxMintPerBlock)
	}

	// ALL kinds mint (upgrade-1 R3, owner 2026-07-16: "every atom is an
	// observation"). The mintable_kinds param is a retired pre-DT-18 artifact;
	// the field survives in the proto for wire compatibility but gates nothing.
	// The peg is photons = distinct atoms, unqualified.
	amt := math.NewInt(int64(newCount)).MulRaw(photons.UphotonPerPhoton)
	coins := sdk.NewCoins(sdk.NewCoin(photons.PhotonDenom, amt))
	if err := k.bank.MintCoins(ctx, photons.ModuleName, coins); err != nil {
		return err
	}
	if p.StorerRewardRecipient != "" {
		addr, err := k.addressCodec.StringToBytes(p.StorerRewardRecipient)
		if err != nil {
			return err
		}
		if err := k.bank.SendCoinsFromModuleToAccount(ctx, photons.ModuleName, addr, coins); err != nil {
			return err
		}
	}
	// Minted counts photons (= distinct atoms), not base units.
	cur, err := k.Minted.Peek(ctx)
	if err != nil {
		return err
	}
	if err := k.Minted.Set(ctx, cur+uint64(newCount)); err != nil {
		return err
	}
	// Advance the per-block accumulator (checked above).
	if err := k.MintHeight.Set(ctx, h); err != nil {
		return err
	}
	if err := k.BlockMinted.Set(ctx, newBlockTotal); err != nil {
		return err
	}
	// Early-warning event well below the hard ceiling: if honest growth ever
	// nears MaxMintPerBlock, this fires long before a legitimate commit would
	// be rejected (raise the ceiling / promote it to a param).
	if newBlockTotal >= photons.MintCeilingSoftWarn {
		sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
			"photon_mint_ceiling_warn",
			sdk.NewAttribute("height", fmt.Sprintf("%d", h)),
			sdk.NewAttribute("block_minted", fmt.Sprintf("%d", newBlockTotal)),
			sdk.NewAttribute("ceiling", fmt.Sprintf("%d", photons.MaxMintPerBlock)),
		))
	}
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		"photon_minted",
		sdk.NewAttribute("kind", kind),
		sdk.NewAttribute("count", fmt.Sprintf("%d", newCount)),
		sdk.NewAttribute("recipient", p.StorerRewardRecipient),
		sdk.NewAttribute("minted_total", fmt.Sprintf("%d", cur+uint64(newCount))),
	))
	return nil
}
