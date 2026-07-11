package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/blong-dev/dreamtree/x/attest"
	reputation "github.com/blong-dev/dreamtree/x/reputation"
)

// maxCoAttestors bounds the propagation fan-out per outcome (no silent runaway).
const maxCoAttestors = 64

// specDec is the fixed-point specificity factor (bps/10000; 0 = full).
func specDec(bps uint32) math.LegacyDec {
	if bps == 0 {
		bps = 10000
	}
	return math.LegacyNewDec(int64(bps)).QuoInt64(10000)
}

// propTargetsFor gathers the parties (besides the contributor) whose reputation
// an outcome on `target` should move: co-attestors on the same work, and the
// contributor's endorsers (liability). The reputation module applies each
// kind's weight at settlement.
func (k Keeper) propTargetsFor(ctx context.Context, target attest.Attestation) []reputation.PropTarget {
	var out []reputation.PropTarget

	// Co-attestors: other non-outcome, non-endorsement attestations on the work.
	coRange := collections.NewPrefixedPairRange[string, uint64](target.Subject)
	_ = k.SubjectIndex.Walk(ctx, coRange, func(key collections.Pair[string, uint64]) (bool, error) {
		id := key.K2()
		if id == target.Id {
			return false, nil
		}
		co, err := k.Attestations.Get(ctx, id)
		if err != nil {
			return false, nil
		}
		if co.ProofType == attest.ProofType_PROOF_TYPE_OUTCOME || co.ProofType == attest.ProofType_PROOF_TYPE_ENDORSEMENT {
			return false, nil
		}
		if co.Attestor == target.Attestor {
			return false, nil // the contributor is handled directly
		}
		out = append(out, reputation.PropTarget{
			Address:    co.Attestor,
			Domain:     co.Domain,
			Kind:       reputation.PropKind_PROP_KIND_COATTESTOR,
			BaseFactor: specDec(co.SpecificityBps),
		})
		return len(out) >= maxCoAttestors, nil
	})
	if len(out) >= maxCoAttestors {
		// No silent truncation — surface that a work's fan-out exceeded the cap
		// so the under-propagation is observable.
		sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
			"propagation_truncated",
			sdk.NewAttribute("target_id", fmt.Sprintf("%d", target.Id)),
			sdk.NewAttribute("cap", fmt.Sprintf("%d", maxCoAttestors)),
		))
	}

	// Endorsers of the contributor (ENDORSEMENT attestations with subject = the
	// contributor's address) — the liability mirror of P4 inheritance.
	enRange := collections.NewPrefixedPairRange[string, uint64](target.Attestor)
	_ = k.SubjectIndex.Walk(ctx, enRange, func(key collections.Pair[string, uint64]) (bool, error) {
		e, err := k.Attestations.Get(ctx, key.K2())
		if err != nil {
			return false, nil
		}
		if e.ProofType != attest.ProofType_PROOF_TYPE_ENDORSEMENT {
			return false, nil
		}
		out = append(out, reputation.PropTarget{
			Address:    e.Attestor,
			Domain:     target.Domain,
			Kind:       reputation.PropKind_PROP_KIND_ENDORSER,
			BaseFactor: math.LegacyOneDec(),
		})
		return len(out) >= 2*maxCoAttestors, nil
	})

	return out
}
