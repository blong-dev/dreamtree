# Creation-credit-forward (citation value signal)

_Shipped 2026-07-12. Status: built, devnet-proven (`scripts/citation-proof.sh`)._

## The idea

A work stands on the works it was built from. When a derivative work B succeeds,
the sources A1..A4 it was built on should share in that success — not as a
hardwired money split (that would violate "the protocol informs; the market
prices; we don't dictate compensation", protocol-spec §8), but as a **value
signal**: a source's value rises with the value of the works built on it.

This is the compensation-free half of the "pay creation forward" idea. The
inventor of B keeps full attribution and full compensation; the sources gain
reputation/value *signal*, which the market can then price however it likes.

## The edge

A `USE` attestation now carries both ends of the derivation edge:

- `subject` = the **prior** work A being built on (so A's value accrues it, via
  the existing same-subject aggregation).
- `used_by` = the **new** work B that builds on A. Directed edge `B → A`.
- USE-only (guarded: `ErrUsedByNonUse`), and `used_by != subject`
  (`ErrSelfUse` — a work cannot build on itself).

`used_by` is `Attestation` field 12 / `MsgAttest` field 8. Empty for every other
proof type; empty or unknown ⇒ no uplift.

## The uplift (read-projection only)

In `x/attest/keeper/projection.go`, work value is:

```
V_base(w)  = 1 - Π(1 - S_i/S_max)         over w's own non-outcome attestations
V(A)       = 1 - Π(1 - share_i)
             where share_i = S_i/S_max, and for a USE attestation i on A:
                 share_i *= (1 + λ · V_base(used_by_i))
```

So a citation from a maximally-valuable work counts up to `(1 + λ)×`. `λ =
citationUpliftLambda = 1.0` today (a constant; **TODO: promote to a governable
attest param** once we regen for it — gov is wired, this is a natural lever).

Key properties:

- **Signal, not money.** It lives entirely in the read-projection (float,
  off-consensus). It never mints, moves, or splits a photon. Consensus
  (`StandingOf`, `S_issuance`, settlement) is untouched.
- **Non-recursive / deterministic.** The uplift reads `V_base(B)` — the citing
  work's *intrinsic* value, computed without uplift — so the citation graph is
  only ever walked one hop. No cycles, no unbounded recursion, terminates.
- **Bounded.** `share_i` is clamped to ≤ 1 after uplift; paper-shape keeps V in
  [0, 1).

## What it does NOT do

- No transitive compensation (photons to sources). That whole layer stays parked
  — see `launch-readiness.md`.
- No multi-hop flow. A source of a source is not uplifted by the grandchild's
  success. One hop by design (determinism + gaming resistance).

## Proof

`scripts/citation-proof.sh`: two identical source works (V = 0.100). One is
built on by a high-value work B (V = 0.271), the other by a valueless work. The
first ends at V = 0.157, the second at 0.145 — the only difference is the value
of what built on them. Being built on by something successful raises the source.
