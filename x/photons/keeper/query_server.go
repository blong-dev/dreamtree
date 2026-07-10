package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/blong-dev/dreamtree/x/photons"
)

var _ photons.QueryServer = queryServer{}

func NewQueryServerImpl(k Keeper) photons.QueryServer { return queryServer{k} }

type queryServer struct{ k Keeper }

func (qs queryServer) Supply(ctx context.Context, _ *photons.QuerySupplyRequest) (*photons.QuerySupplyResponse, error) {
	n, err := qs.k.Minted.Peek(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &photons.QuerySupplyResponse{Minted: n}, nil
}

func (qs queryServer) Params(ctx context.Context, _ *photons.QueryParamsRequest) (*photons.QueryParamsResponse, error) {
	p, err := qs.k.Params.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return &photons.QueryParamsResponse{Params: photons.DefaultParams()}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &photons.QueryParamsResponse{Params: p}, nil
}
