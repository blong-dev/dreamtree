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

// CommitSeed anchors a single commitment on-chain — stored as a batch of one
// (the leaf model; docs/specs/seed-atom-conformance.md).
func (ms msgServer) CommitSeed(ctx context.Context, msg *seeds.MsgCommitSeed) (*seeds.MsgCommitSeedResponse, error) {
	b, err := ms.commit(ctx, commitReq{
		committer:  msg.Committer,
		subject:    msg.Subject,
		merkleRoot: msg.Commitment,
		leafCount:  1,
		newCount:   1,
		kind:       msg.Kind,
		sourceRef:  msg.SourceRef,
		dataType:   msg.DataType,
	})
	if err != nil {
		return nil, err
	}
	return &seeds.MsgCommitSeedResponse{Id: b.FirstSeedId, Height: b.Height}, nil
}

// CommitBatch anchors new_count leaf-seeds under one Merkle root in one tx.
// Convergence rule: re-observed atoms count in leaf_count (provable against
// this root) but not in new_count (no re-mint — sigma accrues, supply doesn't).
func (ms msgServer) CommitBatch(ctx context.Context, msg *seeds.MsgCommitBatch) (*seeds.MsgCommitBatchResponse, error) {
	b, err := ms.commit(ctx, commitReq{
		committer:  msg.Committer,
		subject:    msg.Subject,
		merkleRoot: msg.MerkleRoot,
		leafCount:  msg.LeafCount,
		newCount:   msg.NewCount,
		kind:       msg.Kind,
		sourceRef:  msg.SourceRef,
		dataType:   msg.DataType,
	})
	if err != nil {
		return nil, err
	}
	return &seeds.MsgCommitBatchResponse{FirstId: b.FirstSeedId, BatchId: b.Id, Height: b.Height}, nil
}

type commitReq struct {
	committer, subject, merkleRoot, kind, sourceRef, dataType string
	leafCount, newCount                                       uint32
}

// commit is the shared anchoring path: validates, allocates the leaf-seed id
// range, stores the batch + indexes, emits events, and notifies the photon
// seam with new_count (photons = seeds = distinct atoms).
func (ms msgServer) commit(ctx context.Context, req commitReq) (seeds.Batch, error) {
	if _, err := ms.k.addressCodec.StringToBytes(req.committer); err != nil {
		return seeds.Batch{}, fmt.Errorf("invalid committer address: %w", err)
	}
	if req.merkleRoot == "" {
		return seeds.Batch{}, seeds.ErrEmptyCommitment
	}
	if !isHex(req.merkleRoot) {
		return seeds.Batch{}, seeds.ErrCommitmentNotHex
	}
	if req.kind == "" {
		return seeds.Batch{}, seeds.ErrEmptyKind
	}
	// The aggregate is no longer a seed kind — kind names the LEAF.
	if strings.Contains(strings.ToLower(req.kind), "batch_root") {
		return seeds.Batch{}, seeds.ErrRetiredKind.Wrapf("got kind %q", req.kind)
	}
	// new_count == 0 is a pure-convergence batch (all leaves re-observed):
	// valid provenance — anchors the root, allocates no ids, mints nothing.
	if req.leafCount == 0 || req.newCount > req.leafCount {
		return seeds.Batch{}, seeds.ErrBadCounts.Wrapf("new_count=%d leaf_count=%d", req.newCount, req.leafCount)
	}

	params, err := ms.k.Params.Get(ctx)
	if err != nil {
		return seeds.Batch{}, err
	}
	if uint32(len(req.merkleRoot)) > params.MaxCommitment() {
		return seeds.Batch{}, seeds.ErrCommitmentTooLong.Wrapf("got %d, max %d", len(req.merkleRoot), params.MaxCommitment())
	}
	if uint32(len(req.sourceRef)) > params.MaxSourceRef() {
		return seeds.Batch{}, seeds.ErrSourceRefTooLong.Wrapf("got %d, max %d", len(req.sourceRef), params.MaxSourceRef())
	}
	// new_count is committer-asserted and each new leaf mints a photon: cap
	// the per-tx blast radius (supply-griefing guard; governance-tunable).
	if req.newCount > params.MaxBatchNew() {
		return seeds.Batch{}, seeds.ErrBadCounts.Wrapf("new_count %d exceeds max_batch_new_count %d", req.newCount, params.MaxBatchNew())
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Allocate the leaf-seed id range [first, first+new_count). A pure-
	// convergence batch (new_count == 0) allocates nothing; first stays 0.
	var first uint64
	if req.newCount > 0 {
		cur, err := ms.k.Seq.Peek(ctx)
		if err != nil {
			return seeds.Batch{}, err
		}
		first = cur
		if err := ms.k.Seq.Set(ctx, first+uint64(req.newCount)); err != nil {
			return seeds.Batch{}, err
		}
	}
	batchID, err := ms.k.BatchSeq.Next(ctx)
	if err != nil {
		return seeds.Batch{}, err
	}

	batch := seeds.Batch{
		Id:          batchID,
		FirstSeedId: first,
		NewCount:    req.newCount,
		LeafCount:   req.leafCount,
		MerkleRoot:  req.merkleRoot,
		Committer:   req.committer,
		Subject:     req.subject,
		Kind:        req.kind,
		SourceRef:   req.sourceRef,
		DataType:    req.dataType,
		Height:      sdkCtx.BlockHeight(),
		CommittedAt: sdkCtx.BlockTime().Unix(),
	}
	if err := ms.k.Batches.Set(ctx, batchID, batch); err != nil {
		return seeds.Batch{}, err
	}
	if req.newCount > 0 {
		if err := ms.k.RangeIndex.Set(ctx, first, batchID); err != nil {
			return seeds.Batch{}, err
		}
	}
	if req.subject != "" {
		if err := ms.k.SubjectIndex.Set(ctx, collections.Join(req.subject, batchID)); err != nil {
			return seeds.Batch{}, err
		}
	}

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"seed_batch_committed",
		sdk.NewAttribute("batch_id", fmt.Sprintf("%d", batchID)),
		sdk.NewAttribute("first_id", fmt.Sprintf("%d", first)),
		sdk.NewAttribute("new_count", fmt.Sprintf("%d", req.newCount)),
		sdk.NewAttribute("leaf_count", fmt.Sprintf("%d", req.leafCount)),
		sdk.NewAttribute("committer", req.committer),
		sdk.NewAttribute("subject", req.subject),
		sdk.NewAttribute("kind", req.kind),
		sdk.NewAttribute("merkle_root", req.merkleRoot),
	))
	// Compatibility event for single commits (roots' anchord path).
	if req.newCount == 1 && req.leafCount == 1 {
		sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
			"seed_committed",
			sdk.NewAttribute("id", fmt.Sprintf("%d", first)),
			sdk.NewAttribute("committer", req.committer),
			sdk.NewAttribute("subject", req.subject),
			sdk.NewAttribute("kind", req.kind),
			sdk.NewAttribute("commitment", req.merkleRoot),
		))
	}

	// Ingestion mint: new_count photons (photons = seeds = distinct atoms).
	// Pure-convergence batches mint nothing by definition; otherwise no-op if
	// x/photons is absent or the kind is not a data contribution.
	if req.newCount > 0 && ms.k.Photons() != nil {
		if err := ms.k.Photons().OnRecordBatch(ctx, req.kind, req.newCount); err != nil {
			return seeds.Batch{}, err
		}
	}
	return batch, nil
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

// isHex reports whether s is non-empty and all hex digits (a commitment is a
// digest/Merkle root, never a body).
func isHex(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return len(s) > 0
}
