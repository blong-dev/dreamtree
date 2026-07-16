# DreamTree Protocol — Specification

*Started 2026-05-21. This is the protocol layer that the wallet sits on. Living doc; decisions land here as they're made. Unanswered questions are flagged on purpose, not skipped.*

*Companion docs: [`wallet-v0.md`](./wallet-v0.md) (the buildable Phase-v0 wallet that backs Telekora now — start here for building, not the full protocol), [`data-types.md`](./data-types.md) (the `data_type` registry — the ontology Telekora's tool contracts type against), [`parameters.md`](./parameters.md) (canonical source of truth for all tunable lever values), [`wallet-spec.md`](./wallet-spec.md) (the user-facing wallet on this protocol), `/home/b/quorum/dreamtree/VISION.md` (mission), `/home/b/quorum/dreamtree/manifesto/` (philosophy), `/home/b/quorum/philosphy/` (the academic Attribution-as-Conservation framework, especially §4.3, §4.6, §7, §9), `/home/b/quorum/ARCHITECTURE.md` (DreamTree's place in the four-surface frame).*

---

## Status

- **Scope expansion (2026-05-21)**: the wallet (DTW) is the user-facing layer of a larger protocol. The protocol provides chain-anchored identity, attestation-as-work consensus, license-mediated data access, and a marketplace for owner-directed data sale. The wallet alone is not enough to deliver what the philosophy promises; the protocol is.
- **Manifesto integrated (2026-05-21)**: this spec inherits and operationalizes the commitments in `/home/b/quorum/dreamtree/manifesto/` (Parts I–V) and `/home/b/quorum/philosphy/PRINCIPLES.md` (the ten principles). Where the manifesto names a stance — steward-ownership, AGPL with dual licensing, movement-not-empire, no VC, no ICO, free entry — this spec carries it forward without renegotiation.
- **Treat DreamTree as its own entity.** Despite Blong Enterprises being the current operating umbrella, the protocol is designed *for* DreamTree's mission, not for any consumer surface's convenience. Telekora, Cosmo, HometownWire, and any future surface are consumers of the protocol on equal terms.
- **Vision is settled in shape; implementation specifics are mostly open.** The decisions below land as conversations converge.

---

## Thesis

> **We aren't replacing the real world. We are mimicking it.**

Reputation already exists. Harvard's signature on a chemistry credential carries Harvard's weight, not the chain's. A boss vouching for an employee stakes their own reputation. Institutions confer authority because of the centuries of behavior that built them, not because of the medium they sign in. The protocol's job is to *transport* this existing reputation into a durable, portable, verifiable form — and to record its decay or appreciation over time as reality validates or refutes the attestations.

This collapses three problems most chains struggle with:

- **The chicken-and-egg trust bootstrap** — solved because the protocol piggybacks on existing reputation networks rather than trying to generate trust from scratch
- **The "what is the proof-of-work for" problem** — solved because the work being proven (signing a true attestation against your reputation) is intrinsically valuable, not heat dissipation
- **The gaming problem** — solved because a scammer accumulates many low-value signatures bound to a decaying reputation, not many high-value tokens

The protocol is **record, not financial**. Bad hires cost you on chain because the record persists, not because there's a doubling penalty. Location- and economy-independent: an excellent teacher in the Philippines who shapes thousands of lives earns the same credit as one in Switzerland who does the same. DreamTree does not care where you are.

---

## Design Heuristics

1. **Mimic reality, don't replace it.** When a protocol decision has analogs in existing reputation networks (universities, employers, professional bodies), the analog is the starting point. Novel mechanics are added only when reality has a gap the protocol can fill (durability, portability, verifiability).
2. **Movement, not Empire.** The data-sovereignty movement has years of work behind it. MyData Global, the Solid Project, W3C Verifiable Credentials, AT Protocol (20M+ users via Bluesky), C2PA, IETF SD-JWT, W3C BBS+ — all have standards and communities. Adopt existing standards by default. Contribute upstream where they fall short. Fork only when core principles require it. The ego move is to build everything from scratch; the strategic move is to strengthen the ecosystem that strengthens us.
3. **Cryptography over policy, where the gap matters.** The four hard rules (§Access) are cryptographically enforced for marketplace transactions; reputation dynamics are partly social and partly mechanical. Don't try to crypto-enforce things that are inherently social, but don't policy-enforce things that should be cryptographic.
4. **Permanent record, user-owned.** All non-financial data about a person belongs to them. Financial data has its own regulatory regime; we stay out of that lane.
5. **Algorithm agility from day zero.** Post-quantum signature schemes (Dilithium, Falcon, SPHINCS+) are now standardized. The protocol supports multiple signature algorithms in parallel and migration between them, because the mission's time horizon outlasts current cryptographic assumptions.
6. **Honest about being centralized at launch.** Door #3 trajectory (§Network) — operationally one validator at v0, progressive decentralization as the network matures. Trust DreamTree at v0 → trust the math at v3. No decentralization claim that hasn't been earned.
7. **Parameters are levers, not constants.** Every numeric value in this spec is a stand-in. The engineering goal is *clarity over the complete set of levers and their effects*, not premature precision on their values. Founder sets stand-ins at v0 because the system must run; governance and market signal discover where each lever settles. The complete lever set lives in §Levers. (Exception: producer compensation and per-type data prices are never levers — they're market-determined.)
8. **The protocol informs; the market prices.** We add verified information to the exchange (provenance, reputation, type, validation); we do not reinvent supply and demand. Value is marginal — discovered by the market at the margin, never set by a protocol formula. The protocol's job is to kill the information asymmetry that makes markets misprice human contribution (Akerlof in reverse), then get out of the way. This unifies plural-truth, don't-dictate-compensation, market-discovered levers, and per-type pricing — all the same move.

---

## Architecture Overview

The protocol has six layers. Each is sketched below, then expanded in its own section.

| Layer | What it provides | Status |
|---|---|---|
| Network | Chain, validators, consensus | Direction settled (open chain, progressive decentralization); design open |
| Identity | Human-rooted DIDs, KYC-bound recovery | Direction settled; provider integration open |
| Work & Reputation | Attestation-as-work consensus; signature weight | Direction settled; decay/appreciation math is the hardest open problem |
| Currency & Records | Photons (P, fungible currency; the bond denom since 2026-07-15) + Seeds (S, non-fungible records); access market-priced per type (N_a) | Two-token model settled; leaf model live; issuance/ramp economics open |
| Records | On-chain encrypted user data | Direction settled; storage primitives open |
| Access | License-mediated reads; marketplace | Four hard rules settled; cryptographic primitives open |

## Who This Is For

The protocol is not abstract. From the manifesto's Part III, the human stakes are concrete:

- The **22-year-old who graduates with debt** and a credential that does not distinguish them from a hundred thousand others. They have skills, but no way to prove them. They have stories, but no way to make them visible.
- The **mid-career professional displaced by automation**. Their experience is valuable, but it lives in performance reviews locked in a former employer's database. They start over, invisible.
- The **creator whose work was scraped**, incorporated into a model, regenerated without attribution. Their livelihood erodes. They have no recourse.
- The **mother whose eighteen years of guidance** shaped a child's neural substrate. GDP registered none of it. The contribution was real; the receipt was never written.
- The **maintainer keeping critical infrastructure running**. Their work is noticed only when it fails. While it works, the contribution is invisible.
- The **teacher in the Philippines** who shaped thousands of lives. The protocol does not care where they teach. Their record carries the same weight as the same work done anywhere else.

These are contributors who have been systematically denied recognition for what they contributed. The protocol exists to restore the link between creation and capture — not by redistribution, but by preservation. The contribution was always real. The receipt is what was missing.

---

## Network

**Door #3: open public chain, centralized at launch, progressive decentralization.**

At v0, DreamTree operates the only validator. The protocol is open from day zero — anyone can audit the code (AGPL), and the spec for joining the validator set is public. The schedule for opening validation is published, not deferred.

**Validator-set evolution (proposed)**:
- **v0**: DreamTree operates solo. Honest framing in all user-facing copy: "DreamTree currently operates this network. The protocol is open; here's our roadmap to opening validation."
- **v1**: Federated. A small set of mission-aligned validators (academic institutions, civic nonprofits, identity-verification partners) join. Validator selection criteria are governance-defined.
- **v2**: Permissioned-open. Any entity meeting public criteria can join the validator set. Stake required (reputation or capital, TBD).
- **v3**: Open. Pure protocol participation; criteria are encoded, not gatekept.

**Internal framing at v0**: "really over-engineered database." Externally, no claims of decentralization that haven't been earned. The protocol is *designed for* openness from day zero; that's the only honest decentralization claim until the validator set actually diversifies.

**Ecosystem alignment.** The protocol does not invent identity, credentialing, or content provenance from scratch. It composes existing standards: W3C Decentralized Identifiers (DIDs), W3C Verifiable Credentials 2.0 (May 2025), the AT Protocol for federated identity primitives, MyData Global principles for personal data governance, Solid Project patterns for portable storage, C2PA for content provenance, IETF SD-JWT and W3C BBS+ for selective disclosure. The novel work — the attestation-as-work consensus, the on-chain encrypted records, the license-mediated marketplace — extends what these standards don't yet cover. Forks happen only when core principles require it.

**Chain choice: RESOLVED (2026-05-22) — DreamTree-native chain.** We build our own, not anchor to an existing one. Decision made with eyes open about operational cost.

Justification: the consensus mechanism (attestation-as-work under reputation stake, §Reputation Dynamics) is novel enough that no existing chain expresses it natively. On-chain encrypted record storage, the license registry, and the progressive-decentralization model are specific enough that bending an existing chain to fit would cost more than owning the stack. Owning the chain means owning the protocol's destiny.

**Tension with "Movement, not Empire" (heuristic #2), acknowledged.** We resolve it by layer: *inherit* existing standards where they fit, *build* custom only where the protocol's novelty requires it. Fork only where core principles require.

**"Rolling our own" is a spectrum — we are not reimplementing TCP:**
- **Inherit (don't rebuild)**: p2p networking, gossip, cryptographic-primitive libraries, post-quantum signature libs, BFT ordering plumbing, DID/VC formats. App-chain frameworks (Cosmos SDK + CometBFT, or Substrate) provide the plumbing.
- **Build (the novel core)**: the reputation state machine (R per signer × domain), the attestation-as-work consensus rules, the on-chain encrypted record store with sharding + availability proofs, the license registry, the coin ledger.

**Framework: RESOLVED (2026-05-22) — Cosmos SDK + CometBFT.** App-chain framework, Go. CometBFT provides BFT consensus, p2p, instant finality, validator-set management. The ABCI boundary gives the consensus/value separation for free. The `gov` module gives on-chain parameter governance (the path for moving levers in `parameters.md` from founder-set to community-voted). IBC provides cross-chain messaging (relevant to settlement).

**Consensus layer vs. value layer: RESOLVED (2026-05-22) — separate.** CometBFT validators order events and provide finality (the consensus layer). The reputation/attestation/record/license/coin logic runs in the ABCI app (the value layer). Reputation determines what attestations are *worth*, never who orders blocks — this decouples reputation-capture from chain-capture. The Door #3 validator progression maps onto CometBFT's validator set: solo (v0) → federated (v1) → permissioned-open (v2) → open (v3). Validator Sybil-resistance is *permissioning* through v2; since 2026-07-15 the validator bonds genesis-corpus **photons** (`bond_denom = uphoton` — see the seed=atom decision-log entry; no-ICO preserved: nothing is sold into existence). Economic staking for a permissionless set is a v3-only question, deferred.

### Block structure & the data fabric (RESOLVED 2026-05-22)

The consensus chain is linear (1D); the data is a **fabric** (2D), indexed by wallet.

- **Consensus block (CometBFT, linear)** holds *commitments and small transactions* — seed commitments (hash + metadata, **not bodies**), attestations, outcomes, license purchases, photon transfers, identity ops. It commits to the per-wallet roots via a data-availability root in the header. Small (MB-scale), fast, instant finality.
- **The data fabric (wallet-indexed)** holds the encrypted seed *bodies*. Each wallet has its own append-only chain — its `did:webvh` history extended into its records/seeds. The global structure is a weave of per-wallet chains, cross-linked by attestations and licenses (A attests B → a thread between their chains). This is the "second dimension": cascading, reconstructable, not a single line. Block-lattice lineage (Nano per-account chains); Merkle-DAG lineage (the Great Library Provenance Ledger).
  - **Two lessons from Nano (which proved the per-account-chain model in production):** (1) feelessness killed Nano's node incentive and left it centralized — *we avoid this* because photons reward storage+validation; (2) Nano makes every node store every chain (doesn't scale) — *we avoid this* with sharding + the participation spectrum. Nano confirms both are non-negotiable.
- **Why the split is forced**: you cannot put millions of encrypted record bodies into BFT consensus blocks — they'd be hundreds of MB and every validator would store everything forever. Bodies live in the fabric, sharded; the consensus chain orders events and commits to fabric roots. This refines (does not reverse) "the protocol stores the data" — the data lives in a wallet-indexed fabric distinct from, but committed-to by, the linear consensus chain.
- **Data model**: record/object-based, wallet-partitioned. No global bundle of data types — each wallet holds its own.

### Participation: unified validator-storer, a spectrum not a caste (RESOLVED 2026-05-22)

Participation is a spectrum, not a caste — but the deep research (2026-05-22) sharpened *where* on the spectrum each device sits. Two honest constraints:
1. BFT consensus is O(n²); CometBFT practically caps ~100–200 active consensus validators — "validate on your phone" cannot mean "be a consensus validator."
2. **A phone cannot be a durable storage provider.** The hard part isn't storing a small shard — it's proving non-outsourceable possession *over time* plus uptime and durability. Phones fail on churn, connectivity, and disk reliability, not on compute. (Filecoin's PoRep — 128–192 GiB RAM, GPU, 32 GiB sectors — is the heavyweight extreme; not our model.)

The resolution, by tier:
- **Phones / light nodes**: **data-availability sampling** (Celestia-style — pull headers + commitments, randomly sample ~30 chunks for ~99.9% availability confidence) + verifying other nodes' possession proofs + *optional ephemeral cache/serving*. Near-passive, genuinely phone-class. **Not** slashable durable storage. Note: DAS proves data is *available now*, not *durably stored* — it's the light-verification primitive, paired with the storage tier's possession proofs.
- **Durable storage nodes**: always-on commodity boxes (not GPU rigs), gated on **uptime/durability bonds, not raw power**, using **Storj-style random-audit possession proofs** (no SNARKs, no sealing — return derived stripes; trivial if you hold the data, infeasible if not). This is the reference, not Filecoin.
- **Consensus validators**: the bonded always-on subset that also runs CometBFT.

So the phone dream survives in its important half — light validation, near-passive — while durable storage is honestly a commodity-always-on role.

### Settlement (RESOLVED 2026-05-22)

Two meanings were conflated. **State settlement**: DreamTree is a sovereign L1; CometBFT finalizes its own state; it does *not* settle to any external chain (rejected the rollup framing). **Monetary settlement**: internal in photons (on-chain, instant via CometBFT); fiat touches only the edges (on/off ramps). No external settlement layer; no stablecoin issuance.

### Block cadence

A lever (`timeout_commit`-driven). Cosmos norm 1–6s; stand-in ~2–5s. Reward is per-block but total issuance is governed by the photons-=-seeds rule (§Currency), not raw block count — cadence and issuance are separable.

---

## Identity

**Human-rooted via federated KYC providers.**

In standard crypto wallets, the key is the identity. Lose the key, lose the identity. In DreamTree, the human is the identity. The key is just an authentication artifact bound to a human verified by external attestation. This inverts the recovery problem: you can lose a key, but you can't lose being you.

**Mechanism shape**:
1. At registration, DreamTree requests verification from one or more identity providers (Persona, Onfido, Jumio, Trulioo, or others)
2. The provider verifies documentary identity + (optionally) liveness
3. The provider issues a signed attestation — ideally a W3C Verifiable Credential, when the provider supports it
4. DreamTree creates a wallet bound to that attestation
5. The DID method is `did:webvh` by default (per [`wallet-spec.md`](./wallet-spec.md) L1 Q1), with upgrade paths to other methods
6. Recovery is by re-verification: the user re-runs identity proof with a supported provider; matching identity → wallet recovered

**Privacy posture**: TBD. Two live options — DreamTree learns identity (simpler, more like standard KYC), or DreamTree receives a zero-knowledge attestation of personhood without details (privacy-preserving, more complex). Both are valid points on the trust/UX curve.

**Pre-population**: the dataset doesn't start empty. The protocol maintains records for *every human and every business publicly accounted for* — claimed or not. Public information (corporate registries, public-domain biographical data, openly-published credentials) is reflected into the dataset, with a constant discount applied to denote it's unsigned. When a person or institution claims their wallet, they take ownership of their record; previously-discounted entries can be ratified by their signature, raising the record's weight to claimed-and-verified.

This pattern mimics how credit bureaus pre-populate records about everyone without consent. DreamTree's version adds the user-ownership-on-claim layer. **Open question, flagged honestly**: this has privacy and consent implications even with the unsigned discount. Pre-population is powerful — it solves the cold-start problem and makes the platform useful from day zero — but it inherits some of the structural concerns of the surveillance economy it opposes. Mitigations to design: ingest only public information; let any subject claim and edit; honor takedown requests for incorrect pre-populated data; never pre-populate sensitive categories (health, beliefs, sexuality, etc.) without explicit opt-in.

**For undocumented humans**: fingerprint biometrics + partnerships with UNHCR or equivalent humanitarian-ID programs. Not perfect, but honest about the inclusion problem (per [`wallet-spec.md`](./wallet-spec.md) L2 Q5 — recovery is open about its limits).

---

## Work & Reputation

**The PoW is attestation under reputation stake.**

The work being proven is the act of attesting — signing a credential, vouching for an employment record, certifying a degree, verifying a fact — *as a real-world identified entity, with that entity's reputation at stake*. Three properties:

1. **Bootstrap by piggyback.** Existing reputation networks (universities, employers, professional bodies, governments) carry their reputation into the protocol the moment they sign. The protocol transports existing trust; it does not generate it.
2. **Work is intrinsically valuable.** Unlike hash-puzzle PoW, attestation work *creates* something useful (a portable credential, a verified fact, a reference). The energy isn't dissipated as heat; it's stored as durable provenance.
3. **Truth surfaces over time.** False attestations decay the attestor's signature weight as reality refutes them. The Akerlof dynamic (information asymmetry) works *for* the protocol, not against it — bad signatures lose value; good signatures accumulate.

This is consensus by reputation with cryptographic substrate. The math is Akerlof + Fukuyama + Ostrom mapped to a chain. The protocol's hardest design problem is the dynamics of decay and appreciation — see "Open Problems" below.

**The four canonical proof types** (from the Great Library Protocol; named in the manifesto's Architecture Part III):

- **Proof-of-Origin** — authorship attestation. The contributor or their steward signs that the work is theirs.
- **Proof-of-Rigor** — peer or institutional review attesting to methodology, accuracy, or quality.
- **Proof-of-Use** — a verifiable link when new work references, cites, or semantically depends on prior work.
- **Proof-of-Replication** — independent confirmation that a result is reproducible.

Each proof is a cryptographically signed object attached to the work in the Knowledge Graph and stored on the chain. Together they cover the lifecycle of any contribution: someone made it, someone vouches for its quality, someone built on it, someone verified it works.

**Reputation baseline at zero**:
A new attestor with no prior history isn't reputation-zero in absolute terms — they're situated in a web of pre-existing relationships:
- They were educated by humans whose degrees were signed
- They were hired by people with employment histories
- Their KYC verification itself confers a baseline (verified-human status)

A new manager's first attestation isn't worthless. It carries the signature of someone who has been certified as real, was employed by an entity with reputation, and is signing within a system that tracks consequences. As that manager attests more, two things happen: (a) the protocol accumulates a record of their judgments, and (b) reality validates or refutes those judgments over time. The protocol is not handing out reputation. It is *recording the bets people are making on each other* and letting those bets pay out or fail.

This makes "shell game" attacks costly. A fraudulent university issues high-volume credentials; without meta-attestations from real regulators or ministries, the university's signatures don't accumulate. Real signers don't get downstream payoff from associating with shell signers. The web filters itself, the way it does in reality, only faster because it's measured.

**Reputation propagation**: resolved — see Reputation Dynamics below. Endorsement inheritance is 25% at first hop with geometric multi-hop decay; cred(source) recurses 2 hops. Avoids both overweighting (everyone Harvard touches inherits Harvard's reputation) and underweighting (reputation never propagates).

---

## Reputation Dynamics

*The decay-and-appreciation math — the load-bearing mechanics of the Work & Reputation layer. First-cut design as of 2026-05-22. Parameters marked "lever" are starting points, explicitly tunable.*

### The five objects

| Object | Notation | Definition |
|---|---|---|
| Signer reputation | R(j, k, t) | Signer j's weight in domain k at time t. Domain-indexed, unbounded with logarithmic dampening past saturation. |
| Attestation strength | S(att, t) | Force of a single attestation now. |
| Work value | V(w, t) | A work's accumulated weight from attestations on it. |
| Coin value | C(c, t) | A coin's contextual value. |
| Reputation decay rate | λ_R(j, k) | How fast j's reputation in k atrophies without reinforcement. |

R is load-bearing; everything else derives from it.

### Attestation strength

```
S(att, t) = R(signer, domain, t)
          × specificity(att)
          × type_weight(att.type)          // O / R / U / P weighted differently
          × decay(t - issued)              // proof-type rate × domain obsolescence (see Time Horizons)
          × (1 - refuted_fraction(att, t))
```

### Work value (aggregation — paper shape)

```
V(w, t) = [ 1 - Π_i (1 - S_i / S_max) ] × demand_signal(w, t)
```

Diminishing returns on stacking weak attestations; strong corroboration approaches saturation; no double-counting.

### R update law

```
dR(j,k)/dt = -λ_R(j,k) · R(j,k,t)        // continuous decay toward baseline
           + Σ_events σ · |e| · cred(source(e)) · relevance(k, e.k)
```

- **Asymmetry (2×, lever)**: positive event Δ = +x; negative event Δ = −2x. Two equivalent good acts to recover from one bad one.
- **cred(source) recurses 2 hops.** Unverified public source → cred 0. KYC-verified-unattested → baseline. Verified-with-history → full R. Their endorsers contribute one further hop, then terminate.
- **relevance(k, e.k): domain attenuation.** Upward through the 5-level taxonomy: sub-specialty → specialty ~70%, → discipline ~40%, → field ~15%, → class ~3%. Cross-class spillover ≈ 0.

### Review windows (all signal queued by magnitude)

Every event enters a review window before final application:

```
τ(M) = base_window × log(1 + M / threshold)
```

Trivial events (a single citation): τ ≈ instant. Large events (fraud claim): τ ≈ weeks. During τ, the event accumulates corroboration, refutation, context. Final ΔR is the integrated signal over τ, not the raw claim. Prevents single-point character assassination while keeping small-stakes signal fast.

### Endorsement inheritance (25%, lever)

A endorses B → B inherits 0.25 × R(A, k) at hop 1. Multi-hop geometric decay: 25% / 6.25% / 1.56% / 0.39%. Saturates by hop 4.

### Outcome propagation (both directions; the 2× asymmetry lives only at the contributor R-update — resolved 2026-07-11)

Outcomes are themselves attestations — no central truth oracle. An entity with standing to observe an outcome attests it, staking its own R. When an outcome validates or refutes a contributor's work, the signal propagates up the attestation chain:

```
Outcome O on contributor c, magnitude M_O:
  Contributor c:    ΔR(c,k)   = +M_O   (validated)  |  -2·M_O  (refuted)
  Direct attestor:  ΔR(a_i,k) = ±M_O × attestation_weight(a_i,c)
  Chain (institution behind a_i): × hop_decay (25% chain)
  Hiring/evaluating party: ±M_O × evaluation_factor
```

Contributor moves most; everyone who staked R on them moves proportionally. This makes betting on people rewarding when right, costly when wrong — and makes shell institutions decay as their graduates underperform.

### Outcome magnitude `M_O` (resolved 2026-06-24)

The magnitude of an outcome event — what plugs into the propagation formulas above:

```
M_O = min(M_cap, β · S(att, t_issuance) · √cred(reporter))
```

Five clarifications make the formula behave correctly:

**1. `S(att, t_issuance)`, not `t_outcome`.** Use the attestation's strength **when it was made**, not its decayed strength when the outcome lands. The teacher's 1990 student-quality bet that pays out in 2026 should pay *big*, not tiny — that's the appreciation-compounding mechanic the framework promises. Using current strength would dilute it into nothing for any long-validated bet.

**2. Multiple outcome reports aggregate paper-shape, not sum.** If 10 reporters file the same outcome (the surgeon's case went badly), `M_O` should not 10×. Same aggregation as attestations:

```
M_O_total = M_cap · [1 − Π(1 − M_O_i / M_cap)]
```

Diminishing returns; Sybil-resistant by construction.

**3. Self-reports → `cred ≈ 0`.** A contributor reporting an outcome on their *own* work has cred near zero (the cred-recursion already gives no second-hop endorsement of yourself, but it's worth being explicit). Self-attestation never carries full outcome weight — Akerlof's lemons logic.

**4. Outcomes ARE attestations (a subclass), not a special path.** An outcome is an attestation with `data_type = dt.outcome.*` (validated / refuted / partial). It hits the review window `τ(M)`, the `cred(source)` recursion, the paper-shape aggregation, the time horizons — same machinery. Special *only* in what it does to R: triggers the `M_O` propagation chain (contributor + direct attestors + chain hops + hiring/evaluating party, per the propagation block above).

**5. Outcomes can themselves be refuted, and the 2× asymmetry recurses.** A counter-outcome from a higher-cred source — the medical examiner overruling the hospital, the appellate court overturning the trial court — does two things: it **reverses** the original `M_O` (un-applies it across the contributor + chain + hiring party) **and** hits the original reporter with a **2× R penalty** (their refuted outcome was itself a bad attestation). Without this, false outcomes are sticky and the math is irreversible.

**Stand-ins** (levers, see [`parameters.md`](./parameters.md)):
- `β = 1.0` — when the reporter is baseline-cred (KYC-verified but unaccumulated), `M_O ≈ S(att, t_issuance)`.
- `M_cap = 5 · S(att, t_issuance)` — a single outcome can multiply the original bet by up to 5×, no more. Protects against single-event reputation kills; persistent bad behavior still tanks R all the way to the zero floor (see [The floor is zero](#the-floor-is-zero-no-negative-debt)) — bounded, but bounded at nothing, not at a protected positive value.

Both tunable, governance-evolved.

### Plural truth

No single entity is "the" outcome authority. Multiple reputation networks coexist (e.g., AMA and integrative-medicine bodies both attest, each with their own R in their own domains). The protocol surfaces weighted consensus and explicit contradiction; buyers choose which networks to weight. The protocol surfaces evidence; it does not arbitrate worldviews.

### Cold start (zero-baseline)

```
R_initial(j,k) = baseline_KYC                           // verified-human floor
               + Σ inherited_endorsements(j,k)          // education/employment chain @ 25%
               + Σ early_validated_attestations × ramp  // ramp > 1 for first N attestations
```

Newcomers inherit context (taught by signed humans, hired by entities with history). Early validated work is amplified to clear the dead zone. Established signers' far-higher saturation cap means this isn't unfair to them.

### The floor is zero (no negative debt)

R in a domain lives on `[0, ∞)`. It can decay all the way to zero — you become unable to attest *in that domain* with weight — and reality revealing enough can drive it there. But **zero is the floor: R is never negative, and no negative "debt" is stored.** Zero talent is zero; there is nothing worse than nothing. Recovery is therefore genuinely *from zero* — one good act moves you back into positive standing, not against an accumulated hole you must first dig out of. (The 2× asymmetry still bites while you *have* standing to lose; it has nothing left to take at zero.)

This is what actually stops character assassination. A refutation crowd is bounded twice over: **paper-shape** caps the whole outcome pool at `M_cap` (100 million "he's a bad golfer" claims aggregate to `M_cap`, not 100 million), and the **zero floor** means the bounded blow can at most take you to zero, never into an unrecoverable negative. Both bounds are load-bearing; either alone leaves a hole.

The structural protection layers on top: KYC verification means a verified human can always **start fresh in a new domain**. Refutation in one specialty doesn't bar you from any other. No global civil death; domain-specific loss of standing, plus the always-available ground of being a verified human entering a new domain.

### Shell-institution resistance

Institutional R accumulates only through **meta-attestations** — attestations from outside the institution about the institution (accreditors, regulators, sustained validated graduates). A new institution starts at baseline; without meta-attestations its credentials carry only "an unverified entity signed this" weight. Real institutions accumulate meta-attestations passively. The meta-attestation layer is pre-populated at v0 from public-recognition data, discounted-until-claimed.

### Domain taxonomy (5 levels)

| Level | Granularity | Example |
|---|---|---|
| L1 | Class | Sciences |
| L2 | Field | Sciences/Life-Sciences |
| L3 | Discipline | Sciences/Life-Sciences/Medicine |
| L4 | Specialty | .../Medicine/Cardiology |
| L5 | Sub-specialty | .../Cardiology/Electrophysiology |

Seed at v0 from established taxonomies (LCC + ISCED + ONET/ISCO-08). Taxonomy evolution is a governance question.

### Time horizons (two clocks)

Two distinct decay processes; keep them separate.

**1. Attestation-strength decay** — how fast a single attestation's contribution to work value fades. The `decay(t - issued)` term in S(att, t):

```
effective_λ(att) = proof_type_base_λ × domain_obsolescence_multiplier(k)
```

Proof-type base rates (stand-ins; see §Levers):

| Proof | Base λ | Half-life | Why |
|---|---|---|---|
| Origin | ≈ 0 | permanent | authorship never becomes false |
| Replication | ~0.015/yr | ~45 yr | reproducibility is durable |
| Rigor | ~0.04/yr | ~17 yr | quality standards evolve |
| Use | ~0.08/yr | ~9 yr | citation relevance fades (matches patent-citation half-life) |

Domain-obsolescence multiplier (stand-in): foundational ~0.3×, standard ~1×, frontier-tech ~3×. A Proof-of-Rigor about React proficiency ages far faster than one about statistical reasoning. Set per domain node; governable.

**2. Reputation decay (λ_R)** — how fast a signer's standing fades without reinforcement:

```
λ_R(j, k) = base_λ_R / (1 + validated_attestation_volume(j, k))
```

A long validated track record decays slowly; a one-hit wonder decays fast. **Decays toward baseline_KYC, not zero** — reputation is a stock you keep, not a flow you must maintain. An established signer who goes quiet retains standing; they don't fall back to "just a verified human."

Validated outcomes contribute durable R (half-life ~25 yr stand-in); unvalidated attestations decay at the proof-type rate. Validation locks in standing — this is the appreciation-compounding mechanism. A teacher's 1990 assessment decays as an *attestation*, but once the student's career validated it, the teacher's R as a talent-judge took a durable bump that persists for decades.

### Saturation point — log-dampening on R (resolved 2026-06-24)

R is unbounded but log-dampened past a per-domain saturation point. Two-piece linear + log:

```
effective_R(R) =
  R                                       if R ≤ S
  S + k · log(1 + (R − S) / S)            if R > S
```

- `S` = `reputation.saturation_point` — where compression starts. **Per-domain** (mirrors `domain.obsolescence_multiplier`); each node in the 5-level taxonomy is tagged with one of three tiers.
- `k` = `reputation.dampening_k` — compression strength past `S`. Global (the *shape* of compression doesn't need to vary by domain; only the *threshold* does).

**Tiered stand-ins for `S`:**

| Tier | `S` | Use |
|---|---|---|
| `small` | 5 | niche / sparse domains; typical accumulation is low; saturation reached quickly |
| `standard` | 10 | default for most domains |
| `large` | 50 | hot / dense domains (medicine, AI, mature institutional ecosystems); lots of room before compression |

`k = 5` globally.

Concretely with the standard tier (`S = 10, k = 5`):

| Raw R | Effective R | Note |
|---|---|---|
| 1 (baseline) | 1 | newcomer |
| 10 (S) | 10 | established |
| 20 | ~13.5 | doubling raw → +35% effective |
| 100 | ~21.5 | 10× raw → +115% effective |
| 1000 | ~33 | 100× raw → ~3× effective |

The shape prevents unbounded accumulation while preserving meaningful rank-ordering at high R — Harvard's chemistry rep still beats a no-name university's, but the gap doesn't grow without bound and cred(top) can't dominate the math.

Per-node tier assignment is governance (same pattern as obsolescence). v0 default: `standard`.

### Still open (first-cut math)

- **Cold-start ramp specifics** — the `ramp_factor` and `N` (number of early attestations amplified).

---

## Currency & Records (Photons and Seeds)

Two distinct units. The split is load-bearing: photons are the fungible currency; seeds are the non-fungible records. The protocol records and *informs*; the market *prices* (heuristic #8).

### Photons (P) — the currency

- **Minted as the storage + validation reward.** When a seed is recorded, one photon is minted and awarded to the storer-validators who store and validate it (§Monetary policy). This — not raw seed creation by the author — is how photons enter circulation.
- **Fungible.** Every photon is interchangeable with every other. A clean medium of exchange.
- **Real-world value.** Photons float against fiat; their value is set by demand for data access relative to photon supply. Convertible to fiat at the edges (on/off ramps).
- **Buy access.** Photons buy non-exclusive, time-bound access to seeds for X days.

### Seeds (S) — the records

- **Minted as the participation reward.** Recording a data contribution mints a seed to the creator. The seed *is* the record — not a token pointing at data, the data unit itself. (Distinct from the photon minted to the storer-validators of that same seed — see §Monetary policy.)
- **Non-fungible.** Each seed is unique to the data that produced it.
- **Market-priced per type, equal across creators.** Access to a seed of type `a` costs `N_a` photons — a price the **market** sets through ordinary supply and demand for that type, never the protocol. Value is marginal; we don't reinvent supply and demand. The invariant: `p(c1,s,a) = p(c2,s,a) = p(c3,s,a)` — every creator's contribution of a given type is worth exactly the same. A medical record from Manila is worth what a medical record from Zurich is worth. The market prices the *kind* of data; never the *person*.
- **The value signal is binary within a type, price across types.** For a given type, the only question is whether your seed sells (it costs `N_a` either way) — a producer's earnings are *volume × N_a*. Across types, the market reveals which kinds of data are worth more by what buyers pay for access.

### How data stays creator-equal without the protocol pricing it

The protocol never sets `N_a`. A type's price is discovered in a normal market — the protocol's only contribution is to inject *verified information* into that market (type, provenance, reputation, validation, age/context/place) so buyers price accurately instead of guessing. That's Akerlof in reverse: kill the information asymmetry, let supply and demand work.

The one thing the protocol *does* enforce is the invariant — within a type, all creators are equal. So the market discovers value at the margin (across types, freely) while no individual contributor is ever priced above or below a peer for the same kind of work. **Per-type prices are market outcomes, not levers** — they never appear in `parameters.md`, exactly like producer compensation. The deep commitment that survives from the old global "1 photon" rule is creator-equality-within-type; the global-uniform price is gone, replaced by marginal market pricing per type.

### Monetary policy: photons = seeds

The photon supply is pegged to the corpus: **one photon exists per seed.** As the commons grows by N seeds, N photons mint. The money supply *is* the productive base — no halving schedule, no arbitrary inflation curve; issuance is the growth of the corpus.

**Two minting streams, one event:**
- Recording a seed mints **S** (the seed) to the **creator** — reward for participation; the seed is their record.
- The same seed mints **P** (one photon) to the **storer-validators** who store and validate it — reward for storage + validation.

Each new contribution creates one seed and one photon together; photons = seeds, exactly. *(2026-07-15: counted over distinct atoms — a re-observed contribution strengthens sigma, never re-mints; see the seed=atom decision-log entry.)*

**Why mint P for storage, not creation**: if authoring a seed minted a photon to the author, people would author junk to print money. Routing the photon to storer-validators means junk nobody stores or serves earns nothing — the reward flows to the work of *maintaining* the corpus, which directly funds the phone-participants holding shards.

**The scarcity is the point.** Because photons = seeds and some types cost `N_a > 1`, there are never enough photons to buy all access at once. Buyers must spend selectively — and selective purchasing *is* the demand signal (heuristic #8). Stock vs. flow holds: supply (= seed count) is the stock; total access-payments over time are the flow, larger via velocity.

**Creators earn photons from sales, not from minting.** Minting gives the creator a seed; they earn photons when buyers access it. Two separate streams, no double-count.

### Storage rewards: one-time mint vs. ongoing cost (resolved 2026-05-22)

Storage is an ongoing cost; minting is one-time. Both are true — the resolution is to **fund them separately so the `photons = seeds` peg survives.** New emission happens only at ingestion; all ongoing reward is *redistributed circulating photons*, never new mint.

- **Ingestion (the one-time mint)**: the one photon per seed goes to the storers who first commit to holding the new seed. Rewards bringing data into the corpus. This is the *only* minting event — peg preserved.
- **Ongoing storage rent (circulating, no new mint)**: storers earn continuously for data they hold and serve, proven by availability challenges. Three sources, in priority order:
  - **Endowment (primary — adapted from Arweave)**: the ingestion photon seeds an endowment (per-seed or pooled) that *disburses over time*, calibrated to a conservative storage-cost-decline assumption (Arweave prices at **0.5%/yr "Kryder+"** against a historical ~30–38%/yr; that conservatism is what makes it effectively perpetual). Arweave runs this in production (~95% of an upload fee → reserve → backstop disbursement; drawn only when block rewards + fees fail to cover storage cost). Mint once, disburse over time — `photons = seeds` survives, cold data funded structurally.
    - **Two caveats from the deep research (2026-05-22) — this is borrowed, not free:** (1) Arweave's endowment is *user-pre-paid and byte-priced*; ours is *minted and flat-per-seed*. The math only transfers if **seeds are size-bounded** (small records on-chain; large artifacts off-chain under hashes). If large seeds are ever allowed, storage funding must scale by bytes separately from the 1-photon mint. (2) The load-bearing risk, flagged even in Arweave's own community simulations: the endowment depends on the protocol's **estimate of real-world storage cost** — a storage-cost oracle we must get right, or the endowment drains / over-reserves.
  - **Access cuts** (`economics.access_cut_to_storers`): a slice of each access payment to storers serving a seed — a bonus for serving *popular* data atop the endowment baseline.
  - **Treasury** (fallback): toll/tax revenue if endowment + access cuts underfund a class of storage.
- **Proof mechanism**: **Storj-style random-audit possession proofs** (no SNARK sealing — return derived stripes; trivial if held, infeasible if not), suited to commodity always-on nodes. *Not* Filecoin PoRep — that's GPU/128-GiB-RAM heavyweight and the wrong reference for our accessibility goal. Slashing against uptime/durability bonds.

**The consequence**: endowment funds baseline perpetual storage (the permanent-record promise, structurally); access cuts reward serving popular data. We can honestly say *ongoing* without breaking `photons = seeds`.

**Still open**: per-seed vs. pooled endowment; seed-size cap (forced by the flat-per-seed funding); the disbursement curve + the storage-cost-oracle design; how the ingestion photon / endowment splits among a seed's storers; the `access_cut_to_storers` value.

### Marketplace mechanics (worked example)

A buyer assembles a swath by querying metadata (age, context, place, credentials / R-threshold). Each seed is priced at its type's market rate `N_a`. Example — a swath of 300,000 seeds of a type currently clearing at `N_a = 1 P`:

```
Buyer pays:          315,000 P
Producers receive:   300,000 P   (N_a per seed sold, to each seed's progenitor)
DreamTree toll:       15,000 P   (5% — economics.marketplace_toll)
```

A same-size swath of a higher-value type (say `N_a = 5 P`) clears at 5× the photons; a mixed swath sums each seed's type price. The market sets each `N_a`; the protocol just records and routes.

- **Non-exclusive**: the same seed sells to unlimited buyers; repeat demand is the value signal.
- **Time-bound (1 day default)**: a photon buys access for `access_duration_days`; access expires, the buyer re-buys to renew. Re-buying means re-acquiring a photon — which flows back from owners to buyers through the photon market. The market isn't that direct (you don't literally buy the same photon from the same owner), but that's the underlying pattern: photons circulate, and velocity on a fixed stock is what supports unlimited repeat access.
- **Equal-within-type distribution**: a producer earns by how many of their seeds land in purchased swaths, each at its type rate `N_a`. Across creators of a type, identical; across types, the market differentiates.

### Photon circulation

```
Ingestion mints photons (one per NEW leaf-seed, to the storer recipient — the only mint; no block reward)
  → buyers acquire photons (fiat → P at the on-ramp, or from validators / market)
  → buyers spend photons buying seed-access
  → producers receive photons (N_a per seed sold)
  → producers hold, re-spend (as buyers themselves), or cash out (P → fiat at off-ramp)
```

Internal settlement is in photons, on the DreamTree chain, instant via CometBFT. Fiat touches only the edges (on/off ramps). DreamTree does not settle state externally and does not issue a stablecoin — the photon floats.

### Reputation R never prices the person

The market prices the *type* (`N_a`); R never makes one creator's seed cost more than a peer's of the same type — that's the invariant. R enters two other ways: as a **buyer filter** ("assemble my swath only from producers above R-threshold X in domain k"), which drives *demand* — which seeds get bought; and as **verified information** the market reads when pricing a type. A high-R producer's seeds get included in more swaths → sell more often → earn more photons. Reputation shapes demand and informs price; it never breaches within-type creator equality.

### Real-world leverage

A producer's seed-sales record is portable, verifiable proof of *demonstrated* value. An employee negotiating salary shows that their contributions were bought and used — measured, not asserted. An employer reduces hiring lossiness by buying verified access to a candidate's record. The economic matchmaking the labor market currently does poorly and at high cost.

### Open

- **`economics.access_duration` (X)** — how many days one photon buys access; a lever.
- **Photon issuance schedule** — block-reward size, supply/inflation policy. Minting economics, a lever set.
- **Photon on/off ramps** — the fiat ↔ photon conversion mechanism at the edges.
- **Regulatory posture of the photon** — a minted, floating, fiat-convertible token is an *issued currency*. Not a stablecoin (no peg to manage), but it draws securities/commodity scrutiny depending on jurisdiction. Lighter than issuing a stablecoin, heavier than a pure record. This is the financial layer, explicitly distinct from the non-financial data records (seeds). Flagged honestly; needs counsel.

---

## Records

**Data lives in wallets — a wallet-indexed fabric, sharded across the participant network.**

Each wallet's contents (skills, experiences, attestations, credentials, raw contribution data, etc.) live in that wallet's own append-only chain — its `did:webvh` history extended into its records/seeds. The global structure is the **data fabric** (§Network): a weave of per-wallet chains, cross-linked by attestations and licenses. There is no global bundle of data types; each wallet holds its own. This is what dissolved the "how do I bundle all the data types" problem — you don't, the wallet partition does it.

The encrypted seed *bodies* live in the fabric (not in consensus blocks); the consensus chain holds only commitments and orders events. Storage requirements:

- **Sharding across the unified validator-storer network**: not every node stores every record. Phones hold small amounts; servers hold more. Partial replication with cryptographic availability proofs (challenge-response). Replication factor is a lever.
- **Default custody**: DTW-hosted-by-default (per `wallet-spec.md` L2 Q5) means the participant network hosts wallet data on the owner's behalf, logically owned by the wallet; self-custody opt-out lets a user hold their own.
- **Snapshot/archive paths**: bounded per-node burden as the corpus grows.
- **Size discipline**: small records (text-shaped contributions, attestations, credentials) held directly in the wallet's fabric chain; large artifacts (video, large documents, datasets) stored as content-addressed blobs under fabric-committed hashes + encryption metadata.

**Content provenance for non-text artifacts**: for images, audio, video, and rich documents, the protocol integrates with C2PA (Coalition for Content Provenance and Authenticity). C2PA metadata is embedded at capture or creation time — Leica and Nikon ship cameras that do this; Adobe and OpenAI services tag generated content; BBC News tags its images. When such an artifact enters the wallet, its C2PA provenance becomes part of its on-chain record, and the contributor's signature is linked to the manifest. This lets the protocol distinguish authentic human contribution from synthetic generation natively, which matters increasingly as generative output proliferates.

**Multi-key architecture**:

| Key | Role | Held by |
|---|---|---|
| Identity key | Root of trust; the thing that *is* the DID | User (DTW-hosted-by-default, opt-out to self-custody) |
| Rotation key | Updates the DID document | Same |
| Data master key | Decrypts per-record DEKs | Same |
| Per-record DEKs | Decrypt individual records | Wrapped under master key |
| Re-encryption keys | One per active license; transforms user-cipher → buyer-cipher | Proxy (chain-registered) |
| License signing key | Signs license issuance | User |
| Assertion key | Signs VCs the user issues | User (rare for individuals, common for institutional issuers) |

**Access primitive — revised after deep research (2026-05-22).** The honest finding: *no* primitive cryptographically prevents resale of outputs/plaintext. The 2026-viable stack:

- **Anchor: TEE-attested compute-to-data.** The Compute-to-Data *pattern* (algorithm runs next to the data in an isolated environment, no network egress, only results leave) is right — but done with **hardware attestation (Intel TDX, AMD SEV-SNP, AWS Nitro — production-mainstream in 2026)**, so the node operator can't peek. This is the key correction over Ocean's own C2D, whose security is "trust the operator + an allowlist," *not* hardware-enforced. Borrow Ocean's *architecture pattern*, not its platform (Ocean is unstable post-ASI-Alliance withdrawal, Oct 2025) and not its token.
- **Output minimization** on top: even in a TEE, a result can leak data (a model that memorized rows, a query returning raw records). Constrain what may leave.
- **Proxy re-encryption (tACo / Threshold Network, ex-NuCypher)** — production-ready for delegated/conditional decryption; use for cases where a buyer genuinely needs a *raw* record (identity verification, a specific credential). Honest limit: once decrypted, the buyer has plaintext — resale is then contractual/forensic, not cryptographic.
- **Not viable as the core in 2026**: FHE (2–6 orders of magnitude too slow), general MPC (too narrow/comms-heavy), federated learning alone (leaks via gradients).

**Metering primitive**: Ocean's data-NFT (ERC-721, the record) + datatoken (ERC-20, time-bound transferable access) maps almost exactly onto Seeds + Photons — validated prior art for the two-token, time-bound, non-exclusive licensing model. Borrow the pattern.

---

## Access

**Four hard rules** (carried forward from [`wallet-spec.md`](./wallet-spec.md) L2 Q6):

1. **No resale.** Of any user data, in any form. **Honest correction (2026-05-22 deep research): this is *not* purely cryptographically enforceable** — no primitive (compute-to-data, datatokens, proxy re-encryption) can stop a buyer from re-sharing outputs or plaintext they legitimately received. No-resale is a **stack**: TEE-attested compute-to-data (the operator/buyer never sees raw data, only results) + output minimization + forensic watermarking + contractual terms — *plus DreamTree's structural moat:* a resold copy is a **dead artifact** (stale, unverifiable, stripped of live attestation, reputation, and legal standing), while DreamTree access is **live and verified**. The reseller cannot reproduce that, so the leaked copy is worth far less than live access. The promise holds through economics + a defense-in-depth stack, not through impossible crypto.
2. **No third-party sharing without explicit consent.** Every cross-tenant, cross-service, or external sharing requires the user to authorize each crossing.
3. **No external advertising.** User data cannot be used to target advertising served by external parties.
4. **Owner can leave and take their data.** Portability is unconditional.

**Inside those rules, declared internal use is permitted by default (opt-out)**: services that collect data in the course of providing the service can use it for that service's relationship to the user — personalization, in-session features, product analytics, model training for service improvement, recommendations of the service's own offerings, operational uses. Each service publishes a machine-readable use declaration; users see it at consent time; reads are logged and auditable.

**Internal is bounded by the specific service the user signed up for.** Not the parent company's portfolio. Cross-service inside a corporate family (e.g., Telekora-derived data feeding Cosmo's model) is *still* cross-service and requires explicit consent.

**Marketplace mechanics**:
- Buyers post data-wanted requests (datasets for research, hiring access, identity verification, etc.)
- Users see requests matching their wallet contents
- Users grant scoped licenses; payments flow in photons (internal settlement — §Settlement; no stablecoin)
- The platform takes a brokerage toll (§Economics)
- License is registered on chain; access is mediated via the chosen cryptographic primitive; decay enforced by license expiration

**Issuer restrictions on VCs**: when a wallet VC is signed by an issuer with attached conditions ("valid until 2027," "show only to verified employers," "cannot be aggregated"), both user consent *and* issuer restrictions apply. Issuer restrictions are usually stricter and are honored at the protocol layer, not just policy.

---

## Economics

The protocol distinguishes **three economic flows**, and the distinction matters. Mechanics are in §Currency & Records; this is the principle layer.

1. **Producer compensation** — producers receive **`N_a` per seed sold** (the per-type market rate; creator-equality-within-type invariant). **The protocol does not set `N_a`; the market does.** Total compensation moves through *volume* — how many of a producer's seeds get bought — not through price negotiation. Founder-dictating compensation would replicate the extraction pattern at a new layer; the manifesto Part IV refuses this. Within-type equality is the opposite of dictation: every contribution of a kind is worth exactly the same per access, and the market decides which contributions get accessed.

2. **Marketplace toll** — a fraction of every transaction goes to the infrastructure fund (the 15,000 P in the worked example — 5%). Pays for storage, validator operations, KYC integration, human dispute resolution. An *infrastructure fee*, not compensation distribution.

3. **Value-creation tax** — levied at **sale time, producer-side** (a cut of marketplace revenue — ratified 2026-07-16, superseding the earlier issue-time framing): issuance stays free (free entry — attesting, committing, and credential-signing must never carry a toll that suppresses the observation layer), and the infrastructure fund is financed where value is *realized*. Taxing realized economic activity is what makes shared infrastructure self-sustaining.

**Founder-set initial parameters; governance evolves them.**

At v0, DreamTree-the-org sets the initial toll and tax rates because someone has to bootstrap. The governance mechanism for adjusting them is built in from day zero. As informed-voting infrastructure develops, rate decisions move to community governance. This is *infrastructure-fee setting*, not *compensation dictation* — the distinction is load-bearing.

**Toll rate**: 5% (`economics.marketplace_toll`), reconciled 2026-05-22. A lever; founder-set at v0, governance-evolved.

**What this revenue model rules out**:
- No initial coin offering. No private token sale. **Photons are minted by validation; seeds are generated by creation. Neither is sold into existence.**
- No VC capital. No board seats. No exit pressure.
- No asset-appreciation revenue. DreamTree's income is real economic activity (transaction tolls, work-issuance fees), not speculation.

This is structurally healthier than the standard chain economic model and consistent with the "record, not financial" thesis.

**Free entry for individuals.** The wallet's user-facing surface at `dreamtree.org` is free. The clarity that executives buy for thousands, a 22-year-old in debt deserves too. Institutional revenue (the wallet's enterprise customers, the marketplace toll, the value-creation tax) subsidizes free access. This is the manifesto's commitment: institutional payment funds the mission; the mission is free access.

---

## Stance

The protocol's organizational stance, drawn from the manifesto Part IV:

**Steward-ownership.** DreamTree the entity is structured for indefinite mission alignment via the Purpose Foundation model — the same structure used by Ecosia, Sharetribe, Patagonia (via Holdfast Collective), and ~60% of the Copenhagen stock market by value, including Novo Nordisk and Carlsberg. Profits stay in the mission. The entity cannot be sold to an extractive buyer. Control stays with stewards committed to the mission, not with whoever holds the most shares. Danish research shows steward-owned companies have ~6× higher survival probability after 40 years compared to conventional peers.

**Eventually cooperative.** The long arc gives the entity to the planet — a cooperative structure where users and contributors share ownership of the infrastructure they helped build. We don't have a complete governance design yet. What we have is the commitment to stewardship over extraction with the intention to distribute ownership as we learn what works.

**AGPL-3.0 with dual licensing.** Core code is AGPL: if you use it to build a service, you share your modifications back. This prevents the open-source exploitation pattern — take freely, build proprietary, give nothing back. For enterprise integration cases where AGPL creates genuine obstacles, DreamTree offers alternative commercial licensing. The commons stays common; the accommodation is that not every use fits copyleft.

**Bootstrap, no VC.** No venture capital. No board seats. No exit pressure. Revenue comes from institutions; free access for individuals is the mission. Slower growth in exchange for structural alignment. Mission-driven companies attract talent that conventional companies cannot — the trade is real and favorable.

**We don't dictate compensation distribution.** The protocol enables compensation; the actual distribution mechanisms — who pays what, how revenues split among contributors when work is collective, what counts as fair sharing in disputed cases — emerge from deliberative governance. The protocol's job is to make deliberation possible, not to foreclose it. Founder-dictation here would replicate the surveillance-economy pattern (a small set of actors deciding value distribution without consent of those whose value is being distributed). The spec refuses this explicitly.

**Founder sets infrastructure parameters at v0; governance evolves them.** Initial toll rates, tax rates, validator-set criteria, dispute-resolution procedures are set by DreamTree-the-org at launch because someone has to. The governance mechanism for evolving them is built in from day zero. Informed-voting infrastructure is a priority development item, not an afterthought. *Infrastructure-parameter setting ≠ compensation dictation* — the manifesto refuses the second, not the first.

---

## Relationship to wallet-spec.md

The wallet sits *on* the protocol. The manifesto's Architecture Part III describes three layers; the current spec maps to them as:

- **DreamTree workbook** (manifesto's "entry" layer) → the wallet's user-facing surface at `dreamtree.org`, free for individuals, the on-ramp
- **Value-Tech** (manifesto's "infrastructure" layer) → **this protocol**
- **The Great Library** (manifesto's "foundation" layer) → the four-surface architecture's Gnosis→Library node, consuming the same protocol identity and provenance primitives at civilizational scale

The protocol is the common substrate. The wallet, Telekora-the-LMS, Cosmo-the-Universal-Assistant, HometownWire, and the Great Library are all consumers on equal terms.

Resolutions already landed in [`wallet-spec.md`](./wallet-spec.md):

- **L1 Q1 (DID method)**: `did:webvh` default with pluggable upgrade paths. Still holds. The wallet identifies users; the protocol provides the chain-anchored substrate that makes the DID history log tamper-evident and the data and license layers possible.
- **L2 Q5 (custody)**: hosted custody as default; opt-in self-custody and third-party custody; no invented recovery mechanism. Refined here: the "DTW-mediated recovery" mechanism is *KYC re-verification* — recovery proves you are the same human, not that you have the same key.
- **L2 Q6 (encryption / who can read what)**: four hard rules + declared internal use. Mechanism now upgraded: enforcement is cryptographic via license-mediated access on the chain, not policy via DTW API.

Future wallet-spec updates should reference protocol-spec.md for the underlying mechanics; protocol-spec.md should reference wallet-spec.md for the user-facing layer.

---

## Levers (Parameter Registry)

Every numeric parameter in the protocol is a **lever**, not a constant (heuristic #7). Each is a named variable; the canonical names and current stand-in values live in [`parameters.md`](./parameters.md), which is the **single source of truth for values**. When this section and `parameters.md` disagree on a number, `parameters.md` wins. The table below maps lever → canonical variable for reference; see `parameters.md` for units, constraints, and the machine-readable block.

| Lever | Canonical variable | Stand-in | Disposition |
|---|---|---|---|
| Negative asymmetry | `reputation.neg_asymmetry` | 2× | governance |
| Endorsement inheritance (1st hop) | `reputation.endorsement_inheritance` | 25% | governance |
| cred recursion depth | `reputation.cred_recursion_depth` | 2 hops | governance |
| Domain attenuation (up taxonomy) | `domain.attenuation.*` | 70 / 40 / 15 / 3% | governance |
| Review window | `reputation.review_window_{base,threshold}` | τ(M) = base · log(1 + M/thr) | governance |
| Proof-type base λ (O / R / U / P) | `decay.proof_*` | 0 / 0.015 / 0.04 / 0.08 /yr | governance |
| Domain-obsolescence multiplier | `domain.obsolescence_multiplier.*` | 0.3× / 1× / 3× | per-domain |
| Validated-outcome durability | `decay.validated_outcome_halflife_years` | ~25 yr | governance |
| λ_R base rate | `reputation.lambda_r_base` | TBD | governance |
| λ_R decay target | `reputation.lambda_r_target` | baseline_KYC | **settled** |
| Cold-start ramp factor / N | `coldstart.ramp_{factor,count}` | > 1 / N TBD | governance |
| Saturation point | `reputation.saturation_point` | TBD | governance |
| Outcome magnitude M_O | `reputation.outcome_magnitude` | TBD | governance |
| Aggregation normalizer | `reputation.s_max` | TBD | governance |
| KYC baseline | `reputation.baseline_kyc` | 1.0 | settled (concept) |
| Marketplace toll | `economics.marketplace_toll` | 5% | founder → governance |
| Value-creation tax | `economics.value_creation_tax` | 0.5% (parameters.md wins) | founder → governance |
| Seed-access duration | `economics.access_duration_days` | 1 day | founder → governance |
| Access cut to storers | `economics.access_cut_to_storers` | TBD | founder → governance |
| Storage replication factor | `economics.storage_replication_factor` | TBD | founder → governance |
| Block cadence | `economics.block_cadence_seconds` | ~3s | founder → governance |

**Invariant, not a lever**: `photons = seeds` — supply pegged 1:1 to the corpus; one photon minted per seed (to its storer-validators). No issuance schedule to tune.

**Invariant, not a lever**: `p(c1,s,a) = p(c2,s,a) = p(c3,s,a)` — within a data type, all creators are priced equally. Fixed forever. The market discovers value *across* types (marginal pricing); the protocol guarantees equality *across creators* of the same type. This is the membrane that keeps the protocol from pricing the person.

**Not levers**: per-type data prices (`N_a`) and producer compensation. Both are market outcomes — the market sets `N_a` per type via supply and demand; a producer's compensation is volume × `N_a`. Neither appears in `parameters.md`. Making them tunable would mean the protocol pricing data, which contradicts heuristic #8.

---

## Open Problems

In rough order of difficulty:

1. **Decay & appreciation dynamics.** This is the load-bearing math. How does a signer's reputation change when an attestation is validated or refuted? Counter-attestation mechanism (who can refute, on what standing)? Time-revealed reality (the surgeon's outcomes attest against the certification)? Marketplace signal (buyers stop accepting credentials from a decaying issuer)? Almost certainly all three layered. Getting this wrong collapses into vibes-based reputation or punitive overreach.

2. **Reputation propagation / transitivity.** How does Harvard's weight flow through to its graduates? The graduates' weight flow to those they hire? Avoid both overweighting (everyone Harvard touches inherits Harvard's reputation) and underweighting (Harvard's reputation is purely Harvard's, never propagates).

3. **Shell-institution detection at scale.** Low barrier to entry is a feature; it lets the protocol grow. But it also lets fraudulent institutions accumulate signatures. Mimicking reality means the protocol won't catch all fraud immediately — reality doesn't either — but the protocol's measurability accelerates discovery. The mechanism for meta-attestation chains (regulators → universities → graduates) needs spec.

4. **Pre-population privacy.** Public information aggregation has structural similarity to surveillance economy patterns. The unsigned-discount is a real mitigation but not a complete one. Categories to exclude from pre-population, takedown mechanisms, and the boundary between "publicly accepted" and "private" need careful design.

5. **Coin redemption mechanics.** How exactly does a coin translate to a marketplace price? Is the coin the access right itself, or a separate pricing signal? Probably layered (the coin records work; the access is sold separately, priced partly by the relevant coins).

6. **Cryptographic primitive for license access.** Proxy re-encryption (Umbral / NuCypher lineage) is the leading candidate. TEE-based compute is a worthy augmentation for high-stakes use. Witness encryption is interesting but not production-ready.

7. **Chain choice (anchor / native / hybrid).** Operationally significant; partly downstream of the consensus design.

8. **Identity-verification privacy posture.** Zero-knowledge proof of personhood vs. DreamTree learns identity. Real tradeoff between privacy and operational simplicity.

9. **Governance.** Open chain at v3 implies on-chain governance. Initial governance is DreamTree at v0. The schedule and mechanism for opening governance need spec.

10. **Governance design for evolving founder-set parameters.** Toll rates, tax rates, validator-set criteria, dispute-resolution procedures need a credible transition from founder-set to community-voted. The Stance commits to this evolution; the mechanism is open. Informed-voting infrastructure is itself unsolved at the scale and quality the protocol requires (most chain governance is plutocratic in practice). Foundational work, not afterthought.

11. **Dual-licensing boundary.** AGPL by default; alternative commercial licensing for enterprise cases where AGPL blocks integration. What counts as a "genuine obstacle" that warrants dual licensing? Who decides? How is the commercial license priced? Open. The temptation to over-grant dual licenses leaks the commons; the temptation to under-grant kills enterprise adoption.

---

## Resolutions log

- **2026-05-21 — Protocol-spec scoped.** DreamTree is a protocol, not just a wallet. The wallet is the user-facing layer.
- **2026-05-21 — Network architecture: Door #3 with progressive decentralization.** Open public chain; operationally centralized at v0; published roadmap for opening validation. Internal framing at v0: "over-engineered database." No decentralization claim until earned.
- **2026-05-21 — Consensus is attestation-as-work under reputation stake.** PoW is signing attestations as real-world identified entities. Work is intrinsically valuable, not heat dissipation. Bootstrap by piggyback on existing reputation networks. The four canonical proof types (Origin / Rigor / Use / Replication) from the manifesto are the taxonomy. Decay/appreciation dynamics are the hardest open problem.
- **2026-05-21 — Identity is human-rooted via federated KYC.** Recovery is re-verification of personhood, not seed phrases. Pre-populated dataset with unsigned-public discount for unclaimed humans and institutions.
- **2026-05-21 — Coins are records of work; payment is by sale of access.** Coins valued in context (signer's reputation × work type × time × demand). Anti-gaming: scammer's coins compound their scam, not their wealth.
- **2026-05-21 — On-chain encrypted records.** The chain is the storage layer. Multi-key architecture across identity, data, license, and assertion roles. C2PA integrated for non-text artifact provenance.
- **2026-05-21 — Economics: three distinct flows.** (1) Contributor compensation — market-determined, *not* set by protocol or founder. (2) Marketplace toll — infrastructure fee, founder-set at v0, governable. (3) Value-creation tax — infrastructure fee, founder-set at v0, governable. No ICO, no token sale, no VC. Free entry for individuals; institutional revenue subsidizes.
- **2026-05-21 — Stance carried forward from manifesto Part IV.** Steward-ownership (Purpose Foundation model). Eventually cooperative. AGPL-3.0 with dual licensing for enterprise. Bootstrap, no VC. We don't dictate compensation distribution. Movement, not Empire — adopt existing standards (W3C VC, AT Protocol, MyData, Solid, C2PA, SD-JWT, BBS+) by default.
- **2026-05-22 — Reputation dynamics first-cut.** Five objects (R, S, V, C, λ_R). R domain-indexed, unbounded with log dampening. Update law with 2× negative asymmetry (lever). All signal queued by magnitude-scaled review window τ(M). Endorsement inheritance 25% first-hop, geometric multi-hop (lever). cred(source) recurses 2 hops. Domain attenuation 70/40/15/3% up the 5-level taxonomy. Outcomes are attestations (no central oracle); validation/refutation propagates up the chain to everyone who staked R, 2× asymmetric. Plural truth — protocol surfaces, doesn't arbitrate. No reputation floor; KYC lets you start fresh in new domains. Shell resistance via meta-attestations. Taxonomy seeded from LCC + ISCED + ONET/ISCO-08. Full math in §Reputation Dynamics.
- **2026-05-22 — Time horizons.** Two clocks: attestation-strength decay (proof-type base λ × domain-obsolescence multiplier) and reputation decay (λ_R, modulated by validated-attestation volume). Proof-type base λ stand-ins: Origin 0 (permanent), Replication ~0.015/yr, Rigor ~0.04/yr, Use ~0.08/yr. Domain-obsolescence stand-ins 0.3×/1×/3×. λ_R decays toward baseline_KYC, not zero (reputation is a stock, not a flow) — **settled**. Validated outcomes contribute durable R (~25 yr half-life); unvalidated attestations decay at proof-type rate.
- **2026-05-22 — Parameters are levers (design heuristic #7) + Levers registry created.** Every numeric value is a stand-in; the lever set is the real artifact. §Levers is the single source of truth for what's tunable. Contributor compensation is explicitly NOT a lever (market-determined).
- **2026-05-22 — Every lever is a named variable; canonical values centralized in `parameters.md`.** 26 parameters extracted, namespaced (`reputation.*`, `decay.*`, `domain.*`, `coldstart.*`, `economics.*`). `parameters.md` is the source of truth for values; the YAML block lifts into config/on-chain registry at build. Spec §Levers maps lever → canonical variable name.
- **2026-05-22 — Chain: roll our own.** DreamTree-native chain (not anchored). Justified because the consensus (attestation-as-work) is novel enough no existing chain expresses it. "Roll our own" = inherit plumbing, build the novel core. Framework: **Cosmos SDK + CometBFT**. Consensus layer (CometBFT ordering/finality) and value layer (reputation/attestation/records in the ABCI app) are **separate** — reputation never controls block production. Door #3 validator progression maps to CometBFT validator-set evolution; permissioned through v2 (no staking token, preserves no-ICO), economic staking only at v3. On-chain encrypted storage is the heaviest custom lift; Celestia is reference-only, not a dependency (the data layer IS the product).
- **2026-05-22 — Settlement disambiguated.** Two meanings were conflated. State settlement (Meaning B): DreamTree is a sovereign L1; CometBFT settles its own state; it does NOT settle to any external chain — rejected the rollup framing. Monetary settlement (Meaning A): handled by the two-token model below — internal in photons, fiat only at edges.
- **2026-05-22 — Two-token model: Photons (P) + Seeds (S).** Photons: minted by block production, fungible, real-world floating value, the currency. Seeds: generated by creator participation, non-fungible, ARE the data records. Access is non-exclusive and time-bound (X days/photon). Internal settlement in photons (on-chain, CometBFT); fiat only at on/off ramps. Photon is an issued floating currency (not a stablecoin) — real but lighter regulatory exposure, flagged for counsel. Full detail in §Currency & Records.
- **2026-05-22 — Per-type market pricing (option B) + heuristic #8 "the protocol informs, the market prices."** Data value is marginal — the market sets a per-type price `N_a` via ordinary supply and demand; the protocol never sets it, it only injects verified information (Akerlof in reverse). **Invariant changed**: from global "1 seed = 1 photon" to `creator_equality_within_type` (`p(c1,s,a)=p(c2,s,a)=p(c3,s,a)`) — the market differentiates *across types*, the protocol guarantees equality *across creators of a type*. Never prices the person. Producer compensation = volume × `N_a`; both `N_a` and compensation are market outcomes, not levers. Worked example generalized to per-type rates. parameters.md → v0.3.0.
- **2026-05-22 — Data lives in wallets; the chain is a fabric, not just a line.** Consensus stays linear (CometBFT, 1D); data is a 2D wallet-indexed fabric — each wallet's `did:webvh` history extended into its records/seeds, woven by cross-wallet attestations/licenses (block-lattice / Merkle-DAG lineage). Consensus blocks hold commitments + small txs; encrypted bodies live in the fabric, sharded. This dissolves the "bundle all data types" problem — wallets partition the data. The dimensionalizing intuition was right: the second dimension is the data layer, not consensus.
- **2026-05-22 — Unified validator-storer, participation spectrum.** Same participants store and validate, intensity scaling with device: phones hold shards + do light validation (near-passive), well-resourced nodes also run BFT consensus. Honest about CometBFT's ~100–200 consensus-validator cap — "validate on a phone" means the storage/availability layer, not BFT consensus. One role, a dial, not two castes.
- **2026-05-22 — Monetary policy: `photons = seeds`.** Supply pegged 1:1 to the corpus. **Two minting streams**: S minted to the **creator** (participation reward; the seed is the record); P minted to the **storer-validators** of that seed (storage + validation reward). Mint-P-for-storage-not-creation defeats junk-seed money-printing and funds phone-participants. Scarcity (photons = seeds, some `N_a > 1`) forces buyers to spend selectively — selective purchasing IS the demand signal. Creators earn P from *sales*, not minting. No halving schedule; issuance = corpus growth. parameters.md → v0.4.0.
- **2026-05-22 — Storage rewards (one-time vs. ongoing) + toll + access.** Resolved the storage tension: ingestion is the only *minting* event (one photon per seed → its first storers, peg-preserving); ongoing storage rent is *redistributed circulating photons* (access cuts + treasury subsidy), never new emission — so "ongoing" is honest and `photons = seeds` survives. Active marketplace subsidizes permanence of the whole corpus. **Toll = 5%** (reconciled). **Access duration = 1 day** (re-access = re-buy; photons circulate back from owners via the market). parameters.md → v0.5.0. Open: ingestion-photon split among a seed's storers, `access_cut_to_storers` value, access-vs-treasury rent balance.

---

- **2026-05-22 — Prior-art pass (Movement, not Empire).** Nearly every component has battle-tested prior art; novelty is the combination + attestation-as-work value layer + `photons = seeds`. **Nano block-lattice** validates the per-wallet-chain fabric (two heeded warnings: need participation incentive — photons; need sharding — spectrum). Proof-of-Reputation literature exists but uses reputation *as consensus*; we diverge on purpose, though DAOstack-style reputation-weighted voting is a model for governance-evolution.
- **2026-05-22 — Deep research (3 parallel agents) — corrected three over-optimistic captures:**
  1. **Storage proofs: Storj-style random audits, NOT Filecoin PoRep.** Filecoin sealing is GPU/128-GiB-RAM heavyweight — wrong reference. Storj random-audit possession proofs fit commodity always-on nodes.
  2. **Participation: a phone CANNOT be a durable storage provider** (fails on churn/uptime/durability, not compute). Phones do DAS-style *light validation* + optional ephemeral serving; durable storage = bonded always-on commodity nodes. The phone dream survives in its important half (validation), not durable storage. §Network updated.
  3. **No-resale is NOT purely cryptographically enforceable** — corrected the §Access hard rule. No primitive stops resale of outputs/plaintext. It's a *stack* (TEE-attested compute-to-data + output minimization + forensic watermarking + contractual) PLUS the structural moat (a resold copy is a dead, unverifiable artifact; DreamTree access is live + verified).
  - **Access primitive anchor = TEE-attested compute-to-data** (Intel TDX / AMD SEV-SNP / AWS Nitro — production 2026), borrowing Ocean's C2D *pattern* but not its weak trust model, unstable platform, or token. tACo (Threshold) production-ready for raw-record decryption. FHE/MPC/FL not viable as the core in 2026.
  - **Arweave endowment**: borrow it, but two caveats — Arweave is user-pre-paid + byte-priced while ours is minted + flat-per-seed (forces a **seed-size cap**), and the **storage-cost oracle** is the load-bearing risk (Arweave prices a conservative 0.5%/yr decline vs ~30–38% historical). §Monetary policy updated.
  - Ocean data-NFT + datatoken validates the Seeds + Photons two-token, time-bound, non-exclusive licensing pattern.

---

- **2026-06-24 — Outcome magnitude `M_O` resolved (loose-thread close-out, reputation math 1/3).** Formula: `M_O = min(M_cap, β · S(att, t_issuance) · √cred(reporter))`. Five clarifications: (A) use `S(att, t_issuance)` not current — preserves appreciation-compounding for old-bets-paying-out; (B) multiple reporters aggregate paper-shape, not sum (Sybil-resistant); (C) self-reports cred ≈ 0 (Akerlof); (D) outcomes are attestations of `dt.outcome.*` — same review window + cred recursion + aggregation + time horizons as any attestation, special only in triggering the `M_O` chain; (E) outcomes can themselves be refuted — counter-outcome reverses original `M_O` and applies 2× penalty to the wrong reporter (asymmetry recurses). Stand-ins: `β = 1.0`, `M_cap = 5 · S(att, t_issuance)`. Added `dt.outcome.{validated,refuted,partial}@1` to `data-types.md`. parameters.md → v0.6.0.
- **2026-06-24 — Saturation point resolved (reputation math 2/3).** Two-piece linear + log dampening: `effective_R = R` if `R ≤ S`, `effective_R = S + k · log(1 + (R−S)/S)` if `R > S`. **Per-domain `S` from day zero** (mirrors `domain.obsolescence_multiplier`), three tiers: `small=5`, `standard=10`, `large=50`. Global `k = 5`. Each domain node in the 5-level taxonomy tagged with one tier; v0 default `standard`. Prevents unbounded R accumulation while preserving rank-ordering at high R. parameters.md → v0.7.0.
- **2026-07-11 — Refutation-window integration + zero floor resolved (reputation math 3/3).** A review window integrates a **signed** verdict `M_O_net = V_pool − R_pool`, where each pool is a 1× paper-shape aggregate of its direction's reports, each capped at `M_cap`. The **2× negative asymmetry lives only at the contributor R-update** (`+M_O_net` validated / `2·M_O_net` refuted), never inside the window integration — so a defending report counts 1×, and a **false accusation is neutralized by an equal defense (1:1), not 2:1**. Co-attestors, endorsers, and the reversal penalty move at 1× (signed by the verdict). Fixes a latent double-2× (defenses in a fraud-claim window would have counted 4×, inverting Akerlof). **Zero floor made explicit and debtless**: every R move is capped at the recipient's current standing, so R can reach 0 but never go negative and no debt is stored — recovery is from 0. Paper-shape's `M_cap` (bounds crowd pile-on) plus the zero floor are the two load-bearing anti-assassination bounds. Implemented in `x/reputation/keeper/window.go` (`netVerdict` + `applyFloored`), deterministic unit tests in `window_test.go`. §"The floor is zero" rewritten to match.
- **2026-07-15 — Seed = atom ratified; the leaf model; photon becomes the bond denom; dtvp retired; chain-id `dreamtree` (owner-directed, DT-18).** The seed and the data-model atom are the same object: one data contribution = one seed = one photon = one unit of priced access. Contributions are never collapsed into one seed — **batching is a commitment strategy, not a unit change**: `MsgCommitBatch` registers `new_count` leaf-seeds `[first, first+new_count)` under one Merkle root in one tx (`batch_root` retired as a kind; kind names the LEAF; single commits are batches of one). **Convergence rule**: re-observed atoms count in `leaf_count` (provable against the root) but never in `new_count` — sigma accrues, supply doesn't; `photons = seeds` counts **distinct atoms** exactly. `new_count == 0` is a legal pure-convergence provenance batch. **dtvp retired**: it entered the build 2026-07-10 as staking expedience, was never in this spec, and squatted on the photon's designed native-asset role. Base denom `uphoton` (1 photon = 10⁶; display metadata in bank genesis); `bond_denom = uphoton`; validators bond genesis-corpus photons (the corpus IS the money supply from block 0); voting power = whole photons at default PowerReduction. No slashing/evidence modules are wired, so nothing can burn bonded photons — the peg holds structurally (wiring slashing for external validators must slash-to-treasury or document the deviation). Supply-griefing bounded by `seeds.max_batch_new_count` (default 1,000,000; committer authorization is the long-term gate). The live `dreamtree-1` devnet is wiped once and relaunched as chain-id **`dreamtree`** (no suffix, ever) with the reflow corpus (~11.7M atoms / ~62K generation-batches) carried in genesis; roots re-anchors post-launch via its proven cron path. Proven on live nodes: `scripts/leaf-proof.sh` (batch alloc, photon arithmetic, leaf resolution, pure convergence, roots path, batch_root rejection), `gov-proof.sh`, `e2e-loop.sh` (full economy photon-native). Design + cutover runbook: `docs/specs/seed-atom-conformance.md`. Commits `07367cf`, `2f2ba76`.
- **2026-07-16 — DT-21 canon reconciliation (drift paydown, phase 1).** The conformance comb (`docs/specs/comb-spec-vs-chain.md`, primary-verified) found the build and the documents of record had drifted through four mechanisms: constraint-driven substitutions never surfaced for ratification, sub-spec decisions never flowed up to canon, unlabeled multi-writer documents, and specs living away from their code. This entry records the clear-path repairs: **(1) the √ review window is ratified into canon** — τ(M) = base·√(M/threshold), threshold 4.0 interim (owner-decided 2026-07-10 in `x-reputation-p2-review-windows.md` for consensus determinism; protocol-spec and parameters.md had kept the log curve); **(2) parameters.md v0.8.0** registers every value the build actually runs (s_max 10, type weights incl. Use 0.5, coattestor_weight 0.25, attest_bet_scale 0.1, λ_endorsement 0.08, partial-outcome 0.5, max_coattestors 64, citation_uplift_λ 1.0, mintable_kinds, seeds bounds) — **marked INTERIM: registration is not ratification**, and these are the backtest's first sensitivity targets; **(3) the zero floor is debtless for real** (reversal negations floored + running floor in both read paths — commit `4c69cc6`); **(4) the peg burn hole is being closed by governance** (proposal: burn_vote_veto/prevote/quorum burns off — deposits always return; the chain's first live param-change proposal); **(5) seven spec-internal drift spots fixed** in place (staking-token body text, block-reward circulation line, stablecoin payments line, 1:1 pricing, 30K toll example, 1.5% tax, 2×-heading), plus stale `launch-readiness.md` / `deploy/anchord.service`. **Ratified 2026-07-16 (owner): the value-creation tax is levied at SALE, producer-side, never at issuance** — issuance stays free (free entry; the observation layer must never be tolled), the fund is financed where value is realized; §Economics Flow 3 amended, build conforms as-is. **Still open for owner ratification (DT-21 triage, Group 3):** the interim cred baseline (unverified = baseline until identity lands), authority-set N_a as a v0 stand-in with exit, the mintable-kinds peg gate, ENDORSEMENT as a proof type, citation uplift itself, evaluation_factor/multi-hop propagation (open-flagged), and the MISSING class (meta-attestation pre-population, C2PA, four-hard-rules objects — now honestly flagged rather than silently asserted).

---

*Last updated: 2026-07-16 — DT-21 canon reconciliation phase 1 (drift paydown); seed = atom ratified 2026-07-15 (DT-18); reputation math 3/3 closed 2026-07-11. Open beyond: seed-size cap, storage-cost-oracle, endowment per-seed-vs-pooled, ingestion/endowment split among storers, access_cut_to_storers, on/off ramps, uptime/durability bond design, TEE-specifics, dual-license boundary, governance evolution, formal per-type JSON Schemas, receiver-key handoff API.*
