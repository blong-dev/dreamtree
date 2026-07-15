package keeper

import (
	"context"

	"cosmossdk.io/collections"

	"github.com/blong-dev/dreamtree/x/seeds"
)

// BatchOf resolves a leaf-seed id to its containing batch: the batch with the
// greatest first_seed_id <= id whose range [first, first+new_count) covers id.
// Returns (batch, true) on hit; (zero, false) when the id is unallocated or
// falls in a gap.
func (k Keeper) BatchOf(ctx context.Context, id uint64) (seeds.Batch, bool, error) {
	var batchID uint64
	hit := false
	rng := new(collections.Range[uint64]).EndInclusive(id).Descending()
	if err := k.RangeIndex.Walk(ctx, rng, func(_ uint64, bid uint64) (bool, error) {
		batchID = bid
		hit = true
		return true, nil // first (greatest <= id) is the candidate
	}); err != nil {
		return seeds.Batch{}, false, err
	}
	if !hit {
		return seeds.Batch{}, false, nil
	}
	b, err := k.Batches.Get(ctx, batchID)
	if err != nil {
		return seeds.Batch{}, false, err
	}
	if id < b.FirstSeedId || id >= b.FirstSeedId+uint64(b.NewCount) {
		return seeds.Batch{}, false, nil
	}
	return b, true, nil
}

// SynthesizeSeed materializes the leaf-seed view of id from its batch
// (seeds are stored as batches; docs/specs/seed-atom-conformance.md).
func SynthesizeSeed(b seeds.Batch, id uint64) seeds.Seed {
	return seeds.Seed{
		Id:          id,
		Committer:   b.Committer,
		Subject:     b.Subject,
		Commitment:  b.MerkleRoot,
		Kind:        b.Kind,
		SourceRef:   b.SourceRef,
		DataType:    b.DataType,
		Height:      b.Height,
		CommittedAt: b.CommittedAt,
		BatchId:     b.Id,
		LeafIndex:   uint32(id - b.FirstSeedId),
	}
}
