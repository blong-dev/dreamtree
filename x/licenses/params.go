package licenses

import (
	"fmt"

	"cosmossdk.io/math"
)

func DefaultParams() Params {
	return Params{
		MarketplaceToll:    math.LegacyMustNewDecFromStr("0.05"), // 5%
		AccessDurationDays: 1,
		TreasuryRecipient:  "", // set to dreamtree at launch
	}
}

func (p Params) Validate() error {
	if p.MarketplaceToll.IsNegative() || p.MarketplaceToll.GT(math.LegacyOneDec()) {
		return fmt.Errorf("marketplace_toll must be in [0,1]")
	}
	if p.AccessDurationDays == 0 {
		return fmt.Errorf("access_duration_days must be > 0")
	}
	return nil
}
