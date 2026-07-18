package photons

// Per-block photon mint ceiling (trust-layer W3, docs/specs/trust-layer.md).
//
// Photons are inflationary by design — one per distinct atom — so a global
// supply cap would contradict the peg. The blast-radius risk is instead a
// leaked ANCHORD_TOKEN minting against fabricated roots. A per-block ceiling
// bounds how much a single block can mint, so a flood is capped per block
// (and each ceiling hit is an on-chain event) instead of unbounded.
//
// Sizing (empirical, 2026-07-18): the largest observed batch new_count is 732
// and anchord serializes ~one batch per ~6s block, so honest per-block mint is
// well under 1,000. MaxMintPerBlock is set ~137x the largest observed batch —
// unreachable by honest ingestion (it would take 100k+ atoms committed in one
// block) while capping a flood. MintCeilingSoftWarn emits a warning event well
// below the hard ceiling so honest growth toward the limit is visible long
// before it would ever reject a legitimate commit.
//
// These are code constants for the v0 single-validator chain; promoting
// MaxMintPerBlock to a governance param (so it can be raised without a binary
// redeploy) is the documented follow-up.
const (
	MaxMintPerBlock     uint64 = 100_000
	MintCeilingSoftWarn uint64 = 25_000
)
