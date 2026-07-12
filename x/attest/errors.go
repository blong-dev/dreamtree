package attest

import "cosmossdk.io/errors"

var (
	ErrEmptySubject   = errors.Register(ModuleName, 2, "subject must not be empty")
	ErrBadProofType   = errors.Register(ModuleName, 3, "proof_type must be specified")
	ErrBadSpecificity = errors.Register(ModuleName, 4, "specificity_bps must be <= 10000")
	ErrBadOutcome     = errors.Register(ModuleName, 5, "an OUTCOME attestation requires outcome_kind and an existing target_id")
	ErrOutcomeFields  = errors.Register(ModuleName, 6, "outcome_kind/target_id may only be set on an OUTCOME attestation")
	ErrTargetNotFound = errors.Register(ModuleName, 7, "target attestation not found")
	ErrBadEndorsed    = errors.Register(ModuleName, 8, "an ENDORSEMENT subject must be a valid address")
	ErrSelfEndorse    = errors.Register(ModuleName, 9, "cannot endorse yourself")
	ErrUsedByNonUse   = errors.Register(ModuleName, 10, "used_by may only be set on a USE attestation")
	ErrSelfUse        = errors.Register(ModuleName, 11, "used_by must differ from subject (a work cannot build on itself)")
)
