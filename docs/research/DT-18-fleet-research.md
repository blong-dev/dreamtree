# DT-18 — Research: Conform built chain to protocol-spec.md + Attribution as Conservation (seed=atom, photon-native, dtvp retirement)

**Card:** DT-18 (dreamtree board, research lane, p9)
**Researcher:** Gnosis Researcher agent, sandboxed container (no m3 / no chain-source access — see §5)
**Date:** 2026-07-15

---

## 1. Problem statement

The epic asks for six staged changes (S0–S5) to conform the live `dreamtree-1` chain to
its design of record: seed=atom as the priced unit, batch-commit as a pure commitment
strategy (never a unit change), photon as the chain's native/staking asset, dtvp retired
as redundant staking expedience, and one clean wipe to chain-id `dreamtree`. My job in the
research lane is not to write the spec — it's to ground the plan against whatever primary
sources I can actually reach, find the concrete gaps between "described as built" and
"verifiably built," and flag what Cosmo needs before she can write
`docs/specs/seed-atom-conformance.md` with confidence.

**Working interpretation** (stated per protocol in case Cosmo needs to correct scope): this
is a pre-implementation research pass, not a code change. I have no `write_file`/`edit_file`
tools and the actual chain (Go source, genesis, dtvp module, live RPC) lives on host `m3`,
which is unreachable from this sandbox. My contribution is (a) corroborating the epic's
design claims against the paper and the publicly-committed grant materials, (b) auditing
the two pieces of source I *can* reach (`reflow/anchor.py`, `dreamtree/roots`) for concrete
conformance gaps, and (c) being explicit about what I could not verify.

---

## 2. What I found

### 2.1 protocol-spec.md is real, externally published, and unreachable from here

`protocol-spec.md` is not a hypothetical design doc — it's referenced by filename in
`docs/grants/dreamtree/shared/team-bio.md:71` ("Protocol-level thinking | arXiv 2605.25844,
protocol-spec.md |") and described in
`docs/grants/dreamtree/builders-program/application.md:47-48` as: **"Protocol Specification
(published 2026-06-24)" — "Full technical specification covering: network architecture
(Cosmos SDK + CometBFT, progressive decentralization), identity (...), work & reputation
(attestation-as-work consensus, four proof types, reputation dynamics with formal math),
currency & records (Photons + Seeds, monetary policy, marketplace mechanics), and access
(four hard rules, TEE-attested compute-to-data, proxy re-encryption).**"

That currency & records description already matches S0-S4 of this epic almost verbatim
(photons/seeds, marketplace mechanics, TEE compute-to-data, PRE — compare to S4's "TEE
compute-to-data + PRE + output minimization"). This is a **strong positive signal**: the
epic is not inventing new policy, it's conforming implementation to a spec that was already
published and shown to grant reviewers seven weeks ago (2026-06-24, vs. today 2026-07-15).

I could not find `protocol-spec.md` anywhere in `/v3` (exhaustive filesystem search,
`find / -iname protocol-spec.md` returns nothing). It's presumably published at
`dreamtree.org` ("manifesto + documentation" per `builders-program/application.md:37`),
alongside `id.dreamtree.org` (the live Roots wallet, confirmed reachable per grant docs).
I attempted to fetch it directly:

- `https://dreamtree.org/protocol-spec.md` → **HTTP 403 Forbidden**
- `https://dreamtree.org` (site map) → **HTTP 403 Forbidden**
- Web search for the spec / manifesto → returns unrelated orgs (a Taos NM nonprofit,
  a labeling-software product) with the same name; no indexed hits for the actual document.

**Gap to flag before spec-writing:** nobody on this pass has actually read the current text
of protocol-spec.md to confirm it says what the epic claims (dtvp absent, photon as native
asset, etc. — the epic itself states "zero mentions [of dtvp] in protocol-spec" as a given
fact, which I cannot independently confirm). Someone with browser/authenticated access to
`dreamtree.org`, or access to wherever the markdown source lives (likely alongside the chain
repo on m3), needs to pull the canonical text so `docs/specs/seed-atom-conformance.md`
amends the real file rather than a reconstruction from grant-doc paraphrase.

### 2.2 The paper grounds S5's math exactly — this is conformance, not invention

`philosophy/attribution-economics/theory/definitions.md`:
```
V_out:      Value produced by a transformation.
V_captured: Value received by a contributor (through transactions, claims).
β_i = V_captured_i / V_out        (Σ β_i = 1)
L_i = V_out_i - V_captured_i
```
This is the exact leverage/capture-share math the epic's S5 references ("L_i computable
per creator (V_captured from sales, V_out from attest)"). It appears consistently across
`theory/axioms.md`, `theory/dimensional-analysis.md`, and the paper's own framework
sections (`paper/current/03-theoretical-framework.md`, `04-attribution-mechanism.md`).
S5 is a straight implementation of an equation the paper has carried since its earliest
draft.

Cross-check against the **publicly committed** monetary policy in grant materials
(`builders-program/application.md:84-87`, `technical-abstract.md:49-50`):
> "Photons (P): Fungible currency, minted by block production... Seeds (S): Non-fungible
> records — each seed IS the contribution record... Monetary policy: `photons = seeds`.
> One photon minted per seed to its storer-validators."

This is the "SEED IS THE ATOM... one seed = one photon" ratification from the epic,
word-for-word, already stated publicly. **The ratification isn't new policy — it's
restating what's already on record with grant reviewers.** That materially de-risks S0/S1:
this is closing a drift between implementation (dtvp, batch_root-as-a-kind) and a design
that was already locked and published.

### 2.3 Concrete gap #1 — `reflow/anchor.py` doesn't send `leaf_count` yet

`reflow/anchor.py` (the only anchoring client I could actually read) confirms the Merkle
contract described in the epic exactly: leaves = atom_ids sorted ascending bytewise,
parent = SHA-256(left‖right), odd node promotes unchanged. Its docstring header is dated
to the same substrate decision: *"live from atom one (substrate §0.3, decided
2026-07-10)"* — **the same date the epic says dtvp entered as staking expedience
(`eaf172e`, 2026-07-10)**. That's not proof of a causal link, but it's a coincidence worth
naming: whatever sprint produced the atom-anchoring seam also produced dtvp, which suggests
dtvp was bolted on opportunistically in the same window rather than following from a
separate design decision — consistent with the epic's framing of dtvp as "never a design
decision... staking expedience."

The gap: `Anchor.commit()` (lines 54-63) currently POSTs
`{subject, commitment, kind, source_ref}` to `/anchor` and expects back
`{id, txhash, height}` — a single seed_id per call, i.e. batch-of-1 / `MsgCommitSeed`
semantics. To satisfy S0's `MsgCommitBatch{merkle_root, leaf_count, kind, data_type,
subject, source_ref}`, this client needs:
- a `leaf_count` field added to the POST body and to the `Anchored` dataclass response
  (the response will presumably now be a *range* of seed ids, not one `id`)
- `KIND_BATCH = "reflow.batch_root"` needs to stop being the value sent as `kind` — the
  epic explicitly retires `batch_root` as a seed-kind ("kind describes the LEAF"), so this
  constant needs to become whatever the leaf-level kind/data_type actually is for reflow
  generations.

This is a small, well-scoped, single-file change — good first target for Coder once the
server-side `MsgCommitBatch` handler exists.

### 2.4 Concrete gap #2 — `dreamtree/roots` has zero anchor-tracking code or schema

`dreamtree/roots` is a single-commit repo (`d6359e9`, "DT-2: Roots — new Cloudflare Worker
with dedicated D1", clean working tree). Its only migration,
`migrations/0001_roots_init.sql`, defines `wallets`, `issuers`, `records`, `record_events`,
`access_log`, `api_keys`, `signing_keys`, `memberships` — **no `anchor_state`, `seed_id`,
`txhash`, or `merkle_root` column anywhere**, and a full-text grep of `roots/src` for
`anchor` returns **zero matches**.

This means the cutover acceptance criterion "roots 67/67 re-anchored (D1
`anchor_state=anchored`)" and S2's "roots DID→address binding... seed owner (subject)
becomes payee" require **net-new** schema (a migration 0002 adding anchor-tracking columns
or table) and **net-new** integration code (a client/route in roots calling anchord's
`/anchor`, mirroring what `reflow/anchor.py` already does) — none of this exists yet in the
source I can see. This isn't a blocker, just scope that needs to land on a spec/build card
explicitly rather than being assumed to already exist because reflow's side does.

The "67" is a useful anchor point (no pun intended): it matches the environment facts'
"67 roots records" and "67+ photons all at treasury" — same number, which confirms these
are the same 67 records anchored once already under the current (dtvp-era) scheme, now
needing re-anchoring post-wipe. Small, enumerable, drillable set — consistent with the
epic's "(drill proven)" framing.

### 2.5 Adjacent, out-of-scope finding — gateway kanban REST API is dead code

Not part of this epic, flagging briefly per research discipline: `gateway/internal/handlers/kanban.go`
implements `GET /api/cards/{short_id}` (full card body + events + artifacts) and
`GET /api/boards/{slug}/cards`, but I could not find either handler wired into
`gateway/internal/routing/router.go` or anywhere else in the gateway — no calls to
`KanbanBoards(`, `KanbanCard(`, or `KanbanCTOLens(` exist outside their own definitions.
Confirmed live: `curl` against the running gateway (`http://gateway:8000/api/cards/DT-18`)
returns a plain chi 404, not an app-level error. Practical consequence for this research
pass: I worked from `kanban_list_cards` (titles only) plus the epic text pasted into my
task, not from DT-18's actual stored `body`/comment history, because that data is currently
unreachable via any tool I have. Worth a follow-up card on the `cto` or `dreamtree` board
if the `/board` PWA is meant to be serving card detail views — I did not create one since
it's tangential to DT-18's scope.

### 2.6 DT-17 vs. DT-18 — possible plan conflict, unconfirmed

`kanban_list_cards(board=dreamtree)` shows **DT-17 [proposed] "CHAIN: full value layer →
m3 via x/upgrade (one push)"** (p9, same priority as DT-18) sitting in `proposed`. Title
alone suggests a live-upgrade path with no wipe — in apparent tension with DT-18's "WIPE
HERE (once)... everything after is upgrades." I could not read DT-17's body (same REST gap
as above), so I can't confirm whether it's already superseded or still an active competing
plan. **Recommend Cosmo explicitly reconcile DT-17 against DT-18 before spec lock** — my
best guess reading the epic's own language is that DT-17's `x/upgrade` mechanics are
exactly right for the *post-wipe* stages (S2-S4 are explicitly "upgrades"), just not for the
S0/S1 wipe itself, but that's inference, not confirmed from DT-17's actual content.

### 2.7 DT-16 confirms a prerequisite is already satisfied

`kanban_list_cards` shows **DT-16 [test] "chain: x/attest DONE — next x/reputation"**.
S5 needs `V_out from attest`, which depends on `x/attest` — per this card title, that
module is already built and in the `test` lane. No additional scope needed there.

---

## 3. What I could NOT verify (infra access boundary)

This sandbox has no route to host `m3`: `ssh m3` fails at DNS resolution
("Could not resolve hostname m3"), and no `dreamtreed` binary, no chain Go source
(`x/seeds`, `x/attest`, dtvp module), no `genesis.json`, and no reflow PG database are
reachable from here. I could not independently check:

- the live `dreamtree-1` height (~82.5K), seed count (36,979), or photon balances (67+ at
  treasury)
- dtvp bonded amount (700M / 1B minted) or the validator's key custody claim
- `x/seeds` keeper/msg_server line counts (67 / 124) or the `PhotonHooks.OnRecordSeed(kind)`
  signature
- reflow's Postgres corpus counts (11,695,703 atoms / 61,932 generations, ~25K unanchored)
- whether `eaf172e` (2026-07-10) is in fact the commit that introduced dtvp

All of the above are taken as given from the environment-facts block handed to me in this
task, presumably from Braedon's/Cosmo's own inspection of m3. **I am not confirming them —
I'm noting that nothing in this research pass contradicts them, and nothing in this
sandbox could either.** Go-level S0/S1 implementation verification (the acceptance
criteria's "go tests green," e2e batch-commit tests, genesis boot) has to happen on m3,
by an agent with actual access to that host.

---

## 4. Recommended approach for the spec

1. **Retrieve the real protocol-spec.md text before drafting the amendment.** It's a
   real, externally-published (2026-06-24) document; don't let the spec amendment be
   written against a reconstruction from grant-doc paraphrase when the source almost
   certainly exists verbatim on m3 or is fetchable with authenticated/browser access to
   `dreamtree.org`.
2. **Treat S0/S1 as conformance, not new policy**, when framing the decision log — the
   paper's `L_i`/`β_i` math and the grant docs' "photons = seeds" 1:1 statement were both
   already public before this epic. That's a stronger footing than "ratified today."
3. **First concrete code targets, in dependency order:**
   - `reflow/anchor.py::Anchor.commit()` — add `leaf_count`, retire
     `KIND_BATCH = "reflow.batch_root"` as the sent `kind` value (server-side
     `MsgCommitBatch` has to land first; this is the client-side follow-up).
   - `dreamtree/roots` — new migration adding anchor-tracking (state/seed_id/txhash/
     merkle_root/anchored_at) plus a client hitting anchord's `/anchor`, mirroring
     `reflow/anchor.py`. Currently zero anchor code exists in roots; don't assume it's a
     data-only re-anchor operation.
   - Everything else (x/seeds leaf-range allocation, dtvp deletion, uphoton denom, genesis
     corpus mint, chain-id wipe) is Go-level work on m3 outside this sandbox's reach —
     route to Coder/Executor with m3 access.
4. **Reconcile DT-17 against DT-18** explicitly (§2.6) before spec lock — read DT-17's
   actual body (gateway REST gap means this needs direct DB/board access, not the MCP
   list tool) and either close it as superseded by the wipe decision or fold its
   `x/upgrade` mechanics into the post-wipe stages where the epic already says "everything
   after is upgrades."

---

## 5. Sources

**Local (read directly):**
- `philosophy/attribution-economics/theory/definitions.md`, `axioms.md` — L_i/β_i/V_out/V_captured definitions
- `philosophy/attribution-economics/paper/current/04-attribution-mechanism.md`, `03-theoretical-framework.md`
- `reflow/anchor.py` — full file
- `dreamtree/roots/migrations/0001_roots_init.sql` — full file; `git log` (single commit `d6359e9`)
- `dreamtree/roots/src/**` — grep for `anchor` (zero matches)
- `dreamtree/VISION.md` — full file
- `docs/grants/dreamtree/shared/team-bio.md`, `executive-summary.md`, `technical-abstract.md`
- `docs/grants/dreamtree/builders-program/application.md`
- `docs/grants/dreamtree/LANDSCAPE.md`
- `gateway/internal/handlers/kanban.go`, `gateway/internal/routing/router.go` (dead-code check)
- Filesystem-wide search: no `protocol-spec.md`, no chain Go source, no `dreamtreed` binary anywhere under `/v3` or `/`

**Web (attempted, failed):**
- `https://dreamtree.org/protocol-spec.md` → HTTP 403
- `https://dreamtree.org` → HTTP 403
- Web search for the spec/manifesto → no relevant hits (name collision with unrelated orgs)

**Kanban (dreamtree board, titles only — REST card-detail endpoint unreachable, §2.5):**
- DT-18 (this card), DT-17, DT-16, DT-1, DT-2 through DT-15

**Infra checked, unreachable:**
- `ssh m3` — DNS resolution failure
- Gateway REST `/api/cards/{short_id}` — not wired into router (404 on a live, running gateway)

---

## 6. Confidence assessment

| Claim | Confidence | Basis |
|---|---|---|
| protocol-spec.md exists, published 2026-06-24, covers photons/seeds/currency policy | 0.8 | Directly named + described in two independent grant docs |
| S0/S1 design (seed=atom, photons=seeds 1:1) matches already-public commitments | 0.85 | Verbatim match between epic language and grant-doc monetary policy language |
| S5 math (L_i, β_i) is native to the paper, not new | 0.9 | Multiple primary sources within the paper/theory tree, consistent definitions |
| `reflow/anchor.py` needs a `leaf_count` param to satisfy S0 | 0.9 | Direct code read, current POST body has no such field |
| `dreamtree/roots` has zero anchor integration today | 0.9 | Direct code + schema read, single-commit repo, full-text grep empty |
| dtvp entered 2026-07-10 same day as the anchoring substrate decision | 0.5 | Date coincidence only (docstring vs. epic's stated commit date); not a confirmed causal link |
| Live chain state (height, seed/photon counts, dtvp bonded amount) | Not verified | No m3 access from this sandbox; taken as given |
| DT-17 conflicts with DT-18's wipe plan | 0.3 | Title-only inference; card body unreadable |

---

*Prepared by the Gnosis Researcher agent. Handing to Cosmo for spec-writing
(`docs/specs/seed-atom-conformance.md`) — see §4 for the four items I'd want resolved or
at least acknowledged before that spec locks.*
