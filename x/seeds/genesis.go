package seeds

import "sort"

// NewGenesisState creates a new genesis state with default values.
func NewGenesisState() *GenesisState {
	return &GenesisState{
		Params:      DefaultParams(),
		NextId:      1,
		NextBatchId: 1,
	}
}

// Validate performs basic genesis state validation: unique batch ids, valid
// counts, non-empty roots, and non-overlapping leaf-seed id ranges that stay
// below next_id.
func (gs *GenesisState) Validate() error {
	seen := make(map[uint64]bool, len(gs.Batches))
	ranges := make([]Batch, 0, len(gs.Batches))
	for _, b := range gs.Batches {
		if seen[b.Id] {
			return ErrEmptyCommitment.Wrapf("duplicate batch id %d", b.Id)
		}
		seen[b.Id] = true
		if b.MerkleRoot == "" {
			return ErrEmptyCommitment.Wrapf("batch %d", b.Id)
		}
		// new_count == 0 is a pure-convergence batch (provenance only).
		if b.LeafCount == 0 || b.NewCount > b.LeafCount {
			return ErrBadCounts.Wrapf("batch %d: new_count=%d leaf_count=%d", b.Id, b.NewCount, b.LeafCount)
		}
		if b.NewCount == 0 {
			continue // no seed range to check
		}
		if b.FirstSeedId == 0 {
			return ErrBadCounts.Wrapf("batch %d: first_seed_id must be >= 1", b.Id)
		}
		ranges = append(ranges, b)
	}
	sort.Slice(ranges, func(i, j int) bool { return ranges[i].FirstSeedId < ranges[j].FirstSeedId })
	for _, b := range ranges {
		// Wraparound guard: first + new_count must not overflow uint64, or
		// both this check and read-time resolution would wrap.
		if b.FirstSeedId > ^uint64(0)-uint64(b.NewCount) {
			return ErrBadCounts.Wrapf("batch %d: seed range overflows uint64", b.Id)
		}
	}
	for i := 1; i < len(ranges); i++ {
		prev := ranges[i-1]
		if prev.FirstSeedId+uint64(prev.NewCount) > ranges[i].FirstSeedId {
			return ErrBadCounts.Wrapf("batches %d and %d have overlapping seed ranges", prev.Id, ranges[i].Id)
		}
	}
	// Sequence guards — mirror InitGenesis's 0→1 defaulting, then require the
	// sequences to clear every carried batch: a reissued batch id silently
	// OVERWRITES anchored state; a reissued leaf id corrupts RangeIndex.
	nextID := gs.NextId
	if nextID == 0 {
		nextID = 1
	}
	if n := len(ranges); n > 0 {
		last := ranges[n-1]
		if last.FirstSeedId+uint64(last.NewCount) > nextID {
			return ErrBadCounts.Wrapf("next_id %d does not clear batch %d's range", gs.NextId, last.Id)
		}
	}
	nextBatch := gs.NextBatchId
	if nextBatch == 0 {
		nextBatch = 1
	}
	for _, b := range gs.Batches {
		if b.Id >= nextBatch {
			return ErrBadCounts.Wrapf("next_batch_id %d does not clear batch id %d (reissue would overwrite it)", gs.NextBatchId, b.Id)
		}
	}
	return gs.Params.Validate()
}
