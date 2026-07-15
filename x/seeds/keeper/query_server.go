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

// Seed returns a single leaf-seed by id, synthesized from its batch.
func (qs queryServer) Seed(ctx context.Context, req *seeds.QuerySeedRequest) (*seeds.QuerySeedResponse, error) {
	b, ok, err := qs.k.BatchOf(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !ok {
		return nil, status.Errorf(codes.NotFound, "seed %d not found", req.Id)
	}
	return &seeds.QuerySeedResponse{Seed: SynthesizeSeed(b, req.Id)}, nil
}

// Batch returns a stored anchoring batch by batch id.
func (qs queryServer) Batch(ctx context.Context, req *seeds.QueryBatchRequest) (*seeds.QueryBatchResponse, error) {
	b, err := qs.k.Batches.Get(ctx, req.Id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "batch %d not found", req.Id)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &seeds.QueryBatchResponse{Batch: b}, nil
}

// Seeds lists anchored batches, paginated (one entry per batch — a batch may
// register thousands of leaf-seeds; resolve individual leaves via Seed(id)).
func (qs queryServer) Seeds(ctx context.Context, req *seeds.QuerySeedsRequest) (*seeds.QuerySeedsResponse, error) {
	list, pageRes, err := query.CollectionPaginate(
		ctx, qs.k.Batches, req.Pagination,
		func(_ uint64, value seeds.Batch) (seeds.Batch, error) { return value, nil },
	)
	if err != nil {
		return nil, err
	}
	return &seeds.QuerySeedsResponse{Batches: list, Pagination: pageRes}, nil
}

// SeedsBySubject lists batches for a subject, paginated. The index is keyed
// (subject, batch_id) so pure-convergence batches (no seed range) list too.
func (qs queryServer) SeedsBySubject(ctx context.Context, req *seeds.QuerySeedsBySubjectRequest) (*seeds.QuerySeedsBySubjectResponse, error) {
	list, pageRes, err := query.CollectionPaginate(
		ctx, qs.k.SubjectIndex, req.Pagination,
		func(key collections.Pair[string, uint64], _ collections.NoValue) (seeds.Batch, error) {
			return qs.k.Batches.Get(ctx, key.K2())
		},
		query.WithCollectionPaginationPairPrefix[string, uint64](req.Subject),
	)
	if err != nil {
		return nil, err
	}
	return &seeds.QuerySeedsBySubjectResponse{Batches: list, Pagination: pageRes}, nil
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
