package keeper

import (
	"context"
	"errors"
	"strconv"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/blong-dev/dreamtree/x/attest"
	"github.com/blong-dev/dreamtree/x/attest/projection"
)

var _ attest.QueryServer = queryServer{}

func NewQueryServerImpl(k Keeper) attest.QueryServer { return queryServer{k} }

type queryServer struct{ k Keeper }

func dec(f float64) string { return strconv.FormatFloat(f, 'f', 6, 64) }

func (qs queryServer) Attestation(ctx context.Context, req *attest.QueryAttestationRequest) (*attest.QueryAttestationResponse, error) {
	a, err := qs.k.Attestations.Get(ctx, req.Id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "attestation %d not found", req.Id)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &attest.QueryAttestationResponse{Attestation: a}, nil
}

func (qs queryServer) AttestationsBySubject(ctx context.Context, req *attest.QueryBySubjectRequest) (*attest.QueryBySubjectResponse, error) {
	list, pageRes, err := query.CollectionPaginate(
		ctx, qs.k.SubjectIndex, req.Pagination,
		func(key collections.Pair[string, uint64], _ collections.NoValue) (attest.Attestation, error) {
			return qs.k.Attestations.Get(ctx, key.K2())
		},
		query.WithCollectionPaginationPairPrefix[string, uint64](req.Subject),
	)
	if err != nil {
		return nil, err
	}
	return &attest.QueryBySubjectResponse{Attestations: list, Pagination: pageRes}, nil
}

func (qs queryServer) AttestationsByAttestor(ctx context.Context, req *attest.QueryByAttestorRequest) (*attest.QueryByAttestorResponse, error) {
	list, pageRes, err := query.CollectionPaginate(
		ctx, qs.k.AttestorIndex, req.Pagination,
		func(key collections.Pair[string, uint64], _ collections.NoValue) (attest.Attestation, error) {
			return qs.k.Attestations.Get(ctx, key.K2())
		},
		query.WithCollectionPaginationPairPrefix[string, uint64](req.Attestor),
	)
	if err != nil {
		return nil, err
	}
	return &attest.QueryByAttestorResponse{Attestations: list, Pagination: pageRes}, nil
}

func (qs queryServer) Strength(ctx context.Context, req *attest.QueryStrengthRequest) (*attest.QueryStrengthResponse, error) {
	a, err := qs.k.Attestations.Get(ctx, req.Id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "attestation %d not found", req.Id)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	params, err := qs.k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	pf := projection.LoadParamsF(params)
	now := sdk.UnwrapSDKContext(ctx).BlockTime().Unix()
	return qs.strengthResponse(ctx, a, pf, now)
}

// strengthResponse computes the decomposed strength reading — shared by
// Strength (live params, block clock) and StrengthAt (the M3 dial).
func (qs queryServer) strengthResponse(ctx context.Context, a attest.Attestation, pf projection.ParamsF, now int64) (*attest.QueryStrengthResponse, error) {
	s, rf, err := qs.k.projector(ctx, pf).Strength(a, now)
	if err != nil {
		return nil, err
	}
	return &attest.QueryStrengthResponse{
		Id:              a.Id,
		Strength:        dec(s),
		Reputation:      dec(qs.k.reputationOf(ctx, pf, a.Attestor, a.Domain)),
		Specificity:     dec(projection.SpecificityFactor(a.SpecificityBps)),
		TypeWeight:      dec(pf.Weight[a.ProofType]),
		Decay:           dec(pf.Decay(a, now)),
		RefutedFraction: dec(rf),
		AgeSeconds:      now - a.IssuedAt,
	}, nil
}

// StrengthAt is the dial (backtest M3): the same reading under caller-supplied
// params and/or an as-of clock. Read-only; never on a consensus path.
func (qs queryServer) StrengthAt(ctx context.Context, req *attest.QueryStrengthAtRequest) (*attest.QueryStrengthResponse, error) {
	a, err := qs.k.Attestations.Get(ctx, req.Id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "attestation %d not found", req.Id)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	pf, now, err := qs.dial(ctx, req.ParamsOverride, req.AsOf)
	if err != nil {
		return nil, err
	}
	return qs.strengthResponse(ctx, a, pf, now)
}

func (qs queryServer) WorkValue(ctx context.Context, req *attest.QueryWorkValueRequest) (*attest.QueryWorkValueResponse, error) {
	params, err := qs.k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	pf := projection.LoadParamsF(params)
	now := sdk.UnwrapSDKContext(ctx).BlockTime().Unix()
	v, count, err := qs.k.projector(ctx, pf).WorkValue(req.Subject, now)
	if err != nil {
		return nil, err
	}
	return &attest.QueryWorkValueResponse{Subject: req.Subject, Value: dec(v), AttestationCount: count}, nil
}

// WorkValueAt is the dial (backtest M3) for V(w,t).
func (qs queryServer) WorkValueAt(ctx context.Context, req *attest.QueryWorkValueAtRequest) (*attest.QueryWorkValueResponse, error) {
	pf, now, err := qs.dial(ctx, req.ParamsOverride, req.AsOf)
	if err != nil {
		return nil, err
	}
	v, count, err := qs.k.projector(ctx, pf).WorkValue(req.Subject, now)
	if err != nil {
		return nil, err
	}
	return &attest.QueryWorkValueResponse{Subject: req.Subject, Value: dec(v), AttestationCount: count}, nil
}

// dial resolves the (params, clock) pair for the At queries: override params
// when supplied (validated), live params otherwise; as_of when supplied, block
// time otherwise.
func (qs queryServer) dial(ctx context.Context, override *attest.Params, asOf int64) (projection.ParamsF, int64, error) {
	var p attest.Params
	if override != nil {
		if err := override.Validate(); err != nil {
			return projection.ParamsF{}, 0, status.Errorf(codes.InvalidArgument, "params_override: %v", err)
		}
		p = *override
	} else {
		var err error
		p, err = qs.k.Params.Get(ctx)
		if err != nil {
			return projection.ParamsF{}, 0, err
		}
	}
	now := asOf
	if now == 0 {
		now = sdk.UnwrapSDKContext(ctx).BlockTime().Unix()
	}
	return projection.LoadParamsF(p), now, nil
}

func (qs queryServer) Params(ctx context.Context, _ *attest.QueryParamsRequest) (*attest.QueryParamsResponse, error) {
	params, err := qs.k.Params.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return &attest.QueryParamsResponse{Params: attest.DefaultParams()}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &attest.QueryParamsResponse{Params: params}, nil
}
