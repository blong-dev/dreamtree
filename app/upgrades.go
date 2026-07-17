package app

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"

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
			// R4 (constant pricing): the per-type price table is retired —
			// access is 1 photon per seed per day, a protocol constant.
			if err := app.LicensesKeeper.TypePrices.Clear(ctx, nil); err != nil {
				return nil, err
			}
			// R5 (endorsement paper-shape): stored params predate the
			// e_cap_mult field; write the ratified default so readers don't
			// ride the fallback forever. R2's verified set starts EMPTY by
			// design — every grant is an explicit governed MsgSetVerified.
			p, err := app.ReputationKeeper.Params.Get(ctx)
			if err != nil {
				return nil, err
			}
			if p.ECapMult == "" {
				p.ECapMult = reputation.DefaultECapMult
				if err := app.ReputationKeeper.Params.Set(ctx, p); err != nil {
					return nil, err
				}
			}
			// R1 (Z2 floor) and R3 (all kinds mint) are pure logic changes —
			// they ship with the binary and need no state migration.
			return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
		})
}
