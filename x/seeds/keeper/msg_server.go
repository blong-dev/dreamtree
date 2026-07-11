package keeper

import (
	"context"
	"fmt"
	"strings"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/blong-dev/dreamtree/x/seeds"
)

type msgServer struct {
	k Keeper
}

var _ seeds.MsgServer = msgServer{}

// NewMsgServerImpl returns a seeds MsgServer implementation.
func NewMsgServerImpl(keeper Keeper) seeds.MsgServer {
	return &msgServer{k: keeper}
}

// CommitSeed anchors a commitment on-chain.
func (ms msgServer) CommitSeed(ctx context.Context, msg *seeds.MsgCommitSeed) (*seeds.MsgCommitSeedResponse, error) {
	if _, err := ms.k.addressCodec.StringToBytes(msg.Committer); err != nil {
		return nil, fmt.Errorf("invalid committer address: %w", err)
	}
	if msg.Commitment == "" {
		return nil, seeds.ErrEmptyCommitment
	}
	if msg.Kind == "" {
		return nil, seeds.ErrEmptyKind
	}

	params, err := ms.k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	if uint32(len(msg.Commitment)) > params.MaxCommitment() {
		return nil, seeds.ErrCommitmentTooLong.Wrapf("got %d, max %d", len(msg.Commitment), params.MaxCommitment())
	}
	if uint32(len(msg.SourceRef)) > params.MaxSourceRef() {
		return nil, seeds.ErrSourceRefTooLong.Wrapf("got %d, max %d", len(msg.SourceRef), params.MaxSourceRef())
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	id, err := ms.k.Seq.Next(ctx)
	if err != nil {
		return nil, err
	}

	seed := seeds.Seed{
		Id:          id,
		Committer:   msg.Committer,
		Subject:     msg.Subject,
		Commitment:  msg.Commitment,
		Kind:        msg.Kind,
		SourceRef:   msg.SourceRef,
		DataType:    msg.DataType,
		Height:      sdkCtx.BlockHeight(),
		CommittedAt: sdkCtx.BlockTime().Unix(),
	}
	if err := ms.k.Seeds.Set(ctx, id, seed); err != nil {
		return nil, err
	}
	if msg.Subject != "" {
		if err := ms.k.SubjectIndex.Set(ctx, collections.Join(msg.Subject, id)); err != nil {
			return nil, err
		}
	}

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"seed_committed",
		sdk.NewAttribute("id", fmt.Sprintf("%d", id)),
		sdk.NewAttribute("committer", msg.Committer),
		sdk.NewAttribute("subject", msg.Subject),
		sdk.NewAttribute("kind", msg.Kind),
		sdk.NewAttribute("commitment", msg.Commitment),
	))

	// Ingestion mint: one photon per data-seed (photons = seeds). No-op if
	// x/photons is absent or the kind is not a data contribution.
	if ms.k.Photons() != nil {
		if err := ms.k.Photons().OnRecordSeed(ctx, msg.Kind); err != nil {
			return nil, err
		}
	}
	return &seeds.MsgCommitSeedResponse{Id: id, Height: sdkCtx.BlockHeight()}, nil
}

// UpdateParams sets the module parameters (authority-gated).
func (ms msgServer) UpdateParams(ctx context.Context, msg *seeds.MsgUpdateParams) (*seeds.MsgUpdateParamsResponse, error) {
	if _, err := ms.k.addressCodec.StringToBytes(msg.Authority); err != nil {
		return nil, fmt.Errorf("invalid authority address: %w", err)
	}
	if authority := ms.k.GetAuthority(); !strings.EqualFold(msg.Authority, authority) {
		return nil, fmt.Errorf("unauthorized: got %s, want %s", msg.Authority, authority)
	}
	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}
	if err := ms.k.Params.Set(ctx, msg.Params); err != nil {
		return nil, err
	}
	return &seeds.MsgUpdateParamsResponse{}, nil
}
