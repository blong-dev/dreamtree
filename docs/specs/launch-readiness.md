# Chain launch readiness — what's built, deferred, and open

Living ledger of the dreamtree L1's completeness against `protocol-spec.md`.
Purpose: an honest map of "done vs not" so a fresh-genesis launch is a decision,
not an assumption. Update as items close.

_Last updated: 2026-07-12._

## Built & wired (the value layer)

All five custom modules have message handlers, query servers, genesis handlers,
and are assembled into the app (`app/app.yaml`) atop the standard
auth/bank/staking/distribution/consensus/genutil stack. `reputation` runs in
`end_blockers` (window settlement). The `TestModuleSeamsBind` test guards the
cross-module seams.

| Module | Surface | Status |
|---|---|---|
| `x/seeds` | `CommitSeed` — records/kg_claims committed by hash | built, devnet-proven |
| `x/photons` | per-seed mint to `storer_reward_recipient`; supply pegged 1:1 to corpus (`photons = seeds`), genesis supply 0 | built |
| `x/attest` | `Attest` — attestations, outcomes, endorsements; computes rational `S_issuance` snapshot; routes to reputation | built, devnet-proven |
| `x/reputation` | review windows (τ), signed-verdict settlement, propagation (co-attestor/endorser/reversal), rational `StandingOf`, saturation read-projection | built + unit-tested (math 3/3 closed 2026-07-11) |
| `x/licenses` | `Purchase` marketplace, `SetTypePrice`, time-bound access grants, marketplace toll + value-creation tax | built, devnet-proven |

The economic loop exists end to end: commit a seed → photon mints to the storer
→ attest / endorse / outcome → reputation windows settle and propagate →
marketplace prices and sells access → producers earn, treasury takes toll + tax.

## Not yet built — launch-relevant

- **Cold-start ramp.** Spec §Reputation Dynamics "cold start" wants `ramp > 1`
  for a newcomer's first N validated attestations (clear the dead zone). Not
  coded. Newcomers currently start at `baseline_kyc` + inherited endorsements
  only. Not launch-blocking; affects newcomer bootstrap fairness.
- **`gov` module — DONE 2026-07-12.** Wired (`app/app.yaml` + `app/app.go`);
  the gov module account is the same authority the custom modules already trust,
  so `MsgUpdateParams`/`SetDomainConfig`/`SetTypePrice` now route through
  proposals. Proven end-to-end by `scripts/gov-proof.sh`: a proposal changed
  `x/licenses` `marketplace_toll` 5%→8% on-chain via propose→vote→execute.
  `scripts/init.sh` sets gov `min_deposit` in `dtvp` (the default `stake` denom
  does not exist here — gov would otherwise be undepositable).
- **`x/upgrade` — DONE 2026-07-12.** Wired (`cosmossdk.io/x/upgrade` v0.1.4;
  app.yaml `pre_blockers: [upgrade]` + init_genesis + module config; app.go blank
  import + `UpgradeKeeper` inject, whose `BaseAppOption` store-loader is applied
  by the depinject builder). Proven by `scripts/upgrade-proof.sh`: a gov proposal
  schedules a software-upgrade plan (`q upgrade plan` shows it). In-place upgrades
  are now possible instead of new-genesis migrations.
- **`slashing`.** Absent. Validator misbehavior (downtime, double-sign) is not
  penalized. Acceptable for a permissioned `dtvp`-bond v0; revisit before
  opening the validator set.
- **Domain taxonomy pre-seed.** The 5-level taxonomy (LCC / ISCED / ONET-ISCO)
  is not loaded at genesis; `SetDomainConfig` can set tiers one at a time. Bulk
  seeding is a genesis-data task (gnosis pre-population territory).

## Deferred by design (spec flags these open)

Storage/durability economics: seed-size cap, storage-cost oracle, endowment
(per-seed vs pooled), ingestion/endowment split among storers,
`access_cut_to_storers`, on/off ramps, uptime/durability bond, TEE specifics,
dual-license boundary, receiver-key handoff API. All spec-listed as open; none
required for the core value loop to run.

## Exercised end-to-end

- **Full economic loop — DONE 2026-07-12** (`scripts/e2e-loop.sh`). One scripted
  throwaway devnet proves the whole chain as a single flow: commit seed → photon
  mints to storer (peg holds, supply=3) → bob attests → alice validates the
  outcome → review windows settle → bob's R moves to 1.55 (0.05 bet **+ 0.50
  validated-outcome payout** in the DURABLE_25Y bucket) → marketplace sells
  access (buyer −, producer +, supply unchanged). Exercises the refutation-
  window settlement rewrite live (`settle → applyFloored → contributor +net`).
- **Still worth adding**: a live refutation leg proving R floors at 0 (no debt)
  under a bounded crowd. The floor math is unit-tested (`window_test.go`); the
  live check is not yet scripted.

## Open decisions (parked)

- **Creation-credit-forward — PARKED 2026-07-12 (do not build now).** The
  idea: a boon realized by a derivative work B flows a share to its sources
  A1..A4. NOT hardwired, by design — the manifesto (`03-architecture.md`) names
  transitive compensation the unsolved hard problem, and hardwiring a fixed
  royalty split would violate "the protocol informs; the market prices; we don't
  dictate compensation" (spec §8). Decision (owner, 2026-07-12): do **not**
  define compensation at this time. Two layers to keep distinct if revisited:
  (a) **compensation** (photons flowing to sources) — parked indefinitely;
  (b) **value signal** (a source's reputation/`V` rising when used) — this
  already exists but is weak: a `USE` citation inflates the source's `V(A)` in
  proportion to the *citer's* standing `R` × `WeightUse` (0.5) at citation time,
  and does **not** scale with how successful the citing work B later becomes.
  The compensation-free lever, if wanted later, is to weight a citation's
  contribution to `V(A)` by `V(B)` / B's validated outcomes (eigenvector/
  PageRank flavor) — and it can live entirely in the read-projection
  (`x/attest/keeper/projection.go`), off consensus, so it carries no determinism
  risk. Not started.


- **Refutation friction to reach 0 — DECIDED 2026-07-12: enough friction, no
  change.** `neg_asymmetry = 2.0`; reaching 0 requires only a paper-shape-
  bounded crowd (capped at `M_cap`) plus the review window, not an additional
  cred-weighted quorum. Rationale: paper-shape already bounds crowd pile-on and
  the zero floor makes any bounded blow recoverable-from. Revisit only if live
  data shows bounded crowds driving legitimate contributors to 0 too easily —
  then consider a cred-weighted quorum or longer τ on refuting windows.

## Resolved divergences from earlier spec drafts

- **Single mint stream, not two.** The 2026-05-22 draft had two streams (`S` to
  creator, `P` to storer). The strict tokenomics (2026-07-10) collapsed this:
  one photon mints per seed to the storer-validators; the **creator earns only
  from marketplace sales**, never at ingestion. Selective purchasing is the
  demand signal. `photon` is not the staking/gas token — `dtvp` bonds validators.
