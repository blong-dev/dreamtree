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

	"github.com/blong-dev/dreamtree/x/reputation"
)

var _ reputation.QueryServer = queryServer{}

func NewQueryServerImpl(k Keeper) reputation.QueryServer { return queryServer{k} }

type queryServer struct{ k Keeper }

func dec(f float64) string { return strconv.FormatFloat(f, 'f', 6, 64) }

func (qs queryServer) Reputation(ctx context.Context, req *reputation.QueryReputationRequest) (*reputation.QueryReputationResponse, error) {
	if req.Signer == "" || req.Domain == "" {
		return nil, status.Error(codes.InvalidArgument, "signer and domain required")
	}
	now := sdk.UnwrapSDKContext(ctx).BlockTime().Unix()
	raw, saturation, count, err := qs.k.reputationRaw(ctx, req.Signer, req.Domain, now)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	p, err := qs.k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	eff := effectiveR(raw, saturation, pf(p.DampeningK))
	return &reputation.QueryReputationResponse{
		Reputation:        dec(eff),
		Raw:               dec(raw),
		ContributionCount: count,
	}, nil
}

func (qs queryServer) Contributions(ctx context.Context, req *reputation.QueryContributionsRequest) (*reputation.QueryContributionsResponse, error) {
	list, pageRes, err := query.CollectionPaginate(
		ctx, qs.k.SignerIndex, req.Pagination,
		func(key collections.Pair[string, uint64], _ collections.NoValue) (reputation.Contribution, error) {
			return qs.k.Contributions.Get(ctx, key.K2())
		},
		query.WithCollectionPaginationPairPrefix[string, uint64](req.Signer),
	)
	if err != nil {
		return nil, err
	}
	// optional domain filter (post-filter; index is signer-keyed)
	if req.Domain != "" {
		filtered := list[:0]
		for _, c := range list {
			if c.Domain == req.Domain {
				filtered = append(filtered, c)
			}
		}
		list = filtered
	}
	return &reputation.QueryContributionsResponse{Contributions: list, Pagination: pageRes}, nil
}

func (qs queryServer) DomainConfig(ctx context.Context, req *reputation.QueryDomainConfigRequest) (*reputation.QueryDomainConfigResponse, error) {
	cfg, err := qs.k.DomainConfigs.Get(ctx, req.Path)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "no domain config for %q (defaults apply)", req.Path)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &reputation.QueryDomainConfigResponse{Config: cfg}, nil
}

func (qs queryServer) Params(ctx context.Context, _ *reputation.QueryParamsRequest) (*reputation.QueryParamsResponse, error) {
	p, err := qs.k.Params.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return &reputation.QueryParamsResponse{Params: reputation.DefaultParams()}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &reputation.QueryParamsResponse{Params: p}, nil
}
