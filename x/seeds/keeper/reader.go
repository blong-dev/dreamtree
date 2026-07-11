package keeper

import "context"

// SeedInfo returns a seed's marketplace-relevant facts: its data_type (the
// priceable type) and its producer (the committer, who earns from sales). Used
// by x/licenses via the SeedReader seam. found=false if the seed doesn't exist.
func (k Keeper) SeedInfo(ctx context.Context, id uint64) (dataType string, producer string, found bool) {
	s, err := k.Seeds.Get(ctx, id)
	if err != nil {
		return "", "", false
	}
	return s.DataType, s.Committer, true
}
