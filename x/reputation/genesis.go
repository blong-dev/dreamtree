package reputation

func NewGenesisState() *GenesisState {
	return &GenesisState{Params: DefaultParams(), NextId: 1, NextPendingId: 1}
}

func (gs *GenesisState) Validate() error {
	seen := make(map[uint64]bool, len(gs.Contributions))
	for _, c := range gs.Contributions {
		if seen[c.Id] {
			return ErrEmptySigner.Wrapf("duplicate contribution id %d", c.Id)
		}
		seen[c.Id] = true
		if c.Signer == "" {
			return ErrEmptySigner
		}
		if c.Domain == "" {
			return ErrEmptyDomain
		}
	}
	return gs.Params.Validate()
}
