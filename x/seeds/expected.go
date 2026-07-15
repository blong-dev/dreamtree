package seeds

import "context"

// PhotonHooks is the ingestion seam x/seeds notifies when a batch is committed,
// so x/photons can mint the per-seed photons (photons = seeds = distinct atoms).
// newCount is the batch's NEW distinct contributions — converged (re-observed)
// atoms do not re-mint (sigma accrues, supply doesn't). Defined here (the
// consumer); implemented by x/photons; x/seeds never imports x/photons. nil when
// the module is absent (then no minting — x/seeds runs alone).
type PhotonHooks interface {
	OnRecordBatch(ctx context.Context, kind string, newCount uint32) error
}
