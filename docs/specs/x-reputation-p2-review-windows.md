# x/reputation P2 — review windows (design draft)

*2026-07-10. Builds on the P1 module and [`x-reputation-design.md`](./x-reputation-design.md).
Grounded in protocol-spec §Reputation Dynamics ("Review windows").*

## Decisions locked (2026-07-10, owner)

- **Fold P2 + P3** — build review windows and outcome propagation as one unit
  (they're coupled; windows exist mainly to protect outcome-driven R moves). §1
  below is superseded: no deliberately-empty window phase; outcomes feed the
  windows' corroboration/refutation from day one.
- **τ is a continuous root curve**, not tiers, not hand-rolled `ln`:
  `τ = review_window_base × ApproxSqrt(M / review_window_threshold)` (§2).
  Removes threshold-tier friction. The same move drops the P4 cred tiers
  (cred = continuous fn of rational standing, `√cred` via `ApproxSqrt`). Only
  the spec's *categorical* domain tiers remain.
- **Close by block time**, not height (§2).
- **S_issuance snapshot** = the *rational* S (`standing × spec × type_weight`,
  `LegacyDec`) frozen on the attestation at issuance — `M_O` is consensus, so it
  reads rational standing, never the float display-S.
- **cred in P3 is basic** (continuous fn of the reporter's rational standing);
  the 2-hop recursion is the only thing deferred to P4.

---

## 0. What P2 adds

Today (P1) a reputation-moving event settles **immediately**: making an
attestation writes its "bet" Contribution in the same block. P2 inserts the
review window from the spec:

> Every event enters a review window before final application:
> `τ(M) = base_window × log(1 + M/threshold)`. Trivial events (a citation): τ ≈
> instant. Large events (a fraud claim): τ ≈ weeks. During τ the event
> accumulates corroboration, refutation, context. **Final ΔR is the integrated
> signal over τ, not the raw claim** — prevents single-point character
> assassination while keeping small-stakes signal fast.

So an event becomes a **PendingEvent** with a close time; at close, EndBlock
integrates whatever accumulated and writes the settled Contribution.

---

## 1. The sequencing tension — read this first

The review window's *purpose* is to protect against big, contested R moves —
which are **outcomes** (validation / fraud claims), and those are **P3**. In the
P1/P2 world the only event is the routine attestation-bet, and nothing
corroborates or contests a routine bet. So P2, built strictly before P3, is
**infrastructure with a thin live payload**: bets flow through the window and
settle, but the accumulation slots stay empty until P3 gives them inputs.

Two honest ways to handle it:

- **Recommended — P2 = the window infrastructure, proven by deferred
  settlement.** Build the PendingEvent queue, tiered-τ, the EndBlock drain, and
  the integration formula (2× asymmetry + paper-shape) *ready for inputs*. Route
  the bet through it. It's provable and self-contained: "R now moves at the
  window's close, not at attest time, and τ scales with the magnitude tier."
  P3's outcomes then just (a) enqueue their own PendingEvents and (b) add
  corroboration/refutation to *existing* pending events — a clean surface that
  P2 hands them. Nothing built here is throwaway.
- *Alternative — fold P2 into P3.* Windows and outcomes are genuinely coupled;
  building them together is defensible. But you split them deliberately, and the
  infra-first path de-risks the EndBlock/queue machinery separately from the
  outcome semantics (a smaller, more reviewable change). I'd keep them split.

The rest of this doc assumes the recommended path.

---

## 2. τ(M): a continuous root curve — no tiers, no log (DECIDED 2026-07-10)

τ sets a PendingEvent's close time — **consensus scheduling**, so it must be
deterministic. `τ = base·log(1 + M/threshold)` needs a logarithm, and
`cosmossdk.io/math.LegacyDec` has no `Log`/`Ln` — a float `log` would be a
transcendental in a state transition, forbidden.

But the constraint is "no *transcendental*," not "no continuous function." A
continuous **root** curve clears the bar and avoids threshold-tier friction:

```
τ = review_window_base × ApproxSqrt(M / review_window_threshold)     # seconds = ×86400 (base is days)
```

- `LegacyDec.ApproxSqrt` is deterministic fixed-point (Newton iteration — the
  same primitive AMMs use in consensus). No transcendental, no tiers.
- `sqrt` dampens large events the way `log`'s intent wants (M=1000 → ~31×base,
  not an absurd 1000×base) without a cap; `threshold` tunes how fast trivial
  events approach instant.
- One curve, two existing params (`review_window_base`, `review_window_threshold`),
  no "just-below/just-above" discontinuity.

**Why not tiers (owner call):** threshold tiers bucket a continuous value and
add cliff friction. The continuous curve removes it. The *same* move removes the
cred tiers proposed for P4 (cred = continuous fn of rational standing, `√cred`
via `ApproxSqrt`). The only tiers that remain are the spec's **categorical**
domain saturation/obsolescence labels — governance assigns a domain a label;
nothing continuous is bucketed, so no cliff. Net: the friction-causing tiers are
gone.

**Close by block *time*, not height.** τ is a duration; store
`close_time = opened_at + τ_seconds` and mature when `block_time ≥ close_time`.
This survives variable block cadence (height-based windows drift if block time
changes). base is in days → seconds at enqueue.

---

## 3. State model (consensus)

```
PendingEvent {
  id             uint64          // global sequence (own, or reuse contribution seq namespace)
  signer         string          // whose R moves at settlement
  domain         string
  base_magnitude string          // LegacyDec — the event's own magnitude (the bet, or a P3 outcome's M_O)
  rate_bucket    RateBucket
  source_att_id  uint64
  opened_at      int64           // block time enqueued
  close_time     int64           // opened_at + τ_seconds
  // accumulators — filled by corroboration/refutation during the window (P3):
  corroboration  string          // LegacyDec, paper-shape running aggregate
  refutation     string          // LegacyDec, paper-shape running aggregate
}
// indexes:
//   by close_time:  KeySet Pair[int64 close_time, uint64 id]  → EndBlock drains matured
//   by signer:      KeySet Pair[string, uint64]               → the PendingEvents query
```

Genesis gains `pending_events` (export/import) so an exported chain restores
in-flight windows exactly.

---

## 4. Flow

```
Attest (msg handler)
  └─ rep.OnAttestation → ENQUEUE a PendingEvent   (was: write a Contribution)
        base_magnitude = attest_bet_scale × specificity   (rational, as in P1)
        close_time     = block_time + τ_tier(base_magnitude)

EndBlock (new module EndBlocker)
  └─ drain PendingEvents with close_time ≤ block_time, in id order:
        final = integrate(base_magnitude, corroboration, refutation)
        addContribution(signer, domain, final, rate_bucket, source_att_id)   // the P1 settlement path
        delete the PendingEvent + its index entries
```

`addContribution` is exactly P1's settlement — so the read-projection R is
unchanged; P2 only changes *when* a Contribution appears (at close, not at
attest). Bets with `M < t_trivial` get τ=0 → settle the very next EndBlock, so
in production they're effectively immediate — matching "trivial ≈ instant."

The module becomes an `appmodule.HasEndBlocker`; `reputation` is added to
`end_blockers` in app.yaml.

---

## 5. The integration formula (built now, inputs in P3)

```
integrate(base, corrob, refut):
    signal = base + corrob − neg_asymmetry × refut      # 2× asymmetry (parameters.md)
    return clamp(signal, 0, M_cap)                        # no negative settle in P2; P3 adds reversal
```

Corroboration/refutation aggregate **paper-shape** as they arrive (P3):
`agg' = cap·[1 − (1 − agg/cap)(1 − x/cap)]` — diminishing returns, Sybil-resistant,
all `LegacyDec`. In P2 both accumulators are 0, so `final = base` — but the
formula and the 2× asymmetry ship now, unit-tested, so P3 only has to *feed* it.

---

## 6. R-read semantics

R still sums **settled Contributions only** — a PendingEvent does not move R
while in flight. Correct per spec ("final ΔR is the integrated signal over τ").
Consequence: a signer's R reflects only what has cleared review. The
PendingEvents query (below) surfaces what's in flight so it's observable.

---

## 7. What P2 demonstrably delivers

- New query `PendingEvents(signer?)` → in-flight windows with `close_time`.
- Devnet demo: attest → a PendingEvent appears (R unchanged); advance past
  `close_time` → EndBlock settles it → R moves *then*. Set a larger
  `review_window_base` on devnet to make the delay visibly multi-block; note
  production params make trivial bets ~instant.
- τ scales with the magnitude tier (a high-specificity bet waits longer than a
  low one), shown by two attests with different specificity.

---

## 8. Determinism ledger — P2 delta

| Concern | Where | Arithmetic |
|---|---|---|
| τ tier selection | enqueue (consensus) | integer compare on `LegacyDec` |
| close_time | state | int64 |
| EndBlock drain order | EndBlock (consensus) | by ascending id (deterministic) |
| integration (asymmetry, paper-shape, clamp) | EndBlock (consensus) | `LegacyDec` — no transcendental |
| everything read-side (R decay/saturation) | queries | float (unchanged) |

Still holds: **no transcendental in any state transition.** τ's log is replaced
by tiers precisely to keep that true.

---

## 9. Open questions for you

1. **Route the P1 bet through the window?** → R moves at close, not at attest.
   Trivial τ makes it ~1 block in prod; devnet cranks `base` to show it. (I think
   yes — it's the only way to exercise the machinery pre-P3.)
2. **Tiered-τ vs a hand-rolled fixed-point `ln`?** Given no `LegacyDec.Log`, I
   recommend tiers. Confirm, or ask for the true curve.
3. **Defer all corroboration/refutation inputs to P3** (outcomes are the natural
   source), building only the empty accumulators + integration now? (Lean yes.)
4. **Close by block time (recommended) vs height?** Time survives cadence
   changes; height is simpler to reason about on a fixed-cadence devnet.
5. **Cap in P2.** `M_cap` is an outcome concept (P3). For P2's clamp, use a large
   passthrough cap (bets never approach it) and let P3 introduce the real
   `M_cap`? (Lean yes.)
