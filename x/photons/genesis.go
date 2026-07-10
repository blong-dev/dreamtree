package photons

func NewGenesisState() *GenesisState {
	return &GenesisState{Params: DefaultParams(), Minted: 0}
}

func (gs *GenesisState) Validate() error { return gs.Params.Validate() }
