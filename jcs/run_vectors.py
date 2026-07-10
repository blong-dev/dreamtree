#!/usr/bin/env python3
"""Conformance runner: any canonicalizer must byte-match every vector.

Usage: python3 run_vectors.py          # tests reference.py
Exit 0 = fully conformant. Nonzero = list of failures with byte-level diff.
"""
import json
import sys

from reference import canonical_bytes


def main() -> int:
    doc = json.load(open("vectors.json", encoding="utf-8"))
    failures = []
    for vec in doc["vectors"]:
        expected = vec["expected"].encode("utf-8")
        try:
            got = canonical_bytes(vec["input"])
        except Exception as e:  # noqa: BLE001
            failures.append((vec["name"], f"raised {type(e).__name__}: {e}"))
            continue
        if got != expected:
            failures.append((vec["name"], f"\n  expected: {expected!r}\n  got:      {got!r}"))
    total = len(doc["vectors"])
    if failures:
        print(f"FAIL {len(failures)}/{total}:")
        for name, detail in failures:
            print(f"- {name}: {detail}")
        return 1
    print(f"PASS {total}/{total} vectors")
    return 0


if __name__ == "__main__":
    sys.exit(main())
