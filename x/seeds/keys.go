package seeds

import "cosmossdk.io/collections"

const ModuleName = "seeds"

var (
	// ParamsKey is the prefix for the module params.
	ParamsKey = collections.NewPrefix(0)
	// Prefix 1 is RETIRED (the pre-leaf-model per-seed row map). Never reuse:
	// seeds are stored as Batches; Seed objects are synthesized on read.

	// SeqKey is the prefix for the monotonic leaf-seed id sequence.
	SeqKey = collections.NewPrefix(2)
	// SubjectIndexKey is the prefix for the (subject, batch_id) index — one
	// entry per BATCH, not per leaf (a batch may register thousands), keyed by
	// batch id so pure-convergence batches (no seed range) index too.
	SubjectIndexKey = collections.NewPrefix(3)
	// BatchesKey is the prefix for the batch_id -> Batch map (the stored unit).
	BatchesKey = collections.NewPrefix(4)
	// BatchSeqKey is the prefix for the monotonic batch id sequence.
	BatchSeqKey = collections.NewPrefix(5)
	// RangeIndexKey is the prefix for the first_seed_id -> batch_id ordered
	// index (leaf-id resolution via descending walk).
	RangeIndexKey = collections.NewPrefix(6)
)
