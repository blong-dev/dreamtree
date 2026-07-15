# Seed = Atom Conformance — the leaf model and the photon-native chain

**Status:** ratified design, 2026-07-15 (owner-directed). Kanban: **DT-18**.
**Amends:** `protocol-spec.md` (records the ratification; no doctrine change — this *implements* the doctrine).
**Supersedes nothing** — it conforms the build to the design of record.

## Ratified decisions (owner, 2026-07-15)

1. **The seed is the atom.** One data contribution = one seed = one photon = one
   unit of priced access. The chain docs ("recording a data contribution mints a
   seed... the seed *is* the record"), the reflow data model (the atom), and the
   paper (per-contribution attribution in `L_i`) name the same object.
2. **No collapsing contributions into one seed.** Today's `batch_root` seeds
   (one seed standing for thousands of atoms) violate the unit. Batching is a
   *commitment strategy*, never a unit change: a batch commit registers **N
   leaf-seeds under one Merkle root in one transaction**.
3. **dtvp retires.** It entered the build (`eaf172e`, 2026-07-10) as staking
   expedience, appears nowhere in the protocol spec, and squats on the photon's
   designed role as the native asset. The photon becomes the bond denom.
4. **The chain is `dreamtree`.** No suffix, no successor naming. One wipe +
   fresh genesis gets us there; everything afterward is upgrades.
5. **Photon routing follows the May design**: ingestion photons go to the
   **storer** (`StorerRewardRecipient` = dt-as-storer today; proper first-storer
   distribution when a storer set exists). Mint-to-committer was the rejected
   2026-07-14 plan and stays rejected.

## The convergence rule (photons = *distinct* atoms)

Reflow's D9 replay lands the same atom again when a source is re-fetched:
`atoms_added` (new) vs `atoms_converged` (already known; sigma accrues). The peg
must count **contributions, not observations**:

> A re-observed atom strengthens confidence (sigma); it does not create new
> contribution. **Photons = seeds = distinct atoms.** A batch mints
> `new_count` photons and allocates `new_count` seed ids; its Merkle root may
> cover `leaf_count ≥ new_count` leaves so converged atoms stay provable
> against this batch too.

The chain cannot verify the committer's `new_count` claim (it sees only root +
counts) — the trust boundary is unchanged from today (committer asserts; Merkle
makes it provable; D9 replay is reproducible; attestation and audit catch lies).

**Pure-convergence batches are legal** (implementation delta, 2026-07-15):
`new_count = 0` with `leaf_count ≥ 1` anchors a re-fetch that found nothing new
(D9's 0-new/14-converged shape). It registers provenance — the root is
committed and provable — but allocates no seed ids and mints nothing. The
autonomous worker produces these routinely on re-fetches, so the chain must
accept them or the loop breaks.

---

## Stage 0 — the leaf model (x/seeds)

### Proto

```proto
// New in service Msg:
rpc CommitBatch(MsgCommitBatch) returns (MsgCommitBatchResponse);

message MsgCommitBatch {
  option (cosmos.msg.v1.signer) = "committer";
  string committer   = 1;  // anchoring address (signer)
  string subject     = 2;  // owning wallet DID / producer namespace
  string merkle_root = 3;  // hex; root over the batch's leaf ids (atom ids)
  uint32 leaf_count  = 4;  // total leaves under the root (provable set)
  uint32 new_count   = 5;  // NEW distinct contributions (seed ids + photons minted)
  string kind        = 6;  // the LEAF kind ("record", "kg_claim", ...) — batch_root retired
  string source_ref  = 7;  // opaque locator (e.g. "reflow:gen:61932")
  string data_type   = 8;  // priceable type of the leaves
}
message MsgCommitBatchResponse {
  uint64 first_id = 1;  // seeds [first_id, first_id + new_count)
  uint64 batch_id = 2;
  int64  height   = 3;
}
```

Constraints: `0 < new_count ≤ leaf_count`; `merkle_root` hex-bounded by
`MaxCommitmentBytes`; `kind` non-empty and **must not** be a `*batch_root*`
label (the aggregate is no longer a seed kind).

`MsgCommitSeed` (existing) remains as the batch-of-1 sugar: internally it
becomes `CommitBatch{merkle_root: commitment, leaf_count: 1, new_count: 1}`.
Existing callers (roots via anchord) keep working unchanged.

### State

```go
// Batches: batchId -> Batch{firstSeedId, newCount, leafCount, merkleRoot,
//                           committer, subject, kind, dataType, sourceRef, height, time}
Batches   collections.Map[uint64, seeds.Batch]
BatchSeq  collections.Sequence
// RangeIndex: firstSeedId -> batchId  (ordered; resolves any leaf id by
// descending iteration: greatest firstSeedId <= id, then id < first+newCount)
RangeIndex collections.Map[uint64, uint64]
// Seq (existing) now allocates leaf id ranges: Peek/advance by newCount.
// SubjectIndex (existing) keys (subject, firstSeedId) — PER BATCH, not per
// leaf (11.7M per-leaf rows is the anti-goal). By-subject queries expand.
```

`Seeds` (the per-row map) is **retired**; `Seed` objects are *synthesized* on
read from the containing batch: `Seed{Id: id, Commitment: batch.MerkleRoot,
LeafIndex: id - batch.FirstSeedId, ...batch fields}`. `SeedInfo(id)` (the
licenses seam) resolves through `RangeIndex` — per-leaf pricing and
`AccessGrants` keyed by leaf id work unchanged.

**Implementation deltas (as built, 2026-07-15):**
- `SubjectIndex` is keyed `(subject, batch_id)` — not `first_seed_id` — so
  pure-convergence batches (no seed range) index too, with no key collisions.
- The `Seeds` / `SeedsBySubject` list queries return **batch-level entries**
  (`repeated Batch`): expanding a 10K-leaf batch inline would explode a page.
  Individual leaves resolve via `Seed(id)`; `Batch(id)` returns the raw record.
- `MsgCommitBatch` rejects any kind containing `batch_root` (`ErrRetiredKind`)
  — the aggregate is no longer a seed kind; kind names the leaf.

Genesis import/export moves batches, `Seq`, `BatchSeq`.

### Photon seam

`PhotonHooks.OnRecordSeed(ctx, kind)` → `OnRecordBatch(ctx, kind, newCount)`:
mints `newCount × 10^6 uphoton` to the storer recipient, advances `Minted` by
`newCount`. `Minted` remains the photon (= distinct-atom) count.

### Merkle contract

Unchanged from `reflow/anchor.py` (documented, fixed): leaves are atom ids;
`root(1) = leaf`; pairwise duplicate-last. Per-leaf inclusion proofs derivable
forever off-chain; the chain stores the root and the counts, verifies neither
(same trust model as today's `batch_root`).

---

## Stage 1 — the photon-native chain

- **Base denom `uphoton`**, display `photon`, exponent 6 (bank denom metadata
  in genesis). `1 photon = 10^6 uphoton`; the peg counts photons.
  `app/params/config.go`: `CoinUnit = "uphoton"`; `BondDenom = "uphoton"`;
  dtvp deleted everywhere (`grep -r dtvp` must return zero in code + scripts).
- **Staking:** bond denom uphoton. Default `PowerReduction` (10^6) ⇒ voting
  power = whole photons (~11.7M at genesis). No override needed.
- **Slashing/peg protection (resolved better than planned):** the app wires
  **no x/slashing and no x/evidence module** — nothing in the built chain can
  burn bonded photons, so the photons = seeds peg holds structurally, not by
  parameter. Wiring slashing later (for external validators) must route slashes
  to treasury instead of burning, or accept a documented peg deviation.
- **Gov:** `min_deposit`/`expedited_min_deposit` in uphoton (initial:
  10,000 / 50,000 photons — owner-adjustable at genesis build).
- **No x/mint** (already absent): photon issuance via seeds is the only mint,
  exactly as §Monetary policy demands. Staking rewards = fees (currently 0).
- **Gas:** uphoton denom, `minimum-gas-prices 0uphoton` at launch (lever).

### Genesis (the one wipe)

Chain-id **`dreamtree`**. Genesis app_state carries the corpus:

1. **Exporter** (`scripts/export-genesis-corpus.py`, built): reads reflow PG
   only — one batch per generation with a merkle_root and non-empty leaf set,
   `new_count = atoms_added`, `leaf_count = atoms_added + atoms_converged`,
   pure-convergence generations carried as provenance batches. **Strict peg
   check:** refuses to emit unless `Σ new_count ≡ count(reflow.atoms)`
   (`--force` to override after diagnosis). Emits `corpus.json` (consumed by
   `launch-genesis.sh`) + `mapping.csv` (gen_id → batch_id/first_seed_id, for
   the PG anchor-ref backfill).
   **Roots is NOT in genesis** (delta from the first draft): its proven cron
   re-anchor handles the 67 records post-launch through anchord's unchanged
   single-commit path — simpler cutover, no D1 export machinery.
2. **Keys:** the owner generates/holds all keys, as at the dreamtree-1 launch.
   Claude never generates, reads, or signs with keys.
3. **gentx:** owner bonds a majority of genesis photons (target ~70%,
   owner's call), keeps float liquid for gov deposits + future storers.
4. `launch-genesis.sh` rewritten accordingly (no DTVP_* anywhere).

### Cutover runbook (m3)

1. Pause the reflow worker + anchord (`systemctl stop reflow-worker anchord`
   via the m3 actuator); note last anchored generation.
2. Stop `dreamtreed`; archive `~/.dreamtreed` (post-mortem copy, then delete).
3. Build + install new binary; run exporter → genesis; owner keys + gentx;
   start `dreamtreed` (chain-id `dreamtree`, same ports 16656/16657/16658).
4. **Reset anchor state:** reflow PG `generations.{seed_id, anchor_tx,
   anchor_height, anchored_at} := NULL` for all (genesis carries them now —
   instead re-point: exporter emits a mapping file generation→(batch_id,
   first_seed_id) and a backfill UPDATE marks genesis-carried generations
   anchored with height 0 + genesis batch ids). Roots D1: genesis-carried
   records get anchor_state='anchored' with genesis refs (D1 update via
   Cloudflare API, as drilled 2026-07-1x).
5. Update anchord (batch endpoint pass-through) + `reflow/anchor.py`
   (send `leaf_count`/`new_count`; kind = leaf kind `record`); resume worker.
6. Verify (acceptance below), then re-enable ufw/tunnel surfaces unchanged.

Interim while cut over: new generations anchor as batches on `dreamtree`;
roots' cron sweeps pending records exactly as before via anchord.

### Acceptance (DT-18)

- Unit + e2e green: commit batch of N → seeds `[first, first+N)` queryable
  individually; photon supply +N; converged-only batch (new_count < leaf_count)
  mints only new_count; batch-of-1 sugar identical to old CommitSeed.
- Fresh chain `dreamtree` boots photon-bonded (power = whole photons); gov
  proposal deposit/vote in uphoton passes; `grep -ri dtvp` = 0 hits.
- Cutover: roots 67/67 anchored; all non-empty generations carried in genesis
  (Σ new_count = distinct atom count from PG); worker autonomously lands a new
  observation → atoms → anchored batch on `dreamtree` with no manual step.
- protocol-spec.md gains the ratification + decision-log entry.

---

## Roadmap (subsequent cards; not this implement pass)

- **S2 ownership wiring:** roots DID→address binding credential; seed `subject`
  owner becomes marketplace payee (today: committer). One resolution function.
- **S3 the fabric:** per-wallet chains, DA root, sharding, possession proofs,
  storer set, endowment. The majority of the unbuilt protocol. Until then,
  bodies live in producer stores behind `source_ref` — stated, not hidden.
- **S4 access enforcement:** TEE compute-to-data + PRE + output minimization
  gating body reads against AccessGrants (§Records stack).
- **S5 the instrument:** per-atom seeds + owners + sales + attest projections ⇒
  per-creator `L_i`; the backtest harness (`docs/specs/measurement-backtest.md`)
  studies the levers.
