# DreamTree Data Model

- **Status:** Working design — *not a finished spec.* Layer 1 (the atom) is settled and
  locked to the paper; layers 2–5 are sketched, not defined. We derive one layer at a time,
  getting each right before the next.
- **Date:** 2026-06-05
- **Owner:** Braedon

## Why this exists (and why it comes first)

The data model is **upstream of the wallet and the chain.** The wallet is a *view* of the
data; the chain is an *anchoring* of it. Both are serializations — consequences, not
decisions. You can't settle what the wallet holds or what the chain commits until you know
what a *thing*, a *claim*, and an *identity* truly are. *Persistence is upstream of
everything.* Define the truest model first; the wallet and chain then fall out.

This is the foundation the gnosis KG was empirically surveying (FK-as-proof, gravity, events).
See `gnosis: docs/specs/extend-the-wiki.md` for that survey.

## Method (how we derive, so we don't go into hell)

1. **No pure logic from nothing.** Foundationalism walks into infinite regress (an entity is
   defined by observations; an observation by an entity…) before it finds ground. The escape
   is the **shoulders of giants**: the legitimate real sources (registries, filings, Wikidata)
   are the *base observers* that break the regress. We shape each layer abstractly *just
   enough*, then anchor it in real sources — we don't keep deriving.
2. **Specifics, not thresholds.** At the foundational layers nothing is gated. Legitimacy is a
   *measured specific*, not a bar to clear. Thresholds, if they ever appear, are a *reading*
   of the log far downstream — never a property of the data.
3. **The log is the only ground truth.** Everything else — things, edges, identity, gravity,
   value — is a **derived, recomputable projection** over the observation log. An entity isn't
   a stored fact; it's a current *reading*. Better resolver / better gravity function →
   re-derive; the log never changes. **Reprocessing is structural and free.**

## Layer 1 — The atom: the observation  *(settled, locked to the paper)*

**Axiom:** in the world of data, *it doesn't exist if it isn't observed.* The observation is
the genesis of a datum.

The atom is exactly the attestation tuple from the attribution paper
(`philosphy/04-attribution-mechanism.md` §4.6), in its symbols, unchanged:

> **Observation = (C, A, T, S, σ)**
> - **C** — the referent: the event / effect / thing-that-happened being recorded.
> - **A** — the attestor: the entity doing the observing (itself an entity in the graph —
>   the model is honestly recursive; the observer is the observed next time).
> - **T** — observation time (the genesis).
> - **S** — the statement asserted about C.
> - **σ** — the verifiable proof / provenance (the source record + archive that makes
>   "A asserted S" non-repudiable).

**Properties:**
- **Immutable.** Reality doesn't un-happen. You never edit an observation — only *supersede*
  it with a new one. Append-only underneath; everything above is a view.
- **Provenance-complete by birth.** An observation can't exist without an attestor (A) and a
  proof (σ). No orphan datum, ever.
- **Dual-timed.** T is the *observation* time. The referent's *own* time (when C occurred) is
  a property of **C**, not the atom. Keep them separate or every timeline lies.

**Derived, never stored on the atom:**
- **Legitimacy** `s_{A,k}` — the attestor A's standing for claim-type k. Defined *separately*
  by the paper (§4.3): `weight = s_{A,k} × specificity × decay(recency)`; aggregation
  `A_total = 1 − ∏(1 − weight)`. Domain-indexed (k), recency-decayed. The
  observation→attestation gradient is precisely a gradient of *legitimacy*, and the theory
  already measures it — we store **A** (who); legitimacy is computed from A's standing, never
  frozen onto the observation (standing changes and decays; freezing it would lie).
- **C's referent-time** — a property of the referent, derived/asserted, not an atom field.

**C vs S (the interface that matters at this layer):** C identifies *which* thing happened
(the noun — it points at the referent); S is *what is asserted* about it (the claim). One C
accretes many S from many A — which is the sticky ball beginning to form.  *[Division to
confirm before climbing to layer 2.]*

## Layers ahead — sketched, not defined

2. **Referent / thing.** A thing is the accreted, maintained mass of observations about a C —
   the "sticky ball." Identity *is* the entity.
3. **Identity.** How the ball stays itself: the **Inhabited Library**, not the Archive —
   maintenance, decay (`λ_standing`), outcome-detachment. Real-world referent is the identity;
   FKs/QIDs are *attachments* to it, replaceable, not the thing.
4. **Relationships & events.** Observations whose referent is a relationship or event.
   *Involvement* (observed) vs *effect* (asserted + graded + data-proven). Time lives on events.
5. **Dynamics.** Standing / legitimacy / decay / outcome-propagation. Gravity is a *derived
   reading* — and the only place a threshold could ever live (downstream, never in the data).
6. **Value.** Attribution / settlement — the DreamTree economy — on top of the clean
   observation substrate.

## Grounding (sources this is derived from)

- The atom + legitimacy: `philosphy/04-attribution-mechanism.md` (§4.6 the tuple; §4.3 standing).
- Identity / sticky-ball / Inhabited-Library: the standing–decay essays
  (`v3/philosophy/Cosmo Writings/`) + the DreamTree protocol (human-is-the-identity;
  reputation-as-a-stock with outcome feedback).
- Empirical survey: gnosis KG (`gnosis: docs/specs/extend-the-wiki.md`).

## Open / next

- Confirm the **C vs S** division.
- Then derive **layer 2** (the referent / thing) — one step.
