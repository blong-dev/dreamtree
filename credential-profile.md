# DTW Credential Profile

**Status:** v1 — normative, decided 2026-07-06.
**Owner:** Braedon.
**Reads with:** `data-types.md` (the ontology + credential data types),
`wallet-v0.md` (the holder), `data-model.md` (the attestation substrate),
`telekora/vc-audit.md` (the audit that forced this doc).

This is the **protocol-side credential specification**: the profile every
issuer minting credentials *into* a DTW wallet MUST satisfy — Telekora first
(its "VC Phase 0" = implement this profile), HometownWire's civic bodies next,
third-party issuers after that. DTW owns the standard; issuer surfaces conform
to it as a dependency. Decisions already settled in `data-types.md` /
`wallet-v0.md` are cited, not re-made; exactly two new decisions are recorded
here (D1, D2).

The stance throughout: **compliant with every major standard in the space.**
A DTW credential is never a dialect. If a standards-conformant external
verifier rejects a credential this profile accepts, that is a bug in the
profile.

---

## 1. Envelope — W3C Verifiable Credentials 2.0

- Every credential MUST be a W3C **Verifiable Credentials Data Model 2.0**
  document: `@context` beginning `https://www.w3.org/ns/credentials/v2`,
  `type` including `VerifiableCredential`, `issuer`, `validFrom`,
  `credentialSubject`.
- `validUntil` MUST be present when the credential has any natural expiry
  (compliance certs almost always do); omitted only for genuinely perpetual
  claims.
- Every `@context` URL an issuer introduces MUST be **hosted and stable**
  (immutable once published — version by URL, never edit in place).

## 2. Proofs — Data Integrity, `eddsa-jcs-2022`

- Proofs MUST be W3C **Data Integrity** proofs: `type: DataIntegrityProof`
  with `cryptosuite: eddsa-jcs-2022` (Ed25519 over JCS/RFC 8785 canonical
  JSON). This matches the signing already implemented in Telekora — the fix is
  labeling and required fields, not new cryptography. `eddsa-rdfc-2022` is
  permitted as an additional proof but never required.
- Legacy note: proofs labeled `Ed25519Signature2020` over JCS bytes (the
  pre-profile Telekora shape) are **non-conformant** and MUST NOT be issued
  after an issuer adopts this profile.

### D1 (decided 2026-07-06): dual-proof internally, conformant rendering externally

The dual-proof design (issuer `assertionMethod` + receiver `authentication`
over the same canonical bytes) is a DTW strength — the receiver countersigns
that they actually did the thing. But a second in-credential proof from a
non-issuer trips strict W3C verifiers, which expect holder-binding at
*presentation* time. Resolution — **option (c), both**:

- **Canonical (internal) form:** the stored credential carries BOTH proofs.
  This is what the wallet holds and what DTW-aware verifiers check.
- **Conformant (export) rendering:** any externally-shared copy is the same
  credential with the proof array reduced to the **issuer proof only** —
  byte-identical credential body, so the issuer signature still verifies.
  Holder-binding for external relying parties happens the standard way: the
  holder signs the **Verifiable Presentation** wrapping the credential.
- Reducing proofs MUST NOT re-canonicalize or mutate the credential body;
  the export is a strict subset of the canonical form.

## 3. Identifiers — DIDs

Settled direction (`data-types.md`), made normative:

- **Institutional issuers MUST use `did:web`** (e.g.
  `did:web:telekora.com:tenants:<id>`, or the issuer's own domain — a
  white-label tenant SHOULD graduate to a DID on its own domain). did:web
  gives discoverable DID documents and **key rotation**; issuers MUST serve
  their DID document at the standard well-known path.
- **Receivers (learners/citizens) MAY use `did:key`** as the bootstrap while
  keys are custodied server-side; they migrate to wallet-held identifiers via
  the custody handoff (`telekora/docs/specs/dtw-handoff.md`).
- **Wallets target `did:webvh`** (verifiable history — rotation with an
  auditable key chain), anchored per `wallet-v0.md` ("human-as-key").
- `did:key` is never acceptable for an institutional issuer after this
  profile's adoption: no rotation = one leak poisons every credential.

## 4. Status — Bitstring Status List, revocation AND suspension

### D2 (decided 2026-07-06): both purposes, from day one

Compliance is the launch vertical, and compliance credentials *lapse* (pending
renewal, payment, re-test) as often as they are withdrawn. Therefore:

- Every issued credential MUST carry `credentialStatus` with **two**
  W3C **Bitstring Status List** entries: one `statusPurpose: revocation`
  (permanent withdrawal) and one `statusPurpose: suspension` (temporary —
  reversible).
- Status list credentials are published per issuer, served from the issuer's
  domain, themselves signed per §2.
- Verifiers MUST treat suspended as "not currently valid" and MAY render it
  distinctly from revoked ("lapsed" vs "withdrawn").

## 5. Achievements — Open Badges 3.0

- Completion/achievement credentials (course completed, cert earned) MUST be
  **Open Badges 3.0 `AchievementCredential`s** — the VC-based lingua franca of
  the training/education ecosystem (Credly, Canvas, Accredible interop).
- Fine-grained evidence (e.g. Telekora's per-response, merkle-bound
  credentials) attaches via the OB 3.0 `evidence` property, linking the
  achievement to its proof-of-work. The evidence credentials follow this
  profile too.
- Skill/competency alignment uses OB 3.0 `alignment` targeting a published
  framework (O*NET/ESCO or vertical frameworks, e.g. OSHA competencies);
  optional at first issuance, required for marketplace-listed courses.

## 6. Verification & trust

Signature validity is necessary, never sufficient (`data-types.md` correction:
self-signed did:key credentials otherwise pass). A conformant verifier checks:

1. **Proof** — Data Integrity verification per §2.
2. **Issuer resolution** — DID document fetch (did:web) or decode (did:key).
3. **Status** — both bitstring entries (§4).
4. **Time** — `validFrom`/`validUntil`.
5. **Trust** — the issuer's standing in the **issuer registry**
   (`wallet_issuers`; roadmap: curated list → published did:web registry →
   the public Library registry).

Verification output MUST be tiered honestly, and surfaces MUST render the
tier: **verified-VC** (all five pass, trusted issuer) → **valid-signature**
(1–4 pass, issuer unknown) → **unverified/self-reported**. An unverified
credential must never be rendered as a verified one — the honesty of that
label is the product.

## 7. Presentation & exchange

- Credential sharing uses W3C **Verifiable Presentations**; the holder proof
  lives on the VP (see D1).
- Issuance-into-wallets and presentation-to-verifiers adopt **OID4VCI /
  OID4VP** as the exchange rails when wallet interop ships (EUDI-compatible;
  HAIP profile). Not required for v1 REST issuance, but nothing in an
  implementation may preclude it.
- Selective disclosure (**SD-JWT VC / BBS+**) is reserved for identity-class
  credentials; deferred, tracked, not precluded.

## 8. Relation to the attestation substrate

A profile credential is one serialization of the `data-model.md` atom:
issuer = **A**, `credentialSubject.id` = **C**, claims = **S**,
`validFrom` = **T**, proof = **σ**. The wallet stores the canonical signed
document (`wallet-v0.md`); the substrate's resolution layer (`sameAs` /
`differentFrom`, decay, standing) operates *over* stored credentials and never
mutates them. Nothing in this profile blocks the substrate; nothing in the
substrate excuses non-conformance here.

## 9. Issuer conformance checklist

An issuer is profile-conformant when:

- [ ] VC 2.0 envelope, hosted immutable contexts (§1)
- [ ] `DataIntegrityProof` / `eddsa-jcs-2022` (§2)
- [ ] Dual-proof canonical form + issuer-only export rendering (D1)
- [ ] `did:web` issuer identity with served DID document + rotation plan (§3)
- [ ] `credentialStatus` ×2 (revocation + suspension), signed status lists (D2)
- [ ] Completions issued as OB 3.0 `AchievementCredential` (§5)
- [ ] `validUntil` on expiring credentials (§1)
- [ ] Credentials verify in at least one independent, standards-conformant
      external verifier — the acceptance test for all of the above
