package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/blong-dev/dreamtree/x/licenses"
)

var _ licenses.QueryServer = queryServer{}

func NewQueryServerImpl(k Keeper) licenses.QueryServer { return queryServer{k} }

type queryServer struct{ k Keeper }

func (qs queryServer) TypePrice(ctx context.Context, req *licenses.QueryTypePriceRequest) (*licenses.QueryTypePriceResponse, error) {
	price, err := qs.k.TypePrices.Get(ctx, req.DataType)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return &licenses.QueryTypePriceResponse{Priced: false}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &licenses.QueryTypePriceResponse{Price: price, Priced: true}, nil
}

func (qs queryServer) Access(ctx context.Context, req *licenses.QueryAccessRequest) (*licenses.QueryAccessResponse, error) {
	g, err := qs.k.AccessGrants.Get(ctx, collections.Join(req.Buyer, req.SeedId))
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return &licenses.QueryAccessResponse{HasAccess: false}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	now := sdk.UnwrapSDKContext(ctx).BlockTime().Unix()
	return &licenses.QueryAccessResponse{HasAccess: now < g.ExpiresAt, ExpiresAt: g.ExpiresAt}, nil
}

func (qs queryServer) Params(ctx context.Context, _ *licenses.QueryParamsRequest) (*licenses.QueryParamsResponse, error) {
	p, err := qs.k.Params.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return &licenses.QueryParamsResponse{Params: licenses.DefaultParams()}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &licenses.QueryParamsResponse{Params: p}, nil
}
