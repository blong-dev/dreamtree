# DreamTree Data Model

- **Status:** Working design — *not a finished spec.* Layers 1–2 settled (the atom + the
  thing); Layer 3 (resolution) settled through `decay` — the `read` step is open; layers 4–6
  sketched. We derive one layer at a time, getting each right before the next.
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

**C vs S — settled (2026-06-15).** Resolving the division collapsed the whole atom to one rule.

- **C is a pointer, not an identity.** At observation time C is an unresolved *referent-handle* —
  whatever A pointed at (a name, an external id, an ambiguous string). Turning many handles into
  one thing is Layer 2/3, derived and reprocessable. Putting resolved identity on the atom would
  smuggle Layer 3 into Layer 1 (the regress) and make "same C" a *stored fact* instead of a
  reading. Handle now; identity derived.

- **One C per atom; an event/relationship is a referent in its own right.** A transaction has no
  intrinsic "subject" — picking a party as C imports an angle reality doesn't have. So the *event*
  is its own C, and each party's involvement is a separate single-C observation **predicated on the
  party** (so it accretes onto the party's ball, where identity needs it):

  ```
  event:          (C=→event#123, A=→Reuters, T, S={ →kind:→donation, →magnitude:[$1M], →when:[2025-01-20] }, σ)
  participation:  (C=→Lacy,      A=→Reuters, T, S={ →role:→donor,     →in:→event#123 }, σ)
  participation:  (C=→fund,      A=→Reuters, T, S={ →role:→recipient, →in:→event#123 }, σ)
  ```

  The Lacy↔fund relationship is *derived* from co-reference of `→event#123` — symmetric, n-ary,
  direction recoverable but never privileged. The edge is never stored.

- **Everything is a C or a literal — one law.** The subject is a →C; the attestor **A is a →C**
  (an entity like any other — the recursion is real); every value in S is a literal or a →C.
  **Types, roles, predicates** (`kind`, `role`, `in`, `donor`, `donation`) are *all* Cs — grounded
  if known, provisional if not. There is no special "structural vocabulary"; common predicates are
  just Cs with high grounding (gravity). Nothing is enumerated; everything grounds the same way.

- **The grammar of S (the Layer-1 lock):** `S = { →C : value }`, where `value` is a **literal** or
  a **→C**. A value is a literal or a pointer to a C — period.

- **Raw vs derived (follows from the axiom).** *C = what the attestation is about.* Record A's
  attestation at its natural C and never reshape it — a donation *report* is about the event, so its
  raw atom is `C=event`; a source about Lacy herself is natively `C=Lacy`. Participations are
  **derived projections** (the log is the only ground truth; everything else re-derives); a
  participation is a *raw* atom only when the source genuinely asserts about the party.

## Layer 2 — The referent / thing  *(settled 2026-06-16)*

A thing is its observations — but with a **spine**, not a loose cloud. "Just a bag of
observations" is too loose; solidity is real, it's *earned*, and it has a precise home.

**Structure of a thing:**
- **anchor** — a minted, immutable, *contentless* internal id. A permanent address that
  observations resolve to; it carries no claims (no name, no FK). *This is the solid part:
  append-only, it never moves.*
- **bag** — the observations resolved to the anchor. The body.
- **grounding** — the strongest authoritative proof in the bag (CIK / FEC id / QID / KYC). It
  rides *on top* as observations — never the spine — and it *sets the solidity level*.
- **reading** — the materialized current-best entity (name, attributes), a re-derivable
  projection. The solid-*feeling* surface apps query.

There is no privileged *meaning*-kernel (FKs stay observations, per Layer 1) — but there **is** a
solid *locus*. The anchor is a minted internal id, **not** the strongest FK — so a thing can
exist *before* it's grounded, survive wrong / contested / multiple FKs, and float as provisional
substrate until it earns a proof.

**Solidity is a gradient you earn, not a thing you assume:**
- ungrounded bag (loose mentions) → genuinely provisional, floating — **we refuse to fake
  solidity here** (blank > wrong);
- anchored to an authority → **solid as its source**;
- a KYC-verified human → maximal.

The real-world referent is the solid identity (*human-is-the-identity*); our data is its
reflection. We never claim more solidity than we've grounded.

**Birth, union, death:**
- A thing is **born** by minting an anchor (on a referent not yet seen).
- Two anchors become **one thing** not by merging but by a **`sameAs` observation between them** —
  resolution is itself attestation, with its own author, confidence, and decay:
  ```
  (C=→anchor#7, A=→resolver, T, S={ →sameAs:→anchor#12 }, σ)
  ```
- A **thing = the connected component of anchors under `sameAs`**, read on demand.
- **Merges never destroy.** A bad merge is outvoted or decays — attest `differentFrom`, the
  component re-reads, every original anchor still intact. **Splits are free. Reprocessing is
  free.** Destructive merges are irreversible and are how KGs rot; we never do them.

**The hard / soft split (what makes it not-loose):** the **anchor, the log, and the proof are
hard** (permanent, immutable). Only the **grouping flexes** (confidence-weighted `sameAs`).
Uncertainty is quarantined to the one thing that genuinely *is* uncertain — "are these two
anchors the same referent." Everywhere else is solid.

**Deferred to Layer 3:** *how* `sameAs` is decided — when it's proposed, how its confidence is
computed, how it decays, and mint-vs-reuse on ingest. Layer 2 only fixes *what a thing is*.

## Layer 3 — Identity / resolution (the Inhabited Library)  *(settled through `decay`, 2026-06-16)*

The full resolution lifecycle: **mint → propose → score → decay → read.** All of it is *derived*
over the log — resolution is a reading, never a mutation. The first four are settled; `read` is open.

**mint — what anchors exist.** Mint-vs-reuse is a property of the handle's *namespace*, not a
judgment call:
- **strong handle** — from a uniqueness-guaranteeing authority (FEC id, SEC CIK, QID, DOI, *or our
  own minted ids*) → **content-addressed:** same handle = same anchor, deterministically, forever.
  We *inherit* the uniqueness; we don't assert it.
- **weak handle** — no uniqueness guarantee (a bare name) → **mint a fresh provisional anchor every
  occurrence.** Never presume two weak handles are the same thing at write.

Ingest stays dumb and deterministic — *zero fuzzy decisions at write* — and all "is this the same?"
mess is quarantined into `sameAs`. Strong handles → a few solid content-addressed anchors (the
grounding); weak handles → a cloud of cheap provisional anchors that resolution pulls together (the
substrate). Matches what gnosis converged on: FK → deterministic, FK-less → scored.

**propose — which pairs to consider.** Candidate generation is just *reading the log*: index anchors
by shared features (name tokens, claimed ids, neighbors); any two in a feature-block are candidates.
Recall-oriented, reprocessable, heuristics-OK (it's a reading, not data). Two signal classes, unequal:
- **resemblance** — they *look* alike. Cheap, high-recall, low-trust (two John Smiths). Only ever *proposes.*
- **bridge** — a *shared grounded reference* ties them: same strong id reached two ways, a shared
  *unique* neighbor, an authority **crosswalk**. Evidence of shared *identity* — nearly self-scoring.

Privilege bridges over resemblance (ground in authority; don't trust looks). **Crosswalks aren't
candidates** — an authority's `FEC:C001 = QID:Q123` enters as a *settled* `sameAs` via normal ingest.
Propose/score exists only for pairs no authority has linked.

**score — how sure.** We don't build a `sameAs` scorer; a `sameAs` is a claim, scored by the model's
one law:
```
confidence = 1 − ∏(1 − sₐ · specificity · decay)   over sameAs attestations, net of differentFrom
```
- **standing `sₐ`** = authority vs guesser (the system resolver is just a low-standing attestor — A is a C).
- **specificity** = step 2's bridge-vs-resemblance, *quantified* (shared "chairs fund X" → big; shared
  "lives in the US" → ~0). Not a separate mechanism — just where evidence sits on the axis.
- **decay** = recency.

Resolution self-improves: the resolver's standing is learned from whether its past `sameAs` survived
(outcome feedback → Layer 5); re-aggregate → links sharpen; the log never moves. **No special veto:**
an authoritative `differentFrom` (two distinct FEC ids) simply carries overwhelming negative weight in
the same aggregation and dominates. The veto falls out of the formula.

**decay — the Inhabited Library.** Decay isn't an operation; it's the *time-dependence of the reading*
(taken as-of `now`). Three time-dependent things, kept separate:
1. **claim authority `= sₐ(T)`** — standing *at attestation time*, **frozen forever.** A 1970 expert's
   1970 claim is weighted by their 1970 standing, permanently; the past is never retroactively rewritten.
2. **claim present-relevance `= decay(now − T)`** — a separate axis. The 1970 claim keeps full authority,
   but its contribution to *"what's true now"* ages.
3. **standing `sₐ(t)`** — a time-series that rises on outcomes and **drifts toward a neutral baseline**
   without renewal (decision: neutral, not zero — redemption real, no permanent crown *or* damnation;
   anti-capture). Death = renewal ceases; *go-forward* standing drifts neutral, but the 1970 attestation
   is untouched. This is why "the same claim made at a different time carries different weight" — a later
   attestation uses `sₐ(later)`.

So historical value `= sₐ(T) · specificity` is **fixed**; present contribution `= · decay(now − T)`.
**Decay rate is a grounded property of the predicate** (`birthdate` ≈ none; `job` / `price` ≈ fast), and
composes with explicit **validity intervals** (CEO 2017–2023): the interval governs truth, decay handles
open-ended claims. A thing **"stays itself" by continued attestation**; **forgetting is a vantage point,
never a deletion** — read as-of any time and the old reading is fully intact. It all stays timeless
underneath: `sₐ(t)` is a derived reading, so a better outcome-model re-derives every `sₐ(T)` and re-reads.

**read — OPEN.** The thing = the `sameAs`-connected-component under a confidence cut — the one place in
the whole model a threshold is allowed (a downstream reading, never in the data). *Next.*

## Layers ahead — sketched, not defined
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

- ~~Confirm the **C vs S** division.~~ **Settled 2026-06-15** (Layer 1).
- ~~Derive **layer 2** (the referent / thing).~~ **Settled 2026-06-16** (Layer 2).
- **Layer 3** — identity / resolution: `mint → propose → score → decay` **settled 2026-06-16**;
  the `read` step (the confidence-cut connected component) is **next**.
- Then layer 4 (relationships & events), 5 (dynamics), 6 (value).
