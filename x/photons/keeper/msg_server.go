package keeper

import (
	"context"
	"fmt"
	"strings"

	"github.com/blong-dev/dreamtree/x/photons"
)

type msgServer struct{ k Keeper }

var _ photons.MsgServer = msgServer{}

func NewMsgServerImpl(keeper Keeper) photons.MsgServer { return &msgServer{k: keeper} }

func (ms msgServer) UpdateParams(ctx context.Context, msg *photons.MsgUpdateParams) (*photons.MsgUpdateParamsResponse, error) {
	if _, err := ms.k.addressCodec.StringToBytes(msg.Authority); err != nil {
		return nil, fmt.Errorf("invalid authority address: %w", err)
	}
	if got := ms.k.GetAuthority(); !strings.EqualFold(msg.Authority, got) {
		return nil, fmt.Errorf("unauthorized: got %s, want %s", msg.Authority, got)
	}
	if r := msg.Params.StorerRewardRecipient; r != "" {
		if _, err := ms.k.addressCodec.StringToBytes(r); err != nil {
			return nil, photons.ErrBadRecipient
		}
	}
	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}
	if err := ms.k.Params.Set(ctx, msg.Params); err != nil {
		return nil, err
	}
	return &photons.MsgUpdateParamsResponse{}, nil
}
