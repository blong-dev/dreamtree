package keeper

import (
	"context"

	"cosmossdk.io/collections"

	"github.com/blong-dev/dreamtree/x/attest"
	"github.com/blong-dev/dreamtree/x/attest/projection"
)

// The projection math lives in x/attest/projection (backtest M1) so the
// identical functions run off-chain over an exported log. This file is the
// store-backed adapter: it closes the SDK context into the projection's
// LogReader / ReputationResolver interfaces.

// storeLog backs projection.LogReader with the keeper's collections.
type storeLog struct {
	ctx context.Context
	k   Keeper
}

func (s storeLog) GetAttestation(id uint64) (attest.Attestation, error) {
	return s.k.Attestations.Get(s.ctx, id)
}

func (s storeLog) WalkSubject(subject string, fn func(attest.Attestation) (bool, error)) error {
	rng := collections.NewPrefixedPairRange[string, uint64](subject)
	return s.k.SubjectIndex.Walk(s.ctx, rng, func(key collections.Pair[string, uint64]) (bool, error) {
		a, err := s.k.Attestations.Get(s.ctx, key.K2())
		if err != nil {
			return false, err
		}
		return fn(a)
	})
}

func (s storeLog) WalkTarget(targetID uint64, fn func(attest.Attestation) (bool, error)) error {
	rng := collections.NewPrefixedPairRange[uint64, uint64](targetID)
	return s.k.TargetIndex.Walk(s.ctx, rng, func(key collections.Pair[uint64, uint64]) (bool, error) {
		a, err := s.k.Attestations.Get(s.ctx, key.K2())
		if err != nil {
			return false, err
		}
		return fn(a)
	})
}

// seamRep backs projection.ReputationResolver with the x/reputation seam,
// falling back to baseline_kyc when x/reputation is not wired.
type seamRep struct {
	ctx context.Context
	k   Keeper
	pf  projection.ParamsF
}

func (r seamRep) ReputationOf(signer, domain string) float64 {
	if r.k.rep != nil {
		return r.k.rep.ReputationOf(r.ctx, signer, domain)
	}
	return r.pf.BaselineKyc
}

// projector builds a store-backed Projector for the given context and params.
func (k Keeper) projector(ctx context.Context, pf projection.ParamsF) projection.Projector {
	return projection.Projector{
		Params: pf,
		Log:    storeLog{ctx: ctx, k: k},
		Rep:    seamRep{ctx: ctx, k: k, pf: pf},
	}
}

// reputationOf resolves R(signer,domain) through the reputation seam (kept for
// the query server's decomposed Strength response).
func (k Keeper) reputationOf(ctx context.Context, pf projection.ParamsF, signer, domain string) float64 {
	return seamRep{ctx: ctx, k: k, pf: pf}.ReputationOf(signer, domain)
}
