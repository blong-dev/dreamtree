package licenses

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper — the subset x/licenses needs to move photons buyer → producers + treasury.
type BankKeeper interface {
	SendCoins(ctx context.Context, from, to sdk.AccAddress, amt sdk.Coins) error
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// SeedReader — the subset of x/seeds x/licenses reads (a seed's type + producer).
type SeedReader interface {
	SeedInfo(ctx context.Context, id uint64) (dataType string, producer string, found bool)
}
