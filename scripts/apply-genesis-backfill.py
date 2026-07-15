#!/usr/bin/env python3
"""Mark genesis-carried generations as anchored in the gnosis Postgres.

Run ONCE, after the `dreamtree` chain is live, against the mapping.csv the
exporter emitted (docs/specs/seed-atom-conformance.md, cutover step 4). For
each carried generation it sets:

  seed_id       = first_seed_id   (0 for pure-convergence batches — non-NULL,
                                   so sweep_unanchored never re-selects them)
  anchor_tx     = 'genesis:batch:<batch_id>'
  anchor_height = 0
  anchored_at   = now()
  leaf_count    = atoms_added + atoms_converged (requires v3 migration 016)

Idempotent: rows already carrying a genesis anchor_tx are skipped. Refuses to
touch rows that look live-anchored (anchor_tx set and not genesis) unless
--reset-live is passed — after a wipe, the OLD dreamtree-1 refs are stale and
must be reset, which is exactly what --reset-live is for. Nothing here ever
touches atoms or merkle roots (never edit the log).

Usage:
  POSTGRES_PASSWORD=... ./apply-genesis-backfill.py \
      --pg-host localhost --pg-user gnosis --pg-db gnosis \
      --mapping ./genesis-corpus/mapping.csv [--reset-live] [--dry-run]
"""

import argparse
import csv
import os
import sys

try:
    import psycopg
except ImportError:
    import psycopg2 as psycopg  # type: ignore


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--pg-host", default=os.environ.get("POSTGRES_HOST", "localhost"))
    ap.add_argument("--pg-user", default=os.environ.get("POSTGRES_USER", "gnosis"))
    ap.add_argument("--pg-db", default=os.environ.get("POSTGRES_DB", "gnosis"))
    ap.add_argument("--mapping", required=True)
    ap.add_argument("--reset-live", action="store_true",
                    help="overwrite stale dreamtree-1 anchor refs (post-wipe)")
    ap.add_argument("--dry-run", action="store_true")
    args = ap.parse_args()

    rows = list(csv.DictReader(open(args.mapping)))
    if not rows:
        print("empty mapping", file=sys.stderr)
        return 1

    conn = psycopg.connect(host=args.pg_host, user=args.pg_user, dbname=args.pg_db,
                           password=os.environ.get("POSTGRES_PASSWORD", ""))
    conn.autocommit = False
    cur = conn.cursor()

    updated = skipped_done = blocked_live = 0
    for r in rows:
        pack, gen_id = r["pack"], int(r["gen_id"])
        batch_id, first = int(r["batch_id"]), int(r["first_seed_id"])
        cur.execute(f'select anchor_tx from "{pack}".generations where id=%s', (gen_id,))
        row = cur.fetchone()
        if row is None:
            print(f"WARNING: {pack} gen {gen_id} not found — skipping", file=sys.stderr)
            continue
        tx = row[0]
        if tx and tx.startswith("genesis:"):
            skipped_done += 1
            continue
        if tx and not args.reset_live:
            blocked_live += 1
            continue
        if not args.dry_run:
            cur.execute(
                f'update "{pack}".generations set seed_id=%s, anchor_tx=%s, anchor_height=0, '
                f"anchored_at=now(), leaf_count=coalesce(atoms_added,0)+coalesce(atoms_converged,0) "
                f"where id=%s",
                (first, f"genesis:batch:{batch_id}", gen_id),
            )
        updated += 1

    if args.dry_run:
        conn.rollback()
        print(f"DRY RUN: would update {updated}, already-genesis {skipped_done}, "
              f"live-anchored blocked {blocked_live} (use --reset-live post-wipe)")
    else:
        conn.commit()
        print(f"updated {updated}, already-genesis {skipped_done}, "
              f"live-anchored blocked {blocked_live}")
        if blocked_live:
            print("NOTE: blocked rows carry pre-wipe dreamtree-1 refs; re-run with "
                  "--reset-live to repoint them at genesis.", file=sys.stderr)
    return 0


if __name__ == "__main__":
    sys.exit(main())
