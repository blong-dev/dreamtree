# DreamTree Protocol — Parameters

*The canonical source of truth for every tunable lever in the protocol. Started 2026-05-22.*

*Companion to [`protocol-spec.md`](./protocol-spec.md). The spec describes what each lever **does**; this document holds what each lever **is** (canonical name) and **currently holds** (stand-in value). When the spec and this document disagree on a value, this document wins. When the protocol is built, the canonical block below lifts directly into a config file / on-chain parameter registry.*

---

## The abstraction

Every numeric value in the protocol is a named variable, not a hardcoded constant (see protocol-spec heuristic #7). Three reasons this pays dividends:

1. **One place to change.** A lever moves here; everywhere that references it updates at once.
2. **Governance operates on names.** When parameter-setting moves from founder to community vote, the ballot is "set `reputation.neg_asymmetry` to X" — a named, bounded, auditable change.
3. **Machine-readable path.** The canonical block is valid YAML. It becomes the protocol's config at build time and an on-chain parameter registry at v2+.

**Stand-in values** are placeholders chosen so the system runs. They are not claims about the right value. `null` means "lever identified, value not yet chosen."

**Disposition** codes:
- `settled` — the concept is decided; the value may still tune, but the shape won't change
- `governance` — founder-set at v0, moves to community vote as governance matures
- `founder→governance` — same, explicitly flagged as a launch-bootstrap parameter
- `per-domain` — set per node in the domain taxonomy, not globally

---

## Canonical values

```yaml
# DreamTree Protocol Parameters — canonical source of truth
# All values are stand-ins unless noted settled in the reference table.
version: 0.8.0
updated: 2026-07-16   # DT-21 reconciliation: registry catches up with the build

reputation:
  baseline_kyc: 1.0              # R floor for any verified human entering a new domain
  neg_asymmetry: 2.0             # negative events hit this many times harder than positive
  endorsement_inheritance: 0.25  # fraction of R(A) inherited by B at first hop; geometric per hop
  cred_recursion_depth: 2        # hops the cred(source) recursion traverses before terminating
  review_window_base: 1.0        # base_window in τ(M) = base · √(M/threshold), days (√ ratified 2026-07-10: no transcendental in consensus)
  review_window_threshold: 4.0   # threshold in τ(M) — retuned for √'s fat tail near 0 (build, 2026-07-1x; INTERIM)
  lambda_r_base: null            # TBD — base reputation decay rate, per year
  lambda_r_target: baseline_kyc  # R decays toward baseline, not zero (settled)
  saturation_point:                   # per-domain S in effective_R = S + k·log(1 + (R−S)/S) when R > S
    small: 5                          # niche / sparse domains
    standard: 10                      # default
    large: 50                         # hot / dense domains (medicine, AI, mature ecosystems)
  dampening_k: 5                      # global k — compression strength past saturation
  outcome_magnitude_beta: 1.0         # β in M_O = min(M_cap, β · S(att, t_issuance) · √cred(reporter))
  outcome_magnitude_cap_multiplier: 5 # M_cap = (this) · S(att, t_issuance); single outcome ≤ N× original bet
  coattestor_weight: 0.25             # INTERIM — stands in for spec's attestation_weight(a_i,c) in outcome propagation
  attest_bet_scale: 0.1               # INTERIM — magnitude of the unvalidated-attestation "bet" R-move
  lambda_endorsement: 0.08            # INTERIM — endorsement-contribution decay, per year
  s_max: 10.0                         # normalizer in the V(w) aggregation — INTERIM (build-chosen; backtest target)

decay:                           # attestation-strength decay rates, per year
  proof_origin: 0.0              # permanent
  proof_replication: 0.015       # ~45 yr half-life
  proof_rigor: 0.04              # ~17 yr half-life
  proof_use: 0.08                # ~9 yr half-life
  validated_outcome_halflife_years: 25.0

attest:                          # DT-21 reconciliation — values the build runs, previously unregistered
  type_weight:                   # INTERIM, all — the spec names the lever, never the values
    origin: 1.0
    rigor: 1.0
    use: 0.5                     # a citation is weaker signal than authorship/review
    replication: 1.0
    outcome: 1.0
    endorsement: 1.0             # (proof type itself pending ratification — DT-21 B8)
  partial_outcome_refute_weight: 0.5   # INTERIM — dt.outcome.partial's weight in refuted_fraction
  max_coattestors: 64            # INTERIM — outcome-propagation fan-out cap (2× for endorsers)
  obsolescence_default: 1.0      # attest-side global multiplier; per-domain wiring pending (DT-21 W4)
  citation_uplift_lambda: 1.0    # INTERIM, hardcoded const — promotion to param = backtest M2

seeds:
  max_commitment_bytes: 512      # commitment digest bound (bodies never enter consensus)
  max_source_ref_bytes: 512
  max_batch_new_count: 1000000   # per-batch mint cap (supply-griefing bound; ratified 2026-07-15)

photons:
  mintable_kinds: [record, kg_claim]  # INTERIM — kinds whose NEW leaves mint (peg gate; ratification = DT-21 M6/B3)

domain:
  attenuation:                   # R spillover UP the 5-level taxonomy
    to_specialty: 0.70           # L5 -> L4
    to_discipline: 0.40          # -> L3
    to_field: 0.15               # -> L2
    to_class: 0.03               # -> L1
  obsolescence_multiplier:       # multiplies effective attestation decay by domain volatility
    foundational: 0.3
    standard: 1.0
    frontier: 3.0

coldstart:
  ramp_factor: null              # TBD — >1 amplification on early validated attestations
  ramp_count: null               # TBD — N, number of early attestations amplified

economics:                       # founder-set at v0, governance-evolved
  marketplace_toll: 0.05         # 5% (reconciled 2026-05-22)
  value_creation_tax: 0.005      # 0.5% (owner 2026-07-11, was 1.5%)
  access_duration_days: 1        # 1 day default (2026-05-22); re-access = re-buy
  access_cut_to_storers: null    # TBD — slice of each access payment that funds ongoing storage
  # photon_issuance: RESOLVED — not a free parameter. Supply is pegged: 1 photon per seed
  # (photons = seeds). Minted to the storer-validators of each seed (not the author).
  # No halving/inflation schedule; issuance = corpus growth.
  storage_replication_factor: null  # TBD — how many nodes hold each shard
  block_cadence_seconds: 3       # stand-in (~2–5s); timeout_commit-driven

# INVARIANTS (not levers — fixed, never tunable):
#   creator_equality_within_type # p(c1,s,a) = p(c2,s,a) = p(c3,s,a): within a data type,
#                                # all creators priced equally. The market sets price ACROSS
#                                # types (marginal); the protocol guarantees equality ACROSS
#                                # creators of a type. The protocol never prices the person.
```

---

## Parameter reference

| Variable | Stand-in | Units | Governs | Constraint | Disposition |
|---|---|---|---|---|---|
| `reputation.baseline_kyc` | 1.0 | R | floor for a verified human in a new domain | > 0 | settled (concept) |
| `reputation.neg_asymmetry` | 2.0 | ratio | how much harder bad signal hits than good | ≥ 1 | governance |
| `reputation.endorsement_inheritance` | 0.25 | fraction | reputation flow A→B per hop (geometric) | [0, 1] | governance |
| `reputation.cred_recursion_depth` | 2 | hops | credential-laundering resistance | integer ≥ 1 | governance |
| `reputation.review_window_base` | 1.0 | days | base of τ(M) = base·√(M/threshold) | > 0 | governance |
| `reputation.review_window_threshold` | 4.0 | magnitude | threshold of τ(M) | > 0 | **INTERIM** (build retune) |
| `reputation.lambda_r_base` | null | 1/yr | reputation atrophy speed | ≥ 0 | governance |
| `reputation.lambda_r_target` | baseline_kyc | R | reputation as stock vs. flow | — | **settled** (baseline) |
| `reputation.saturation_point.small` | 5 | R | log-dampening threshold for niche / sparse domains | > 0 | per-domain governance |
| `reputation.saturation_point.standard` | 10 | R | log-dampening threshold; default | > 0 | per-domain governance |
| `reputation.saturation_point.large` | 50 | R | log-dampening threshold for hot / dense domains | > 0 | per-domain governance |
| `reputation.dampening_k` | 5 | scalar | compression strength past saturation in `S + k·log(1 + (R−S)/S)` | > 0 | governance |
| `reputation.outcome_magnitude_beta` | 1.0 | scalar | β in `M_O = min(M_cap, β · S(att, t_issuance) · √cred(reporter))` | ≥ 0 | governance |
| `reputation.outcome_magnitude_cap_multiplier` | 5 | scalar | `M_cap = (this) · S(att, t_issuance)` — single-outcome ceiling as a multiplier of the original bet | ≥ 1 | governance |
| `reputation.s_max` | 10.0 | S | normalizer in V(w) aggregation | > 0 | **INTERIM** (build-chosen) |
| `reputation.coattestor_weight` | 0.25 | fraction | outcome propagation to co-attestors (× specificity) | [0, 1] | **INTERIM** |
| `reputation.attest_bet_scale` | 0.1 | scalar | unvalidated-bet R-move magnitude | ≥ 0 | **INTERIM** |
| `reputation.lambda_endorsement` | 0.08 | 1/yr | endorsement-contribution decay | ≥ 0 | **INTERIM** |
| `attest.type_weight.use` | 0.5 | fraction | Use-proof strength weight (others 1.0) | [0, 1] | **INTERIM** |
| `attest.partial_outcome_refute_weight` | 0.5 | fraction | PARTIAL outcome's weight in refuted_fraction | [0, 1] | **INTERIM** |
| `attest.max_coattestors` | 64 | count | propagation fan-out cap | ≥ 1 | **INTERIM** |
| `attest.citation_uplift_lambda` | 1.0 | scalar | creation-credit-forward strength (hardcoded; promote = backtest M2) | ≥ 0 | **INTERIM** |
| `seeds.max_batch_new_count` | 1000000 | count | per-batch mint cap | ≥ 1 | ratified 2026-07-15 |
| `photons.mintable_kinds` | record, kg_claim | set | leaf kinds whose NEW atoms mint | — | **INTERIM** (peg gate) |
| `decay.proof_origin` | 0.0 | 1/yr | Proof-of-Origin aging | = 0 | **settled** (permanent) |
| `decay.proof_replication` | 0.015 | 1/yr | Proof-of-Replication aging (~45 yr ½) | ≥ 0 | governance |
| `decay.proof_rigor` | 0.04 | 1/yr | Proof-of-Rigor aging (~17 yr ½) | ≥ 0 | governance |
| `decay.proof_use` | 0.08 | 1/yr | Proof-of-Use aging (~9 yr ½) | ≥ 0 | governance |
| `decay.validated_outcome_halflife_years` | 25.0 | yr | how long a validated success keeps paying | > 0 | governance |
| `domain.attenuation.to_specialty` | 0.70 | fraction | L5→L4 R spillover | [0, 1] | governance |
| `domain.attenuation.to_discipline` | 0.40 | fraction | →L3 R spillover | [0, 1] | governance |
| `domain.attenuation.to_field` | 0.15 | fraction | →L2 R spillover | [0, 1] | governance |
| `domain.attenuation.to_class` | 0.03 | fraction | →L1 R spillover | [0, 1] | governance |
| `domain.obsolescence_multiplier.foundational` | 0.3 | multiplier | decay scaling for durable domains | > 0 | per-domain |
| `domain.obsolescence_multiplier.standard` | 1.0 | multiplier | decay scaling, baseline | > 0 | per-domain |
| `domain.obsolescence_multiplier.frontier` | 3.0 | multiplier | decay scaling for fast-moving domains | > 0 | per-domain |
| `coldstart.ramp_factor` | null | multiplier | newcomer early-win amplification | > 1 | governance |
| `coldstart.ramp_count` | null | integer | N early attestations amplified | ≥ 0 | governance |
| `economics.marketplace_toll` | 0.05 | fraction | infrastructure funding from transactions | [0, 1] | founder→governance |
| `economics.value_creation_tax` | 0.005 | fraction | infrastructure funding from work issuance | [0, 1] | founder→governance |
| `economics.access_duration_days` | 1 | days | how long one photon buys access to a seed | > 0 | founder→governance |
| `economics.access_cut_to_storers` | null | fraction | slice of an access payment funding ongoing storage | [0, 1] | founder→governance |
| `economics.storage_replication_factor` | null | count | how many nodes hold each shard | ≥ 1 | founder→governance |
| `economics.block_cadence_seconds` | 3 | seconds | block time | > 0 | founder→governance |

---

## Invariants (fixed, never tunable)

- **`creator_equality_within_type`** — `p(c1,s,a) = p(c2,s,a) = p(c3,s,a)`. Within a data type, every creator is priced identically. The market discovers value *across* types at the margin; the protocol guarantees equality *across creators* of the same type. The protocol never prices the person. (This replaces the earlier global "1 seed = 1 photon" rule — value is now marginal and market-set per type; what survives is creator-equality-within-type.)
- **`photons = seeds`** — the photon supply is pegged 1:1 to the count of **distinct atoms** (seed = atom, ratified 2026-07-15): one photon mints per NEW leaf-seed at ingestion, to the storer recipient; converged re-observations accrue sigma and never re-mint. No halving schedule, no inflation curve; supply IS the corpus. Caveat pending ratification (DT-21 M6/B3): the `photons.mintable_kinds` gate means only whitelisted leaf kinds mint — the peg as built reads "photons = mintable-kind seeds".

---

## Not parameters

Some values that look like levers are deliberately **not** in this registry:

- **Per-type data prices (`N_a`).** Market outcomes — supply and demand set what access to a data type costs, in photons. The protocol injects verified information so the market prices accurately (heuristic #8), but never sets the price. Never appears here.
- **Producer compensation.** Equals volume × `N_a` — how many of a creator's seeds sell, at their type's market price. A market outcome, not a parameter. Never appears here.
- **Domain taxonomy contents.** The taxonomy (which classes/fields/disciplines/specialties exist) is a governance-maintained dataset, not a numeric lever. Seeded at v0 from LCC + ISCED + ONET/ISCO-08.
- **The pre-population unsigned discount** — currently lives in protocol-spec §Identity; should be promoted here as a lever once its shape firms up. Flagged for migration.

---

## Change log

- **2026-05-22 — v0.1.0.** Initial registry. 26 parameters extracted from protocol-spec §Levers and §Reputation Dynamics. Two settled (`reputation.lambda_r_target`, `decay.proof_origin`); five `null` pending design (lambda_r_base, saturation_point, outcome_magnitude, s_max, coldstart.*); rest are governance-evolved stand-ins.
- **2026-05-22 — v0.2.0.** Two-token model (Photons + Seeds). Added `economics.access_duration_days` and `economics.photon_issuance`. Added the first **invariant**: `seed_access_per_photon = 1` (fixed, not a lever). Renamed the "not a parameter" entry from contributor-compensation-rate to producer-compensation-rate and clarified it's volume-driven (1 P/seed fixed). `economics.marketplace_toll` flagged unreconciled (worked example ~10% vs. earlier 1–2%).
- **2026-05-22 — v0.3.0.** Per-type market pricing adopted (data value is marginal; market sets price per type). Invariant **changed** from `seed_access_per_photon = 1` to `creator_equality_within_type` (`p(c1,s,a)=p(c2,s,a)=p(c3,s,a)`) — global uniform price gone; creator-equality-within-type survives. Added per-type data prices (`N_a`) to "not parameters" (market outcome). Producer compensation restated as volume × `N_a`.
- **2026-05-22 — v0.4.0.** Monetary policy resolved: **`photons = seeds`** invariant (supply pegged 1:1 to corpus). Two minting streams — S to creators (participation), P to storer-validators (storage + validation). `economics.photon_issuance` removed as a free parameter (determined by the peg). Added `economics.storage_replication_factor` and `economics.block_cadence_seconds`. Data lives in wallets (wallet-indexed fabric); unified validator-storer participation spectrum.
- **2026-05-22 — v0.5.0.** `marketplace_toll` reconciled to **5%**. `access_duration_days` set to **1**. Storage rewards resolved: one-time ingestion mint (peg-preserving) + ongoing rent from circulating photons (access cuts + treasury subsidy), never new emission — added `economics.access_cut_to_storers`. Open: ingestion-photon split among storers, access-cut value.
- **2026-06-24 — v0.6.0.** Outcome magnitude `M_O` resolved (formula + 5 clarifications, see `protocol-spec.md` §Reputation Dynamics). Replaced `reputation.outcome_magnitude: null` with `reputation.outcome_magnitude_beta: 1.0` and `reputation.outcome_magnitude_cap_multiplier: 5`. Uses `S(att, t_issuance)` (not current); multiple reporters aggregate paper-shape; self-reports cred ≈ 0; outcomes are attestations of `dt.outcome.*`; outcome refutation reverses + 2× penalty.
- **2026-06-24 — v0.7.0.** Saturation point resolved (two-piece linear + log dampening). Replaced `reputation.saturation_point: null` with per-domain tiered structure: `small=5`, `standard=10`, `large=50` (each domain node tagged with a tier; v0 default `standard`). Added global `reputation.dampening_k: 5`. Per-domain from day zero (mirrors `domain.obsolescence_multiplier`), not deferred to v1+.
- **2026-07-16 — v0.8.0 (DT-21 reconciliation).** The registry catches up with the build — closing the comb's "unregistered-lever cluster" (comb item 10). Review window corrected to the **ratified √ curve** (owner, 2026-07-10, `docs/specs/x-reputation-p2-review-windows.md`; parameters.md had drifted) with the build's threshold retune (4.0, INTERIM). Registered, all **INTERIM = build-chosen, pending owner ratification and backtest sensitivity study**: `reputation.{s_max: 10, coattestor_weight: 0.25, attest_bet_scale: 0.1, lambda_endorsement: 0.08}`, new `attest.*` section (type weights incl. Use=0.5, partial-outcome 0.5, max_coattestors 64, citation_uplift_lambda 1.0), `photons.mintable_kinds`, `seeds.*` bounds (`max_batch_new_count` ratified 2026-07-15). `photons = seeds` invariant restated for the leaf model (distinct atoms; mintable-kinds caveat flagged for ratification). INTERIM rows are exactly the values `docs/specs/measurement-backtest.md`'s sensitivity curves exist to study — registration is not ratification.
