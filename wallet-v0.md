# DreamTree Wallet v0 — Buildable Structure

*Started 2026-05-22. The concrete, build-it-now slice of the wallet that Telekora's next phase is built on. This is **Phase v0** of the roadmap implied by [`protocol-spec.md`](./protocol-spec.md): the "over-engineered database" that backs Telekora silently, with zero chain / photons / marketplace / TEE. Every structure here is designed to migrate forward into the full protocol with no conceptual rewrite — the forward map is at the bottom.*

*Grounded in recon (2026-05-22) of the shipped DreamTree app (`/home/b/DreamTree/dreamtree/`) and Telekora (`/home/b/quorum/telekora/`). Stack for both: Cloudflare Workers + D1 (SQLite). v0 stays on that stack.*

---

## What v0 is and isn't

**Is**: a typed, per-wallet, encrypted record store with a durable identity anchor, a consent-gated read API (smart content), and an issuer write API (credentials). A standalone Workers + D1 service that DreamTree-the-app and Telekora both consume over HTTP.

**Isn't (deferred to v1+)**: the sovereign chain, photons/seeds economy, the marketplace, TEE compute, reputation consensus, the data fabric, did:webvh, KYC-rooted recovery. None of it is needed to back Telekora.

**The one new capability v0 must add**: credential issuance. Neither app has any VC/attestation/signing today. Telekora needs to issue course completions *into* the wallet; that's the headline new thing.

---

## 1. Identity anchor — human-as-key, real-world proof first-class

The identity model follows the thesis we settled: **the human is the identity; the key is an attachment.** You can lose a key; you can't lose being you. Real-world proof is **first-class in v0**, not deferred — because it reuses the credential layer (§5), it costs almost nothing to design in now.

- **`wallet_id`** — stable opaque anchor (nanoid). Both apps already treat `users.id` as the de facto anchor; v0 promotes it to an explicit `wallet_id`. Designed to bind to a **`did:webvh`** in v1 (`wallets.did` column present now); the verified human is what that DID will anchor to.

- **Verification tier (first-class in v0):**
  - `unverified` — IdP-bound only (Telekora = Google). The **silent-wallet default**: a learner gets a wallet at signup with zero friction. Credentials still issue, but carry no proof-of-personhood.
  - `verified_human` — the wallet holds a valid **proof-of-personhood credential** issued by a registered KYC provider (Persona / Onfido / Jumio / Trulioo). This is the human-as-key anchor.

- **Real-world proof = the credential layer applied to identity.** A KYC provider is just a high-trust **registered issuer** (`wallet_issuers`); a proof-of-personhood is a `wallet_record` of `data_type = 'identity.proof_of_personhood'` signed by the provider. No new machinery — §5's issuer/credential mechanism *is* the real-world-proof mechanism. `verification_tier` is **derived** from holding a valid one.

- **Promotion (silent → verified)** is `wallet-spec.md` L1 Q3, "make explicit what was always there": the learner upgrades when they want their credentials to carry weight — e.g., carrying a completion to an employer. Low-friction onboarding *and* real proof on demand. A course-completion credential issued to a `verified_human` wallet is worth far more to an employer than one on an anonymous Google account — this is why proof-of-personhood belongs in v0 even though learning doesn't require it.

- **Recovery tracks the tier:** `unverified` → IdP recovery + the password/data-key model. `verified_human` → **re-prove you're the same human** (re-run KYC; the provider re-attests, matching the prior attestation). This is the human-as-key recovery — the key is replaceable because the human is the anchor.

- **Custody**: hosted (`wallet-spec.md` L2 Q5). Server can decrypt during an active session. Honestly *not* zero-knowledge (§3).

- **Privacy posture (open, flagged in `protocol-spec.md` §Identity):** does the proof-of-personhood reveal identity to DreamTree, or is it a **zero-knowledge attestation** (verified-real, distinct-from-others, no details)? Both are valid points on the trust/UX curve; pick before wiring a provider.

- **Staging — what's v0 vs. turn-on:** the **model, the verification tier, the issuer-as-KYC-provider mechanism, and the upgrade/recovery flows are all v0** (the schema and credential layer support them today). The actual KYC-*provider* integration is a v0 feature you switch on when you pick a provider — "export that proof at first" (consume an existing verifier; build our own KYC++ capability later). The first Telekora sprint can ship with the structure in place and verification available as the upgrade.

---

## 2. The typed record model — the core

Every piece of wallet data is a typed record. This single abstraction subsumes DreamTree's 40-table `user_*` ontology **and** Telekora's planned `learner_data` table. A v0 record is a v0 **seed**.

**The `data_type` ontology lives in [`data-types.md`](./data-types.md).** That document is the authoritative registry — naming convention (`dt.<category>.<subcategory>@<version>`), the v0 seed list (grounded in DreamTree's 21 `DataSourceType` + Telekora's `learner_data` types + W3C VC credential types + KYC identity types), PII classification, and external-standard alignment. The wallet service loads it as a JSON seed at boot; every `POST /records` validates `data_type` against it; every record's `encrypted` flag is derived from the registry's PII classification. Telekora tool contracts (`writes_data_types` / `reads_data_types`) validate against the same registry — single source of truth across the stack.

```sql
CREATE TABLE wallets (
  id                 TEXT PRIMARY KEY,   -- wallet_id (nanoid) — durable anchor
  created_at         INTEGER NOT NULL,
  verification_tier  TEXT NOT NULL DEFAULT 'unverified',  -- 'unverified' | 'verified_human'
                                         -- DERIVED from holding a valid identity.proof_of_personhood record
  did                TEXT                -- nullable; did:webvh in v1, anchors to the verified human
);
-- Note: no kyc_attestation_hash column — the KYC proof is a wallet_record (data_type
-- 'identity.proof_of_personhood') signed by a KYC-provider issuer; verification_tier derives from it.

CREATE TABLE wallet_identities (            -- IdP attachments (Google now; more later)
  wallet_id     TEXT NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
  provider      TEXT NOT NULL,              -- 'google' | 'password' | ...
  provider_uid  TEXT NOT NULL,
  PRIMARY KEY (provider, provider_uid)
);

CREATE TABLE wallet_records (               -- THE core: each row is a typed record = a v0 seed
  id            TEXT PRIMARY KEY,
  wallet_id     TEXT NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
  data_type     TEXT NOT NULL,              -- ontology key: 'skill', 'story.soared',
                                            -- 'value.work', 'credential.course_completion', ...
  payload       TEXT NOT NULL,              -- JSON; if encrypted, the {v,iv,ciphertext} envelope
  encrypted     INTEGER NOT NULL DEFAULT 0, -- driven by the data_type's PII classification
  source_type   TEXT,                       -- 'tool' | 'issuer' | 'import' | 'self'
  source_ref    TEXT,                       -- tool_id / lesson_id / import id (provenance)
  issuer_id     TEXT,                       -- nullable; set for issued credentials/attestations
  signature     TEXT,                       -- nullable; issuer Ed25519 sig over payload
  created_at    INTEGER NOT NULL,
  updated_at    INTEGER NOT NULL
);
CREATE INDEX idx_records_wallet_type ON wallet_records(wallet_id, data_type);

CREATE TABLE wallet_issuers (               -- registered issuers + pubkeys (v0 trust = a known list)
  id            TEXT PRIMARY KEY,
  name          TEXT NOT NULL,
  public_key    TEXT NOT NULL,              -- Ed25519 pubkey to verify signatures
  created_at    INTEGER NOT NULL
);

CREATE TABLE wallet_access_log (            -- consent/audit: every consumer read
  id            TEXT PRIMARY KEY,
  wallet_id     TEXT NOT NULL,
  reader_id     TEXT NOT NULL,              -- the consuming service/tenant
  data_type     TEXT,
  purpose       TEXT NOT NULL,              -- declared purpose
  at            INTEGER NOT NULL
);
```

**The ontology = the set of `data_type` keys.** Seed it from the DreamTree 40-table schema (skill, value.work, story.soared, experience, idea_tree, competency_score, flow_log, …) plus the credential types Telekora needs (credential.course_completion, attestation.*). This is the v0 answer to `wallet-spec.md` L3 Q8 ("the v0 wallet ontology"): the existing schema, promoted to a typed-record `data_type` registry rather than 40 bespoke tables.

---

## 3. Encryption (v0 — honest)

Keep the existing scheme (`src/lib/auth/encryption.ts`): password → PBKDF2 (SHA-256) → wrapping key → AES-GCM-256 data key, wrapped in `auth.wrapped_data_key`. PII-typed records encrypted at rest as `{v,iv,ciphertext}`. The data key lives in the session so the server can decrypt during an active session.

**This is hosted custody, server-readable during your session — NOT zero-knowledge. Say that honestly.** (Per `wallet-spec.md` L2 Q6 and the 2026-05-22 deep-research honesty correction.)

**Fix the footguns recon found before building on it:**
- `encryptPII` silently stores **plaintext** if no session key is present — **fail closed instead.**
- AT Protocol session tokens in `user_atp_connections.session_data` are documented as encrypted but stored **plaintext** — encrypt them.
- `PII_FIELDS` references `user_contacts` columns that don't exist — reconcile the metadata.
- Duplicate-numbered migrations (two `0008`/`0009`/`0010`) — clean up before forking the schema.

---

## 4. Read API — smart content, consent-gated

`GET /wallet/{id}/records?data_type=<key>&reader=<consumer>&purpose=<declared>`

- Returns typed records (decrypting PII types during the wallet owner's active session).
- Gated by the four hard rules + declared internal use (`wallet-spec.md` L2 Q6): Telekora reads as a first-party consumer with declared purpose `personalize_telekora`. Every read appended to `wallet_access_log`.
- This is the v0 of `wallet-spec.md` L6 Q21 (Read API for smart content) and satisfies Telekora's `data_contract_json.reads_data_types`.

---

## 5. Write / issuer API — records and credentials

**Self/tool data** (no signature): `POST /wallet/{id}/records` — tool outputs (DreamTree-29 tools: skills, SOARED stories, flow logs) land as typed records.

**Credential issuance** (the new capability): `POST /wallet/{id}/credentials`
- Body: `{ vc }` — a full **W3C Verifiable Credential 2.0** document (the canonical JSON-LD VC with `@context`, `type`, `issuer`, `credentialSubject`, `proof[]`).
- The wallet stores it as a record with `data_type` resolved from the VC's `type` array (e.g. `dt.credential.learner_response@1`), the full VC as the payload, `issuer_id` = the VC's `issuer` DID, and `signature` = the issuer-proof's `proofValue` (parsed out for indexing).
- Verification is **local**: DIDs are `did:key:z6Mk…` (Ed25519 — the key IS the DID), so no resolver is needed. Verify by `ed25519.verify(pubkey, canonical-vc-without-proof, proofValue)`.

**Correction (2026-05-22):** earlier sketch had v0 = "Ed25519 signed JSON, W3C VC later." That was wrong. **Telekora has already designed and is implementing W3C VC 2.0** (`/home/b/quorum/telekora/docs/specs/verifiable-credentials.md`, task #67): full content-addressable merkle tree (tool → lesson → course → path), `did:key` identifiers, dual-proof (issuer `assertionMethod` + receiver `authentication`). The wallet must consume W3C VC 2.0 **from day one** — anything less is throwaway. The seed of `wallet-spec.md` L4 is the W3C VC stack itself.

**Dual proofs and the receiver-key handoff** (`dtw-handoff.md`): Telekora today server-holds *both* keys (issuer + receiver). The receiver key migration to the wallet is one of v0's load-bearing jobs — when a wallet is linked, Telekora flips `receiver_keys.source` from `'server'` → `'dtw'`, deletes the encrypted server-side private key, and from then on the wallet signs the receiver proof. Past credentials remain verifiable under the historical key (Telekora archives the old pubkey publicly per `dtw-handoff.md` §4).

**Verify**: `GET /wallet/{id}/records/{rid}/verify` → for VC-typed records, runs the local W3C VC verification (issuer proof + receiver proof + optional merkle-root check against current curriculum roots, per `verifiable-credentials.md`).

**Export** (data sovereignty): `GET /wallet/{id}/export` → full decrypt + dump (the existing `api/profile/export` generalized).

**Register issuer**: `POST /issuers` → `{ name, public_key }`.

**KYC providers are issuers too.** Real-world identity proof reuses this exact layer: a KYC provider (Persona/Onfido/Jumio/Trulioo) registers as a high-trust issuer, and a `POST /wallet/{id}/credentials` with `data_type = 'identity.proof_of_personhood'` binds proof-of-personhood to the wallet. `wallets.verification_tier` derives from holding a valid one. No separate identity machinery — the human-as-key anchor is just the most important credential in the wallet.

---

## 6. How Telekora integrates

1. **Learner signup** → `POST /wallet` (create wallet, bind Google identity). `wallet_id` becomes the durable anchor that augments Telekora's local `users.id`.
2. **Replace the planned `learner_data` table with wallet API calls.** Telekora's `alpha.md` seam (`user_id`-scoped, nullable-tenant, `data_type`+`data_json`+provenance) maps 1:1 onto `wallet_records`. Building against the wallet API now means **no future migration** — exactly what `alpha.md` wanted, achieved by calling the service instead of growing a local table.
3. **Tool outputs** (when the `tools`/`tool_types` registry ships) → `POST /wallet/{id}/records`, typed by the tool's `writes_data_types`.
4. **Course completion** → `POST /wallet/{id}/credentials` with Telekora's tenant as the signing issuer. This is the credential issuance Telekora explicitly lacks today.
5. **Smart content** → `GET /wallet/{id}/records?...&purpose=personalize_telekora` for personalization, matching the tool's `reads_data_types`.

---

## 7. Deployment shape (DECIDED 2026-05-22)

**Standalone wallet service** — CF Workers + a dedicated D1 database, consumed over HTTP. Telekora is a client from day one, so the kora.md migration concern never arises.

**DreamTree-the-app sits invisible for now.** It does *not* migrate onto the wallet service in v0 — it keeps running on its own substrate. **Telekora is v0's only consumer.** Design the API to Telekora's needs first, keep it general enough that DreamTree-the-app and other surfaces slot in later (v1+). The wallet is the silent backbone Telekora builds on; nothing user-facing ships under the DreamTree brand in this phase.

---

## 8. Forward-migration map (v0 → protocol-spec)

Every v0 structure has a protocol-spec counterpart, so v0 is a foundation, not a throwaway:

| v0 structure | Migrates to (protocol-spec) |
|---|---|
| `wallet_records` row | a **Seed** in the wallet-indexed data fabric |
| `data_type` registry | the type taxonomy / 5-level domain taxonomy |
| `wallet_id` | the wallet's **`did:webvh`** (anchor binds to the DID, which anchors to the verified human) |
| `verification_tier` + `identity.proof_of_personhood` record | KYC-rooted DID identity; the KYC provider becomes a high-reputation issuer / meta-attestor (R) |
| `wallet_identities` (IdP) | IdP-at-surface → DID-at-persistence (L1 Q6) |
| Ed25519 signed-JSON credential | **W3C VC** + the attestation-as-work reputation layer |
| `wallet_issuers` pubkeys | issuer **DIDs + reputation R** (meta-attestation) |
| hosted custody (`sessions.data_key`) | hosted default **+ opt-out** self/third-party custody |
| `wallet_access_log` | the on-chain **consent/license registry** |
| D1 per-wallet rows | per-wallet chains in the **fabric**, sharded |
| (none in v0) | photons, seeds-as-currency, marketplace, TEE access, endowment storage |

---

## 9. Build order (suggested)

1. Stand up the wallet service: `wallets`, `wallet_identities`, `wallet_records`, the read/write routes, the existing encryption (with footgun fixes).
2. Seed the `data_type` ontology from DreamTree's schema + Telekora's credential needs.
3. Add the issuer layer: `wallet_issuers`, `POST /credentials`, verify route, Ed25519 signing.
4. Point Telekora at it: wallet-on-signup, completion → credential, smart-content reads.
5. Migrate DreamTree-the-app's `user_*` writes to the wallet service (or leave dual-running until v1).

---

*Last updated: 2026-05-22 — v0 structure + human-as-key identity (real-world proof via the credential layer) + W3C VC 2.0 credentials from day one (correcting the earlier "signed JSON, VC later" sketch — Telekora is already shipping W3C VC 2.0). DECIDED: standalone wallet service; DreamTree-the-app sits invisible (Telekora is v0's only consumer); `data_type` ontology in [`data-types.md`](./data-types.md). OPEN: formal per-type JSON Schemas (loose-validation v0, tighten v1); the receiver-key handoff with Telekora (`dtw-handoff.md` §3-4); on/off ramps (deferred to protocol).*
