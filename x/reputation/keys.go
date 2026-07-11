package reputation

import "cosmossdk.io/collections"

const ModuleName = "reputation"

var (
	ParamsKey        = collections.NewPrefix(0)
	ContributionsKey = collections.NewPrefix(1)
	SeqKey           = collections.NewPrefix(2)
	SignerIndexKey   = collections.NewPrefix(3) // (signer, id)
	DomainConfigKey  = collections.NewPrefix(4) // path -> DomainConfig
	SourceIndexKey   = collections.NewPrefix(5) // (source_att_id, contribution_id) — for reversal

	PendingKey         = collections.NewPrefix(6) // id -> PendingEvent
	PendingSeqKey      = collections.NewPrefix(7)
	CloseTimeIndexKey  = collections.NewPrefix(8)  // (close_time, id) — EndBlock drains matured
	PendingByTargetKey = collections.NewPrefix(9)  // target_att_id -> open outcome pending id
	ReversedKey        = collections.NewPrefix(10) // set of overturned outcome att ids (idempotent reversal)
)
