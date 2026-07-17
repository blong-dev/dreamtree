# DT-21 — Conformance comb: protocol-spec.md vs the built chain

**Status: research phase, 2026-07-16. Agent-produced. PENDING OWNER TRIAGE — nothing here is a decision.**

Method: every substantive claim/mechanism in `protocol-spec.md` (read in full, decision log
included; the 2026-07-15 seed=atom entry treated as ratified truth) checked against the actual
code in `x/{seeds,attest,reputation,photons,licenses}`, `app/`, `scripts/`, `cmd/anchord`.
Companion docs consulted: `parameters.md` (canonical values), `docs/specs/seed-atom-conformance.md`,
`x-reputation-design.md`, `x-reputation-p2-review-windows.md`, `citation-value.md`,
`launch-readiness.md`, `measurement-backtest.md`.

Classes:
- **CONFORMS** — spec mechanism present and faithful (possibly via a different but equivalent mechanism; noted).
- **PARTIAL** — present but incomplete or diverging in a bounded way.
- **MISSING** — spec asserts it; build has nothing (and spec does *not* flag it open).
- **DEFERRED-BY-SPEC** — spec itself marks it open/TBD/roadmap; absence is honest.
- **CONTRADICTS** — build does something the spec (or parameters.md, which the spec says wins) says otherwise. Loudest class.
- **BUILD-NOT-IN-SPEC** — the build carries a mechanism/policy value the protocol spec never ratified (the dtvp class). Listed in §9.

All file:line references are to repo `/home/b/quorum/dreamtree/dreamtree`.

---

## 1. Network

| # | Spec claim | Spec | Build evidence | Class |
|---|---|---|---|---|
| N1 | Door #3: v0 = DreamTree operates the only validator; open code | protocol-spec.md:77-87 | `scripts/launch-genesis.sh:130` single owner gentx; AGPL LICENSE at repo root | CONFORMS |
| N2 | Cosmos SDK + CometBFT app-chain | protocol-spec.md:101 | `app/app.yaml:1-76`, `go.mod` (cosmos-sdk) | CONFORMS |
| N3 | Consensus layer vs value layer separate; reputation never orders blocks | protocol-spec.md:103 | Reputation lives in ABCI app modules only (`x/reputation`); staking voting power = bonded uphoton (`app/params/config.go:20`), no reputation input to validator set | CONFORMS |
| N4 | Validator Sybil-resistance is *permissioning* through v2 — "no staking token needed, preserving no-ICO" | protocol-spec.md:103 | Build DOES have a bonded staking token: `bond_denom = uphoton` (`app/params/config.go:20`, `scripts/launch-genesis.sh:130`). Superseded by the 2026-07-15 decision-log entry (protocol-spec.md:745: "validators bond genesis-corpus photons") — the *body text* at :103 was never updated | CONFORMS to decision log; body-text drift (see §8-D1) |
| N5 | Consensus blocks hold commitments + small txs, never bodies | protocol-spec.md:109 | `x/seeds/params.go:6-7` (512-byte commitment cap); `x/seeds/keeper/msg_server.go:80-101` hex-digest validation; comment "bodies never enter consensus" | CONFORMS |
| N6 | Data fabric: per-wallet append-only chains, DA root in header, sharding, cross-links | protocol-spec.md:110-113 | Nothing. `seed-atom-conformance.md:213-214` (roadmap S3): "bodies live in producer stores behind `source_ref` — stated, not hidden" | DEFERRED-BY-SPEC (spec :57 "storage primitives open"; conformance doc names it S3) |
| N7 | Participation spectrum: phones/DAS light nodes, durable storage nodes, Storj-style possession proofs, uptime/durability bonds | protocol-spec.md:115-126 | Nothing (no storage-node, DAS, or possession-proof code anywhere in `x/`) | DEFERRED-BY-SPEC (rides with S3; spec :749 lists "uptime/durability bond design" open) |
| N8 | Sovereign L1, no external settlement, no stablecoin | protocol-spec.md:128-130 | No IBC wired (`app/app.yaml` has no ibc module), settlement all in uphoton (`x/licenses/keys.go:7`) | CONFORMS |
| N9 | Block cadence a lever, `economics.block_cadence_seconds` ~3s, timeout_commit-driven | protocol-spec.md:132-134; parameters.md:84 | Not a chain parameter anywhere; proof scripts hand-sed `timeout_commit = "1s"` (`scripts/e2e-loop.sh:72` etc.); launch-genesis.sh does not set it (CometBFT default ~5s) | PARTIAL — cadence exists only as CometBFT config, never surfaced as the named lever |
| N10 | Chain-id `dreamtree`, no suffix ever | protocol-spec.md:745 | `scripts/launch-genesis.sh:27` `CHAIN_ID=${CHAIN_ID:-dreamtree}` — CONFORMS. **But** `deploy/anchord.service:14` still ships `ANCHORD_CHAIN_ID=dreamtree-devnet-1` | PARTIAL — stale deploy artifact |
| N11 | IBC "provides cross-chain messaging (relevant to settlement)" | protocol-spec.md:101 | No IBC module in app.yaml | DEFERRED-BY-SPEC (spec only says the framework provides it) |

## 2. Identity

| # | Spec claim | Spec | Build evidence | Class |
|---|---|---|---|---|
| I1 | Human-rooted identity via federated KYC; key is an artifact, not the identity | protocol-spec.md:140-151 | No identity module; no KYC/verification concept on chain. `subject` fields are free-form strings (`proto/dreamtree/attest/v1/types.proto:41-43`) | DEFERRED-BY-SPEC (spec :54 "provider integration open") — but see I2, which is NOT honest-absent |
| I2 | cred ladder: "Unverified public source → cred 0. KYC-verified-unattested → baseline. Verified-with-history → full R" | protocol-spec.md:239 | `x/reputation/keeper/standing.go:35-58` — **every** bech32 address gets `baseline_kyc = 1.0` standing; `credOf` (:63-65) = StandingOf. There is no verified/unverified distinction, so an unverified address gets baseline cred (spec says 0), can attest, report outcomes, and endorse with weight | **CONTRADICTS** — the cred ladder's bottom rung is unimplementable without identity, and the build silently gives everyone the middle rung |
| I3 | Recovery by KYC re-verification | protocol-spec.md:150 | Nothing on chain (wallet-layer; keys held by owner per seed-atom-conformance.md:167-168) | DEFERRED-BY-SPEC |
| I4 | Pre-population: records for every human/business, unsigned-public discount, ratify-on-claim | protocol-spec.md:154-156 | Data corpus pre-population exists (genesis carries ~11.7M atoms as seeds batches, `scripts/launch-genesis.sh:75-79`), but no entity/person pre-population, no unsigned discount mechanism, no claim/ratify flow | PARTIAL — the discount lever is also flagged "should be promoted" in parameters.md:151, never done |
| I5 | DID method `did:webvh`; wallet DID histories anchor into the chain | protocol-spec.md:149,519 | `subject` on batches/attestations carries DIDs as opaque strings (`proto/dreamtree/seeds/v1` Batch.subject); no DID verification or binding. seed-atom-conformance.md:211-212 names S2 "roots DID→address binding credential" | DEFERRED-BY-SPEC (S2 roadmap) |

## 3. Work & Reputation (x/attest + x/reputation)

### 3a. Attestation machinery

| # | Spec claim | Spec | Build evidence | Class |
|---|---|---|---|---|
| W1 | Four canonical proof types O/R/U/P | protocol-spec.md:174-181 | `proto/dreamtree/attest/v1/types.proto:13-21` — all four present; plus OUTCOME (=spec's `dt.outcome.*` subclass, :292) and ENDORSEMENT | CONFORMS (OUTCOME as ProofType instead of `data_type = dt.outcome.*` is a representation choice; ENDORSEMENT as a 6th proof type is a build addition — see §9-B8) |
| W2 | S(att,t) = R × specificity × type_weight × decay × (1−refuted_fraction) | protocol-spec.md:215-221 | `x/attest/keeper/projection.go:82-87,126-133` — exact shape | CONFORMS |
| W3 | type_weight: "O / R / U / P weighted differently" | protocol-spec.md:218 | `x/attest/params.go:17-21`: O=R=P=Outcome=1.0, **Use=0.5**. Values appear in NO spec/parameters.md registry; no `type_weight` row in the §Levers table (protocol-spec.md:646-668) | PARTIAL — mechanism conforms, but the actual weights are unregistered policy (§9-B4) |
| W4 | decay = proof-type base λ × domain obsolescence multiplier | protocol-spec.md:344-359 | λ values conform (`x/attest/params.go:23-26` = parameters.md:54-57). Obsolescence: x/attest uses a single global `ObsolescenceDefault = 1.0` (`x/attest/params.go:28`, projection.go:60) and does NOT read the per-domain `DomainConfig` that x/reputation maintains (`x/reputation/keeper/projection.go:76-87`) | PARTIAL — per-domain obsolescence shapes R but not attestation-strength S; spec :359 says "Set per domain node" |
| W5 | Work value V = [1 − Π(1 − S_i/S_max)] × demand_signal | protocol-spec.md:225-229 | `x/attest/keeper/projection.go:148-176`; comment :136 "demand_signal = 1.0 at v0" — demand term not implemented | PARTIAL — aggregation conforms; demand_signal absent (spec gives no v0 value; arguably deferred) |
| W6 | s_max normalizer — parameters.md holds it `null` (TBD) | parameters.md:51,113 | Build defaults `SMax = "10.0"` (`x/attest/params.go:30`) | **CONTRADICTS** parameters.md (canonical source says value-not-chosen; build chose 10.0 without registry update) — §9-B5 |
| W7 | refuted_fraction: outcomes reduce work strength | protocol-spec.md:220 | `x/attest/keeper/projection.go:91-123` paper-shape over REFUTED outcomes; PARTIAL outcomes weighted **0.5** (:107) — a constant found nowhere in spec/parameters.md | PARTIAL (mechanism conforms; 0.5 unratified, §9-B6) |

### 3b. R update law & standing

| # | Spec claim | Spec | Build evidence | Class |
|---|---|---|---|---|
| R1 | R(j,k,t) domain-indexed, derived from event log | protocol-spec.md:203-211 | Contribution log + read projection (`x/reputation/keeper/projection.go:104-152`; design ratified in x-reputation-design.md §1, owner-locked 2026-07-10) | CONFORMS |
| R2 | dR/dt = −λ_R·R + Σ σ·\|e\|·cred·relevance | protocol-spec.md:233-236 | Implemented as: contributions accrue via events (hooks.go), each decays at its bucket λ at read (projection.go:113-135), attenuated by relevance. Equivalent discrete form | CONFORMS (different but equivalent mechanism) |
| R3 | 2× negative asymmetry, contributor-only (per 2026-07-11 resolution) | protocol-spec.md:238,744 | `x/reputation/keeper/window.go:174-177` — 2× applied only at the contributor R-update; window integration symmetric (netVerdict :115-135) | CONFORMS |
| R4 | cred(source) recurses **2 hops**, then terminates; `reputation.cred_recursion_depth = 2` | protocol-spec.md:239; parameters.md:39,102 | `x/reputation/keeper/standing.go:63-65` — cred = StandingOf, which sums *settled* endorsement contributions. Multi-hop inheritance is emergent and geometric but has **no 2-hop termination**: hop 3, 4, 5… all contribute (0.25ⁿ). The `cred_recursion_depth` lever has no code binding at all | PARTIAL — geometric decay yes; explicit depth-2 termination no; named lever unimplemented |
| R5 | relevance(k,e.k): 70/40/15/3% up the 5-level taxonomy, cross-class 0 | protocol-spec.md:240; parameters.md:61-65 | `x/reputation/keeper/projection.go:40-55` and fixed-point twin `standing.go:16-31` — exact table | CONFORMS (values hardcoded in two switch statements, not read from params — see §9-B7) |
| R6 | λ_R(j,k) = base_λ_R / (1 + validated_attestation_volume) — long track record decays slowly | protocol-spec.md:361-365 | No volume modulation anywhere; each contribution decays at a fixed bucket rate (`x/reputation/keeper/projection.go:57-73`). `lambda_r_base` is null/TBD in parameters.md:42 | PARTIAL — the *value* is deferred-by-spec, but the *shape* (1/(1+volume)) is spec'd and absent |
| R7 | R decays toward baseline_KYC, not zero (settled) | protocol-spec.md:367; parameters.md:43 | Structural: R = baseline + Σ decayed contributions (`projection.go:139`); as contributions decay to 0, R → baseline | CONFORMS (emergent) |
| R8 | Validated outcomes durable ~25 yr; unvalidated attestations decay at proof-type rate | protocol-spec.md:369 | Outcome settlements land in `RATE_BUCKET_DURABLE_25Y` (`window.go:180,206`; λ=0.0277≈ln2/25, `x/reputation/params.go:17`); bets land in proof-type buckets (`hooks.go:17-30`) | CONFORMS |
| R9 | Saturation: effective_R = R if R≤S else S + k·log(1+(R−S)/S); tiers small/standard/large = 5/10/50; k=5 global; v0 default standard; per-node governance | protocol-spec.md:371-406 | `x/reputation/keeper/projection.go:93-101` exact formula; defaults standard=10, k=5 (`params.go:13-14`); per-domain via `MsgSetDomainConfig` (`x/reputation/keeper/msg_server.go:30-41`, proto types.proto:50-64) | CONFORMS |

### 3c. Review windows

| # | Spec claim | Spec | Build evidence | Class |
|---|---|---|---|---|
| V1 | **τ(M) = base_window × log(1 + M/threshold)** | protocol-spec.md:246; §Levers :652; parameters.md:40 ("τ(M) = base · log(1 + M/threshold)") | `x/reputation/keeper/window.go:23-38`: **τ = base × √(M/threshold)** (ApproxSqrt). Owner-decided 2026-07-10 (x-reputation-p2-review-windows.md:12-16, determinism: no transcendental in consensus) — but protocol-spec.md AND parameters.md still state the log curve; the spec decision log has **no entry** ratifying the √ substitution | **CONTRADICTS** the documents of record (ratified only in a sub-spec) — top-tier flag |
| V2 | `review_window_threshold` = **1.0** (parameters.md, which "wins") | parameters.md:41,104 | `x/reputation/params.go:27`: default **"4.0"**, comment "tuned up: √ has a fat tail near 0…" — a code-side retune never reflected into parameters.md | **CONTRADICTS** parameters.md (canonical values doc) |
| V3 | Every event enters a review window; trivial ≈ instant, large ≈ weeks; final ΔR is the integrated signal | protocol-spec.md:242-250 | All bets + outcomes enqueue as PendingEvents; EndBlock settles matured (`window.go:56-113`; app.yaml:11 wires reputation end_blocker). Trivial bets → τ≈0 → next block | CONFORMS (with V1/V2 curve caveats) |
| V4 | During τ the event accumulates corroboration, refutation, context | protocol-spec.md:249 | Outcome windows accumulate both directions (`hooks.go:106-117`); **bet** windows have accumulators but no path ever fills them (nothing corroborates a bet) | PARTIAL — spec says "every event"; only outcomes integrate |

### 3d. Outcomes, M_O, propagation

| # | Spec claim | Spec | Build evidence | Class |
|---|---|---|---|---|
| O1 | M_O = min(M_cap, β·S(att,t_issuance)·√cred(reporter)); β=1, M_cap=5·S | protocol-spec.md:270-300; parameters.md:49-50 | `x/reputation/keeper/hooks.go:72-85` exact formula; `S_issuance` frozen at attest time (`x/attest/keeper/msg_server.go:78-101`, proto types.proto:57-64); defaults β=1.0, cap_mult=5.0 (`x/reputation/params.go:24-25`) | CONFORMS |
| O2 | Uses S at issuance, not decayed-now | protocol-spec.md:280 | s_issuance snapshot, rational (standing × spec × type_weight, `x/attest/sissuance.go:33-37`) | CONFORMS |
| O3 | Multiple reports aggregate paper-shape, capped M_cap, never sum | protocol-spec.md:282-288 | `window.go:42-53` paperShapeAdd; corroboration path `hooks.go:109-115` | CONFORMS |
| O4 | Self-reports → cred ≈ 0 | protocol-spec.md:290 | `hooks.go:67-70` — reporter == target attestor hard-rejected (no effect at all) | CONFORMS (stronger than spec) |
| O5 | Outcomes ARE attestations; same window/cred/aggregation/time-horizon machinery | protocol-spec.md:292 | OUTCOME is a ProofType in the same log with the same indexes (`x/attest/keeper/msg_server.go:36-49,113-117`) | CONFORMS |
| O6 | Counter-outcome reverses original M_O across the chain AND 2× penalty to the wrong reporter | protocol-spec.md:294 | `window.go:215-252` — negation contributions via SourceIndex + 2× floored penalty; idempotent (Reversed set) | CONFORMS — **but** the negation contributions (:233-241) go through `addContribution`, NOT `applyFloored`; if the beneficiary's standing meanwhile decayed/was spent, the stored ledger sum can go net-negative (stored debt). See Z2 |
| O7 | Propagation: contributor ±M_O; direct attestors × attestation_weight; institution chain × 25% hop; **hiring/evaluating party × evaluation_factor** | protocol-spec.md:256-268 | Contributor: `window.go:171-181`. Co-attestors × coattestor_weight(0.25)×specificity: `propagation.go:30-57` + `window.go:198-199`. Endorsers × endorse_inherit(0.25): `propagation.go:70-86` + `window.go:200-201`. **Hiring/evaluating party: nothing** — no concept in code | PARTIAL — 3 of 4 propagation lines built; `evaluation_factor` MISSING (no spec-open flag for it) |
| O8 | Institution chain: multi-hop 25% up the endorsement chain | protocol-spec.md:265 | Endorser propagation is single-hop (only direct endorsers of the contributor are captured, `propagation.go:68-86`); endorsers-of-endorsers don't move | PARTIAL |
| O9 | Propagation weight "attestation_weight(a_i,c)" for direct attestors | protocol-spec.md:264 | Build uses `coattestor_weight = 0.25 × specificity` (`x/reputation/params.go:28`) — a parameter that exists in no spec/parameters.md registry | PARTIAL + §9-B2 |

### 3e. Plural truth, cold start, floor, shells, taxonomy

| # | Spec claim | Spec | Build evidence | Class |
|---|---|---|---|---|
| P1 | Plural truth: weighted consensus + explicit contradiction surfaced; buyers choose networks | protocol-spec.md:302-304 | Both-direction outcomes coexist and aggregate (V-pool/R-pool, `window.go:121-135`; refutedFraction in attest projection); no arbitration. Buyer-side network weighting: nothing (no query surface for "weight by network") | PARTIAL — structural non-arbitration conforms; surfacing/weighting UX absent |
| C1 | Cold start: R_initial = baseline_KYC + inherited endorsements + early ramp | protocol-spec.md:306-314 | baseline + endorsement inheritance work (`standing.go`, `hooks.go:142-163`); ramp (`coldstart.ramp_{factor,count}`) not coded | DEFERRED-BY-SPEC (spec :408-410 "Still open"; parameters.md nulls; launch-readiness.md:24-27 honest) |
| Z1 | Zero floor: every R move capped at current standing; no negative debt stored; recovery from 0 | protocol-spec.md:316-322,744 | `window.go:141-152` applyFloored on contributor, propagation, reversal penalty; `standing.go:54-56` clamps rational standing ≥ 0; `projection.go:93-95` floors float R at 0. Unit-tested (`window_test.go`) | CONFORMS on the settlement path — with two leaks, Z2 |
| Z2 | "no negative 'debt' is stored … recovery is genuinely from zero" | protocol-spec.md:318 | Two leaks: (a) reversal **negation** contributions bypass the floor (`window.go:233-241` uses addContribution directly), so a signer whose earlier gain decayed/was offset can be pushed to a net-negative stored sum; (b) `x/reputation/keeper/projection.go:92` comment says the quiet part: "The underlying contribution debt persists, so recovery is real work" — the float projection floors the *display* at 0 while the ledger sum can sit negative (decay-rate mismatch between positive and negative rows makes this reachable even without reversals) | **CONTRADICTS** (bounded, but the spec's "debtless" claim is not strictly true in the build; the projection comment even asserts the opposite) |
| S1 | Shell-institution resistance: institutional R only via meta-attestations from outside | protocol-spec.md:324-326 | Generic endorsement/outcome machinery gives the *substrate* (self-endorsement rejected, `x/attest/keeper/msg_server.go:56-58`; endorser liability, propagation.go:68-86); no institution concept, no meta-attestation distinction | PARTIAL |
| S2 | Meta-attestation layer pre-populated at v0 from public-recognition data | protocol-spec.md:326 | Nothing (no genesis attestations; launch-readiness.md:38-41 notes taxonomy/pre-seed as an open genesis-data task) | MISSING (spec asserts "is pre-populated at v0"; no open-flag) |
| T1 | 5-level taxonomy, seeded at v0 from LCC + ISCED + ONET/ISCO-08 | protocol-spec.md:328-338 | Path-string domains + prefix-depth relevance (`projection.go:26-36`) conform structurally; **no taxonomy seed loaded at genesis**, no validation of paths (any string accepted) | PARTIAL — mechanism yes; v0 seed data no (launch-readiness.md:38-41 admits it) |
| T2 | Domain-obsolescence multiplier per domain node (0.3/1/3) | protocol-spec.md:359; parameters.md:66-69 | `DomainConfig.obsolescence_multiplier` settable per node (reputation only, see W4); defaults standard=1.0 | PARTIAL (x/attest ignores it) |

## 4. Currency & Records — Photons & Seeds (the leaf model)

| # | Spec claim | Spec | Build evidence | Class |
|---|---|---|---|---|
| M1 | Seed = atom; one contribution = one seed = one photon = one unit of priced access | protocol-spec.md:745; seed-atom-conformance.md:9-12 | Leaf model: `MsgCommitBatch` registers new_count leaf-seeds under one root (`x/seeds/keeper/msg_server.go:45-192`); per-leaf resolution `resolve.go:15-55`; per-leaf pricing via SeedInfo (`reader.go`) | CONFORMS |
| M2 | Batching is a commitment strategy, not a unit change; `batch_root` retired as a kind; kind names the LEAF | protocol-spec.md:745 | Kind containing "batch_root" rejected (`msg_server.go:87-89`, ErrRetiredKind); CommitSeed = batch-of-1 sugar (:28-42) | CONFORMS |
| M3 | Convergence rule: re-observed atoms in leaf_count, never new_count; `new_count == 0` legal pure-convergence batch | protocol-spec.md:745 | `msg_server.go:90-93,114-126,183-190` — new_count==0 allocates no ids, mints nothing; leaf_count ≥ new_count enforced | CONFORMS |
| M4 | photons = seeds, counted over distinct atoms; ingestion is the ONLY mint | protocol-spec.md:440-452,745; parameters.md:140 | `x/photons/keeper/mint.go:18-58` mints exactly new_count photons per batch; no x/mint module in app.yaml; Minted sequence counts atoms (:44-48) | CONFORMS — with M6 caveat |
| M5 | "No slashing/evidence modules wired, so nothing can burn bonded photons — the peg holds structurally" | protocol-spec.md:745 | app.yaml has no slashing/evidence — TRUE for bonded photons. **BUT** the gov module account has `burner` permission (`app/app.yaml:33`) and SDK gov defaults `burn_vote_veto = true`; `launch-genesis.sh:92-96` sets deposits (10,000/50,000 photons) without touching burn params. A vetoed proposal burns its uphoton deposit → total supply < Minted count → **the photons = seeds peg can silently break via governance** | **CONTRADICTS** (latent) — the structural-peg claim has an unclosed burn path the decision log doesn't mention |
| M6 | "one photon exists per seed" — every recorded seed mints | protocol-spec.md:440 | Mint is gated on `MintableKinds = ["record","kg_claim"]` (`x/photons/params.go:5,17-28`; mint.go:26-28). A batch of any *other* leaf kind allocates seed ids but mints **zero** photons → seeds > photons. The mintable-kinds gate appears in no spec/parameters.md; its comment ("Batch roots aggregate many and do NOT mint") is pre-DT-18 stale | PARTIAL + §9-B3 — peg holds only for the two whitelisted kinds |
| M7 | Two minting streams: S (seed) to the **creator**; P (photon) to storer-validators | protocol-spec.md:442-448 | Photon → `StorerRewardRecipient` (dt-as-first-storer, ratified seed-atom-conformance.md:23-25): CONFORMS. Seed "minted to the creator": batch records `subject` (owning wallet) but the marketplace producer/payee is the **committer** (`x/seeds/keeper/reader.go:8-14`) — S2 ownership wiring is roadmap (seed-atom-conformance.md:211-212). Note launch-readiness.md:104-108 ("Single mint stream, not two… dtvp bonds validators") is stale pre-DT-18 text contradicting both spec and build | PARTIAL (subject recorded; creator-as-payee deferred to S2; stale doc flagged) |
| M8 | Mint-to-committer rejected; photon routes to storer | protocol-spec.md:745; seed-atom-conformance.md:23-25 | `mint.go:34-41` sends to params recipient, never the committer | CONFORMS |
| M9 | Supply-griefing bounded by `seeds.max_batch_new_count` (default 1,000,000); committer authorization the long-term gate | protocol-spec.md:745 | `x/seeds/params.go:8-13` + `msg_server.go:106-110`. Committer authorization: none (any address can commit; acknowledged deferred) | CONFORMS (cap); DEFERRED-BY-SPEC (authz). Note: `max_batch_new_count` is ratified in the decision log but absent from parameters.md's registry — §9-B10 |
| M10 | Base denom `uphoton`, 1 photon = 10⁶, display metadata in bank genesis; bond_denom = uphoton; voting power = whole photons; dtvp retired everywhere | protocol-spec.md:745 | `app/params/config.go:12-56`; `x/photons/keys.go:7-13`; `launch-genesis.sh:104-126` (metadata + dtvp denom guard); `grep -r dtvp` in code returns only retirement comments | CONFORMS |
| M11 | Storage rewards: endowment (Arweave-adapted), access cuts to storers, treasury fallback, Storj audits, seed-size cap, storage-cost oracle | protocol-spec.md:454-468 | Nothing (no endowment, no access_cut_to_storers split in purchase.go, no possession proofs) | DEFERRED-BY-SPEC (spec :468 "Still open" + :749) |
| M12 | Photon circulation: "Validators mint photons (block reward)" | protocol-spec.md:489 | No block reward exists (no x/mint; staking rewards = fees = 0, per seed-atom-conformance.md:148-149). Mint is per-ingestion | CONFORMS to §Monetary policy; :489 is stale spec-internal text (§8-D2) |
| M13 | Photon on/off ramps, regulatory posture | protocol-spec.md:508-511 | Nothing | DEFERRED-BY-SPEC (listed §Open) |

## 5. Marketplace (x/licenses)

| # | Spec claim | Spec | Build evidence | Class |
|---|---|---|---|---|
| K1 | Per-type market price N_a; protocol never sets it; never a lever; never in parameters.md | protocol-spec.md:429-436,674 | `TypePrices` keyed by data_type only (`x/licenses/keeper/msg_server.go:27-38`); N_a absent from parameters.md ✓. **But** v0 price discovery = `MsgSetTypePrice` by the gov **authority** (proto tx.proto:35-41) — the protocol (founder/governance) literally sets N_a today. Code comment admits "v0 market stand-in; real price discovery is a later mechanism" (types.proto:37-38). No spec/decision-log ratification of authority-set prices | PARTIAL, loud — pragmatic stand-in for a "never" invariant, unratified (§9-B1) |
| K2 | creator_equality_within_type invariant: p(c1,s,a)=p(c2,s,a)=p(c3,s,a) | protocol-spec.md:429,672; parameters.md:139 | Price is a pure function of data_type (`purchase.go:50`); no per-creator pricing possible structurally | CONFORMS |
| K3 | Buyer assembles swath by metadata query (age/context/place/credentials/R-threshold) | protocol-spec.md:472 | Swath = caller-supplied seed_ids (`tx.proto:25` "resolved off-chain by metadata query"); no on-chain metadata/R-threshold query surface | PARTIAL — purchase path yes; discovery/filter surface absent |
| K4 | Toll: 5% `economics.marketplace_toll`, buyer-side atop N_a; producers get N_a each | protocol-spec.md:474-480,591; parameters.md:76 | `x/licenses/params.go:11` 0.05; `purchase.go:66-71` toll atop totalSale, buyer pays N_a+toll; producers paid per-seed price | CONFORMS — caveat: toll (and tax) apply **only when a treasury recipient is configured** (`purchase.go:66-68`), a fail-open not in spec; launch sets treasury (`launch-genesis.sh:83`) |
| K5 | Non-exclusive access | protocol-spec.md:482 | Grants keyed (buyer, seed_id) (`purchase.go:117-121`); any number of buyers | CONFORMS |
| K6 | Time-bound: `access_duration_days = 1`; expiry then re-buy | protocol-spec.md:483; parameters.md:78 | `params.go:12`; `purchase.go:28-29` expires = now + days·86400; AccessGrant.expires_at recorded | CONFORMS (as a record; nothing yet *serves* bodies to enforce expiry against — S4) |
| K7 | R never prices the person; R as buyer filter + verified info only | protocol-spec.md:498-500 | Pricing path never touches reputation (purchase.go has no rep import) ✓; R-threshold buyer filter not built (see K3) | CONFORMS (invariant) / PARTIAL (filter) |
| K8 | Payments "flow in stablecoin" (§Access marketplace bullets) | protocol-spec.md:569 | Payments in uphoton (`purchase.go:74-113`) | CONFORMS to §Settlement/:496 (no stablecoin); :569 is stale spec-internal text (§8-D3) |
| K9 | License registered on chain; access mediated by cryptographic primitive; decay by expiration | protocol-spec.md:571 | Grant registered ✓, expiry recorded ✓; cryptographic mediation (TEE/PRE) — nothing (S4, seed-atom-conformance.md:215-216) | PARTIAL — registry yes, enforcement deferred (spec :58 "cryptographic primitives open") |

## 6. Records layer

| # | Spec claim | Spec | Build evidence | Class |
|---|---|---|---|---|
| D1 | Data lives in wallets; per-wallet append-only chains; sharding; replication lever | protocol-spec.md:517-525 | Not built; bodies in producer stores behind source_ref (interim stated in seed-atom-conformance.md:213-214); `storage_replication_factor` null in parameters.md:83 | DEFERRED-BY-SPEC (S3) |
| D2 | Size discipline: small records on-chain, large artifacts as content-addressed blobs | protocol-spec.md:526 | Only commitments on chain (512-byte cap) — stronger than the spec's minimum; blob layer absent | PARTIAL (safe direction) |
| D3 | C2PA integration for non-text artifacts | protocol-spec.md:528 | Nothing (`grep -ri c2pa x/ app/ cmd/` = 0) | MISSING — spec asserts "the protocol integrates with C2PA" with no open-flag (roadmap doesn't name it either) |
| D4 | Multi-key architecture (identity/rotation/data-master/DEKs/re-encryption/license-signing/assertion) | protocol-spec.md:530-540 | Nothing on chain (single tx-signing keys only) | DEFERRED-BY-SPEC-ish — spec :57 "storage primitives open" covers it loosely; no roadmap item names the key architecture. Borderline MISSING |
| D5 | Access primitive: TEE compute-to-data anchor + output minimization + PRE (tACo) | protocol-spec.md:542-547 | Nothing; named S4 (seed-atom-conformance.md:215-216) | DEFERRED-BY-SPEC |

## 7. Access & Economics

| # | Spec claim | Spec | Build evidence | Class |
|---|---|---|---|---|
| A1 | Four hard rules (no resale / no 3rd-party sharing w/o consent / no external ads / owner can leave with data) | protocol-spec.md:555-560 | Nothing on chain enforces or encodes any of the four (no use-declaration registry, no consent objects, no export path). Spec itself says no-resale is a stack (economics + TEE + watermark + contract), mostly off-chain | MISSING as on-chain objects (the spec's own §Relationship :636 claims "enforcement is cryptographic via license-mediated access on the chain, not policy" — the build has grants but zero enforcement, so that claim is aspirational) |
| A2 | Declared internal use, machine-readable use declarations, logged/auditable reads | protocol-spec.md:562-564 | Nothing | MISSING (no open-flag in spec) |
| A3 | Issuer restrictions on VCs honored at protocol layer | protocol-spec.md:573 | Nothing (no VC objects on chain) | MISSING (silent gap; VC layer never reached the chain design docs) |
| E1 | Flow 1 — producer compensation: "one photon per seed sold, fixed (the 1:1 invariant)" | protocol-spec.md:581 | Build pays per-type N_a (`purchase.go:50-57`) — conforming to the *ratified* per-type model (:429-436, decision log :723). :581 is stale pre-v0.3.0 text | CONFORMS to ratified model; spec-internal drift (§8-D4) |
| E2 | Flow 2 — marketplace toll "(the 30,000 P in the worked example)" | protocol-spec.md:583 | Worked example says 15,000 P at 5% (:474-480); :583 still says 30,000 (the old 10% draft). Code: 5% | CONFORMS (code); spec-internal drift (§8-D5) |
| E3 | Flow 3 — value-creation tax levied **"when work is issued onto the chain (a credential signed, an attestation made, a record verified)"** | protocol-spec.md:585 | Build levies it at **sale time**, producer-side, on marketplace revenue (`purchase.go:92-98`; proto types.proto:25-27 "a cut on the producer's sale revenue"). `MsgAttest` and `MsgCommitBatch`/`CommitSeed` carry **no fee at all** (gas is 0uphoton at launch). Same fund, different taxable event — issuance is free, sales are taxed | **CONTRADICTS** — the tax point diverges from the spec with no ratifying entry |
| E4 | value_creation_tax rate: spec §Levers 1.5% (:664) vs parameters.md 0.005 (0.5%, owner 2026-07-11, :77) | protocol-spec.md:664; parameters.md:77,129 | `x/licenses/params.go:14` = 0.005 | CONFORMS (parameters.md wins by its own rule :5); protocol-spec §Levers table stale (§8-D6) |
| E5 | No ICO, no token sale; photons minted by validation, seeds by creation, neither sold into existence | protocol-spec.md:594 | No sale mechanism; genesis supply = the pre-existing corpus minted to dt-as-first-storer (`launch-genesis.sh:47-67`), ratified :745 | CONFORMS |
| E6 | Founder-set params at v0, governance evolves; gov built in from day zero | protocol-spec.md:587-589 | x/gov wired (app.yaml:45-47); every module's params authority = gov account; proven `scripts/gov-proof.sh` (launch-readiness.md:29-35) | CONFORMS |
| E7 | Free entry for individuals; institutional revenue subsidizes | protocol-spec.md:600 | Off-chain/product-layer; chain-side: gas 0uphoton at launch is consistent | N/A on chain (no contradiction) |

## 8. Spec-internal drift (fix the document, not the chain)

The spec contradicts *itself* in places where the code follows the ratified/later section.
Owner should reconcile the text:

- **D1** :103 "no staking token needed, preserving no-ICO" vs :745 photon-bonded validators (code follows :745).
- **D2** :489 "Validators mint photons (block reward)" vs :440-452 ingestion-only mint (code follows the latter).
- **D3** :569 "payments flow in stablecoin" vs :496/:130 no stablecoin, photons internal (code follows the latter).
- **D4** :581 "one photon per seed sold, fixed" vs :429-436 per-type N_a (code follows the latter).
- **D5** :583 "the 30,000 P in the worked example" vs the 15,000 P / 5% example at :474-480.
- **D6** §Levers :664 tax 1.5% vs parameters.md :77 0.5% (parameters.md wins; table stale).
- **D7** :56 architecture table still says "1 S-access = 1 P" (superseded by per-type pricing :723).
- **D8** :256 heading "Outcome propagation (both directions, 2× asymmetric)" vs the 2026-07-11 resolution (:744) that the 2× lives only at the contributor update.
- **D9** `launch-readiness.md:104-108` ("Single mint stream", "dtvp bonds validators") and `deploy/anchord.service:14` (chain-id `dreamtree-devnet-1`) are stale pre-DT-18 artifacts.

## 9. BUILD-NOT-IN-SPEC — unratified build decisions (the dtvp class)

Things the build carries as policy that `protocol-spec.md`/`parameters.md` never ratified.
Each needs either a decision-log entry + parameters.md row, or removal.

| # | Item | Where | Why it's policy |
|---|---|---|---|
| B1 | **Authority-set type prices** (`MsgSetTypePrice`) | `proto/dreamtree/licenses/v1/tx.proto:35-41`, `x/licenses/keeper/msg_server.go:27-38` | Spec invariant: the protocol *never* sets N_a (:432-436, :674). The v0 stand-in inverts that; self-described as temporary in a code comment only |
| B2 | `coattestor_weight = 0.25` | `x/reputation/params.go:28` | Substitutes for the spec's `attestation_weight(a_i,c)` (:264); value not in parameters.md |
| B3 | `MintableKinds = ["record","kg_claim"]` | `x/photons/params.go:5` | Gates the photons=seeds peg per kind; not in any registry; stale pre-DT-18 comment ("Batch roots… do NOT mint") |
| B4 | type_weight values O/R/P/Outcome=1.0, **Use=0.5**; Endorsement hardcoded 1.0 | `x/attest/params.go:17-21`, `x/attest/sissuance.go:20` | Spec names the lever concept (:218) but the §Levers table and parameters.md carry no `type_weight` rows |
| B5 | `s_max = 10.0` | `x/attest/params.go:30` | parameters.md:51 says `null` (TBD); build chose a value silently |
| B6 | PARTIAL-outcome weight **0.5** in refuted_fraction | `x/attest/keeper/projection.go:106-107` | Hardcoded constant; `dt.outcome.partial` semantics never quantified in spec |
| B7 | Relevance table hardcoded (two copies) | `x/reputation/keeper/projection.go:40-55`, `standing.go:16-31` | Values match parameters.md:61-65 but are compile-time constants, not the governable `domain.attenuation.*` levers |
| B8 | `PROOF_TYPE_ENDORSEMENT` as a 6th proof type | `proto/dreamtree/attest/v1/types.proto:20` | Owner-leaned in x-reputation-design.md Q5, never ratified in the spec (which names four canonical types + outcome subclass) |
| B9 | `citationUpliftLambda = 1.0` const + the whole used_by/creation-credit-forward mechanism | `x/attest/keeper/projection.go:136-143`, types.proto:65-71; `docs/specs/citation-value.md` | Shipped 2026-07-12, devnet-proven, self-flagged TODO — but **protocol-spec.md has no decision-log entry for citation uplift at all**; the mechanism itself is unratified in the document of record |
| B10 | `max_batch_new_count = 1,000,000` | `x/seeds/params.go:13` | Ratified in the 2026-07-15 decision log (:745) ✓ but missing from parameters.md's canonical registry |
| B11 | `maxCoAttestors = 64` (and 2× for endorsers) propagation fan-out cap | `x/attest/keeper/propagation.go:16,56,85` | Hardcoded consensus-affecting bound (who moves on an outcome); event-surfaced on truncation but not a registered lever |
| B12 | `attest_bet_scale = 0.1` (the P1 "bet" magnitude) | `x/reputation/params.go:22` | The whole unvalidated-bet-moves-R mechanic quantifies σ·\|e\| in a way the spec leaves abstract; value in no registry |
| B13 | `lambda_endorsement = 0.08` | `x/reputation/params.go:21` | Endorsement-contribution decay rate; spec/parameters.md silent on endorsement decay |
| B14 | `review_window_threshold = 4.0` retune + √ curve | `x/reputation/params.go:26-27`, `window.go:23-38` | See V1/V2 — the loudest instance; parameters.md still says log + 1.0 |
| B15 | `MaxCommitmentBytes/MaxSourceRefBytes = 512` | `x/seeds/params.go:6-7` | Sensible plumbing bounds; not in parameters.md (low stakes) |
| B16 | Outcomes don't age as work-value inputs (λ=0) | `x/attest/keeper/projection.go:47` | Reasonable, but a decay-policy choice the spec doesn't make |
| B17 | Toll + tax silently skipped when `treasury_recipient` unset | `x/licenses/keeper/purchase.go:66-68,93-96` | Fail-open on an economics invariant; spec assumes the toll always applies |
| B18 | Gov deposits 10,000/50,000 photons; SDK default `burn_vote_veto=true` untouched | `scripts/launch-genesis.sh:92-96`; `app/app.yaml:33` | Deposit sizing is owner-adjustable per conformance doc ✓; the *burn* default is the M5 peg hole — needs an explicit decision (route to treasury, or document the deviation, exactly as :745 demands for slashing) |

---

## 10. Tallies

Counting the classified rows above (§1–§7; §9 items counted once under their §1–7 row where cross-referenced, else as BUILD-NOT-IN-SPEC):

| Class | Count | Rows |
|---|---|---|
| CONFORMS | 33 | N1-3, N5, N8, W1, W2, R1, R2, R3, R5, R7, R8, R9, V3, O1-O6 (O6 caveated), M1-M4, M8, M10, M12, K2, K4-K6, K8, E1, E2, E4-E6, Z1 |
| PARTIAL | 19 | N9, N10, I4, W3, W4, W5, W7, R4, R6, V4, O7, O8, O9, P1, S1, T1/T2, D2, K1, K3/K7, K9, M6, M7 |
| MISSING | 5 | S2, D3, A1, A2, A3 (D4 borderline) |
| DEFERRED-BY-SPEC | 12 | N6, N7, N11, I1, I3, I5, C1, M9(authz), M11, M13, D1, D5 |
| **CONTRADICTS** | **6** | **I2, V1, V2, W6, M5, E3, Z2** (7 rows; W6/V2 same root cause — parameters.md not honored) |
| BUILD-NOT-IN-SPEC | 18 | B1–B18 |

(Exact counts depend on row-splitting; the classes and members are what matter.)

## 11. Top 10 most consequential gaps (triage order)

1. **V1/V2 — review-window curve and threshold contradict the documents of record.** Code: τ = base·√(M/threshold), threshold 4.0 (`x/reputation/keeper/window.go:23-38`, `params.go:26-27`). Spec :246/:652 and parameters.md:40-41 still say log + 1.0. The √ substitution was owner-decided in a sub-spec (p2 doc) but never landed in protocol-spec.md's decision log or parameters.md — parameters.md is by its own rule the value source of truth and it is wrong on both curve and value.
2. **M5/B18 — the photons=seeds peg has an unclosed burn path.** Decision log :745 claims nothing can burn photons; gov has burner permission (`app/app.yaml:33`) and default `burn_vote_veto=true` will burn a vetoed proposal's 10,000-photon deposit. Needs slash-to-treasury-style handling or a documented deviation.
3. **E3 — value-creation tax is levied at the wrong event.** Spec :585 taxes work *issuance* (attestations, credentials); build taxes producer *sale revenue* (`x/licenses/keeper/purchase.go:92-98`). Different economics (issuance free, sales double-touched), unratified.
4. **I2 — no verified/unverified distinction: every address gets baseline cred.** Spec's cred ladder (:239) puts unverified sources at 0; `standing.go:35-58` gives any bech32 address baseline 1.0, so unverified reporters/endorsers carry weight the spec says they must not. Bounded today only by the solo-operator reality.
5. **B1 — authority-set type prices vs the "protocol never sets N_a" invariant.** The v0 stand-in (`MsgSetTypePrice`) inverts a fixed-forever invariant (:432-436) with only a code comment acknowledging it. Needs a ratified interim-mechanism entry with an exit path.
6. **Z2 — the "no stored debt" floor claim leaks.** Reversal negations bypass `applyFloored` (`window.go:233-241`), and the float projection comment (`projection.go:92`) outright says "contribution debt persists, recovery is real work" — the spec's §floor-is-zero (:318) says the opposite.
7. **O7 — outcome propagation is 3 of 4 lines: the hiring/evaluating party (`evaluation_factor`) doesn't exist**, and the institution chain is single-hop (O8). The spec's shell-institution decay story partially rests on these.
8. **B9 — citation uplift (used_by + citationUpliftLambda) is a shipped consensus-adjacent mechanism with zero presence in protocol-spec.md.** It has its own sub-spec and proof script, but the document of record has no decision-log entry — exactly the dtvp failure mode (entered the build without entering the spec).
9. **M6/B3 — MintableKinds silently forks the peg per seed kind.** Non-whitelisted leaf kinds allocate seeds without photons (seeds > photons). Either every leaf kind mints, or the kind gate gets ratified with its peg implications stated.
10. **W6 + the unregistered-lever cluster (W3/B2/B4/B5/B12/B13/B14).** parameters.md claims to be "the canonical source of truth for every tunable lever," but the build runs ~8 policy values it doesn't carry (s_max=10, type weights, coattestor_weight, attest_bet_scale, lambda_endorsement, review threshold, mintable kinds, partial-outcome 0.5). A single parameters.md reconciliation pass would close most of this comb's noise.

---

*Companion stale-doc fixes (no code change): §8 list — spec body text at :103/:489/:569/:581/:583/:664/:56/:256, `launch-readiness.md` dtvp-era paragraphs, `deploy/anchord.service` chain-id.*
