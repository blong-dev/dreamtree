package seeds

// NewGenesisState creates a new genesis state with default values.
func NewGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
		NextId: 1,
	}
}

// Validate performs basic genesis state validation.
func (gs *GenesisState) Validate() error {
	seen := make(map[uint64]bool, len(gs.Seeds))
	for _, s := range gs.Seeds {
		if seen[s.Id] {
			return ErrEmptyCommitment.Wrapf("duplicate seed id %d", s.Id)
		}
		seen[s.Id] = true
		if s.Commitment == "" {
			return ErrEmptyCommitment
		}
	}
	return gs.Params.Validate()
}
