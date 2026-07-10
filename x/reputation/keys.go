package reputation

import "cosmossdk.io/collections"

const ModuleName = "reputation"

var (
	ParamsKey        = collections.NewPrefix(0)
	ContributionsKey = collections.NewPrefix(1)
	SeqKey           = collections.NewPrefix(2)
	SignerIndexKey   = collections.NewPrefix(3) // (signer, id)
	DomainConfigKey  = collections.NewPrefix(4) // path -> DomainConfig
)
