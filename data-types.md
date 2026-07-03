# DreamTree Wallet — `data_type` Registry (v0 seed)

*Started 2026-05-22. The canonical registry of `data_type` keys for wallet records. Every wallet record's `data_type` resolves to an entry here; every Telekora tool's `data_contract_json.reads`/`writes` keys against these.*

*Grounded in (a) DreamTree's shipped `DataSourceType` enum (`/home/b/DreamTree/dreamtree/src/lib/connections/types.ts`), (b) Telekora's `data-architecture.md` (the `learner_data` ontology) and `verifiable-credentials.md` (W3C VC 2.0 issuance design), and (c) the four-surface architecture's wallet-as-data-plane commitment.*

*Forward-compatible with `protocol-spec.md`'s 5-level domain taxonomy and the seeds-as-records model — every `data_type` here becomes a type-node in the protocol's domain taxonomy without rename.*

---

## Naming convention (the un-retconnable part — settle first)

```
<namespace>.<segment>[.<segment>]*@<version>
```

- **Open-ended depth.** Dotted segments after the namespace can be any length — `dt.a.b.c.d.e.f@1` is valid. The protocol's 5-level domain taxonomy (Class → Field → Discipline → Specialty → Sub-specialty, per `protocol-spec.md` §Reputation Dynamics) is a *guideline* for typical depth, not a cap. Use the depth the domain actually needs:
  - Shallow when the domain doesn't have rich hierarchy: `dt.value.work@1` (2 segments).
  - Moderate, taxonomy-aligned: `dt.skill.transferable@1` (2), `dt.identity.proof_of_personhood@1` (2).
  - Deep when the domain genuinely has sub-sub-specialty: `dt.health.medicine.cardiology.diagnostic.ecg.holter@1` (6); `dt.law.criminal.federal.title_18@1` (4).
- **Segment rules**: each segment matches `[a-z][a-z0-9_]*` — starts with a lowercase letter; lowercase letters, digits, underscores only. No hyphens, no slashes. Dots are the only separator. Version trailer is `@<integer>`.
- **Namespace**: `dt` for core types owned by the DreamTree wallet protocol. Third-party / tenant types use their own namespace later (e.g. `org.telekora.*`, `gov.example.*`) — this answers `wallet-spec.md` L3 Q9 ("who can register types") without solving governance now.
- **Version**: explicit, integer, appended with `@`. New schema = new version (`@2`); records carry their version forever, so old payloads always parse. Never retcon `@1`.

### Examples
- `dt.skill.transferable@1` — a transferable skill (DreamTree)
- `dt.story.soared@1` — a SOARED story (DreamTree)
- `dt.credential.learner_response@1` — a W3C VC `TelekoraLearnerResponse` (Telekora)
- `dt.identity.proof_of_personhood@1` — a KYC attestation (DreamTree wallet)
- `dt.health.medicine.cardiology.diagnostic.ecg@1` — sub-specialty depth where the domain warrants it (KYC++ trajectory)

### Rules
- The first segment after the namespace is **stable forever** (it's the type's anchor). Deeper segments and variants can be added; never removed.
- The protocol must reject records with unknown `data_type` keys (fail-closed); the registry is authoritative.
- **No hard depth cap**, but discourage gratuitous nesting. Match the depth to the domain's actual structure, not aesthetic preference.

---

## The seed list (v0)

Grouped by category. Each entry: **key**, **payload sketch**, **PII** (drives `encrypted` flag), **external standard** (where it cleanly maps), and **source** (which existing artifact it generalizes).

### `dt.skill.*` — skills

| key | payload sketch | PII | external | source |
|---|---|---|---|---|
| `dt.skill.transferable@1` | `{ skill_id, name, category, mastery: 1..5, rank, evidence?: text }` | no | ESCO / O\*NET | DT `RankedSkill` + `transferable_skills` |
| `dt.skill.soft@1` | same shape, `category: 'self_management'` | no | ESCO / O\*NET | DT `soft_skills` |
| `dt.skill.knowledge@1` | same shape, `category: 'knowledge'` | no | ESCO / O\*NET | DT `knowledge_skills` |

Note: `dt.skill.transferable`/`soft`/`knowledge` are distinct subcategories rather than a single `dt.skill@1` with a category field, because buyers query by subcategory and the protocol's domain taxonomy splits them at L5.

### `dt.story.*` — narrative

| key | payload sketch | PII | external | source |
|---|---|---|---|---|
| `dt.story.soared@1` | `{ experience_id?, title?, situation, obstacle, action, result, evaluation, discovery }` | **yes** (free-text personal narrative) | — | DT `SOAREDStory` |

### `dt.experience.*` — work / education / projects

| key | payload sketch | PII | external | source |
|---|---|---|---|---|
| `dt.experience@1` | `{ title, organization?, type: 'job'|'education'|'project'|'other', start_date?, end_date?, description? }` | partial (description may be PII) | schema.org Role / Occupation | DT `Experience` |
| `dt.experience.employment@1` | as above, `type: 'job'`, plus `{ skills: [skill_id], outcomes? }` | partial | schema.org WorkExperience | DT `employment_history` |
| `dt.experience.education@1` | as above, `type: 'education'`, plus `{ degree?, field? }` | partial | schema.org EducationalCredential | DT `education_history` |

### `dt.value.*` — what energizes / drains / matters

| key | payload sketch | PII | external | source |
|---|---|---|---|---|
| `dt.value.work@1` | `{ rank, content }` | no | — | DT `work_values`, `RankedItem` shape |
| `dt.value.life@1` | `{ rank, content }` | no | — | DT `life_values`, `RankedItem` |
| `dt.value.compass_statement@1` | `{ statement: text }` | **yes** (personal articulation) | — | DT `values_compass`, `valueCompassStatement` |

### `dt.flow.*` — energy / focus signal (DreamTree differentiator, no clean external standard)

| key | payload sketch | PII | external | source |
|---|---|---|---|---|
| `dt.flow.activity@1` | `{ activity, energy: -2..2, focus: 1..5, logged_date, is_high_flow }` | partial | — | DT `FlowActivity`, `flow_tracking` |

### `dt.career.*` — career planning

| key | payload sketch | PII | external | source |
|---|---|---|---|---|
| `dt.career.option@1` | `{ title, description?, rank: 1..3, coherence_score?, work_needs_score?, life_needs_score?, unknowns_score? }` | no | — | DT `CareerOption`, `career_options` |
| `dt.career.location@1` | `{ city?, region?, country?, preference_rank? }` | **yes** (location is identifying) | — | DT `locations` |

### `dt.budget@1` — financial needs

| key | payload sketch | PII | external | source |
|---|---|---|---|---|
| `dt.budget@1` | `{ monthly_expenses, annual_needs, hourly_batna, benefits_needed?, notes? }` | **yes** (financial PII) | — | DT `BudgetData`, `budget` |

### `dt.personality.*` — type indicators

| key | payload sketch | PII | external | source |
|---|---|---|---|---|
| `dt.personality.mbti@1` | `{ code: 'INTJ'|… (16 codes) }` | no | MBTI 16-type | DT `mbti_code` |

### `dt.competency.*` — assessed competencies

| key | payload sketch | PII | external | source |
|---|---|---|---|---|
| `dt.competency.score@1` | `{ competency_id, level_id, score, assessed_at }` | no | OECD competencies | DT `competency_scores` |

### `dt.idea_tree@1` — brainstorming structure

| key | payload sketch | PII | external | source |
|---|---|---|---|---|
| `dt.idea_tree@1` | `{ root, nodes: [{ id, parent_id, content }], edges: [{ from, to, label? }] }` | partial | — | DT `idea_trees` |

### `dt.list@1` — generic ranked or freeform lists

| key | payload sketch | PII | external | source |
|---|---|---|---|---|
| `dt.list@1` | `{ list_id, name, items: [{ rank?, content }] }` | partial | — | DT `lists`, generic storage |

### `dt.profile.*` — synthesized identity-shaped text

| key | payload sketch | PII | external | source |
|---|---|---|---|---|
| `dt.profile.headline@1` | `{ text }` | **yes** | — | DT `profile_text`, professionalHeadline |
| `dt.profile.summary@1` | `{ text }` | **yes** | — | DT `profile_text`, professionalSummary |
| `dt.profile.display_name@1` | `{ text }` | **yes** | — | DT `display_name` (encrypted) |
| `dt.profile.identity_story@1` | `{ text }` | **yes** | — | DT `identityStory` |

### `dt.dashboard.life@1` — the life-dashboard snapshot

| key | payload sketch | PII | external | source |
|---|---|---|---|---|
| `dt.dashboard.life@1` | `{ … the life_dashboard_* fields }` | **yes** | — | DT `life_dashboard` |

### `dt.response.*` — Telekora learner responses (the silent-wallet writes)

| key | payload sketch | PII | external | source |
|---|---|---|---|---|
| `dt.response.quiz@1` | `{ card_type: 'quiz_mc'|'quiz_tf', question, options?, chosen_index?, chosen_answer?, correct_index?, correct_answer?, correct }` | no | — | Telekora `quiz_response` |
| `dt.response.text@1` | `{ prompt, body }` | **yes** (free-text response) | — | Telekora `text_response` |

### `dt.credential.*` — issued credentials (W3C VC 2.0)

**These are W3C Verifiable Credentials. The payload IS the full VC document.** The wallet stores the canonical signed JSON.

**Verification is two checks, not one — do not conflate them.** The `did:key` embeds its public key, so *signature integrity* ("these exact bytes were signed by the key named in the proof") needs no resolver. But that is only half of verification. It does **not** establish *issuer authenticity* — that the named key actually belongs to Telekora-tenant-Acme rather than to whoever generated the credential. A `did:key` is self-asserting: anyone can mint a keypair, sign a VC claiming `issuer: did:key:<their-key>` + `"…attested by tenant Acme"`, and it passes a signature-only check. Authenticity requires resolving the issuer DID against a **trusted issuer registry** and confirming that issuer has standing to attest the claim (e.g. owns the referenced course). Concretely, a verifier must:

1. **Signature** — verify each proof against the key in its `did:key` (no resolver; local).
2. **Issuer authenticity** — resolve the issuer DID against the issuer registry; reject DIDs that aren't registered issuers, and confirm standing (issuer's tenant owns the referenced course/claim).
3. **Proof completeness** — require BOTH proofs present and role-bound: exactly one `assertionMethod` proof whose DID is the (registered) issuer, and exactly one `authentication` proof whose DID equals `credentialSubject.id`. A single-proof or wrong-role credential must fail.

The **issuer registry is a resolver with a swappable backend**, in increasing order of decentralization: **v0** — the issuer's keystore (Telekora `issuer_keys` / `receiver_keys`), so verification through the platform is sound today; **vNext** — a published `did:web` / signed issuer registry, verifiable off-platform without trusting a platform API; **vision** — the public provenance ledger (per `ARCHITECTURE.md`, the Library, not the per-wallet chain) as the issuer + attestation registry. Same interface; the ledger is one backend, not a prerequisite. Design verification against this interface now — a credential minted with a stable issuer DID, a standard proof suite, and a content-addressable id (the merkle tree already provides this) is registry-portable and can be adopted by the ledger later with no re-issue. *This corrects the earlier "verification is local, no resolver needed" framing, which described only step 1 and would ship a forgery hole (self-signed credentials pass) into the wallet-v0 build. See Telekora `verifiable-credentials.md` and the credential-trust-anchor work (gnosis GNS-810); the stable-DID choice is the `did:webvh`-vs-`did:key` question in GNS-808.*

| key | payload | PII | external | source |
|---|---|---|---|---|
| `dt.credential.learner_response@1` | the full Telekora `TelekoraLearnerResponse` VC (issuer + receiver dual-proof, content-addressable merkle tree) | yes-at-rest (encrypt under user's data key in the wallet; selective disclosure at presentation is v1+) | **W3C VC 2.0** | Telekora `verifiable-credentials.md` |
| `dt.credential.course_completion@1` | a W3C VC asserting course completion (issued by Telekora tenant) | yes-at-rest | **W3C VC 2.0** + Open Badges 3.0 | Telekora roadmap |
| `dt.attestation@1` | a generic W3C VC envelope for non-course attestations (manager vouches, peer review, etc.) | yes-at-rest | **W3C VC 2.0** | future — protocol-spec attestation layer |

Note: **encrypted-at-rest in the wallet by default** even though VCs are externally verifiable, because the wallet's owner-sovereignty stance says cleartext-on-server is the wrong default (per `wallet-spec.md` L2 Q5 hosted-with-blindness-when-logged-out). At presentation time the holder decrypts and presents (or in v1+, uses selective disclosure to present subsets). This is the wallet's privacy posture above and beyond the VC's signature properties.

### `dt.outcome.*` — outcomes (validating or refuting prior attestations)

Outcomes are themselves attestations, signed by a reporter who has standing to observe the outcome — the hospital reporting a patient outcome, the employer reporting a hire's performance, the appellate court overturning the trial court. Per `protocol-spec.md` §Reputation Dynamics, outcomes use the same machinery as any attestation (review window, cred recursion, paper-shape aggregation, time horizons) — special only in what they do to R: they trigger the `M_O` propagation chain (contributor + the attestation chain that staked on them). Outcomes can themselves be refuted; the 2× asymmetry recurses (a refuted outcome reverses + penalizes the original reporter).

| key | payload sketch | PII | external | source |
|---|---|---|---|---|
| `dt.outcome.validated@1` | `{ subject_attestation_ref, contributor_wallet_id, description, evidence?: text, observed_at }` — confirms the underlying attestation / work | yes-at-rest (context-dependent; default encrypted) | **W3C VC 2.0** (outcomes can be issued as VCs) | protocol-spec §Outcome propagation + `M_O` |
| `dt.outcome.refuted@1` | same payload + `refutation_reason: text` — refutes the underlying attestation | yes-at-rest | **W3C VC 2.0** | protocol-spec |
| `dt.outcome.partial@1` | same + `partial_scope: text` — partial confirmation/refutation. Deferred to v1; v0 forces binary. | yes-at-rest | **W3C VC 2.0** | future |

Note: an outcome's `M_O` magnitude is **computed by the protocol** (formula in `protocol-spec.md` §Reputation Dynamics), not declared by the reporter — the reporter can include a `magnitude_hint`, but the protocol clamps it to `min(M_cap, β · S(att, t_issuance) · √cred(reporter))`.

### `dt.identity.*` — identity proof (the human-as-key anchor)

| key | payload sketch | PII | external | source |
|---|---|---|---|---|
| `dt.identity.proof_of_personhood@1` | a W3C VC issued by a KYC provider (Persona / Onfido / Jumio / Trulioo). Two subtypes possible: **full** (provider attests identity details — name, document) or **zero-knowledge** (provider attests `is_verified_real_human + distinct_from_others` with no details). The privacy posture is open. | **yes** | **W3C VC 2.0** + provider attestation schema | wallet-v0 §1 (verification tier) |
| `dt.identity.did_key@1` | `{ did: 'did:key:z6Mk…', public_key_multibase, generated_at }` — the wallet's own receiver key, when the wallet holds it (post Telekora handoff per `dtw-handoff.md` §4) | partial (DID is correlatable) | **W3C did:key** | dtw-handoff.md §3 hybrid signing |

---

## PII classification — what drives `encrypted = 1`

The `encrypted` flag on each `wallet_record` is **derived from the type**, not set per record. The registry holds the classification. v0 rule:

- **PII (always encrypted at rest)**: types marked PII in the tables above. Free-text personal narrative (`dt.story.soared`, `dt.response.text`, profile.*), financial (`dt.budget`), locating (`dt.career.location`), identity-revealing (`dt.identity.proof_of_personhood`).
- **Non-PII (clear at rest)**: typed structured data (skills, values rankings, MBTI code, competency scores, career options).
- **Credentials (encrypted at rest by default)**: even though VCs are externally verifiable, encrypt-at-rest in the wallet preserves owner sovereignty. Decrypt on presentation; selective disclosure later.
- **Partial**: types where some fields are PII (e.g. `dt.experience` description) and some aren't. v0: encrypt the whole payload if any field is PII. v1+ refinement: field-level encryption inside the payload, leaving structured fields queryable.

---

## External standards alignment ("Movement, not Empire")

Where the mapping is clean, we align — so a wallet record can be exported / federated to standard-aware consumers without translation:

| DreamTree category | External standard | Mapping notes |
|---|---|---|
| `dt.skill.*` | **ESCO** (European Skill/Competence/Qualification) + **O\*NET** + **ISCO-08** | Map `skill_id` → ESCO concept URI when available |
| `dt.experience.*` | **schema.org** (`WorkExperience`, `EducationalCredential`, `Occupation`) | Add `@context` references at export time |
| `dt.credential.*` / `dt.attestation@1` | **W3C VC 2.0** (May 2025) + **Open Badges 3.0** | Payload IS the VC document; native alignment |
| `dt.identity.*` | **W3C DID Core 1.0** + **W3C VC 2.0** | `did:key` resolution is local |
| `dt.personality.mbti@1` | MBTI 16-type code | Trivial enumeration |
| `dt.competency.score@1` | **OECD competencies** | Map `competency_id` → OECD taxonomy node |

DreamTree differentiators with **no clean external standard** — preserved as `dt.*` natives:
- `dt.story.soared@1` (SOARED is DreamTree's framework)
- `dt.flow.activity@1` (flow-tracking shape is DreamTree's)
- `dt.idea_tree@1` (the structural graph)
- `dt.dashboard.life@1` (the life-dashboard composite)

---

## Implementation — the `data_type_registry` table / seed file

The registry can live either in D1 (a `data_type_registry` table) or as a JSON seed bundled with the wallet service. Recommendation: **JSON seed file checked into the wallet repo**, hot-loaded at boot. Reasons:
- It's authoritative protocol metadata, not user-mutable runtime state.
- Versioning the registry = git history.
- Telekora's `tool_types` registry (`tool_types.writes_data_types`, `reads_data_types`) validates against this same JSON at tool-publish time — single source of truth.

Schema per entry (the JSON file format):

```jsonc
{
  "key": "dt.skill.transferable",
  "version": 1,
  "category": "skill",                     // top-level group
  "display_name": "Transferable skill",
  "payload_schema": { /* JSON Schema */ },
  "pii": false,                            // drives wallet_records.encrypted
  "encrypt_credential_payloads": false,    // overrides PII for VC types
  "external_standards": ["esco", "onet", "isco-08"],
  "source": { "dreamtree": "transferable_skills", "telekora": null },
  "deprecated": false,                     // never delete; mark dead
  "successor": null                        // when @1 → @2, point successor here
}
```

JSON Schemas for each payload are TBD as follow-up work (loose-validation in v0, tightened in v1) — same staging as Telekora's tool `input_schema_json` (`alpha.md`: "loose validation in Alpha; formal JSON Schemas are a follow-up that doesn't gate the build").

---

## Versioning & evolution policy

- **New schema = new version.** Add fields → bump version. Old records keep parsing under their `@N`.
- **Never retcon a version.** Once `dt.skill.transferable@1` is in the wild, its schema is frozen. If you need to change it, publish `@2` and point `@1.successor = "dt.skill.transferable@2"`. Readers should accept both.
- **Deprecation, not deletion.** A retired type stays in the registry with `deprecated: true` so historical records still type-check.
- **Subcategory additions** (e.g. adding `dt.skill.creative@1`) don't bump existing versions.
- **The successor field is the migration path** — tools can read `successor`'s schema if they encounter an old version.

---

## Open questions (worth surfacing; don't block the build)

1. **Selective disclosure for credentials at presentation time** — v0 stores VC payloads encrypted-at-rest and decrypts whole for export/presentation. SD-JWT and BBS+ let the holder present a subset without revealing the rest. Spec-wise this is a v1+ enhancement on top of the `dt.credential.*` family.
2. **Field-level encryption inside `payload`** — v0: encrypt whole payload if PII. v1+: structured encryption so non-PII fields stay queryable while PII fields stay encrypted (matches the existing PII_FIELDS pattern in the DreamTree app, but applied per-record-type rather than per-table).
3. **Third-party / tenant namespace registration** — `org.telekora.*`, `gov.example.*`, etc. v0: not enabled (only `dt.*` accepted). v1+: a registration flow, possibly with the protocol's reputation layer gating "who can register a type."
4. **Cross-type relationships** — a `dt.story.soared` references an `dt.experience` via `experience_id`. A `dt.skill.transferable` references the master skill via `skill_id`. v0: foreign keys are just text fields; v1+ the protocol's per-wallet chain makes these provenance links cryptographic.
5. **The `dt.identity.proof_of_personhood@1` payload shape** — full identity vs. ZK-attestation-of-personhood. Resolved separately (`protocol-spec.md` §Identity privacy posture).

---

## Wire to `wallet-v0.md`

This registry is the substance behind `wallet-v0.md` §2 ("the ontology = the set of `data_type` keys"). Wire as:
- Wallet service loads `data-types.json` at boot.
- Every `POST /wallet/{id}/records` validates `data_type` against the registry; rejects unknown.
- Every record's `encrypted` flag is set from `registry[data_type].pii` (overridden to `true` for credential types via `encrypt_credential_payloads: true`).
- Telekora tools' `data_contract_json.writes_data_types` / `reads_data_types` validate against the registry at publish.
- Wallet export annotates each record with `external_standards` for federated consumers.

---

*Last updated: 2026-05-22 — v0 seed registry, grounded in DreamTree's 21 `DataSourceType` + Telekora's `learner_data` ontology + W3C VC 2.0 from `verifiable-credentials.md`. Open: formal per-type JSON Schemas; third-party namespace registration; selective-disclosure at presentation; field-level encryption.*
