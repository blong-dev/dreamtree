package app

import (
	"context"
	"time"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/blong-dev/dreamtree/x/attest"
	"github.com/blong-dev/dreamtree/x/reputation"
)

// Upgrade1 — the chain's first governed in-place upgrade (DT-20,
// docs/specs/upgrade-1.md). The launch wipe of 2026-07-16 was the last one;
// every chain change from here rides a MsgSoftwareUpgrade through this path.
const Upgrade1 = "upgrade-1"

// RegisterUpgradeHandlers wires every known upgrade. Handlers must be
// registered before Load so a node restarted at the halt height finds them.
func (app *DreamtreeApp) RegisterUpgradeHandlers() {
	app.UpgradeKeeper.SetUpgradeHandler(Upgrade1,
		func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			logger := sdk.UnwrapSDKContext(ctx).Logger().With("upgrade", Upgrade1)
			// R4 (constant pricing): the per-type price table is retired —
			// access is 1 photon per seed per day, a protocol constant.
			if err := app.LicensesKeeper.TypePrices.Clear(ctx, nil); err != nil {
				return nil, err
			}
			logger.Info("R4: TypePrices cleared")
			// R5 (endorsement paper-shape): stored params predate the
			// e_cap_mult field; write the ratified default so readers don't
			// ride the fallback forever. R2's verified set starts EMPTY by
			// design — every grant is an explicit governed MsgSetVerified.
			p, err := app.ReputationKeeper.Params.Get(ctx)
			if err != nil {
				return nil, err
			}
			logger.Info("R5: reputation params read", "e_cap_mult_before", p.ECapMult)
			if p.ECapMult == "" {
				p.ECapMult = reputation.DefaultECapMult
				if err := app.ReputationKeeper.Params.Set(ctx, p); err != nil {
					return nil, err
				}
				readBack, _ := app.ReputationKeeper.Params.Get(ctx)
				logger.Info("R5: e_cap_mult migrated", "now", readBack.ECapMult)
			}
			// Backtest M2 (rides upgrade-1): citation_uplift_lambda promoted
			// from a compile-time const to a governable attest param — fill
			// the ratified default on stored params.
			ap, err := app.AttestKeeper.Params.Get(ctx)
			if err != nil {
				return nil, err
			}
			if ap.CitationUpliftLambda == "" {
				ap.CitationUpliftLambda = attest.DefaultParams().CitationUpliftLambda
				if err := app.AttestKeeper.Params.Set(ctx, ap); err != nil {
					return nil, err
				}
				logger.Info("M2: citation_uplift_lambda promoted", "value", ap.CitationUpliftLambda)
			}
			// Gov clock (owner 2026-07-17): the 48h/24h genesis defaults made
			// same-day governed deploys impossible on a single-validator
			// chain. Shorten to 1h voting / 30m expedited. Done HERE (not a
			// parallel MsgUpdateParams) because gov params are full-replace:
			// proposal #1's burn-flag update executes first and would revert
			// a duration change that landed before it — the plan height is
			// chosen after #1's execution so this ordering holds.
			gp, err := app.GovKeeper.Params.Get(ctx)
			if err != nil {
				return nil, err
			}
			hour, half := time.Hour, 30*time.Minute
			gp.VotingPeriod = &hour
			gp.ExpeditedVotingPeriod = &half
			if err := app.GovKeeper.Params.Set(ctx, gp); err != nil {
				return nil, err
			}
			logger.Info("gov clock shortened", "voting", hour, "expedited", half)
			// R1 (Z2 floor) and R3 (all kinds mint) are pure logic changes —
			// they ship with the binary and need no state migration.
			return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
		})
}
