package keeper

import (
	"context"
	"fmt"
	"strings"

	"github.com/blong-dev/dreamtree/x/licenses"
)

type msgServer struct{ k Keeper }

var _ licenses.MsgServer = msgServer{}

func NewMsgServerImpl(keeper Keeper) licenses.MsgServer { return &msgServer{k: keeper} }

func (ms msgServer) Purchase(ctx context.Context, msg *licenses.MsgPurchase) (*licenses.MsgPurchaseResponse, error) {
	if _, err := ms.k.addressCodec.StringToBytes(msg.Buyer); err != nil {
		return nil, fmt.Errorf("invalid buyer address: %w", err)
	}
	if len(msg.SeedIds) == 0 {
		return nil, licenses.ErrNoPricedSeeds
	}
	return ms.k.Purchase(ctx, msg.Buyer, msg.SeedIds)
}

// SetTypePrice is RETIRED (upgrade-1 R4, owner 2026-07-16): access costs a
// constant 1 photon per seed per day — there is no price to set. The msg type
// stays registered so pre-upgrade history still decodes; the handler refuses.
func (ms msgServer) SetTypePrice(ctx context.Context, msg *licenses.MsgSetTypePrice) (*licenses.MsgSetTypePriceResponse, error) {
	return nil, licenses.ErrRetired
}

func (ms msgServer) UpdateParams(ctx context.Context, msg *licenses.MsgUpdateParams) (*licenses.MsgUpdateParamsResponse, error) {
	if err := ms.assertAuthority(msg.Authority); err != nil {
		return nil, err
	}
	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}
	if err := ms.k.Params.Set(ctx, msg.Params); err != nil {
		return nil, err
	}
	return &licenses.MsgUpdateParamsResponse{}, nil
}

func (ms msgServer) assertAuthority(authority string) error {
	if _, err := ms.k.addressCodec.StringToBytes(authority); err != nil {
		return fmt.Errorf("invalid authority address: %w", err)
	}
	if got := ms.k.GetAuthority(); !strings.EqualFold(authority, got) {
		return fmt.Errorf("unauthorized: got %s, want %s", authority, got)
	}
	return nil
}
