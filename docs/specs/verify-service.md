# verify.dreamtree.org — the verification service (v0)

Status: BUILD (owner-approved 2026-07-18: "do it. I don't love the idea of a
door into M3 from the public though.")
Basis: the observation-is-attestation live proofs (2026-07-17: new-atom proof,
converged-atom proof vs on-chain batches) + the 2026-07-17 market/agent
research (no incumbent for hosted zero-auth machine-first verification;
stamper standing is the unoccupied property).

## The one call

```
GET /verify/{sha256-hex}          (zero-auth, public)
→ 200 verdict | 202 pending | 404 unknown-after-resolution
```

Graded verdict — never a bare boolean, never a probability:

```json
{
  "status": "observed" | "converged" | "not_found",
  "atom_id": "…", "c_handle": "…", "t": "…",
  "sigma_arrivals": 17,
  "generation": 82698, "owner_generation": 76930,
  "merkle_root": "…", "proof": [["R","…"], null, …],
  "anchor": {"chain_id": "dreamtree", "txhash": "…", "height": 54132,
              "batch_id": 63520, "subject": "did:web:id.dreamtree.org:tenants:gnosis",
              "committer": "dream1…"},
  "standing": {"address": "dream1…", "reputation": "0.000000", "note": "…"},
  "resolved_at": "…"
}
```

`status=converged` is first-class: it means independently re-observed
(σ > 1 arrivals), the strongest signal the log offers. The `proof` is a
self-contained Merkle path any caller can re-verify offline against the
`merkle_root`, which is the value committed on-chain in `anchor` — auditable,
not an oracle (the SCITT/Rekor lesson).

## Architecture — m3 dials OUT, never listens

```
agent → Worker (verify.dreamtree.org, CF)   ←  public, zero-auth reads
             │  D1: verdicts cache + pending queue
             ▲  (Bearer VERIFY_SYNC_TOKEN, both directions initiated by m3)
   m3 verify-resolver (container, poll loop) — OUTBOUND HTTPS only
             │ reads reflow PG (manifests, generations, anchors)
             │ standing via anchord GET /standing (localhost, Bearer)
```

- **No inbound path to m3.** The Worker never contacts m3; the resolver polls
  `POST /queue/pull` (drain pending hashes) and answers `POST /queue/resolve`
  (store verdicts), both outbound from m3, both Bearer-gated.
- **Lazy index**: D1 stores only hashes someone actually asked about. First
  query answers in one resolver poll cycle (~3-5s, HTTP 202 + Retry-After in
  between); every later query is a cache hit. `not_found` verdicts carry a
  short TTL (the atom may be observed later); positive verdicts are immutable
  except `sigma_arrivals`/`standing`, which re-resolve on a long TTL.
- **What the resolver reads**: `generation_leaves` (GNS-934 manifest) →
  inclusion proof; `generations` → root/anchor_tx/height/seed_id; committer
  standing via a new read-only `GET /standing` on anchord (anchord already
  holds chain access + auth; ~30 lines, keeps the chain's RPC loopback-only).

## MCP

The same Worker speaks MCP (streamable HTTP, authless — the canonical
Cloudflare remote pattern; roots' /mcp is the in-house precedent): one tool,
`verify_observation(hash)`, returning the identical verdict object as
structured content. Registry publication (`server.json`, io.github namespace)
after the endpoint has burned in.

## What v0 is NOT

- Not a C2PA validator yet — DT-22 folds manifest validation into the same
  verdict as a second envelope type; the verdict schema already leaves room
  (`status` vocabulary extends, `anchor` stays).
- Not a stamping API — stamping stays anchord-side (paid writes come later;
  free reads are the network effect, per sigstore).
- Not serving giant-generation proofs at the edge: proofs are computed by the
  resolver on m3 where the manifest lives; the Worker only stores/serves.

## Ops

- Worker + D1 in the dreamtree repo (`verify/`), deployed with wrangler;
  route `verify.dreamtree.org` (own project — never touches other projects'
  routes).
- Resolver = reflow module (`reflow/verify_resolver.py`), m3 compose service
  on the ingest image, actuator-allowlisted like its siblings.
- `VERIFY_SYNC_TOKEN`: operational bearer (ANCHORD_TOKEN class), generated at
  deploy, stored as a wrangler secret + m3 env; never printed.
