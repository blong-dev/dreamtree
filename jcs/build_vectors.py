#!/usr/bin/env python3
"""Regenerate vectors.json with guaranteed-valid escaping.

The `expected` strings are HAND-AUTHORED against RFC 8785 + ECMA-262
§7.1.12.1 — they are the normative artifact. This script only guarantees the
container file is valid JSON; it never computes expected values.
"""
import json

V = []
def v(name, inp, expected, why=None):
    d = {"name": name, "input": inp, "expected": expected}
    if why: d["_why"] = why
    V.append(d)

# ── key sorting ──────────────────────────────────────────────────────────
v("keys-basic-sort", {"b": 1, "a": 2}, '{"a":2,"b":1}')
v("keys-nested-sort", {"z": {"d": 4, "c": 3}, "y": [{"b": 2, "a": 1}]},
  '{"y":[{"a":1,"b":2}],"z":{"c":3,"d":4}}')
v("keys-utf16-vs-codepoint-DISCRIMINATOR", {"￿": 2, "\U0001F600": 1},
  '{"\U0001F600":1,"￿":2}',
  "U+1F600 is UTF-16 D83D DE00; 0xD83D < 0xFFFF so the emoji key sorts FIRST "
  "under RFC 8785 (UTF-16 code units). Code-point sorters put it after — and fail.")
v("keys-utf16-discriminator-with-prefix",
  {"a￿": 2, "a\U0001F600": 1, "a": 0},
  '{"a":0,"a\U0001F600":1,"a￿":2}')
v("keys-rfc-style-ordering",
  {"€": "Euro Sign", "\r": "Carriage Return", "1": "One",
   "": "Control", "ö": "Latin Small Letter O With Diaeresis"},
  '{"\\r":"Carriage Return","1":"One","":"Control",'
  '"ö":"Latin Small Letter O With Diaeresis","€":"Euro Sign"}',
  "UTF-16 unit order: 0x000D < 0x0031 < 0x0080 < 0x00F6 < 0x20AC")
v("keys-empty-string-key", {"": 0, "a": 1}, '{"":0,"a":1}')

# ── numbers (ECMA-262 7.1.12.1 incl. Note 2) ─────────────────────────────
v("num-zero", [0], "[0]")
v("num-minus-zero", [-0.0], "[0]", "-0 serializes as 0")
v("num-whole-float", [100.0], "[100]")
v("num-one", [1.0], "[1]")
v("num-tenth", [0.1], "[0.1]")
v("num-negative", [-1.5], "[-1.5]")
v("num-max-safe-int", [9007199254740991], "[9007199254740991]")
v("num-2pow53", [9007199254740992], "[9007199254740992]")
v("num-exp-threshold-below", [1e20], "[100000000000000000000]", "1e20 stays positional")
v("num-exp-threshold-at", [1e21], "[1e+21]", "≥1e21 → exponent, lowercase e, explicit +")
v("num-small-positional", [0.000001], "[0.000001]", "RFC sample ieee 3eb0c6f7a0b5ed8d")
v("num-small-exponent", [1e-7], "[1e-7]",
  "crosses ECMA n≤-7 → exponent, NO zero-padded exponent (Python repr says 1e-07: wrong)")
v("num-min-subnormal", [5e-324], "[5e-324]")
v("num-rfc-sample-thirds", [333333333.3333333], "[333333333.3333333]",
  "RFC sample ieee 41b3de4355555555")

# ── strings ──────────────────────────────────────────────────────────────
v("str-short-escapes", ["\b\t\n\f\r"], '["\\b\\t\\n\\u000b\\f\\r"]'.replace("\\n\\u000b", "\\n\\u000b"),
  "wait-see-fix")
# fix: input must include \x0b to exercise the no-short-escape control
V[-1]["input"] = ["\b\t\n\x0b\f\r"]
V[-1]["expected"] = '["\\b\\t\\n\\u000b\\f\\r"]'
V[-1]["_why"] = "\\b \\t \\n \\f \\r short escapes; U+000B has none → lowercase \\u000b"
v("str-nul-and-c0", ["\x00\x1f"], '["\\u0000\\u001f"]')
v("str-quote-backslash-slash", ["\"\\/"], '["\\"\\\\/"]', "forward slash is NOT escaped")
v("str-non-ascii-literal", ["é \U0001F600"], '["é \U0001F600"]',
  "é and 😀 literal UTF-8, never \\u-escaped")
v("str-del-not-escaped", ["\x7f"], '["\x7f"]', "only C0 controls escape; U+007F literal")

# ── literals & structure ─────────────────────────────────────────────────
v("literals", [True, False, None], "[true,false,null]")
v("empty-containers", {"a": {}, "b": []}, '{"a":{},"b":[]}')
v("array-order-preserved", [3, 1, 2, "b", "a"], '[3,1,2,"b","a"]')
v("no-whitespace", {"a": [1, 2], "b": {"c": 3}}, '{"a":[1,2],"b":{"c":3}}')
v("escaped-input-key-equals-literal", {"é": 1}, '{"é":1}',
  "canonicalization operates on parsed values: \\u00e9 in input == literal é")

doc = {
    "_comment": ("Normative JCS (RFC 8785) golden vectors for the dreamtree quorum. "
                 "Every canonicalizer (roots-TS, reflow-Go, reflow-Py, chain tooling) "
                 "MUST byte-match `expected` (compared as UTF-8 bytes) for each `input`. "
                 "Hand-authored against RFC 8785 §3.2.2/§3.2.3 and ECMA-262 §7.1.12.1 "
                 "incl. Note 2. Never edit an existing vector; append new ones."),
    "version": 1,
    "vectors": V,
}
with open("vectors.json", "w", encoding="utf-8") as f:
    json.dump(doc, f, ensure_ascii=False, indent=1)
    f.write("\n")
print(f"wrote {len(V)} vectors")
json.load(open("vectors.json"))  # self-check
print("re-parse OK")
