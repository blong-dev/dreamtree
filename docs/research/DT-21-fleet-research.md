# DT-21 — CONFORMANCE COMB: protocol-spec + paper vs built system (ratified gap matrix)

> **STATUS: SUPERSEDED IN PART — see §6 (Reconciliation, 2026-07-16) before using
> any cell of this matrix.** This document was produced under an honest,
> self-declared sandbox boundary (§2.1: protocol-spec.md unread; no chain
> source). Primary-source combs with full filesystem access exist and override
> it cell-by-cell:
> `dreamtree/docs/specs/comb-spec-vs-chain.md` (spec read in full, spec-line +
> Go-line evidence) and the research repo's
> `reviews/comb-paper-vs-instrument.md` (paper v0.2 read in full). This file is
> NOT "the ratified output" — it is a proxy-based research record, retained for
> its two original findings (§6.3) and as the record of the sandbox boundary
> itself.

**Card:** DT-21 (dreamtree board, research lane, p8)
**Researcher:** Gnosis Researcher agent, sandboxed container (no m3 / no chain-source access — same boundary as DT-18)
**Date:** 2026-07-16
**Method:** Two parallel sub-agent combs (spec-vs-chain; paper-vs-instrument) + one direct verification of the highest-stakes discrepancy they surfaced. Full sub-agent transcripts available on request; this file is the synthesized, ratified output.

---

## 1. Problem statement

DT-18 (2026-07-15) found `seed=atom` conformance by accident of conversation and did a first-pass research/implement cycle under a hard infra boundary: no route to host `m3`, where the live `dreamtree-1` chain's Go source, genesis, and the canonical `protocol-spec.md` all live. This card makes that discovery systematic: comb the two documents of record — `protocol-spec.md` (chain design) and *Attribution as Conservation* (the theory, `/v3/philosophy/attribution-economics/`) — against the BUILT system (chain `x/` modules + reflow + roots), and produce a CONFORMS/PARTIAL/MISSING/CONTRADICTS gap matrix with file:line evidence on both sides, for the owner to triage into build/defer/amend decisions.

**Scope correction, stated up front per research protocol:** the infra boundary that limited DT-18 is unchanged today, re-verified independently three times (DT-18 on 2026-07-15, and both of this card's sub-agents on 2026-07-16): `ssh m3` fails DNS resolution, `https://dreamtree.org/protocol-spec.md` returns HTTP 403, and no chain Go source (`x/seeds`, `x/attest`, `projection.go`) exists anywhere in this filesystem. **A true primary-source spec-vs-chain comb is not possible from this sandbox.** What follows is the best achievable comb given that boundary: primary evidence where this sandbox has it (reflow, roots, the paper), clearly-labeled secondary/paraphrase evidence where only that exists (grant docs describing protocol-spec.md), and an explicit `UNVERIFIABLE-LOCAL` rating — never a guess — everywhere the answer lives only in protocol-spec.md or chain Go source. Fixing that boundary is DT-19's job, not this card's; DT-19 is still sitting in `proposed`, unactioned (see §5).

---

## 2. Section A — Spec vs. Built (chain side)

### 2.1 Proxy-evidence caveat

`protocol-spec.md` itself was not read by anyone on this pass (or DT-18's). All "spec-side" evidence below is one of: (a) grant-doc paraphrase (`docs/grants/dreamtree/**`, all describing a document "published 2026-06-24" and shown to grant reviewers), or (b) the epic's own conformance plan (`docs/specs/seed-atom-conformance.md`), which is itself a reconstruction, not an amendment of the real file (its own header says so). Treat every "spec-side" cell below as secondary until DT-19 lands.

### 2.2 Gap matrix

| # | Item | Rating | Spec-side evidence | Built-side evidence | Reasoning |
|---|---|---|---|---|---|
| 1 | seed-size cap | **UNVERIFIABLE-LOCAL** | none found in any local doc | none found | Zero local hits for "seed size"/"size cap"; concept only reachable in protocol-spec.md or `x/seeds` on m3. |
| 2 | storage-cost oracle | **UNVERIFIABLE-LOCAL** | none found | none found | No local document mentions any storage-cost-oracle mechanism. |
| 3 | endowment: per-seed vs. pooled | **UNVERIFIABLE-LOCAL** | `docs/specs/seed-atom-conformance.md:149` names "endowment" once, no structural detail | none | Named but not designed anywhere reachable; the per-seed-vs-pooled choice lives in protocol-spec.md/m3 only. |
| 4 | access_cut_to_storers | **MISSING** | `docs/grants/dreamtree/builders-program/application.md:84,93` — photons mint to storer-validators at seed-recording time; separately "Marketplace toll (5%) → infrastructure fund" (not storers by name) | grep for `access_cut`/`storer.*cut`/`storer.*payout` across `docs`, `dreamtree`, `reflow` → only the one mint-time line; zero revenue-distribution code anywhere | The reachable proxy describes storer comp only at mint time and routes *ongoing* marketplace revenue to an "infrastructure fund," not storers — the "access_cut_to_storers" framing may already be in tension with the grant-doc proxy itself, and no code implements any split either way. |
| 5 | on/off ramps (fiat↔photon) | **UNVERIFIABLE-LOCAL** | `builders-program/application.md:84` — photons "float against fiat" (implies a rate, not a ramp) | none | Thin signal only; no local text confirms or denies a ramp mechanism exists. |
| 6 | uptime/durability bonds | **UNVERIFIABLE-LOCAL — flagged tension** | `docs/specs/seed-atom-conformance.md:138-140` — S1 scope explicitly includes "**zero slash fractions**" alongside the `uphoton` bond denom | none (chain/Go-side concept, expected absent locally) | The local conformance plan states slashing is being *zeroed*, post-dtvp. If protocol-spec.md calls for durability bonds backed by slashing, this is a live CONTRADICTS — but which way it resolves can't be confirmed without the real text. Worth flagging to the owner explicitly rather than filing as routine MISSING. |
| 7 | TEE access stack | **MISSING** | `docs/grants/dreamtree/shared/technical-abstract.md:57`, `builders-program/application.md:48` — "TEE-attested compute-to-data (Intel TDX / AMD SEV-SNP / AWS Nitro)"; also listed under seed-atom-conformance.md's own "Explicitly NOT done" | grep `TEE`\|`compute-to-data`\|`SEV-SNP`\|`Nitro`\|`TDX` across `dreamtree/`, `reflow/` → zero real hits | Clearly and repeatedly stated intent; zero implementation anywhere reachable; the epic's own spec doc agrees it's unbuilt. |
| 8 | per-type JSON Schemas / `data_type` | **MISSING (concrete)** | `seed-atom-conformance.md:13-14` defines `MsgCommitBatch{merkle_root, leaf_count, kind, data_type, subject, source_ref}` — `data_type` is a named field of the chain's own commit message; `team-bio.md:31`/`executive-summary.md:51` separately claim a "data_type registry" is "✅ Complete (Phase 2)" | `reflow/anchor.py:99-102`'s actual POST payload omits `data_type` entirely; grep `data_type` across `dreamtree/roots/src`, `dreamtree/roots/migrations`, `reflow/` → **zero matches**; `roots/migrations/0001_roots_init.sql` has only a `format` column (credential wire-format, e.g. `vc-di`), not a data_type taxonomy | This is the sharpest concrete gap in the matrix: the spec's own commit-message schema names a required field that literally no reachable client sends, and a grant doc's "complete" claim doesn't match what's in the one schema file that would carry it. Recommend: **build** (add `data_type` to `anchor.py`'s POST + a real registry), not defer — it blocks S0 conformance as specified. |
| 9 | receiver-key handoff (custody migration) | **MISSING** | Kanban DT-6 title: "roots P5 — custody handoff (receiver-key migration)" (`proposed`, p6); `executive-summary.md:55` — "Custody handoff \| 🔵 Planned" | grep `custody`\|`receiver.?key` in `dreamtree/roots/src` → zero matches | Explicitly named, explicitly still planned on the board, confirmed zero implementation in reachable Roots source. Already tracked (DT-6) — no new card needed, just re-priority if the owner wants it. |
| 10 | domain taxonomy tiers | **UNVERIFIABLE-LOCAL — likely concept mismatch** | Only "domain-indexed reputation" language found (`icf/proposal.md:39,64`, `technical-abstract.md:44`) — a reputation-scoring domain index, not a data/seed classification taxonomy | none | Can't confirm without protocol-spec.md whether "domain taxonomy tiers" means the reputation domains found here or a distinct data/access taxonomy. Flag for whoever retrieves the real text to disambiguate explicitly. |
| 11 | committer authorization (`max_batch_new_count` interim cap) | **UNVERIFIABLE-LOCAL** | none — `max_batch_new_count` returns **zero hits** anywhere under `/v3` | `reflow/anchor.py:77-79` shows bearer-token client auth (`ANCHORD_TOKEN`) — a different concept (client auth, not an on-chain committer-authorization/cap rule) | The specific named mechanism is entirely chain-side (`x/seeds` keeper/msg_server), unreachable. |
| 12 | fabric layer (per-wallet chains, DA root, possession proofs, storer set) | **MISSING** | `technical-abstract.md:31,79` — "wallet-indexed data fabric... per-wallet append-only chains... sharded"; `seed-atom-conformance.md:148-149` lists this whole item under "Explicitly NOT done" | grep `wallet.indexed`\|`DA root`\|`possession proof`\|`storer set` in `dreamtree/roots/src`, `reflow/` → zero matches | Clear design intent, explicitly and repeatedly flagged unbuilt by the epic's own spec, confirmed absent in reachable code. Largest single deferred item — this is S3+ scope, appropriately **defer**, not build-now. |
| 13 | owner payout (S2 — DID/address→payee resolution) | **MISSING** | `seed-atom-conformance.md:144-147` — explicitly "not done... roots' side only got the anchor-tracking schema, not the ownership wiring"; kanban DT-20 title: "S2 owner-payout rides first" (`proposed`) | `roots/migrations/0002_roots_anchor_tracking.sql` adds only `anchor_state/seed_id/anchor_tx/anchor_height/merkle_root/anchored_at` — **no payee/owner/address column**; grep `payee`\|`owner.as.payee` in `roots/src` → zero matches | Confirmed not-done by both the spec's own text and direct schema/grep read. DT-20 already proposes prioritizing this — matrix corroborates that's the right call. |
| 14 | access enforcement (S4 — TEE compute-to-data + PRE + output minimization) | **MISSING** | `technical-abstract.md:55-59` — four hard rules, TEE, PRE (tACo/Threshold Network), license registry; `seed-atom-conformance.md:148-150` lists explicitly as not done | grep `proxy.re.?encryption`\|`tACo`\|`Threshold Network`\|`output.minimization` → zero matches | Same pattern as #7/#12: clear intent, zero reachable implementation, epic agrees it's unbuilt. Appropriately **defer** — S4 scope, sequenced behind S0-S2. |

### 2.3 Material finding outside the checklist — same-day document staleness on `MsgCommitBatch`

Two local artifacts make a **factual claim this sandbox cannot verify**, and one local artifact **contradicts** them, all dated the same day (2026-07-15):

- `reflow/anchor.py:22-28` (comment): *"The chain-side `MsgCommitBatch` handler landed (dreamtree `07367cf`) and actively REJECTS any kind containing `batch_root` (`ErrRetiredKind`)."*
- `docs/BUILD_LOG.md:122-166` ("Addendum, 2026-07-15, same day"): repeats the same claim — *"DT-18's `MsgCommitBatch` handler landed on the chain (dreamtree `07367cf`)"* — and describes downstream client changes (`new_count` now required) made on the strength of that claim.
- `docs/specs/seed-atom-conformance.md:135-136`, written the **same day**, lists *"`x/seeds`' `MsgCommitBatch` handler (leaf-range allocation, retiring `batch_root`...)"* under **"Explicitly NOT done (out of this sandbox's reach)."**

Neither the "it landed" claim nor the "not done" claim is independently checkable from here — no Go source, no m3 route, confirmed three times now. The BUILD_LOG addendum sits chronologically *after* the original DT-18 entry in the log (reverse-chronological file), so at minimum the spec doc is now stale relative to whatever later session produced that addendum. **Recommend: amend the doc** — reconcile `seed-atom-conformance.md`'s "Explicitly NOT done" list against the addendum's claim, ideally by having someone with m3 access confirm commit `07367cf` actually exists and behaves as described, before either document is trusted at face value. This is a documentation-integrity gap, not a build gap — cheap to fix, high value (multiple downstream client changes, e.g. `new_count` becoming required, were made on the strength of an unverified claim).

---

## 3. Section B — Paper vs. Built (instrument side)

### 3.1 Framing

The paper's core formalism — `V_out`, `V_captured`, `β_i = V_captured_i/V_out` (Σβ_i = 1), `L_i = V_out_i - V_captured_i` — is DT-18's S5 target. The real S5 deliverable, `docs/specs/measurement-backtest.md`, does **not exist** (confirmed by `find /v3 -iname "measurement-backtest*"` → empty, independently on both 2026-07-15 and 2026-07-16). Reflow is not that instrument — it's Gnosis's own separate attribution pipeline — but it's the only locally-reachable analog of "is this math built anywhere yet," so it's the best available proxy for the built side.

### 3.2 Gap matrix

| # | Claim | Rating | Paper-side evidence | Built-side (reflow) evidence | Reasoning |
|---|---|---|---|---|---|
| 1 | Core equations: V_out / V_captured / β_i / L_i | **MISSING** | `theory/definitions.md:47-60`; `theory/axioms.md:90-100`; `theory/equations/_master.md:121-187`; `paper/current/03-theoretical-framework.md:334-460` | grep `V_out\|V_captured\|beta_\|β\|leverage\|decay_rate` across `/v3/reflow/*.py` → zero matches | No file anywhere in reflow computes or stores any of these under any name. S5 gap confirmed still fully open. |
| 2 | Σβ_i = 1 conservation constraint | **MISSING** | `axioms.md:96` ("Σβ_i = 1, claims sum to total output"); `_master.md:126` | none — β doesn't exist to be normalized (see #1) | Cannot be enforced when the underlying quantity isn't computed. **Do not conflate** with reflow's `resolution.py:128-135` `confidence() = 1-∏(1-w)` — that's a *different* paper equation (Appendix A.7, attestation aggregation), not the capture-ratio conservation constraint (Axiom 7). Flagged explicitly so downstream spec-writing doesn't mistake one for the other. |
| 3 | Decay rates (λ = λ_physical + λ_obsolescence + λ_information, V(t)=V₀e^(-λt)) | **PARTIAL** | `axioms.md:76-87` (Axiom 6); `03-theoretical-framework.md §3.5:221-330`; `Appendix A §A.2:131-267`; `Appendix C §C.2:124-210` | `resolution.py:128-135` uses a per-edge `decay` scalar inside its confidence formula (structurally matches §4.3/A.7, not Axiom 6) | Vocabulary and "decay as a quality-weighting factor" idea overlap with one paper equation; the paper's central value-decay dynamics (λ decomposition, exponential decay) is absent everywhere in reflow. |
| 4 | Threshold dynamics (§05 — Θ, δ, V_unlock, activation energy) | **PARTIAL** | `05-threshold-dynamics.md §5.1-5.5:9-260`; `Appendix A §A.6:589-687` | `changepoints.py:79-110` (BIC binary segmentation); `detectors.py:30,32` (`_R_STRONG=0.7`, `_Z_ANOMALY=3.0`); `enhance.py:141-142` (`_ESTABLISHED_CUT=1.5` binary gate) | Reflow's threshold tooling is statistical regime/anomaly detection on time series — a different mathematical object from the paper's phase-transition/V_unlock model (no activation energy, no hysteresis). `enhance.py`'s gravity cutoff is the closest structural analog but gates graph-confidence tiers, not economic value-unlock. |
| 5 | Paper's own measurement methodology (§07/Appendix C) vs. `measurement-backtest.md` | **MISSING** | `07-methodology.md` (full); `Appendix C` (protocols C.1-C.6, full) | `find /v3 -iname "measurement-backtest*"` → empty | The actual S5 deliverable does not exist anywhere in this sandbox. Unchanged from DT-18's finding yesterday — zero progress in 24h. |
| 6 | `demand_signal=1.0` stand-in | **UNVERIFIABLE-LOCAL** | n/a (epic-asserted Go implementation detail, not a paper claim) | `find / -xdev -iname "projection.go"` → empty; `grep -rn "demand_signal" /v3` → empty; `ssh m3` → DNS failure (reverified) | Lives on `m3`, unreachable by any tool available here. |
| 7 | citation-uplift → configurable param (TODO in `projection.go`) | **UNVERIFIABLE-LOCAL**, with adjacent context | n/a — paper's closest concept is β_proof/demonstrability (`_master.md:145-151`), not "citation uplift" | Same empty grep/find as #6. Adjacent-but-distinct: `materialize.py:532,563,571,580,595` stores OpenAlex `cited_by_count` as a mutable, COALESCE-protected field — citation-*count* storage, not a citation-*uplift* multiplier | The Go-side constant can't be assessed from here. The adjacent reflow mechanism is relevant context but is not the same mechanism — do not treat it as evidence the Go TODO is already solved elsewhere. |

### 3.3 Material finding outside the checklist — no contradiction, but one genuine positive signal

Nowhere does reflow's attribution-adjacent math structurally violate the paper's conservation principle — it simply doesn't implement the β/L machinery, so there's nothing to contradict (no CONTRADICTS rating in this section). The one genuine positive: reflow's attestation/standing/specificity/decay machinery (`resolution.py:128-135`, `enhance.py:25-45`) is a working, production implementation of the paper's §4.3/Appendix A.7 attestation-aggregation equation (`1-∏(1-w)`) — just applied to entity-resolution confidence, not economic attribution. Worth naming to the owner as evidence the paper's formalism is implementable and already has one working precedent in this codebase, even though it's not the S5-relevant core.

---

## 4. Cross-cutting observations

1. **The root blocker is singular and already has a card.** 7 of 21 matrix rows (33%) are `UNVERIFIABLE-LOCAL` for the same reason: no m3 access, no protocol-spec.md text. DT-19 ("Retrieve canonical protocol-spec.md text before DT-18 spec amendment lands") is the correct, already-proposed vehicle for closing this — it is still sitting in `proposed`, unactioned, as of this pass. Every future conformance-comb card will re-derive the same boundary and the same UNVERIFIABLE-LOCAL rows until DT-19 (or equivalent m3-connected access) lands. Recommend the owner prioritize DT-19 over spawning further sandbox-bound research passes on this topic.
2. **DT-17/DT-18/DT-20 plan reconciliation is still open.** DT-18 §2.6 flagged a possible conflict between DT-17 ("full value layer → m3 via `x/upgrade`, one push") and DT-18 ("one wipe, then upgrades"); DT-20's title ("governed in-place upgrades, no more wipes; S2 owner-payout rides first") reads like a resolution in DT-17's direction, but its body is unreachable (kanban card-detail REST endpoint still dead code, unchanged from DT-18's finding) so this is inference, not confirmed.
3. **One concrete, cheap, high-value fix identified: §2.3's `MsgCommitBatch` staleness.** This doesn't need m3 access to fix — it needs someone to reconcile two local documents (`seed-atom-conformance.md` vs. the BUILD_LOG addendum + `anchor.py` comment) against whichever is actually current, which likely does need one m3-side confirmation of commit `07367cf`.
4. **One concrete, buildable-now gap: `data_type` (§2.2 #8).** The spec's own `MsgCommitBatch` schema names this field; no reachable client sends it. This is fixable inside this sandbox's reach (same shape as DT-18's original `leaf_count` fix) without needing m3 access at all — good next Coder target.

---

## 5. Recommended owner triage (build / defer / amend-doc)

| Bucket | Items | Suggested decision |
|---|---|---|
| **Build now (no m3 needed)** | `data_type` field on `anchor.py`/`roots/src/anchor.ts` POST bodies + a minimal registry (#8) | Build — same shape as DT-18's leaf_count fix, closes a spec-vs-client gap that's fully within this sandbox's reach. |
| **Amend doc (no m3 needed)** | `MsgCommitBatch` staleness (§2.3) | Amend — reconcile `seed-atom-conformance.md` against the BUILD_LOG addendum's claim; flag for one m3-side confirmation of commit `07367cf`. |
| **Escalate (needs m3/owner-arranged access)** | Items #1, #2, #3, #5, #10, #11 (Section A) — seed-size cap, storage-cost oracle, endowment design, on/off ramps, domain taxonomy disambiguation, committer authorization; `demand_signal`/citation-uplift (Section B #6-7) | Defer, but bottlenecked on DT-19 — recommend prioritizing DT-19 explicitly since it unblocks all of these at once rather than one at a time. |
| **Flagged tension, needs a read of the real text** | uptime/durability bonds vs. "zero slash fractions" (#6) | Do not build either direction yet — genuine risk of building the wrong thing; resolve via DT-19 first. |
| **Already tracked, no new card needed** | receiver-key handoff (#9, = DT-6), owner payout (#13, = DT-20), fabric layer (#12, = DT-17/S3+), access enforcement (#14, = S4) | Defer at current sequencing — matrix corroborates existing board priority, no action needed beyond what's already proposed. |
| **Build the instrument** | S5 core equations (paper §3.2 #1-2, #5) | Defer — this is the real `measurement-backtest.md` deliverable; still fully unbuilt, no local blocker other than it not having been scoped/built yet. Distinct from the m3 boundary — this one just needs to be picked up. |

---

## 6. Sources

**Read directly, this pass:** `docs/research/DT-18.md`, `docs/specs/seed-atom-conformance.md`, `docs/BUILD_LOG.md` (targeted + the 2026-07-15 addendum in full), `reflow/anchor.py` (full), all `docs/grants/dreamtree/**/*.md` (13 files), `dreamtree/roots/migrations/{0001,0002}*.sql`, `dreamtree/roots/src/{anchor.ts,routes/admin.ts,types.ts}`, `dreamtree/roots/wrangler.toml`, `philosophy/attribution-economics/theory/{definitions,axioms,dimensional-analysis}.md`, `theory/equations/_master.md`, `paper/current/{03-theoretical-framework,04-attribution-mechanism,05-threshold-dynamics,07-methodology}.md`, `paper/appendices/{A-equation-derivations,C-measurement-protocols}.md`, `reflow/{materialize,relationships,reconcile,resolution,enhance,changepoints,detectors}.py`.

**Verified directly, this pass:** `ssh m3` (DNS failure), `https://dreamtree.org/protocol-spec.md` and `https://dreamtree.org` (HTTP 403), web search for the spec (name-collision noise only), `find` for `protocol-spec.md`/`projection.go`/`measurement-backtest*` (all empty), `kanban_list_cards(board=dreamtree)` (21 cards, titles only — DT-19/DT-17/DT-20 body text still unreachable, dead REST endpoint per DT-18 §2.5, not re-verified this pass but nothing suggests it changed).

**Sub-agent transcripts (full detail beyond this synthesis):** Spec-vs-chain comb (agent `a1f918b2a58d08d51`), Paper-vs-instrument comb (agent `a2e929c1422bc6dbd`) — both run 2026-07-16, available for follow-up questions via SendMessage if the owner wants deeper drill on any single row.

---

## 7. Confidence assessment

| Claim | Confidence | Basis |
|---|---|---|
| Infra boundary (no m3, no protocol-spec.md) is unchanged from DT-18 | 0.95 | Independently re-verified 3x across 2 days by 3 separate passes, identical results each time |
| `data_type` field gap (#8) is real and fixable without m3 access | 0.9 | Direct code read (zero matches) + spec doc's own schema definition, unambiguous |
| `MsgCommitBatch` same-day document staleness (§2.3) is real | 0.85 | Direct read of both conflicting documents, exact line citations; only the *resolution* (which claim is true) is unverifiable, not the *existence* of the conflict |
| Paper's core S5 equations (V_out/V_captured/β/L) are unimplemented anywhere reachable | 0.9 | Exhaustive grep across all of `reflow/*.py`, zero matches, corroborated by DT-18's independent finding the prior day |
| `measurement-backtest.md` does not exist | 0.95 | Direct `find`, empty, confirmed twice across two days |
| Items rated UNVERIFIABLE-LOCAL genuinely require m3/protocol-spec.md access (not researcher shortfall) | 0.85 | Each has a documented exhaustive local search returning zero results, plus the independently-confirmed access boundary |
| DT-17/DT-20 reconciliation status | 0.3 | Title-only inference; card bodies unreachable |

---

*Prepared by the Gnosis Researcher agent, synthesizing two parallel sub-agent combs plus direct verification of the highest-stakes local discrepancy found. Handing to the owner for triage per DT-21's own design: each gap above is one explicit build/defer/amend-doc decision (§5).*

---

## 6. Reconciliation (2026-07-16, integrating session with full filesystem access — owner-directed)

Everything below was verified first-hand against the primary sources this
sandbox could not reach. Where this section conflicts with the matrix above,
this section wins.

### 6.1 The §2.3 integrity question — ANSWERED

Commit `07367cf` ("seed = atom: the leaf model + photon-native chain") **exists
on `origin/main` of github.com/blong-dev/dreamtree** (verified via
`git log`/`git branch --contains`). `MsgCommitBatch` is live on the `dreamtree`
chain (relaunched 2026-07-15, chain-id `dreamtree`, 13,463,772 photons = corpus
atoms at genesis) and was exercised end-to-end by `scripts/leaf-proof.sh` on a
live node. `reflow/anchor.py`'s and `BUILD_LOG.md`'s claims were correct; the
"Explicitly NOT done" list in v3's `docs/specs/seed-atom-conformance.md` was a
sandbox reconstruction that is WRONG about this — see §6.4.

### 6.2 Matrix cells superseded by the primary comb

Items 1, 2, 3, 5, 6, 10, 11 (`UNVERIFIABLE-LOCAL`) are all resolvable from the
real `protocol-spec.md`: seed-size cap, storage-cost oracle,
endowment-per-seed-vs-pooled, ramps, uptime/durability bonds,
`access_cut_to_storers`, and taxonomy tiers appear in the spec's own
"Still open" ledger — **DEFERRED-BY-SPEC**, not unknown. Item 6's flagged
tension resolves cleanly: the chain wires NO slashing/evidence modules at all
(nothing can burn bonded photons; documented in the spec's 2026-07-15
decision-log entry), so zero-slash is ratified design, not a contradiction.
Item 11: `max_batch_new_count` exists in `x/seeds/params.go`
(default 1,000,000 — the interim supply-griefing cap). §3's claim that
`measurement-backtest.md` "does not exist" is wrong for the system (it exists
in the chain repo at `docs/specs/measurement-backtest.md`); it was true only of
this sandbox. The full primary matrix (33 CONFORMS / 19 PARTIAL / 5 MISSING /
12 DEFERRED-BY-SPEC / 7 CONTRADICTS / 18 build-not-in-spec) is in
`dreamtree/docs/specs/comb-spec-vs-chain.md`.

### 6.3 This document's findings that SURVIVE (credited, first-hand confirmed)

1. **The `data_type` gap (item 8) is real and actionable**: `reflow/anchor.py`'s
   POST payload omits `data_type` while `MsgCommitBatch` carries the field —
   every reflow batch anchors unpriceable for the marketplace. Fix: payload +
   anchord passthrough + per-pack data_type values.
2. **Document bifurcation (§2.3, generalized) is real and worse than stated**:
   v3 carries doppelgänger reconstructions of primary documents —
   `v3/docs/specs/seed-atom-conformance.md` (contradicts the canonical
   chain-repo spec of the same name) and `v3/philosophy/attribution-economics/`
   (a PRE-v0.2 copy of the research repo: its paper still carries the old §6
   and old loss decomposition — any pass reading it combs superseded theory).

### 6.4 Standing decisions this reconciliation puts to the owner

(i) doppelgänger documents: delete / replace with pointer stubs / one-way sync
with freshness stamps; (ii) fleet sandbox access to the chain repo + canonical
paper (read-only mount, published spec, or scoping dreamtree research away from
fleet dispatch); (iii) a document-provenance convention (CANONICAL / MIRROR-of
/ RECONSTRUCTION / PROXY-RESEARCH headers) so no future reader — human, fleet,
or session — has to guess a document's epistemic basis again.
