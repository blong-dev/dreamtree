package seeds

import "cosmossdk.io/collections"

const ModuleName = "seeds"

var (
	// ParamsKey is the prefix for the module params.
	ParamsKey = collections.NewPrefix(0)
	// SeedsKey is the prefix for the id -> Seed map.
	SeedsKey = collections.NewPrefix(1)
	// SeqKey is the prefix for the monotonic id sequence.
	SeqKey = collections.NewPrefix(2)
	// SubjectIndexKey is the prefix for the (subject, id) secondary index.
	SubjectIndexKey = collections.NewPrefix(3)
)
