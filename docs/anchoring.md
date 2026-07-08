# Anchoring: roots and gnosis → the chain

*How off-chain producers write commitments to dreamtree. Status: v0 devnet,
working end to end (`x/seeds` + `anchord`).*

## The shape

Producers do not hold keys and do not speak gRPC. They make one authenticated
HTTP call to **anchord**, a small service co-located with the chain that holds
the single anchor key and broadcasts `MsgCommitSeed`. Broadcasts are serialized
so the account sequence never races.

```
roots (CF Worker, TS) ─┐
                       ├─POST /anchor──▶ anchord ──MsgCommitSeed──▶ dreamtree chain
gnosis (Python)      ──┘   (Bearer)      (m3)         (gRPC :9090)     (x/seeds)
```

What lands on-chain is a **commitment**, never a body: a sha256 digest of a
record, or a Merkle root of a batch. Bodies stay in their home store (roots'
D1, gnosis' warehouse/KG). This is the spec's consensus/data-fabric split.

## The contract

```
POST /anchor          Authorization: Bearer <ANCHORD_TOKEN>
{
  "subject":     "did:web:id.dreamtree.org:w:<uuid>",   // what it's about (a wallet DID)
  "commitment":  "<hex sha256 or merkle root>",          // the anchor; hex, ≤512 chars
  "kind":        "record",                               // producer-owned label
  "source_ref":  "roots:record:<id>"                     // opaque off-chain locator
}
→ 200 { "id": 4, "txhash": "A23F…", "height": 154 }

GET /healthz  → 200 { "ok": true, "height": "153" }
```

Errors: `400` bad input (non-hex commitment, missing kind), `401` bad token,
`502` broadcast/inclusion failure (with the chain's reason). The call blocks
until the tx is included (~2–4s) and returns the assigned seed `id` — store it.

### kinds in use

| producer | kind | subject | source_ref |
|----------|------|---------|------------|
| roots | `record` | wallet DID | `roots:record:<record_id>` |
| roots | `batch_root` | wallet DID (or `""` for cross-wallet) | `roots:batch:<n>` |
| roots | `attestation` | attested wallet DID | `roots:attest:<id>` |
| gnosis | `kg_claim` | claim/entity URI | `gnosis:claim:<id>` |
| gnosis | `kg_batch` | `""` | `gnosis:batch:<cycle>` |

Producers own their kinds; add rows here when you add one.

## Wiring roots (Cloudflare Worker)

roots already produces `record_events`. On create (or a periodic sweep of
un-anchored events), compute the digest over the canonical record bytes and
anchor it, then persist `seed_id` + `txhash` back on the record so the wallet
can display "anchored at height N, tx …".

```ts
// roots/src/anchor.ts
export async function anchorRecord(env: Env, rec: { id: string; walletDid: string; canonical: string }) {
  const digest = await sha256Hex(rec.canonical);           // hex of the record bytes
  const r = await fetch(`${env.ANCHOR_URL}/anchor`, {
    method: 'POST',
    headers: { 'authorization': `Bearer ${env.ANCHOR_TOKEN}`, 'content-type': 'application/json' },
    body: JSON.stringify({
      subject: rec.walletDid,
      commitment: digest,
      kind: 'record',
      source_ref: `roots:record:${rec.id}`,
    }),
  });
  if (!r.ok) throw new Error(`anchor failed: ${r.status} ${await r.text()}`);
  const { id, txhash, height } = await r.json();
  await env.DB.prepare(
    `UPDATE records SET seed_id=?, anchor_tx=?, anchor_height=? WHERE id=?`,
  ).bind(id, txhash, height, rec.id).run();
}

async function sha256Hex(s: string): Promise<string> {
  const buf = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(s));
  return [...new Uint8Array(buf)].map(b => b.toString(16).padStart(2, '0')).join('');
}
```

`env.ANCHOR_URL` = the tunnelled anchord (e.g. `https://anchor.dreamtree.org`),
`ANCHOR_TOKEN` a Worker secret. Batch mode: collect N digests, compute a Merkle
root, anchor once with `kind:'batch_root'`, and keep the leaf→root paths in D1
so any single record stays independently provable.

## Wiring gnosis (Python)

gnosis synthesizes KG claims. A handler (or the nightly enrichment tail) anchors
a per-cycle batch root so the graph's state is provable without putting claims
on-chain.

```python
# gnosis: anchor a KG batch root
import hashlib, httpx

def anchor(commitment_hex: str, kind: str, subject: str = "", source_ref: str = "") -> dict:
    r = httpx.post(
        f"{ANCHOR_URL}/anchor",
        headers={"authorization": f"Bearer {ANCHOR_TOKEN}"},
        json={"subject": subject, "commitment": commitment_hex, "kind": kind, "source_ref": source_ref},
        timeout=30,
    )
    r.raise_for_status()
    return r.json()  # {id, txhash, height}

# e.g. at end of an enrichment cycle:
root = merkle_root(claim_digests)                      # your Merkle over claim hashes
res = anchor(root, "kg_batch", source_ref=f"gnosis:batch:{cycle_id}")
# persist res["id"]/res["txhash"] next to the cycle record
```

Follows the structured-output rule: Python computes the digest and writes the
returned ids to its own store; no model in the loop.

## Deploy (m3)

The chain and anchord run on m3 (10.0.0.2). anchord is reachable from
Cloudflare via a cloudflared tunnel at `anchor.dreamtree.org`.

1. Build from a clean tree at HEAD (per the m3 build lesson): `make install`
   for `dreamtreed`, `go install ./cmd/anchord`.
2. `dreamtreed init` + genesis already done for `dreamtree-devnet-1`; copy the
   home to `/var/lib/dreamtreed`.
3. Create a dedicated `anchor` key (not alice): `dreamtreed keys add anchor
   --keyring-backend test`, fund it in genesis / from the validator.
4. Install the systemd units in `deploy/` (`dreamtreed.service`,
   `anchord.service`), set `ANCHORD_TOKEN` from `infra/.env`.
5. Point a cloudflared tunnel ingress: `anchor.dreamtree.org` → `localhost:9110`.
6. Set roots' `ANCHOR_URL`/`ANCHOR_TOKEN` secrets; set gnosis' env likewise.

Health: `curl https://anchor.dreamtree.org/healthz`. Liveness is measured by
seeds actually landing (per the agent-liveness rule), not by the health ping.

## Honest limits (v0)

- One validator; the chain is "a really over-engineered database" until the
  validator set diversifies. Anchoring is real; decentralization is not claimed.
- anchord shells `dreamtreed` to broadcast. Fine for batched volume; the native
  SDK-client broadcast is the tracked upgrade and doesn't change this contract.
- The anchor key is hosted (like wallet custody). Documented, not hidden.
