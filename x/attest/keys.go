package attest

import "cosmossdk.io/collections"

const ModuleName = "attest"

var (
	ParamsKey        = collections.NewPrefix(0)
	AttestationsKey  = collections.NewPrefix(1)
	SeqKey           = collections.NewPrefix(2)
	SubjectIndexKey  = collections.NewPrefix(3) // (subject, id)
	AttestorIndexKey = collections.NewPrefix(4) // (attestor, id)
	TargetIndexKey   = collections.NewPrefix(5) // (target_id, id) — outcomes/refutations of an attestation
)
