# Trust Layer — making the chain enforce what the protocol claims

**Status:** DESIGN, 2026-07-18. Agent-produced from the roots/dreamtree security
audit (same date). PENDING OWNER TRIAGE — nothing here is a decision.
**Basis:** the 2026-07-18 audit (roots credential layer + chain modules + anchord/
verify seams), read against `comb-spec-vs-chain.md` (DT-21), `seed-atom-conformance.md`
(DT-18), `verify-service.md` (verify.dreamtree.org, BUILD).
**File:line refs are to** `/home/b/quorum/dreamtree/dreamtree` unless noted.

---

## 0. The one question

> Can a relying party who does **not** trust the operator verify a dreamtree
> claim — that a specific piece of content was contributed, by an identified
> party, at a point in time, and is committed on an immutable ledger?

Today the honest answer is **no**. The chain is a well-built ledger that trusts
its single operator completely. The credential layer in roots (`eddsa-jcs-2022`,
Ed25519, issuer-binding) is real cryptography; the chain underneath it verifies
no signatures and no content. The trust layer is the set of cryptographic
bindings that make the answer **yes** — and it is exactly the property the
C2PA/CAWG earworm strategy sells to the outside world (transported trust). We
must not invite external scrutiny of the chain (which CAWG participation does)
until the gap between the protocol's language and the chain's enforcement is
closed, or explicitly and publicly scoped as "v0, operator-trusted."

---

## 1. Current trust model (what is verified vs trusted)

| Property | Mechanism today | Verified by a third party? |
|---|---|---|
| Tx author is the account holder | Cosmos SDK tx signature | **Yes** (SDK built-in) |
| A seed's merkle root commits to real content | none — root stored as opaque hex ≤512 chars (`x/seeds/keeper/msg_server.go:77-101`) | **No** |
| `new_count` (→ photons minted) is truthful | committer-asserted, minted on trust (`x/photons/keeper/mint.go:18-30`) | **No** |
| An attestation reflects examined work | none — `MsgAttest` has no proof/digest/signature over content (`proto/dreamtree/attest/v1/tx.proto`) | **No** |
| Reputation reflects verified work | accrues from well-formed message receipt, cred-weighted (`x/attest/keeper/hooks.go:35`) | **No** |
| The anchored commitment came from an authorized party | `ANCHORD_TOKEN` bearer; anchor key in **unencrypted** `test` keyring (`deploy/anchord.service`) | **No** (token-gated pass-through) |
| An atom is included under the on-chain root | **emerging**: `verify-service.md` returns a self-contained merkle proof — but the resolver that produces it is not yet running, and merkle verification is off-chain | **partial / in flight** |
| Canonical bytes signatures are made over | roots does `eddsa-jcs-2022`; the Go chain has no JCS (the `jcs/` package is golden vectors + a **Python** reference; reflow-Go unwritten) | **No parity** |

The single most load-bearing fact: **whoever holds `ANCHORD_TOKEN` can mint up
to `MaxBatchNewCount` (default 1,000,000) photons per tx against a fabricated
root, with no global supply cap.** Economic integrity rests entirely on trusting
that one token holder.

---

## 2. Target trust model

A relying party, holding only (a) the content bytes and (b) the chain's public
state, can verify without trusting the operator:

1. **Inclusion** — the content's leaf hash is in the batch whose merkle root is
   committed on-chain (merkle path re-verification). *verify-service.md already
   designs this; it needs to run and its inputs need to be trustworthy.*
2. **Authorship** — the commitment carries a signature by a key bound to a
   resolvable DID, and that DID is the claimed contributor. *Missing.*
3. **Mint integrity** — the photons minted for a batch equal the number of
   genuinely new leaves, bounded by a cap the chain enforces. *Missing.*
4. **Attestation binding** — an attestation carries a signature over a digest of
   the work it references, so "who vouched for what" is cryptographic, not
   asserted. *Missing.*
5. **Canonicalization parity** — every signature verifies against the same
   canonical bytes regardless of which language produced it. *Missing (Go JCS).*

Inclusion is the property the earworm story most needs and is closest to done.
Authorship + mint integrity are the properties that stop the operator (or a
leaked token) from fabricating value. Attestation binding is the deepest and can
follow.

---

## 3. Workstreams

Each: **what**, **why**, **effort** (single-dev engineering days, ±50%),
**dependencies**, **owner decisions**.

### W1 — Signed commitments (authorship) · ~5–8d
**What:** `MsgCommitSeed`/`MsgCommitBatch` carry a `committer_sig` and a
`committer_did`; the seeds keeper verifies the signature over a canonical digest
of `(merkle_root, leaf_count, new_count, subject, source_ref)` against the key
resolved from `committer_did`, and stores the DID on the batch. anchord signs
with a key whose DID is registered (W7 gives it a real key).
**Why:** binds "who committed" to "what was committed" — property 2. Without it,
authorship is just "whoever held the token."
**Depends on:** W5 (canonical digest), W7 (real signing key), a key→DID registry
(a `Verified`-set extension or an on-chain key record).
**Owner decisions:** (a) verify on-chain (every validator recomputes — costs
block time) vs at anchord with the signature recorded on-chain for later audit;
(b) key resolution source — reuse the `x/reputation` `Verified` set, or a new
key registry module.

### W2 — Content binding / merkle verification · ~4–6d
**What:** on `CommitBatch`, verify that `new_count ≤ leaf_count` **and** that the
submitted root is well-formed for the declared `leaf_count` (a structural merkle
check), and record enough that verify-service can prove any leaf's inclusion.
Full "the leaves are real content" can't live on-chain (the chain never sees
content) — so the binding is: root ↔ leaf_count is structurally sound, and
verify-service (W4) closes content↔leaf off-chain with a re-verifiable proof.
**Why:** property 1 + the precondition for honest minting. Today the root is
opaque.
**Depends on:** the reflow batch builder emitting a proof-friendly root
(seed-atom-conformance.md already moves to N-leaves-under-one-root).
**Owner decisions:** how much to enforce on-chain vs delegate to the auditable
off-chain proof (the SCITT/Rekor "auditable not oracle" line verify-service
already takes).

### W3 — Mint integrity · ~2–3d
**What:** a global photon supply cap (or a per-epoch mint ceiling) enforced in
`x/photons`, and mint tied to verified new leaves from W1/W2 rather than the
bare asserted `new_count`. At minimum, ship the cap now (small, independent).
**Why:** caps the blast radius of a fabricated commitment from 1M/tx-unbounded
to a governed ceiling — property 3.
**Depends on:** nothing for the cap; W1/W2 for full "mint follows verified work."
**Owner decisions:** cap value / epoch schedule (governance param).

### W4 — The verifier, real and running · ~6–10d
**What:** stand up the verify-service resolver (`verify.dreamtree.org`): the
process that pulls jobs, recomputes merkle inclusion + resolves the anchor +
reads standing, and writes back the self-contained proof. Bring its code into
the repo and audit it — today the worker (`verify/src/index.ts`) stores whatever
the off-repo resolver posts, **verbatim and unre-checked** (`index.ts:207-209`).
**Why:** this is where inclusion actually gets proven for a relying party. The
worker is a cache; the resolver is the trust primitive.
**Depends on:** W2 (proof-friendly roots). Aligns with the in-flight
`verify-service.md` BUILD.
**Owner decisions:** the owner already flagged the concern — "I don't love the
idea of a door into M3 from the public." The dial-out architecture answers it;
confirm the resolver never listens.

### W5 — Go JCS (canonicalization parity) · ~3–5d
**What:** implement RFC 8785 JCS in Go against the existing golden vectors
(`jcs/vectors.json`, 30/30), wire it into a CI target (today nothing runs the
vectors), and use it for every canonical digest the chain/anchord signs or
verifies.
**Why:** a signature made by roots (`eddsa-jcs-2022`) or by anchord must verify
against **byte-identical** canonical input on the chain/verifier. Divergent
canonicalization = signatures that silently fail or, worse, a second
canonicalizer that disagrees on an edge case. The ES6 number rules are the trap;
the vectors already encode the discriminators.
**Depends on:** nothing. Unblocks W1, W6.
**Owner decisions:** none — pure implementation. **Recommend doing this first**
(it's cheap, it's the substrate everything signed depends on, and the vectors
already exist).

### W6 — Attestation content binding · ~5–8d
**What:** `MsgAttest` carries a `subject_digest` (a hash of the attested content
or its canonical descriptor) and a signature over it; the keeper verifies before
enqueuing any reputation effect. Reputation then flows from a signed statement
about a specific artifact, not from message receipt.
**Why:** turns "assertion-as-work" into "attestation-as-work" — the CRITICAL gap
between the protocol's name and its behavior. Deepest change; touches the
reputation hook path.
**Depends on:** W5, W1's key-resolution.
**Owner decisions:** what the digest commits to (raw content hash vs a canonical
claim descriptor) — a protocol-doctrine call, not just engineering.

### W7 — Mint-authority key custody · ✅ DONE 2026-07-18
**What (done):** the `dt-validator` account/mint key (`dream10ahjsyp8…`) was
moved off the plaintext `test` keyring onto a dedicated TPM-unlocked LUKS2
vault, mirroring the hwsignd pattern:
- Vault `/var/lib/anchord/keys.luks` (LUKS2), TPM2 auto-unlock (PCR 7, sha256),
  mounted at `/opt/anchord/secure`; `crypttab`+`fstab` `nofail`.
- The keyring moved to the vault; `<ANCHORD_HOME>/keyring-test` is now a symlink
  to `/opt/anchord/secure/keyring-test`. The plaintext original was shredded
  from the unencrypted disk — the key exists **only** inside the vault.
- `anchord.service` gained `RequiresMountsFor=/opt/anchord/secure` so the
  symlink resolves at boot; backup.sh tars the **encrypted** container only.
- Recovery key enrolled (keyslot 2) and stored off-host at
  `infra/anchord-luks-recovery.txt` (0600) on m1; **verified** it unlocks the
  volume (wrong-key control fails). Same verified for the hwsign vault.
- Proven: TPM boot-unlock cycle works with no passphrase; a live anchor after
  migration minted a seed with committer `dream10ahjsyp8…` (unchanged address).

**Residual (owner decisions, not blockers):**
1. **Rotation.** The key was plaintext-at-rest historically (and in old
   backups). Encrypt-at-rest fixes it *going forward*; the only true
   remediation for the historical exposure is rotating to a fresh
   mint/validator account key — a chain-governance operation
   (`StorerRewardRecipient` routing + validator operator account), not a
   custody swap. Deferred as a separate decision.
2. **`file`-backend hardening.** The vault keeps the `test` backend (key
   plaintext *within* the encrypted volume, readable while mounted — same model
   as hwsignd). If the threat model tightens (multi-operator, live-compromise
   resistance), upgrade to the Cosmos `file` backend (key blob always encrypted,
   decrypted only transiently per-sign) with a TPM-sealed systemd credential fed
   to dreamtreed on stdin. Requires an anchord code change; scoped, not done.
3. **Consensus key.** `config/priv_validator_key.json` (block-signing) is still
   plaintext at rest. Lower value on a single-validator chain (double-sign
   slashing is moot), but the same vault could hold it later.

### W8 — did:webvh history verification · ~3–5d
**What:** verify the SCID + entry-hash chain in `resolve.ts` /
`resolveWebvhLatest` (roots) instead of TOFU-trusting the latest entry
(`historyVerified: false` today, `roots/src/credentials/resolve.ts:224-233`).
**Why:** completes DID rotation trust — a resolved key is provably the current
one in an unbroken signed history. Lower urgency (roots already flags it
honestly), but required before did:webvh is load-bearing.
**Depends on:** W5 (the entry hashes are over canonical JSON).
**Owner decisions:** none.

---

## 4. Sequencing

Two independent quick wins ship immediately, no dependencies, high value:
- **W7 (key custody)** — ✅ DONE 2026-07-18. The mint key is off plaintext, on a
  TPM-LUKS vault; recovery verified.
- **W3 cap (1d slice)** — bounds fabricated-mint blast radius while the rest
  lands. **Now the top remaining quick win** — with the key encrypted at rest,
  the residual mint risk is a leaked `ANCHORD_TOKEN` minting against a fabricated
  root; a supply/epoch cap bounds that blast radius. Recommend next.

Then the substrate:
- **W5 (Go JCS, 3–5d)** — everything signed depends on it.

Then the trust core, in dependency order:
- **W1 (signed commitments)** → **W2 (content binding)** → **W4 (verifier real)**
  — these three together deliver properties 1–3: a relying party can verify
  inclusion + authorship + bounded mint. This is the minimum that lets the CAWG
  story survive outside scrutiny.
- **W6 (attestation binding)** and **W8 (webvh history)** follow — they deepen
  the model but aren't gating for the earworm claim.

Rough total to "a relying party can verify inclusion + authorship + mint without
trusting the operator": **~20–30 engineering days**, of which the first ~3
(W7 + W3-cap + start W5) remove the scariest risks.

## 5. What this unlocks for the strategy

The C2PA/CAWG positioning transports trust from dreamtree identities into
Content Credentials. The moment W1+W2+W4 land, a HometownWire signed clip's
`org.dreamtree.anchor` assertion (already in every hwsignd manifest) points at a
commitment a stranger can independently verify — inclusion **and** authorship —
which is the actual product: "provenance you don't have to take our word for."
Until then, the anchor assertion is honest-but-operator-trusted, and any public
CAWG claim about it must say so.

## 6. Owner decisions blocking a start

1. **On-chain vs at-anchord signature verification** (W1) — block time vs
   audit-later.
2. **Key resolution source** (W1) — extend `Verified` set vs new key registry.
3. **How much content binding lives on-chain** vs the auditable off-chain proof
   (W2/W4).
4. **Mint cap value / epoch** (W3).
5. **Attestation digest semantics** (W6) — content hash vs canonical claim
   descriptor. *(Doctrine, not engineering.)*
6. **Custody mechanism** for the mint key (W7).

None of W7, W3-cap, or W5 needs a decision to start — they are the recommended
first moves.
