package licenses

func NewGenesisState() *GenesisState {
	return &GenesisState{Params: DefaultParams()}
}

func (gs *GenesisState) Validate() error { return gs.Params.Validate() }
