package seeds

import "context"

// PhotonHooks is the ingestion seam x/seeds notifies when a seed is committed,
// so x/photons can mint the per-seed photon (photons = seeds). Defined here (the
// consumer); implemented by x/photons; x/seeds never imports x/photons. nil when
// the module is absent (then no minting — x/seeds runs alone).
type PhotonHooks interface {
	OnRecordSeed(ctx context.Context, kind string) error
}
