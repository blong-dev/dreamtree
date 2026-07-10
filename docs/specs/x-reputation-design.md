# x/reputation — design draft

*Draft 2026-07-10. Replaces the `R = baseline_kyc` stub in `x/attest` with real
`R(j,k,t)`. Grounded in protocol-spec §Reputation Dynamics and parameters.md.*

## Decisions locked (2026-07-10, owner)

- **Build P1 alone first** (contributions + decay + saturation + the seam;
  immediate settlement, no review windows).
- **Hybrid confirmed** (§1): consensus stores integrated contributions; the
  decay/saturation shaping is a read-time float projection.
- **Snapshot `S(att, t_issuance)`** on each attestation — a need. In P1 this is
  rational (R = baseline constant); in later phases it uses the *rational
  standing* view, never the float R (see next bullet).
- **Endorsements reuse the `x/attest` log** later as `proof_type = ENDORSEMENT`,
  not a separate object.
- **cred / float resolution (the load-bearing rule): consensus never reads the
  float R.** The contribution log has two views: a **read view** (float —
  `exp` decay + `log` saturation, for S/V/display) and a **consensus view**
  (`Σ Mᵢ·relevance`, a `LegacyDec` sum — no transcendentals). cred = integer
  tiers on the rational sum; `√cred` via `LegacyDec.ApproxSqrt`. Correct because
  settlement fires within days–weeks while decay half-lives are years, so
  undecayed standing ≈ decayed R at settlement. **No transcendental ever runs in
  a state transition.** This is a P3/P4 concern; P1 is unaffected (its
  magnitudes are already rational).

---

## 0. What exists, what's missing

`x/attest` computes attestation strength `S = R × specificity × type_weight ×
decay × (1-refuted)` — but `R(signer,domain)` is a flat `baseline_kyc = 1.0`.
Everything downstream (S, work-value V) is real; only R is stubbed. This module
makes R the accumulated, decayed, domain-indexed thing the spec defines:

> R(j,k,t) — signer j's weight in domain k at time t. Load-bearing; everything
> else derives from it.

The spec hands us a lot: continuous decay toward baseline, 2× negative
asymmetry, review windows τ(M), cred-recursion (2 hops), endorsement inheritance
(25% geometric), outcome propagation with reversal, log-dampened saturation,
cold-start. Not all of it should land at once — see the phased plan (§7).

---

## 1. The central decision: what is consensus state, what is a projection?

This is the one to get right; everything else follows from it.

`x/attest` set the precedent: **store the log, compute value at read time as a
float projection** — because S and V touch only a bounded set (attestations on
one subject) and never gate a state transition, so they needn't be
deterministic across validators.

R is different in two ways:
1. **Reads fan out.** A single work-value query needs S for every attestation on
   a work — each by a *different* signer — so it needs *many* signers' R. If R
   were a pure from-scratch replay of each signer's whole history, one query
   could trigger unbounded work.
2. **R is path-dependent over time.** Review windows mean an event's effect isn't
   known at submission — it integrates corroboration/refutation over τ, then
   settles. That's inherently stateful and time-ordered.

But data-model.md is also clear: *value is a derived, reprocessable projection,
never stored.* So a naive "store R as a mutable number" violates the model and
buries the exponential-decay math inside consensus (forcing deterministic
fixed-point `exp`/`log` — painful and error-prone).

### Recommendation — the hybrid (store integrated *contributions*; project the shaping)

Split R into two layers:

- **Consensus state = an append-only log of settled *contributions*.** When a
  reputation-moving event finishes its review window, EndBlock writes one
  immutable record: `(signer j, domain k, magnitude M, rate_bucket, t_settled,
  source)`. Magnitude M is a sum/aggregation of stake-weighted votes — **rational
  arithmetic only** (sums, products of `(1 - x/cap)` for paper-shape). No
  transcendental functions in consensus. This log *is* part of the app state
  (deterministic, in the app hash), but it only ever records integrated
  magnitudes — never an exponential.

- **Read projection (float, non-consensus) = the shaping.**
  ```
  R(j,k,t) = effective_R(
      baseline_kyc
    + Σ_contributions M_i · relevance(k, k_i) · exp(-λ_bucket(k_i) · (t - t_settled_i))
  )
  ```
  Decay (`exp`), saturation (`effective_R` log-dampening), and taxonomy
  `relevance` all live here, in float, exactly like `x/attest`'s S/V. Bounded
  because we iterate a signer's *settled contributions* (paginable, cappable) —
  not a raw replay of every vote.

Why this satisfies both constraints:
- **Faithful to data-model:** R is still a projection — a pure deterministic
  function of the contributions log, rebuildable from genesis. The contributions
  log is itself a projection of the attestation/outcome log (produced
  deterministically in EndBlock). Nothing "stores R."
- **No deterministic-exp problem:** consensus only ever does rational magnitude
  math; the exponentials are read-side float.
- **Bounded reads:** contributions are pre-integrated, so a read is a sum over a
  signer's contribution list, not a re-run of review windows.

**This is the decision I most want your sign-off on.** The alternatives:
- *Pure projection (no stored contributions):* recompute R by replaying the raw
  attestation+outcome log per read. Cleanest data-model story, but reads are
  unbounded and review-window integration on every read is wasteful. Rejected
  for performance.
- *Materialized mutable R (store one number per (j,k), lazily decayed):* O(1)
  reads, but forces deterministic fixed-point `exp` in consensus and reads as a
  stored value, not a projection. Rejected for complexity + model friction.

---

## 2. State model (consensus)

```
Contribution {
  id            uint64        // global sequence
  signer        string        // j (address)
  domain        string        // k — 5-level taxonomy path
  magnitude     string        // M, signed fixed-point decimal (LegacyDec); may be negative
  rate_bucket   RateBucket    // which λ decays this term at read time
  t_settled     int64         // when the review window closed (decay clock start)
  source_att_id uint64        // the attestation/outcome that produced it (provenance)
}
RateBucket = { PERMANENT, DURABLE_25Y, RIGOR, USE, REPLICATION, ENDORSEMENT }

// indexes: by (signer, domain) for the read projection; by source_att_id for reversal.

PendingEvent {                // an event inside its review window τ(M)
  id            uint64
  source_att_id uint64
  close_height  int64         // block at which τ elapses → settle in EndBlock
  // accumulators (rational): corroboration / refutation stake gathered during τ
  ...
}
// index: by close_height, so EndBlock processes only what matures this block.

DomainConfig {                // governance registry; default when absent
  path                string  // taxonomy node
  saturation_tier     Tier    // small(5) | standard(10) | large(50)
  obsolescence_tier   Tier    // foundational(0.3) | standard(1.0) | frontier(3.0)
}
Params { baseline_kyc, neg_asymmetry, dampening_k, outcome_beta, outcome_cap_mult,
         λ per rate_bucket, base_review_window, review_threshold, endorse_inherit,
         cred_hops=2, saturation tiers, obsolescence tiers }   // all from parameters.md
```

`magnitude` uses `math.LegacyDec` (fixed-point) because it's consensus state and
the settlement math must be identical across validators. The *decay* applied to
it is float, at read time only.

---

## 3. The read projection (float, non-consensus)

Mirrors `x/attest/keeper/projection.go`:

```
reputationOf(j, k, now):
    sum = 0
    for c in contributions[j] (any domain k_i):          # paginate/cap
        sum += toFloat(c.magnitude)
             * relevance(k, c.domain)                     # taxonomy attenuation
             * exp(-λ(c.rate_bucket, k) * years(now - c.t_settled))
    raw = baseline_kyc + sum
    return effectiveR(raw, saturationTier(k))             # log-dampening past S

relevance(k, ki): longest-common-prefix depth d of the two paths →
    d=5 →1.0, 4 →0.70, 3 →0.40, 2 →0.15, 1 →0.03, 0 →0    # spec table

effectiveR(R, S): R if R ≤ S else S + k_damp·log(1 + (R−S)/S)
```

`x/attest` gets a new query `Reputation(signer, domain)` and its S/V projection
calls `reputationOf` instead of the baseline constant (see §6).

---

## 4. EndBlock settlement (consensus, rational math)

Per block, process `PendingEvent`s whose `close_height == now`:

1. **Integrate the window.** Final signal = accumulated corroboration −
   `neg_asymmetry` × accumulated refutation (2× asymmetry). Paper-shape multiple
   reporters: `M = M_cap·[1 − Π(1 − M_i/M_cap)]` — products of rationals, no exp.
2. **cred(source) recursion (2 hops).** Weight each reporter by cred: unverified
   public → 0; KYC-unattested → baseline; verified-with-history → full R; their
   endorsers contribute one further hop, then terminate. Bounded 2-hop walk over
   the endorsement graph. (cred uses R, which is float — but at settlement we can
   use a *snapshot* R read; see open question Q3.)
3. **Emit Contributions.** For the contributor and — for outcomes — the
   propagation chain (direct attestors × attestation_weight, institution × 25%
   chain hop, hiring party × factor). Each becomes a `Contribution` row with the
   right `rate_bucket` and `t_settled = now`.
4. **Reversal.** If this is a counter-outcome refuting an earlier outcome: write
   negation contributions un-applying the original chain, and a 2×-penalty
   contribution against the original reporter. (Provenance via `source_att_id`.)

`M_O = min(M_cap, β · S(att, t_issuance) · √cred(reporter))` uses S *at issuance*
— so we snapshot S when the attestation is made (store it, or recompute from the
attestation's frozen inputs + the R-snapshot at issuance). See Q2.

---

## 5. Domain taxonomy

- **Domain = path string** (`class/field/discipline/specialty/subspecialty`).
  `relevance` is a pure function of two paths (common-prefix depth) — no registry.
- **Tiers** (saturation, obsolescence) come from `DomainConfig`, a
  governance-set map; **v0 defaults everything to `standard`**, overrides added
  by governance. Seeds from LCC+ISCED+ONET are a later data task, not blocking.
- No taxonomy *validation* at write time in v0 — any path is accepted; unknown
  nodes just get default tiers. (Tightening is governance.)

---

## 6. The module seam (no import cycle)

- **`x/attest` defines an expected interface** it reads R through:
  ```go
  type ReputationSource interface {
      ReputationOf(ctx, signer, domain string) float64
  }
  ```
  Its S/V projection calls this instead of the baseline constant. Wired via
  depinject; if unset (reputation module absent), it falls back to `baseline_kyc`
  so `x/attest` still runs standalone.
- **`x/reputation` imports `x/attest`** (reads attestations/outcomes to build
  PendingEvents and settle Contributions) and *provides* `ReputationSource`.
  One-directional dependency: reputation → attest. No cycle.
- **Trigger:** when `x/attest` records an attestation that moves reputation
  (an OUTCOME, or a validated/endorsement event), it emits a typed event /
  hook that `x/reputation` picks up (EndBlock scan of new attestations, or a
  keeper hook). Prefer an explicit keeper hook over event-scraping.

---

## 7. Phased build (each phase compiles, runs, is provable on devnet)

- **P1 — contributions + decay + saturation (the spine).** Contribution log,
  the read projection (decay toward baseline + log-dampened saturation +
  taxonomy relevance), the `x/attest` seam. R moves when contributions exist;
  no review windows yet (events settle immediately). Delivers: real
  domain-indexed R feeding S/V. *This is the module's backbone and probably the
  right first PR.*
- **P2 — review windows.** PendingEvent queue, EndBlock settlement, 2× negative
  asymmetry, paper-shape aggregation. Delivers: events integrate over τ before
  hitting R; single-point assassination resistance.
- **P3 — outcome propagation.** M_O formula, the propagation chain (contributor
  + attestors + chain hop + hiring party), reversal + 2× penalty. Delivers: the
  appreciation-compounding mechanic — validated bets pay out up the chain.
- **P4 — cred recursion + endorsement inheritance.** 2-hop cred weighting,
  endorsement objects, geometric inheritance. Delivers: shell-institution
  resistance, bootstrap-by-piggyback.

Cold-start ramp (spec-open) and the taxonomy seed data ride alongside as their
specifics settle.

---

## 8. Determinism ledger (the discipline that keeps this sound)

| Concern | Where | Arithmetic |
|---|---|---|
| Attestation log | `x/attest` state | exact (strings/ints) |
| Contributions log | `x/reputation` state | `LegacyDec` (fixed-point) |
| Review-window integration, M_O, paper-shape | EndBlock (consensus) | `LegacyDec` — sums, products of rationals, **no exp/log** |
| Decay `exp(-λΔt)`, saturation `log`, relevance | read projection | `float64`, non-consensus |
| S, V, R values returned | queries | `float64` strings |

Rule: **no transcendental function ever runs in a state transition.** If a phase
tempts me to put `exp` in EndBlock, that's the signal the split is wrong.

---

## 9. Open questions for you

1. **Sign-off on the hybrid (§1)** — store integrated contributions, project the
   decay/saturation. This is the load-bearing call.
2. **`S(att, t_issuance)` snapshotting (§4).** M_O needs an attestation's
   strength *at issuance*. Store a frozen `S_issuance` on each attestation when
   made (needs R-at-issuance, so this co. coincides with P1), or recompute from
   frozen inputs? Storing is simpler and matches "use issuance strength, not
   decayed." Lean: store it.
3. **cred snapshot vs live (§4).** cred(reporter) feeds settlement magnitude
   (consensus). But cred derives from R (float projection). Feeding a float into
   consensus magnitude breaks determinism. Options: (a) quantize cred to a small
   fixed set of tiers (unverified=0 / baseline=1 / established=2 …) decided by
   integer thresholds on a *snapshotted* raw-contribution sum — keeps it rational;
   (b) store an R-snapshot as `LegacyDec` at settlement. Lean (a): cred as
   discrete tiers, so consensus never touches float. **This is a real subtlety —
   worth a look.**
4. **Phasing.** Is P1 (contributions + decay + saturation, immediate settlement)
   the right first shippable module, with review windows/propagation/cred as
   follow-ons? Or do you want P1+P2 together (windows from the start)?
5. **Endorsements as first-class attestations or a separate object?** The spec
   treats outcomes as attestations; endorsements could be
   `proof_type = ENDORSEMENT` in `x/attest` (reuses the log + indexes) rather
   than a new object. Lean: reuse `x/attest`.

---

## 10. Faithful-to-spec checklist (where each mechanic lands)

| Spec mechanic | Lands in | Phase |
|---|---|---|
| Domain-indexed R(j,k,t) | contributions + read projection | P1 |
| Decay toward baseline (not zero) | read projection | P1 |
| λ_R = base/(1+validated_volume) | rate_bucket + volume-scaled λ | P1/P2 |
| Log-dampened saturation, per-domain tiers | read projection + DomainConfig | P1 |
| Taxonomy relevance attenuation | read projection (prefix depth) | P1 |
| Durable validated R (25y) vs proof-type decay | rate_bucket assignment | P1/P3 |
| Review windows τ(M) | PendingEvent + EndBlock | P2 |
| 2× negative asymmetry | settlement | P2 |
| Paper-shape aggregation (M_O, votes) | settlement (rational) | P2/P3 |
| Outcome propagation chain | settlement | P3 |
| M_O = min(cap, β·S_issuance·√cred) | settlement + S snapshot | P3 |
| Reversal + 2× penalty | settlement (negation contributions) | P3 |
| cred(source) 2-hop recursion | settlement (tiered) | P4 |
| Endorsement inheritance 25% geometric | endorsement contributions | P4 |
| No reputation floor / start-fresh-in-new-domain | baseline in projection; domain-keyed | P1 |
| Shell-institution resistance (meta-attestations) | cred + endorsement | P4 |
| Cold-start ramp | settlement multiplier | open |
