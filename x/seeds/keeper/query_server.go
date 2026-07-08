package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/blong-dev/dreamtree/x/seeds"
)

var _ seeds.QueryServer = queryServer{}

// NewQueryServerImpl returns a seeds QueryServer implementation.
func NewQueryServerImpl(k Keeper) seeds.QueryServer {
	return queryServer{k}
}

type queryServer struct {
	k Keeper
}

// Seed returns a single commitment by id.
func (qs queryServer) Seed(ctx context.Context, req *seeds.QuerySeedRequest) (*seeds.QuerySeedResponse, error) {
	seed, err := qs.k.Seeds.Get(ctx, req.Id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "seed %d not found", req.Id)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &seeds.QuerySeedResponse{Seed: seed}, nil
}

// Seeds returns all commitments, paginated.
func (qs queryServer) Seeds(ctx context.Context, req *seeds.QuerySeedsRequest) (*seeds.QuerySeedsResponse, error) {
	list, pageRes, err := query.CollectionPaginate(
		ctx, qs.k.Seeds, req.Pagination,
		func(_ uint64, value seeds.Seed) (seeds.Seed, error) { return value, nil },
	)
	if err != nil {
		return nil, err
	}
	return &seeds.QuerySeedsResponse{Seeds: list, Pagination: pageRes}, nil
}

// SeedsBySubject returns commitments for a subject, paginated.
func (qs queryServer) SeedsBySubject(ctx context.Context, req *seeds.QuerySeedsBySubjectRequest) (*seeds.QuerySeedsBySubjectResponse, error) {
	list, pageRes, err := query.CollectionPaginate(
		ctx, qs.k.SubjectIndex, req.Pagination,
		func(key collections.Pair[string, uint64], _ collections.NoValue) (seeds.Seed, error) {
			return qs.k.Seeds.Get(ctx, key.K2())
		},
		query.WithCollectionPaginationPairPrefix[string, uint64](req.Subject),
	)
	if err != nil {
		return nil, err
	}
	return &seeds.QuerySeedsBySubjectResponse{Seeds: list, Pagination: pageRes}, nil
}

// Params returns the module parameters.
func (qs queryServer) Params(ctx context.Context, req *seeds.QueryParamsRequest) (*seeds.QueryParamsResponse, error) {
	params, err := qs.k.Params.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return &seeds.QueryParamsResponse{Params: seeds.DefaultParams()}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &seeds.QueryParamsResponse{Params: params}, nil
}
