package keeper

import (
	"context"
	"fmt"
	"strings"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/blong-dev/dreamtree/x/attest"
)

type msgServer struct{ k Keeper }

var _ attest.MsgServer = msgServer{}

func NewMsgServerImpl(keeper Keeper) attest.MsgServer { return &msgServer{k: keeper} }

func (ms msgServer) Attest(ctx context.Context, msg *attest.MsgAttest) (*attest.MsgAttestResponse, error) {
	if _, err := ms.k.addressCodec.StringToBytes(msg.Attestor); err != nil {
		return nil, fmt.Errorf("invalid attestor address: %w", err)
	}
	if strings.TrimSpace(msg.Subject) == "" {
		return nil, attest.ErrEmptySubject
	}
	if msg.ProofType == attest.ProofType_PROOF_TYPE_UNSPECIFIED {
		return nil, attest.ErrBadProofType
	}
	if msg.SpecificityBps > 10000 {
		return nil, attest.ErrBadSpecificity
	}

	isOutcome := msg.ProofType == attest.ProofType_PROOF_TYPE_OUTCOME
	var target attest.Attestation
	if isOutcome {
		if msg.OutcomeKind == attest.OutcomeKind_OUTCOME_KIND_UNSPECIFIED || msg.TargetId == 0 {
			return nil, attest.ErrBadOutcome
		}
		t, gerr := ms.k.Attestations.Get(ctx, msg.TargetId)
		if gerr != nil {
			return nil, attest.ErrTargetNotFound.Wrapf("id %d", msg.TargetId)
		}
		target = t
	} else if msg.OutcomeKind != attest.OutcomeKind_OUTCOME_KIND_UNSPECIFIED || msg.TargetId != 0 {
		return nil, attest.ErrOutcomeFields
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	id, err := ms.k.Seq.Next(ctx)
	if err != nil {
		return nil, err
	}

	// Snapshot the RATIONAL strength-at-issuance (standing × spec × type_weight),
	// frozen so an outcome's M_O reads S(att, t_issuance) deterministically.
	params, err := ms.k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	standing := math.LegacyOneDec()
	if ms.k.Rep() != nil {
		standing = ms.k.Rep().StandingOf(ctx, msg.Attestor, msg.Domain)
	}
	sIssuance := params.SIssuance(standing, msg.ProofType, msg.SpecificityBps)

	a := attest.Attestation{
		Id:             id,
		Attestor:       msg.Attestor,
		Subject:        msg.Subject,
		ProofType:      msg.ProofType,
		Domain:         msg.Domain,
		SpecificityBps: msg.SpecificityBps,
		OutcomeKind:    msg.OutcomeKind,
		TargetId:       msg.TargetId,
		IssuedAt:       sdkCtx.BlockTime().Unix(),
		Height:         sdkCtx.BlockHeight(),
		SIssuance:      sIssuance,
	}
	if err := ms.k.Attestations.Set(ctx, id, a); err != nil {
		return nil, err
	}
	if err := ms.k.SubjectIndex.Set(ctx, collections.Join(a.Subject, id)); err != nil {
		return nil, err
	}
	if err := ms.k.AttestorIndex.Set(ctx, collections.Join(a.Attestor, id)); err != nil {
		return nil, err
	}
	if isOutcome {
		if err := ms.k.TargetIndex.Set(ctx, collections.Join(a.TargetId, id)); err != nil {
			return nil, err
		}
	}

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"attested",
		sdk.NewAttribute("id", fmt.Sprintf("%d", id)),
		sdk.NewAttribute("attestor", a.Attestor),
		sdk.NewAttribute("subject", a.Subject),
		sdk.NewAttribute("proof_type", a.ProofType.String()),
		sdk.NewAttribute("domain", a.Domain),
	))

	// Notify the reputation seam (no-op if x/reputation is absent).
	if ms.k.Rep() != nil {
		if isOutcome {
			targetIsOutcome := target.ProofType == attest.ProofType_PROOF_TYPE_OUTCOME
			refutes := a.OutcomeKind == attest.OutcomeKind_OUTCOME_KIND_REFUTED
			if err := ms.k.Rep().OnOutcome(ctx, a.Attestor, refutes, a.TargetId, target.Attestor, target.Domain, target.SIssuance, targetIsOutcome, id); err != nil {
				return nil, err
			}
		} else if err := ms.k.Rep().OnAttestation(ctx, a.Attestor, a.Domain, int32(a.ProofType), a.SpecificityBps, id); err != nil {
			return nil, err
		}
	}
	return &attest.MsgAttestResponse{Id: id, Height: a.Height}, nil
}

func (ms msgServer) UpdateParams(ctx context.Context, msg *attest.MsgUpdateParams) (*attest.MsgUpdateParamsResponse, error) {
	if _, err := ms.k.addressCodec.StringToBytes(msg.Authority); err != nil {
		return nil, fmt.Errorf("invalid authority address: %w", err)
	}
	if authority := ms.k.GetAuthority(); !strings.EqualFold(msg.Authority, authority) {
		return nil, fmt.Errorf("unauthorized: got %s, want %s", msg.Authority, authority)
	}
	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}
	if err := ms.k.Params.Set(ctx, msg.Params); err != nil {
		return nil, err
	}
	return &attest.MsgUpdateParamsResponse{}, nil
}
