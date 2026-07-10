package photons

// DefaultMintableKinds — the x/seeds kinds that are atomic data contributions
// (each mints one photon). Batch roots aggregate many and do NOT mint.
var DefaultMintableKinds = []string{"record", "kg_claim"}

func DefaultParams() Params {
	return Params{
		StorerRewardRecipient: "", // set to dreamtree's address at launch
		MintableKinds:         DefaultMintableKinds,
	}
}

func (p Params) Validate() error { return nil }

// Mintable reports whether a seed kind mints a photon.
func (p Params) Mintable(kind string) bool {
	kinds := p.MintableKinds
	if len(kinds) == 0 {
		kinds = DefaultMintableKinds
	}
	for _, k := range kinds {
		if k == kind {
			return true
		}
	}
	return false
}
