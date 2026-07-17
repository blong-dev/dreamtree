# upgrade-1 — the first governed upgrade (DT-20)

Status: SPEC (ratified riders; mechanism in build)
Owner ruling basis: 2026-07-16/17 Group 3 rulings, canon commits `f6aa361`, `1157c0b`, `13e38ba`.
Card: DT-20. Prior art: DT-17 (folded into this card — its x/upgrade item IS this spec).

## Why this exists

The chain `dreamtree` launched 2026-07-16 with a one-time wipe, explicitly the last one:
**all future chain changes go through governed in-place upgrades** — same chain-id, same
history, seeds preserved. This spec defines the mechanism and the first payload.

## The mechanism

`x/upgrade` is already wired in the app (`app/app.yaml`: module registered, `pre_blockers:
[upgrade]`; `app/app.go`: `UpgradeKeeper`). What's missing is the handler and the procedure.

1. **Handler registration** (`app/upgrades.go`, new): `SetUpgradeHandler("upgrade-1", ...)`
   runs module migrations plus the rider migrations below. Store loader registers any
   added stores for the upgrade height.
2. **Proposal**: `MsgSoftwareUpgrade` (gov authority) naming `upgrade-1` and a halt height.
   Submitted and voted by dt-validator. GOTCHA (proven on proposal #1): anchord signs with
   the same account continuously — every tx needs the sequence-retry loop (query sequence,
   pass `--sequence` explicitly, retry on code 32).
3. **Halt + swap**: at the height, FinalizeBlock panics with `UPGRADE "upgrade-1"
   NEEDED` and CometBFT logs `CONSENSUS FAILURE` — **the process hangs rather than
   exits** (proven in rehearsal). Ops: watch the journal for the NEEDED line, then
   `systemctl stop dreamtree`, build new binary on m1, ship to m3, replace
   `/home/b/dreamtree/bin/dreamtreed`, `systemctl start dreamtree`. New binary sees the upgrade info, runs the handler,
   resumes. No cosmovisor (single validator, native systemd; the actuator recipe
   `restart_dreamtree` — add alongside existing verbs if absent — is the ops-fabric hook).
4. **Verify**: peg intact (photons = distinct atoms), anchord loop resumes (SEED commits),
   each rider's behavior spot-checked (checklist per rider below).

## The payload — five ratified riders

### R1 — Z2 zero-floor fix (already in code, commit `4c69cc6`)
Reversal negations floored, running floor in `StandingOf` + `reputationRaw`. Read-time +
settlement logic only; no state migration. Ships with the binary.
Verify: none live-observable yet (no negative excursions on chain); unit tests carry it.

### R2 — Standing starts at ZERO; baseline granted by verification
Canon: decision log 2026-07-16. Default standing for an unknown address = 0; `baseline_kyc`
is GRANTED via a verified set.
- New state in `x/reputation`: `VerifiedSet` (addr → member), `MsgSetVerified{authority,
  address, verified}` gated to the gov authority (the v0 stand-in for the identity layer;
  durable home = did:webvh, DT-4).
- `StandingOf` / `reputationRaw` / cred: baseline term = `baseline_kyc` if verified else 0.
- Migration: seed the set with the validator/anchor account so the live loop and gov
  keep working through the flip.
Verify: query standing of a random address (0), of dt-validator (baseline+).

### R3 — ALL kinds mint
Canon: 2026-07-16 ("every atom is an observation"). Remove the `mintable_kinds` gate check
in `x/photons/keeper/mint.go`; the proto field stays (reserved/deprecated — proto field
removal is a wire break), logic ignores it. No retroactive mint needed: every kind
committed pre-upgrade (record, kg_claim) was already mintable, so the peg carries over
exact. Verify: peg check post-upgrade; commit a non-whitelisted kind on devnet and see it mint.

### R4 — Constant pricing: 1 photon per seed per day
Canon: 2026-07-16, reaffirmed 2026-07-17 after the owner's free-market challenge (the
constant is a unit definition, not a price control; value discovery lives in volume).
- `x/licenses/keeper/purchase.go`: price per seed = `AccessDurationDays × 1_000_000
  uphoton` (1 photon/seed/day), replacing the `TypePrices` lookup. Every seed with a
  resolvable producer is purchasable (no "unpriced type" skip). Toll + value-creation-tax
  plumbing unchanged (tax at SALE — ratified).
- `MsgSetTypePrice`: handler returns `ErrRetired` (msg stays registered — decoding
  history must not break). Migration: delete all `TypePrices` entries.
Verify: devnet purchase of a 3-seed swath costs 3 photons + toll.

### R5 — Endorsement breadth bounded paper-shape
Canon: 2026-07-17 (`13e38ba`). Per (endorsee, domain): `E_total = E_cap·[1−Π(1−eᵢ/E_cap)]`,
`E_cap = e_cap_mult × max(eᵢ)`, new param `reputation.e_cap_mult = 2.0` (shape ratified,
value INTERIM — backtest tunes).
- `StandingOf` + `reputationRaw`: ENDORSEMENT-bucket contributions are collected during
  the walk (relevance/decay applied per-contribution as today) and folded paper-shape
  after it, instead of summing inline. `hooks.go` enqueue unchanged — aggregation is a
  read/settle law, not a write-path change.
- Migration: param addition only (`e_cap_mult`); stored contributions untouched.
Verify: devnet — two endorsers vs forty endorsers of equal R; forty ≤ 2× the strongest.

## Sequencing

1. **Before anything ships**: gov proposal #1 (burn flags off) executes 2026-07-18 19:13Z —
   verify, then submit upgrade-1.
2. Build riders on `main` (deploy is explicit — binary swap only at the halt height).
3. Devnet rehearsal: run the full proposal → halt → swap → verify loop on a local devnet
   before touching m3.

## The larger order (owner-approved 2026-07-17)

1. **DT-20 (this)** — mechanism + upgrade-1. Unblocks everything below.
2. **Backtest M1–M4** (`measurement-backtest.md`) — promotes every INTERIM lever
   (review threshold 4.0, s_max, coattestor_weight, e_cap_mult, citation λ) and tests
   whether volume alone carries the value signal under constant pricing.
3. **DT-22 C2PA** at roots intake — roots itself is LIVE in production (DT-1 epic
   closed complete 2026-07-17; intake shapes are already stable), so C2PA is the
   next real product build. Roots' remaining cards: DT-6 custody handoff (owner-
   parked until real users), DT-8 auth fix, DT-4 did:webvh (also the verified
   set's durable home).
4. Planned-not-built spec items, as they come due: meta-attestation pre-population
   (rides taxonomy seeding), four-hard-rules enforcement stack, use declarations,
   VC issuer restrictions.
