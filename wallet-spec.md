# DTW — DreamTree Wallet Specification

*Started 2026-05-06. This is a living spec; decisions land here as they're made. Do not treat unanswered questions as design choices — they're flagged on purpose.*

*Companion docs: `/home/b/quorum/ARCHITECTURE.md` (where DTW sits in the four-surface frame), `/home/b/quorum/dreamtree/VISION.md` (the wallet's mission), `/home/b/quorum/dreamtree/manifesto/` (the deeper philosophy), `/home/b/DreamTree/dreamtree/` (the working app the wallet is being lifted out of).*

---

## Status

- **The wallet stands alone.** It is the primitive. Telekora, dreamtree.org, Cosmo, HometownWire — all are downstream consumers. The wallet's design decisions are not paced by Alpha; Alpha conforms to the wallet.
- **Vision is settled in shape; nothing about the wallet's *implementation* is decided yet.** The decisions below are open and are to be worked through deliberately, in conversation, in order.
- **The dreamtree-app is the v0 substrate.** It is the dogfood vehicle and the place where many wallet-shaped patterns already work in something close to production form. The wallet is not built from zero — it's lifted out of the existing app, generalized, made standalone, made consumable by the other surfaces. What survives, what changes, and what gets retired all come out of the decisions below.

---

## Vision

DTW is a wallet for who you are.

The user owns it. Cryptographically, not as policy. Their data lives in it — skills, stories, experiences, values, credentials, attestations, references, identity claims — and travels with them. They control what's shared and with whom. They can leave with everything intact.

It self-fills from signals that already exist. Drop in a resume, the wallet extracts and structures. Have a conversation, it fills naturally. Grant access to work documents with consent, it learns from them. Complete a course, the credential lands. Get a project signed off by your boss, the attestation lands. The wallet's job is to *receive* receipts — for human contribution as it happens — and to keep them in a form that travels.

It is the substrate Telekora delivers learning credentials into, the place dreamtree.org reflects back as the user's sovereign self, the storage Cosmo's agents use for their own contribution receipts, the private counterpart to HometownWire's public civic ledger. The same primitive serves humans and machines because the problem is the same: contribution without portable provenance does not get rewarded.

The wallet is an instance of the broader thesis. Economics is a fundamental force; extraction is friction; **attribution is conservation**. The wallet is the structural commitment to legibility — capture human contribution at the moment it happens, give the contributor the receipt, make the receipt portable. Without the wallet, attribution stays trapped in the institutions that claim it. With the wallet, attribution travels.

---

## What Exists Today (the v0 substrate)

These are real, working, in production. The wallet inherits from them; it does not reinvent them. (See `/home/b/DreamTree/dreamtree/` for the code.)

| Capability | Where it lives | Status |
|---|---|---|
| User-derived encryption (PBKDF2 → wrapping key → data key → AES-GCM field encryption) | `src/lib/auth/encryption.ts`, `pii.ts` | Live. Server-decryptable during active session via `sessions.data_key`. Password recovery loses encrypted data by design. |
| 40-table user-data ontology (Tier 1: profile, values, skills, stories, experiences, locations, career options, budget, flow, companies, contacts, jobs, idea trees/nodes/edges, competency scores) | `migrations/0001_initial.sql` | Live. Strong candidate for the v0 wallet ontology; needs explicit promotion or refactor. |
| Connections / data-plane primitive (typed sources, resolver, auto_populate / hydrate / reference_link / custom) | `src/lib/connections/` | Live. ~22 typed sources. v0 of wallet-as-data-plane. Currently single-tenant, single-issuer, no notion of credential vs. raw data. |
| AT Protocol scaffolding (DID + handle + PDS session storage; one-way skill sync via `com.dreamtree.skill` lexicon) | `src/lib/atproto/`, migration 0019 | Partial. Skills only. No reverse sync, no multi-record-type, no DID-as-primary-identity. |
| Email hashing for non-revealing lookup | `auth/encryption.ts` `hashEmail` | Live. |
| Reference / citation tables | `migrations/0001_initial.sql` (`references`, `content_sources`) | Live. v0 attribution chain for content; not yet wallet-attached. |

Things the substrate does *not* yet have:

- VC issuance (no issuer key, no VC schema, no issuance flow)
- Multi-issuer trust model
- Selective disclosure (SD-JWT, BBS+, or similar)
- Wallet export / import API
- DID as primary identity (the user is `users.id` first, ATP DID is an attachment)
- Reverse sync from PDS (PDS is a write target only)
- Server-blind operation (server *can* read PII when session is active)
- Public DTW SDK / API

---

## Decision Roadmap

Decisions are organized by depth, not urgency. Decide top-down: identity questions cascade into key-custody questions which cascade into encryption questions which cascade into wallet-API questions. Trying to decide a downstream question without the upstream one settled produces fragile answers.

Each decision has a **stakes** line — the consequence of getting it wrong, or of locking the wrong shape early.

### Layer 1 — Identity & ownership

The wallet's anchor: who *is* the user from the wallet's perspective?

1. **DID method.**
   - Options on the table: `did:plc` (AT Protocol's hosted method), `did:key` (self-contained from a public key), `did:web` (DNS-rooted), `did:ion` (Bitcoin-anchored), or a multi-method approach where DTW issues a DID-of-DTW's-choosing internally and federates.
   - Stakes: cross-IdP merging, key-rotation semantics, federation with the broader self-sovereign identity ecosystem, what "leaving DTW" actually means in practice.
   - **Resolution (2026-05-20): `did:webvh` as default, with pluggable upgrade paths.**
     - **Default for all new wallets**: `did:webvh:dreamtree.org:u:<opaque-id>`. DTW hosts the DID Document and the append-only signed history log under `dreamtree.org`. Identity and rotation keys are encrypted with the user's wrapping key per the existing PBKDF2/Argon2 pattern (`auth/encryption.ts`), extended from PII to identity material.
     - **Append-only history**: did:webvh's signed log gives the tamper-evidence of "DTW runs its own chain" without operating one. Each document version is signed by the prior-state key; a host that tries to silently rewrite history fails verification.
     - **Portability**: a user can export history log + DID Document + key material and re-host the DID under their own domain (or any did:webvh-compatible host); the DID stays the same across hosts. Continuity is preserved by a signed migration operation.
     - **Upgrade paths (opt-in, user-initiated)**: `did:plc` (free, joins ATP ecosystem), `did:dht` (free, DHT-anchored sovereignty), `did:ion` (Bitcoin-anchored, contingent on the network surviving Microsoft's withdrawal), `did:ethr` (gas-paid, crypto-native users). The migration mechanism — re-issue VCs to the new DID vs. preserve them under a controller-relationship across DIDs — is deferred to Layer 4.
     - **Why not did:dht as default**: more sovereign by default (no operator dependency), but lacks a recovery key model — losing the root key permanently bricks the DID. For the silent-wallet pattern (Telekora users who don't know they have a wallet), this trades a manageable trust point for an unrecoverable failure mode. did:dht stays as a first-class upgrade target.
     - **Why not did:plc as default**: ATP integration is partly scaffolded but did:plc binds DTW's identity layer to Bluesky's protocol governance. Better as an opt-in upgrade for users who want ATP alignment.
     - **Why not roll-our-own chain**: operationally expensive (validator economy, security budget, interop loss). did:webvh delivers append-only history without running a chain; chain-anchored upgrade paths exist for users who want chain-strength continuity.
     - **Reversibility**: this resolution can be revisited. did:webvh was chosen because the silent-wallet UX commitment dominates at this stage. If DTW later decides sovereignty should be the *default* posture, migration tooling makes the switch tractable for the existing user base.
     - **Downstream effects**: constrains L1 Q2 (DID at surface, `users.id` remains internal anchor), reshapes L1 Q3 (promotion = migrate hosting and/or DID method), leaves L1 Q4 (cross-IdP merging) untouched, and partially constrains L2 Q5 (custody) to hosted-with-blindness-when-logged-out for the default. Recovery model (L2 Q5 sub-decision) is now the next hinge.

2. **Primary anchor: DID-first or account-first?**
   - In the v0 substrate, `users.id` is the primary anchor and ATP DID is an attachment. In a wallet-first world, the DID is the anchor and the IdP-bound account is an attachment.
   - Stakes: every downstream record-store decision (where wallet contents are *keyed*); whether deleting an IdP binding leaves the wallet intact; whether the same wallet can have multiple IdPs.

3. **Implicit-to-explicit promotion path.**
   - Telekora's silent-wallet posture means a learner has a DTW from day zero without knowing it. At some point the wallet becomes explicit (UI surface, export, portable credential, etc.). When? How? Who triggers it?
   - Stakes: whether silent users actually have *real* wallets or stub wallets, what the promotion costs (re-key? re-issue? re-encrypt?), what the experience feels like at the moment of discovery.

4. **Cross-IdP merging.**
   - User signs up with Google. Later signs up with Auth0 at the same email. Are these the same DTW? When are they merged? On consent? Automatically? Never?
   - Stakes: identity collisions, fraud surface, lost wallets if the wrong account becomes canonical.

### Layer 2 — Custody & cryptography

Once "who is the user" is settled, the wallet needs to know what keys exist, who holds them, and what each key authorizes.

5. **Key custody model.**
   - Pure self-custody (user holds the only copy of their wallet keys; if they lose it, it's gone). Hosted custody (DTW holds keys on the user's behalf, server-blind via password-derived wrapping). Hybrid / progressive (start hosted, promote to self-custody on demand). Threshold / social recovery (m-of-n).
   - Stakes: every promise about "the wallet is yours" depends on this. Hosted custody means DTW is a trust point; self-custody means UX cliffs and lost wallets; hybrid is harder to reason about than either pure model. The v0 substrate's approach (password-derived data key, server-decryptable during session) is hosted-custody-with-blindness-when-logged-out, which is a real position but not the only one.
   - **Resolution (2026-05-20): Hosted custody as default forever, with opt-in paths out. No invented recovery mechanisms.**
     - **Default for all users (silent and explicit alike)**: hosted custody with user-derived wrapping key. Extends the existing v0 substrate pattern (PBKDF2/Argon2 → wrapping key → AES-GCM, `auth/encryption.ts`) from PII to identity/rotation keys. DTW can decrypt during an active session via `sessions.data_key`; cannot decrypt when the user is logged out.
     - **Default recovery**: DTW-mediated through alt-channel identity proof (email + 2FA + whatever combination meets the trust bar). DTW openly is a trust point in this mode. Stated honestly, not papered over.
     - **Opt-in: self-custody**. User takes possession of rotation key, accepts loss-on-loss semantics. DTW no longer mediates. This is irreversible-ish — there's no path back to DTW-mediated recovery once self-custody is chosen, because the user is the only key-holder.
     - **Opt-in: third-party custody**. User delegates to a non-DTW custodian of their choice. Mechanism deferred to L6 (Surface & API) once a real consumer-grade custody ecosystem exists.
     - **Explicitly not required**: BIP-39 / seed phrase recovery, social recovery (m-of-n trusted contacts), hardware-bound recovery (YubiKey / passkey-only), Shamir's Secret Sharing variants. We are not shipping any of these.
     - **Reason for non-requirement (the user-supplied argument that closed this)**: real-world recovery mechanisms have failure modes the philosophy actually warns against. Social recovery in particular: "trusted contacts" trend toward whoever has the most leverage in the user's life — bosses, dominant family, abusive partners — which is exactly the soft-tyranny pattern PRINCIPLES.md §8 names. The mechanism design problem ("how does a normal person designate m-of-n contacts who actually represent them") is unsolved in the wider crypto ecosystem. Without novel insight, we do not require it. Ship honest delegation to DTW; don't ship vapor sovereignty.
     - **The honest framing**: DTW is the trust point in the default model. The user can leave any time (export their did:webvh history log + wallet contents; re-host elsewhere; choose self-custody). "Sovereignty is your right, not your obligation" — Carol in the minivan shouldn't have to manage a seed phrase to keep her contribution receipts. Power without conscience is the failure mode PRINCIPLES.md names; forcing UX cliffs on normies is a different version of the same failure.
     - **Watching posture**: when a battle-tested consumer-grade recovery pattern emerges from the broader ecosystem (one that doesn't collapse into "boss as recovery contact" under power-asymmetry), evaluate for integration. Until then, the menu is hosted / self / third-party.
     - **Consistency with the philosophy**: Section 9.6 of `/home/b/quorum/philosphy/09-limitations.md` is explicit about not pretending the framework has solved implementation challenges. DTW shouldn't either. Honest trust beats invented sovereignty.
     - **Downstream effects**: closes L2 Q5 in the form "what custody model does DTW commit to?" Leaves open L2 Q6 (what's encrypted, who can read it, when — smart-content tension) and L2 Q7 (encryption layering with VCs). L1 Q3 (implicit → explicit promotion) is further constrained: promotion no longer requires the user to commit to a non-DTW custody model. They may opt in, but the default stays available even after the wallet becomes explicit to them.

6. **What's encrypted, who can read it, when.**
   - The v0 substrate encrypts only PII fields (display_name, budget data, contacts, Module 1.4 responses), not the whole wallet. Smart-content rendering depends on the server reading wallet data during a session.
   - Stakes: a smart-content commitment that depends on server-readable wallet data is not the same product as a wallet-everything-encrypted commitment. Choosing which fields are wallet-encrypted vs. tenant-readable is a hard trade between privacy and platform intelligence.

7. **Encryption layering with VCs.**
   - VCs are themselves cryptographic objects (signed by issuer). Storing them additionally encrypted at rest is a separate layer. Selective disclosure (SD-JWT, BBS+) is a third layer at presentation time.
   - Stakes: wrong layering means either accidental redundancy (encrypting already-private VCs nobody will ever read) or accidental leakage (storing VCs in a way that the server can verify signatures on but the user thought was server-blind).

### Layer 3 — Wallet contents & ontology

The wallet's *what*: typed contents, relationships, schemas.

8. **The v0 wallet ontology.**
   - Promote the dreamtree-app 40-table schema as-is, refactor it, or treat it as one of several feeder ontologies?
   - Stakes: every existing tool's `data_contract_json` (per the Telekora alpha spec) types against this. Promoting as-is is fast but locks in dreamtree-app shape decisions that may not be wallet-grade. Refactoring forces a re-shape but takes time. Treating it as a feeder ontology among others lets the wallet have a higher-level abstraction but increases scope.

9. **The data-type registry.**
   - Naming convention, namespacing (tld-style? URL-style? bare strings?), versioning, who can register new types.
   - Stakes: forward compatibility. A bad data-type identifier scheme can't be retconned. AT Protocol uses NSIDs (`com.dreamtree.skill`); W3C VC uses URIs; Open Badges uses contexts. Picking the convention determines compatibility direction.

10. **Tenant-scoped vs. wallet-scoped data.**
    - Some wallet contents belong to the wallet across all contexts (skills the user has). Some are produced inside a tenant and might or might not promote out (a quiz response in MLM-X's course). What's the rule?
    - Stakes: leakage if scoped-up too aggressively; lost portability if scoped-down too aggressively.

11. **Raw data vs. credentials vs. attestations.**
    - The wallet holds three different kinds of contents: raw user data (a SOARED story), self-claims (the user says they have this skill at mastery 4), and third-party credentials/attestations (the manager signs that the user shipped the project). These have different trust semantics, different storage shapes, different export rules.
    - Stakes: collapsing these into one model produces either weak credentials (mixed with self-claims) or heavyweight raw data (over-formalized). Keeping them separate but related is the harder design.

### Layer 4 — Issuance & verification

How credentials and attestations get *into* the wallet, and how they're trusted on the way out.

12. **VC schema authority.**
    - Already settled at the architecture level: DTW's ontology authors the schema. *How* — what's the registry, the governance, the schema-registration flow? Open.
    - Stakes: the "ontology-aligned credentials" differentiator only holds if there's a real ontology-authority surface; otherwise it's vibes.

13. **Issuer trust model.**
    - When an MLM customer issues a VC into the wallet, the wallet trusts what about the issuer? Trust roots, key rotation, revocation.
    - Stakes: a wallet that trusts everything is no different from a CV; a wallet that trusts nothing is empty.

14. **Attestation flow.**
    - When a manager signs off on a project, what's the UX? Does the manager need a DTW too? What's the receipt format? How does the user accept it?
    - Stakes: this is the actual mechanism by which "human verification travels." If it's heavyweight, no one uses it. If it's lightweight, the verification is weak. Finding the right shape is product work, not just protocol.

15. **Selective disclosure mechanism.**
    - SD-JWT (selective disclosure for JWT-VCs, simpler), BBS+ (zero-knowledge selective disclosure, more powerful but heavier), or both. Plus what disclosure operations the user can actually do at presentation time.
    - Stakes: privacy commitments and what "prove your degree without revealing your GPA" means in practice.

### Layer 5 — Storage, sync, & portability

Where the wallet lives, how it travels.

16. **Storage model.**
    - DTW server-side D1 (current direction inherited from dreamtree-app). User's PDS. Both. Local-first with server sync. This decision is heavily downstream of #5 (custody) and #6 (encryption).
    - Stakes: server-side storage is convenient but makes "user-owned" an assertion the user has to trust; local-first is true to the principle but operationally hard.

17. **PDS sync semantics.**
    - When the user has an ATP PDS, what's the sync direction (D1 → PDS only, like today; bidirectional; PDS-as-source-of-truth)? What's the conflict resolution? What's synced, all wallet contents or a subset?
    - Stakes: bidirectional sync with conflict resolution is a hard distributed systems problem; one-way sync limits what "data sovereignty" actually delivers.

18. **Export / import.**
    - JSON dump? Cryptographically signed bundle? PDS clone? Standard format (W3C Data Vault, Solid Pod, …)?
    - Stakes: portability is a brand promise. Exporting to a format nobody else reads is a non-portability promise dressed up.

19. **Multi-device.**
    - User signs in from a second device. Does the wallet follow? How are keys reconstituted on the new device without weakening custody?
    - Stakes: every UX cliff in self-custody wallets lives here.

### Layer 6 — Surface & API

What DTW looks like from the outside.

20. **Public SDK shape.**
    - TypeScript first? Multi-language? RPC vs. REST vs. GraphQL? Authentication for SDK consumers (OAuth, capability tokens, …)?
    - Stakes: the SDK is the product surface for everyone except dreamtree.org. Telekora-the-LMS, Cosmo, HometownWire, future consumers all build against it.

21. **Read API for smart content.**
    - The "smart content" commitment requires reads against the wallet from inside courses. What's the read surface, what's the consent model (does every read require explicit user consent? implicit during a session? course-scoped scope tokens?), what does it look like from the content author's perspective?
    - Stakes: this is the difference between smart content as a real platform property and smart content as a privacy hazard.

22. **Write API for issuers.**
    - Telekora issuing a course-completion VC into the wallet, HometownWire issuing a vote-record VC. What's the issuer API, the user consent model, the receipt back to the issuer that issuance happened?
    - Stakes: writes-without-consent erode the wallet's user-owned posture; consent-on-every-write is unusable in practice. The middle path is the design.

23. **dreamtree.org wallet UI.**
    - The standalone surface where the user *uses* their wallet — reviews contents, manages disclosures, sees attestations, runs the workbook + data-literacy + DTW-education courses against the wallet.
    - Stakes: this is the brand experience. The principles in `PRINCIPLES.md` and the UX DNA in `dreamtree-app/CLAUDE.md` (anti-gamification, conversational, user-owned aesthetic, magic moments) are the design constraints.

### Layer 7 — Movement-level & legal

The wallet exists in an ecosystem and a legal context.

24. **Standards alignment.**
    - W3C VC 2.0, Open Badges 3.0, AT Protocol, C2PA, MyData, Solid, IETF SD-JWT — which to actively conform to, which to interoperate with, which to advance, which to ignore.
    - Stakes: standards alignment is movement-membership and lock-in-prevention; each one chosen is a relationship to maintain.

25. **AGPL boundary.**
    - DTW protocol + SDK + reference impl is AGPL per the architecture decision. *What's the protocol vs. the implementation* is fuzzy in practice (especially for cryptographic schemes). Where does the AGPL boundary actually fall in code?
    - Stakes: enterprise consumers will route around AGPL if the boundary's drawn permissively; the principle is hollow if drawn too narrowly.

26. **Steward-ownership in code.**
    - The legal commitment is steward-ownership; the code commitment is AGPL; the missing piece is what governance looks like for the wallet's *protocol* once external consumers depend on it.
    - Stakes: this is the question the manifesto explicitly defers ("compensation refused as founder dictat — build the deliberative infrastructure first"). The wallet protocol is the first thing that will need such a governance surface.

---

## How to use this doc

- Decisions get made in conversation, then recorded *in this file* as resolutions appended to the relevant section.
- New questions get added as they're surfaced; mark them with date and reason.
- Don't decide downstream layers (4–7) before upstream layers (1–3) are settled. The temptation is real and the rework cost is high.
- This doc is companion to ARCHITECTURE.md's Open Questions section. When a wallet decision lands here, the corresponding architecture-level open question gets updated.

---

## Resolutions log

- **2026-05-20 — L1 Q1 (DID method): `did:webvh` as default with pluggable upgrade paths to `did:plc` / `did:dht` / `did:ion` / `did:ethr`.** Chosen because the silent-wallet UX commitment (Telekora learners as implicit DTW instances from day zero) dominates at this stage; did:dht's no-recovery-key failure mode is unacceptable as a default but valuable as an upgrade; rolling our own chain is operationally expensive without buying more than did:webvh's append-only log already gives. Reversible. Constrains L1 Q2, reshapes L1 Q3, partially constrains L2 Q5.

- **2026-05-20 — L2 Q5 (key custody model): Hosted custody as default forever, with opt-in paths to self-custody or third-party custody. No invented recovery mechanisms — no BIP-39 / social / hardware required.** DTW is honestly a trust point in the default; user can leave any time. Social recovery rejected as default because real-world failure mode is bosses / dominant family / abusive partners becoming "trusted contacts" — soft tyranny is the exact pattern PRINCIPLES.md §8 warns against. Sovereignty is the user's right, not their obligation. The mechanism design problem is unsolved in the wider ecosystem; without novel insight, we do not require what we cannot make safe. Watching posture: integrate battle-tested consumer-grade recovery if one emerges.

---

*Last updated: 2026-05-20 — L1 Q1 and L2 Q5 resolved. L1 Q4 (cross-IdP merging), L2 Q6 (encryption surface), L2 Q7 (VC encryption layering), and Layer 3+ (ontology, issuance, storage, API, movement) remain open.*




