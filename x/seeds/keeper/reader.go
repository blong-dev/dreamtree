package keeper

import "context"

// SeedInfo returns a leaf-seed's marketplace-relevant facts: its data_type (the
// priceable type) and its producer (the committer, who earns from sales). Used
// by x/licenses via the SeedReader seam. found=false if the id doesn't resolve.
func (k Keeper) SeedInfo(ctx context.Context, id uint64) (dataType string, producer string, found bool) {
	b, ok, err := k.BatchOf(ctx, id)
	if err != nil || !ok {
		return "", "", false
	}
	return b.DataType, b.Committer, true
}
