package keeper

import (
	"context"
	"sort"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/blong-dev/dreamtree/x/licenses"
)

// Purchase buys time-bound access to a swath. For each priced seed the buyer
// pays that type's N_a to the seed's producer; a marketplace toll (atop N_a)
// goes to the treasury. All seeds of a type pay the same N_a
// (creator_equality_within_type). Non-exclusive: the same seed sells to any
// number of buyers. The protocol records + routes; it never sets N_a.
func (k Keeper) Purchase(ctx context.Context, buyer string, seedIDs []uint64) (*licenses.MsgPurchaseResponse, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	buyerAddr, err := k.addressCodec.StringToBytes(buyer)
	if err != nil {
		return nil, err
	}
	now := sdk.UnwrapSDKContext(ctx).BlockTime().Unix()
	expires := now + int64(params.AccessDurationDays)*86400

	producerTotals := map[string]uint64{}
	var totalSale uint64
	var bought uint32
	type grant struct {
		id    uint64
		price uint64
	}
	var grants []grant

	for _, id := range seedIDs {
		dataType, producer, found := k.seeds.SeedInfo(ctx, id)
		if !found || dataType == "" || producer == "" {
			continue
		}
		price, err := k.TypePrices.Get(ctx, dataType)
		if err != nil || price == 0 {
			continue // unpriced type → not purchasable
		}
		producerTotals[producer] += price
		totalSale += price
		bought++
		grants = append(grants, grant{id: id, price: price})
	}
	if bought == 0 {
		return nil, licenses.ErrNoPricedSeeds
	}

	toll := params.MarketplaceToll.MulInt64(int64(totalSale)).TruncateInt().Uint64()
	total := totalSale + toll

	// Fail early (deterministically) if the buyer can't cover the swath.
	bal := k.bank.GetBalance(ctx, buyerAddr, licenses.PhotonDenom)
	if bal.Amount.LT(math.NewIntFromUint64(total)) {
		return nil, licenses.ErrInsufficient.Wrapf("need %d, have %s", total, bal.Amount)
	}

	// Route to producers (sorted for determinism), then the toll to treasury.
	producers := make([]string, 0, len(producerTotals))
	for p := range producerTotals {
		producers = append(producers, p)
	}
	sort.Strings(producers)
	for _, p := range producers {
		pAddr, err := k.addressCodec.StringToBytes(p)
		if err != nil {
			return nil, err
		}
		if err := k.bank.SendCoins(ctx, buyerAddr, pAddr, coins(producerTotals[p])); err != nil {
			return nil, err
		}
	}
	if toll > 0 && params.TreasuryRecipient != "" {
		tAddr, err := k.addressCodec.StringToBytes(params.TreasuryRecipient)
		if err != nil {
			return nil, err
		}
		if err := k.bank.SendCoins(ctx, buyerAddr, tAddr, coins(toll)); err != nil {
			return nil, err
		}
	}

	// Grant time-bound access.
	for _, g := range grants {
		if err := k.AccessGrants.Set(ctx, collections.Join(buyer, g.id), licenses.AccessGrant{
			Buyer: buyer, SeedId: g.id, ExpiresAt: expires, PricePaid: g.price,
		}); err != nil {
			return nil, err
		}
	}

	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		"seed_access_purchased",
		sdk.NewAttribute("buyer", buyer),
		sdk.NewAttribute("seeds", uintToStr(uint64(bought))),
		sdk.NewAttribute("producers_paid", uintToStr(totalSale)),
		sdk.NewAttribute("toll", uintToStr(toll)),
	))
	return &licenses.MsgPurchaseResponse{
		ProducersPaid:   totalSale,
		Toll:            toll,
		TotalPaid:       total,
		SeedsPurchased:  bought,
		AccessExpiresAt: expires,
	}, nil
}

func coins(amt uint64) sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin(licenses.PhotonDenom, math.NewIntFromUint64(amt)))
}

func uintToStr(n uint64) string { return math.NewIntFromUint64(n).String() }
