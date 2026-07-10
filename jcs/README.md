# jcs — the quorum's one canonicalization

**RFC 8785 (JCS) golden vectors + conformant reference.** Everything in the
dreamtree quorum that signs, hashes, or anchors JSON — roots credentials and
exports (`eddsa-jcs-2022`), reflow atoms (`atom_id = SHA-256(JCS(assertion))`),
chain commitments re-verified by tooling — MUST canonicalize to byte-identical
output. Divergent canonicalization is a silent hash-fork: "verification failed"
on honest data.

**`vectors.json` is the normative artifact.** An implementation is conformant
iff it byte-matches (UTF-8) every vector's `expected`. Vectors are append-only;
never edit an existing one.

## Files

- `vectors.json` — 30 normative vectors, hand-authored against RFC 8785
  §3.2.2/§3.2.3 and ECMA-262 §7.1.12.1 (incl. Note 2). Includes negative
  discriminators for the two classic failure modes.
- `reference.py` — conformant Python implementation (passes 30/30). Seed for
  reflow-Py; port for reflow-Go.
- `run_vectors.py` — conformance runner (exit 0 = conformant).
- `build_vectors.py` — regenerates the container file's escaping only; the
  `expected` values are hand-authored, never computed.

## The two places implementations die (both covered by discriminating vectors)

1. **Property sorting is UTF-16 code units** (RFC 8785 §3.2.3), *not* Unicode
   code points. JS default `.sort()` is CORRECT here (JS strings are UTF-16);
   naive Python `sorted()` and explicit code-point sorts are WRONG for
   astral-plane keys. Vector: `keys-utf16-vs-codepoint-DISCRIMINATOR`.
2. **Numbers are ECMA-262 `Number::toString`**: `-0` → `0`, `100.0` → `100`,
   exponent from `1e21` / below `1e-6`, `1e-7` **not** `1e-07`. Python's
   `repr`/`json.dumps` layout is nonconformant and must be re-laid-out (see
   `reference.py:_number`). Vectors: `num-*`.

## Status of quorum implementations (2026-07-10)

| impl | sort | numbers | verdict |
|---|---|---|---|
| roots primary (`src/credentials/canonical.ts`) | ✅ UTF-16 (default `.sort()`) | ❌ none (relies on caller discipline) | nonconformant — adopt full number layout |
| roots v3 copy (`di.ts jcsCanonicalise`) | ❌ explicit code-point sort | partial | nonconformant — sort order backwards for astral keys |
| `reference.py` (here) | ✅ | ✅ | **30/30** |
| reflow-Go | — | — | to be written against these vectors |

## Migration rules

- **roots**: adopt a conformant canonicalizer, then **re-verify every stored
  credential under the new implementation before cutover** — verification
  re-canonicalizes, so any credential whose bytes differ under the new impl
  (astral keys, exotic numbers; expected ≈ none) must be identified, not
  discovered later. Wire `vectors.json` into vitest.
- **reflow**: atoms additionally obey the D8 number rule (integers within
  ±2^53 or lexical decimal *strings*; floats prohibited in canonical atoms) —
  a stricter profile ON TOP of JCS, not a deviation from it. Credentials and
  other wallet documents still need full float conformance, which is why the
  vectors cover it.
- **CI**: every implementation runs its vector suite on every build. A vector
  failure is a build failure, nowhere lower.
