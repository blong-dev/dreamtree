package keeper

import (
	"context"
	"fmt"
	"strings"

	"github.com/blong-dev/dreamtree/x/reputation"
)

type msgServer struct{ k Keeper }

var _ reputation.MsgServer = msgServer{}

func NewMsgServerImpl(keeper Keeper) reputation.MsgServer { return &msgServer{k: keeper} }

func (ms msgServer) UpdateParams(ctx context.Context, msg *reputation.MsgUpdateParams) (*reputation.MsgUpdateParamsResponse, error) {
	if err := ms.assertAuthority(msg.Authority); err != nil {
		return nil, err
	}
	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}
	if err := ms.k.Params.Set(ctx, msg.Params); err != nil {
		return nil, err
	}
	return &reputation.MsgUpdateParamsResponse{}, nil
}

func (ms msgServer) SetDomainConfig(ctx context.Context, msg *reputation.MsgSetDomainConfig) (*reputation.MsgSetDomainConfigResponse, error) {
	if err := ms.assertAuthority(msg.Authority); err != nil {
		return nil, err
	}
	if strings.TrimSpace(msg.Config.Path) == "" {
		return nil, reputation.ErrEmptyDomain
	}
	if err := ms.k.DomainConfigs.Set(ctx, msg.Config.Path, msg.Config); err != nil {
		return nil, err
	}
	return &reputation.MsgSetDomainConfigResponse{}, nil
}

// SetVerified — governance grants or revokes membership in the verified set
// (upgrade-1 R2). Each grant is an explicit governed act: this IS the v0
// ratification mechanism that hands out the baseline_kyc floor.
func (ms msgServer) SetVerified(ctx context.Context, msg *reputation.MsgSetVerified) (*reputation.MsgSetVerifiedResponse, error) {
	if err := ms.assertAuthority(msg.Authority); err != nil {
		return nil, err
	}
	if _, err := ms.k.addressCodec.StringToBytes(msg.Address); err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}
	if msg.Verified {
		if err := ms.k.Verified.Set(ctx, msg.Address); err != nil {
			return nil, err
		}
	} else {
		if err := ms.k.Verified.Remove(ctx, msg.Address); err != nil {
			return nil, err
		}
	}
	return &reputation.MsgSetVerifiedResponse{}, nil
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
