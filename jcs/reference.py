"""JCS (RFC 8785) — conformant Python reference implementation.

The normative artifact is vectors.json; this implementation exists to prove
the vectors are satisfiable and to serve as reflow-Py's canonicalizer seed.

The two places naive implementations die, handled explicitly:

1. **Property sorting is UTF-16 code units** (RFC 8785 §3.2.3), NOT Unicode
   code points. Python's native str ordering is code points and diverges on
   astral-plane keys — so keys are sorted by their UTF-16-BE encoding, whose
   byte order equals code-unit order.
2. **Numbers follow ECMA-262 §7.1.12.1** (Note 2 included). Python's repr()
   provides the shortest-round-trip digits but formats differently
   (1e-07 vs 1e-7, 100.0 vs 100, exponent thresholds), so the ECMA layout
   rules are applied over repr's digits.
"""
from __future__ import annotations

import math

_ESCAPES = {
    "\b": "\\b", "\t": "\\t", "\n": "\\n", "\f": "\\f", "\r": "\\r",
    '"': '\\"', "\\": "\\\\",
}


def _string(s: str) -> str:
    out = ['"']
    for ch in s:
        if ch in _ESCAPES:
            out.append(_ESCAPES[ch])
        elif ch < "\x20":
            out.append(f"\\u{ord(ch):04x}")
        else:
            out.append(ch)
    out.append('"')
    return "".join(out)


def _number(x: float) -> str:
    """ECMA-262 §7.1.12.1 Number::toString(x, 10)."""
    if math.isnan(x) or math.isinf(x):
        raise ValueError("NaN/Infinity are not permitted in JCS (RFC 8785 §3.2.2.3)")
    if x == 0:
        return "0"  # covers -0.0
    if x < 0:
        return "-" + _number(-x)

    # Shortest round-trip digits via repr, then re-layout per ECMA.
    r = repr(x)
    if "e" in r:
        mant, _, e = r.partition("e")
        exp = int(e)
    else:
        mant, exp = r, 0
    int_part, _, frac = mant.partition(".")

    if int_part != "0":
        n = len(int_part) + exp
        digits = (int_part + frac).rstrip("0") or "0"
    else:
        stripped = frac.lstrip("0")
        n = exp - (len(frac) - len(stripped))
        digits = stripped.rstrip("0")
    k = len(digits)

    if k <= n <= 21:
        return digits + "0" * (n - k)
    if 0 < n <= 21:
        return digits[:n] + "." + digits[n:]
    if -6 < n <= 0:
        return "0." + "0" * (-n) + digits
    # exponent form
    e10 = n - 1
    head = digits[0] + ("." + digits[1:] if k > 1 else "")
    return f"{head}e{'+' if e10 >= 0 else '-'}{abs(e10)}"


def canonicalize(value) -> str:
    """Return the RFC 8785 canonical JSON text for a parsed JSON value."""
    if value is None:
        return "null"
    if value is True:
        return "true"
    if value is False:
        return "false"
    if isinstance(value, str):
        return _string(value)
    if isinstance(value, int):
        # ±2^53 itself is exactly representable; beyond it integers collide.
        if abs(value) > 2 ** 53:
            raise ValueError(f"integer outside IEEE-754 exact range: {value}")
        return str(value)
    if isinstance(value, float):
        return _number(value)
    if isinstance(value, (list, tuple)):
        return "[" + ",".join(canonicalize(v) for v in value) + "]"
    if isinstance(value, dict):
        # RFC 8785 §3.2.3: sort property names by UTF-16 code units.
        items = sorted(value.items(), key=lambda kv: kv[0].encode("utf-16-be"))
        return "{" + ",".join(f"{_string(k)}:{canonicalize(v)}" for k, v in items) + "}"
    raise TypeError(f"not a JSON value: {type(value).__name__}")


def canonical_bytes(value) -> bytes:
    return canonicalize(value).encode("utf-8")
