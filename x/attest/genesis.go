package attest

func NewGenesisState() *GenesisState {
	return &GenesisState{Params: DefaultParams(), NextId: 1}
}

func (gs *GenesisState) Validate() error {
	seen := make(map[uint64]bool, len(gs.Attestations))
	for _, a := range gs.Attestations {
		if seen[a.Id] {
			return ErrEmptySubject.Wrapf("duplicate attestation id %d", a.Id)
		}
		seen[a.Id] = true
		if a.Subject == "" {
			return ErrEmptySubject
		}
	}
	return gs.Params.Validate()
}
