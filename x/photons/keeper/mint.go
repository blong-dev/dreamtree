package keeper

import (
	"context"
	"fmt"

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
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		"photon_minted",
		sdk.NewAttribute("kind", kind),
		sdk.NewAttribute("count", fmt.Sprintf("%d", newCount)),
		sdk.NewAttribute("recipient", p.StorerRewardRecipient),
		sdk.NewAttribute("minted_total", fmt.Sprintf("%d", cur+uint64(newCount))),
	))
	return nil
}
