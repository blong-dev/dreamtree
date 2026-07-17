package keeper

import (
	"context"
	"sort"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/blong-dev/dreamtree/x/licenses"
)

// Purchase buys time-bound access to a swath at the protocol CONSTANT: one
// photon per seed per day (upgrade-1 R4, owner 2026-07-16 — supersedes the
// per-type N_a market). The constant is a unit definition, not a price
// control: scarcity allocates (photons = seeds), differentiation across types
// expresses as volume, and the photon's real-world value finds its own
// equilibrium at the edges. Creator equality is absolute — across creators AND
// types. Each seed's payment routes to its producer; a marketplace toll (atop)
// goes to the treasury. Non-exclusive: the same seed sells to any number of
// buyers.
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
	// 1 photon per seed per day of access, in base units.
	pricePerSeed := uint64(params.AccessDurationDays) * licenses.UphotonPerPhoton

	producerTotals := map[string]uint64{}
	var totalSale uint64
	var bought uint32
	type grant struct {
		id    uint64
		price uint64
	}
	var grants []grant

	seen := make(map[uint64]struct{}, len(seedIDs))
	for _, id := range seedIDs {
		if _, dup := seen[id]; dup {
			continue // dedup: a seed in the swath twice is bought (and paid) once
		}
		seen[id] = struct{}{}
		_, producer, found := k.seeds.SeedInfo(ctx, id)
		if !found || producer == "" {
			continue // unresolvable seed → not purchasable (nobody to pay)
		}
		producerTotals[producer] += pricePerSeed
		totalSale += pricePerSeed
		bought++
		grants = append(grants, grant{id: id, price: pricePerSeed})
	}
	if bought == 0 {
		return nil, licenses.ErrNoPricedSeeds
	}

	// Taxes apply only when a treasury is configured. marketplace_toll is
	// buyer-side (added atop N_a); value_creation_tax is producer-side (deducted
	// from each producer's earnings). uint64→Int (no int64 narrowing).
	hasTreasury := params.TreasuryRecipient != ""
	toll := uint64(0)
	if hasTreasury {
		toll = params.MarketplaceToll.MulInt(math.NewIntFromUint64(totalSale)).TruncateInt().Uint64()
	}
	total := totalSale + toll // buyer outlay (the value-creation tax comes out of producers)

	// Fail early (deterministically) if the buyer can't cover the swath.
	bal := k.bank.GetBalance(ctx, buyerAddr, licenses.PhotonDenom)
	if bal.Amount.LT(math.NewIntFromUint64(total)) {
		return nil, licenses.ErrInsufficient.Wrapf("need %d, have %s", total, bal.Amount)
	}

	// Route to producers (sorted for determinism), net of the value-creation
	// tax; accumulate that tax + the toll to the treasury.
	producers := make([]string, 0, len(producerTotals))
	for p := range producerTotals {
		producers = append(producers, p)
	}
	sort.Strings(producers)
	valueTaxTotal := uint64(0)
	for _, p := range producers {
		pAddr, err := k.addressCodec.StringToBytes(p)
		if err != nil {
			return nil, err
		}
		gross := producerTotals[p]
		vct := uint64(0)
		if hasTreasury {
			vct = params.ValueCreationTax.MulInt(math.NewIntFromUint64(gross)).TruncateInt().Uint64()
		}
		valueTaxTotal += vct
		if net := gross - vct; net > 0 {
			if err := k.bank.SendCoins(ctx, buyerAddr, pAddr, coins(net)); err != nil {
				return nil, err
			}
		}
	}
	treasuryTotal := toll + valueTaxTotal
	if treasuryTotal > 0 && hasTreasury {
		tAddr, err := k.addressCodec.StringToBytes(params.TreasuryRecipient)
		if err != nil {
			return nil, err
		}
		if err := k.bank.SendCoins(ctx, buyerAddr, tAddr, coins(treasuryTotal)); err != nil {
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

	producersPaid := totalSale - valueTaxTotal
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		"seed_access_purchased",
		sdk.NewAttribute("buyer", buyer),
		sdk.NewAttribute("seeds", uintToStr(uint64(bought))),
		sdk.NewAttribute("producers_paid", uintToStr(producersPaid)),
		sdk.NewAttribute("treasury", uintToStr(treasuryTotal)),
	))
	return &licenses.MsgPurchaseResponse{
		ProducersPaid:   producersPaid,
		Toll:            treasuryTotal,
		TotalPaid:       total,
		SeedsPurchased:  bought,
		AccessExpiresAt: expires,
	}, nil
}

func coins(amt uint64) sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin(licenses.PhotonDenom, math.NewIntFromUint64(amt)))
}

func uintToStr(n uint64) string { return math.NewIntFromUint64(n).String() }
