#!/usr/bin/env python3
"""Export the reflow corpus as dreamtree genesis batches (the leaf model).

Reads reflow.generations from the gnosis Postgres and emits:
  corpus.json   — app_state fragments: seeds batches + photons minted count
  mapping.csv   — gen_id,batch_id,first_seed_id,new_count (for the PG backfill
                  that marks genesis-carried generations as anchored)

Rules (docs/specs/seed-atom-conformance.md):
  * covers ALL packs (default: reflow, ta) — one batch per generation with a
    merkle_root and a non-empty leaf set
  * leaf_count = atoms_added + atoms_converged (the root's leaf set)
  * new_count  = atoms_added (photons = seeds = DISTINCT atoms; convergence
    does not re-mint — sigma accrues, supply doesn't)
  * pure-convergence generations (atoms_added = 0) are carried as provenance
    batches: no seed range, no photons
  * id-assignment convention: within a batch, the NEW atoms take ids in
    ascending-bytewise atom_id order (re-derivable from PG forever)
  * strict peg check: sum(new_count) must equal sum over packs of
    count(<pack>.atoms); refuse to emit otherwise (--force after diagnosis)

Usage:
  POSTGRES_PASSWORD=... ./export-genesis-corpus.py \
      --pg-host localhost --pg-user gnosis --pg-db gnosis \
      --committer dream1... \
      --subject did:web:id.dreamtree.org:tenants:gnosis \
      --out-dir ./genesis-corpus [--force]
"""

import argparse
import csv
import json
import os
import sys

try:
    import psycopg
except ImportError:  # psycopg2 fallback
    import psycopg2 as psycopg  # type: ignore


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--pg-host", default=os.environ.get("POSTGRES_HOST", "localhost"))
    ap.add_argument("--pg-user", default=os.environ.get("POSTGRES_USER", "gnosis"))
    ap.add_argument("--pg-db", default=os.environ.get("POSTGRES_DB", "gnosis"))
    ap.add_argument("--committer", required=True, help="dream1... address anchoring the corpus (dt as first storer)")
    ap.add_argument("--subject", default="did:web:id.dreamtree.org:tenants:gnosis")
    ap.add_argument("--kind", default="record", help="LEAF kind for the corpus atoms")
    ap.add_argument("--data-type", default="", help="dt.*@v type of the leaves (empty = unpriced)")
    ap.add_argument("--packs", default="reflow,ta", help="comma-separated pack schemas to export")
    ap.add_argument("--out-dir", default=".")
    ap.add_argument("--force", action="store_true", help="emit even if the peg check fails")
    args = ap.parse_args()

    pw = os.environ.get("POSTGRES_PASSWORD", "")
    conn = psycopg.connect(host=args.pg_host, user=args.pg_user, dbname=args.pg_db, password=pw)
    cur = conn.cursor()
    # One REPEATABLE READ snapshot for every query: the peg check compares
    # counts across statements, and a concurrent writer (worker, backfill,
    # enhance) would otherwise race the comparison into false failures.
    cur.execute("begin isolation level repeatable read")
    packs = [p.strip() for p in args.packs.split(",") if p.strip()]

    distinct_atoms = 0
    for pack in packs:
        cur.execute(f'select count(*) from "{pack}".atoms')
        distinct_atoms += cur.fetchone()[0]

    rows = []
    for pack in packs:
        cur.execute(
            f"""
            select %s::text, id, encode(merkle_root,'hex'), atoms_added, atoms_converged,
                   extract(epoch from created_at)::bigint
            from "{pack}".generations
            where merkle_root is not null
              and (coalesce(atoms_added,0) + coalesce(atoms_converged,0)) > 0
            order by id
            """,
            (pack,),
        )
        rows.extend(cur.fetchall())
    if not rows:
        print("no anchorable generations found", file=sys.stderr)
        return 1
    # Deterministic batch order across packs: (created_at, pack, gen_id).
    rows.sort(key=lambda r: (r[5], r[0], r[1]))

    batches = []
    mapping = []
    next_seed = 1
    batch_id = 0
    minted = 0
    skipped_no_root = 0

    for pack, gen_id, root_hex, added, converged, created_epoch in rows:
        added = added or 0
        converged = converged or 0
        if not root_hex:
            skipped_no_root += 1
            continue
        batch_id += 1
        leaf_count = added + converged
        first = next_seed if added > 0 else 0
        b = {
            "id": str(batch_id),
            "first_seed_id": str(first),
            "new_count": added,
            "leaf_count": leaf_count,
            "merkle_root": root_hex,
            "committer": args.committer,
            "subject": args.subject,
            "kind": args.kind,
            "source_ref": f"{pack}:gen:{gen_id}",
            "data_type": args.data_type,
            "height": "0",
            "committed_at": str(created_epoch),
        }
        batches.append(b)
        mapping.append((pack, gen_id, batch_id, first, added))
        if added > 0:
            next_seed += added
            minted += added

    # The peg check: photons = seeds = distinct atoms, exactly.
    if minted != distinct_atoms:
        msg = (
            f"PEG CHECK FAILED: sum(atoms_added over anchorable generations) = {minted} "
            f"but count(reflow.atoms) = {distinct_atoms} (diff {distinct_atoms - minted}). "
            "Atoms exist whose generation has no merkle_root (crashed runs?) or counts drifted. "
            "Diagnose before genesis — the supply must equal the corpus."
        )
        if not args.force:
            print(msg, file=sys.stderr)
            return 2
        print("WARNING (--force): " + msg, file=sys.stderr)

    os.makedirs(args.out_dir, exist_ok=True)
    corpus_path = os.path.join(args.out_dir, "corpus.json")
    with open(corpus_path, "w") as f:
        json.dump(
            {
                "batches": batches,
                "next_id": str(next_seed),
                "next_batch_id": str(batch_id + 1),
                "minted": str(minted),
            },
            f,
        )
    map_path = os.path.join(args.out_dir, "mapping.csv")
    with open(map_path, "w", newline="") as f:
        w = csv.writer(f)
        w.writerow(["pack", "gen_id", "batch_id", "first_seed_id", "new_count"])
        w.writerows(mapping)

    conv_only = sum(1 for m in mapping if m[4] == 0)
    print(f"exported {len(batches)} batches ({conv_only} pure-convergence), "
          f"{minted} seeds/photons (peg vs atoms: {distinct_atoms}), next_id={next_seed}")
    if skipped_no_root:
        print(f"note: {skipped_no_root} generations skipped (empty merkle_root)")
    print(f"wrote {corpus_path} and {map_path}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
