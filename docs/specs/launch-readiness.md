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
- **`gov` module.** Not wired. `UpdateParams`/`SetDomainConfig`/`SetTypePrice`
  handlers exist and are authority-gated, but there is no on-chain proposal
  path — params are set at genesis / by the authority address. Every "lever,
  governance-evolved" in the spec is genesis-set until gov lands.
- **`x/upgrade`.** Not wired. Fine for a fresh genesis; needed before the first
  in-place upgrade (otherwise upgrades = new-genesis migrations).
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

## Not yet exercised

- **Full economic loop as one end-to-end devnet flow.** Components are each
  devnet-proven individually (marketplace 20-seed purchase; reputation R moves;
  photon mint). The whole chain seed→mint→attest→outcome→settle→sale→income has
  not been run as a single scripted flow. Worth doing once before / at launch.

## Open decisions (parked)

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
