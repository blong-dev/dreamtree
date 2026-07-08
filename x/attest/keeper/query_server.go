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
	pf := loadParamsF(params)
	now := sdk.UnwrapSDKContext(ctx).BlockTime().Unix()
	s, rf, err := qs.k.strength(ctx, a, pf, now)
	if err != nil {
		return nil, err
	}
	return &attest.QueryStrengthResponse{
		Id:              a.Id,
		Strength:        dec(s),
		Reputation:      dec(pf.reputation(a.Attestor, a.Domain)),
		Specificity:     dec(specificityFactor(a.SpecificityBps)),
		TypeWeight:      dec(pf.weight[a.ProofType]),
		Decay:           dec(pf.decay(a, now)),
		RefutedFraction: dec(rf),
		AgeSeconds:      now - a.IssuedAt,
	}, nil
}

func (qs queryServer) WorkValue(ctx context.Context, req *attest.QueryWorkValueRequest) (*attest.QueryWorkValueResponse, error) {
	params, err := qs.k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	pf := loadParamsF(params)
	now := sdk.UnwrapSDKContext(ctx).BlockTime().Unix()
	v, count, err := qs.k.workValue(ctx, req.Subject, pf, now)
	if err != nil {
		return nil, err
	}
	return &attest.QueryWorkValueResponse{Subject: req.Subject, Value: dec(v), AttestationCount: count}, nil
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
