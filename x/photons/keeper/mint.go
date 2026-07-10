package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/blong-dev/dreamtree/x/photons"
)

// OnRecordSeed is the x/seeds ingestion hook. For a data-contribution kind it
// mints EXACTLY ONE photon (photons = seeds) and routes it to the storer-reward
// recipient. This is the ONLY minting event in the protocol.
func (k Keeper) OnRecordSeed(ctx context.Context, kind string) error {
	p, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}
	if !p.Mintable(kind) {
		return nil
	}
	one := sdk.NewCoins(sdk.NewInt64Coin(photons.PhotonDenom, 1))
	if err := k.bank.MintCoins(ctx, photons.ModuleName, one); err != nil {
		return err
	}
	if p.StorerRewardRecipient != "" {
		addr, err := k.addressCodec.StringToBytes(p.StorerRewardRecipient)
		if err != nil {
			return err
		}
		if err := k.bank.SendCoinsFromModuleToAccount(ctx, photons.ModuleName, addr, one); err != nil {
			return err
		}
	}
	n, err := k.Minted.Next(ctx) // returns the pre-increment count
	if err != nil {
		return err
	}
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		"photon_minted",
		sdk.NewAttribute("kind", kind),
		sdk.NewAttribute("recipient", p.StorerRewardRecipient),
		sdk.NewAttribute("minted_total", fmt.Sprintf("%d", n+1)),
	))
	return nil
}
