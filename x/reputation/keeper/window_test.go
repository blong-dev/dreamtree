package keeper

import (
	"testing"

	"cosmossdk.io/math"

	"github.com/blong-dev/dreamtree/x/reputation"
)

// dec is a test helper mirroring window.go's d().
func ld(s string) math.LegacyDec { return math.LegacyMustNewDecFromStr(s) }

func params() reputation.Params {
	return reputation.Params{
		NegAsymmetry:  "2.0",
		OutcomeCapMult: "5.0",
	}
}

// netVerdict is a pure function of (Params, PendingEvent) — no store touched —
// so a zero-value Keeper suffices to exercise the refutation-window math.
func TestNetVerdict(t *testing.T) {
	var k Keeper
	p := params()

	cases := []struct {
		name    string
		pe      reputation.PendingEvent
		wantStr string // expected signed M_O_net
	}{
		{
			// Uncontested validation: net = base (positive).
			name:    "validation_uncontested",
			pe:      reputation.PendingEvent{OutcomeRefutes: false, BaseMagnitude: ld("3"), Corroboration: ld("0"), Refutation: ld("0"), TargetSIssuance: ld("2")},
			wantStr: "3.000000000000000000",
		},
		{
			// Uncontested refutation (fraud claim): net = −base (negative).
			name:    "refutation_uncontested",
			pe:      reputation.PendingEvent{OutcomeRefutes: true, BaseMagnitude: ld("3"), Corroboration: ld("0"), Refutation: ld("0"), TargetSIssuance: ld("2")},
			wantStr: "-3.000000000000000000",
		},
		{
			// A fraud claim (base=3) fully defended by an equal defense (3) nets to
			// 0 — false accusation neutralized 1:1, NOT 2:1. This is the F8 fix:
			// the 2× is not inside the window integration.
			name:    "refutation_defended_1to1_neutral",
			pe:      reputation.PendingEvent{OutcomeRefutes: true, BaseMagnitude: ld("3"), Corroboration: ld("0"), Refutation: ld("3"), TargetSIssuance: ld("2")},
			wantStr: "0.000000000000000000",
		},
		{
			// Defense exceeds the fraud claim → net flips positive (validated).
			name:    "refutation_overdefended_flips_positive",
			pe:      reputation.PendingEvent{OutcomeRefutes: true, BaseMagnitude: ld("3"), Corroboration: ld("0"), Refutation: ld("5"), TargetSIssuance: ld("2")},
			wantStr: "2.000000000000000000",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := k.netVerdict(p, tc.pe)
			if got.String() != tc.wantStr {
				t.Fatalf("netVerdict = %s, want %s", got, tc.wantStr)
			}
		})
	}
}

// A crowd pile-on cannot breach the paper-shape ceiling M_cap = cap_mult ×
// S_issuance. 100 million corroborations ≈ M_cap, not 100 million — the bound
// that (with the 0-floor) makes character assassination impossible.
func TestNetVerdictCrowdBounded(t *testing.T) {
	var k Keeper
	p := params()
	// base fraud claim + an enormous corroboration accumulator.
	pe := reputation.PendingEvent{
		OutcomeRefutes:  true,
		BaseMagnitude:   ld("1"),
		Corroboration:   ld("100000000"),
		Refutation:      ld("0"),
		TargetSIssuance: ld("2"), // M_cap = 5 × 2 = 10
	}
	net := k.netVerdict(p, pe)
	capM := ld("10")
	// rPool saturates AT M_cap (inclusive); net = −rPool, so |net| ≤ M_cap. The
	// crowd is bounded to 10, not 100 million — that is the whole point.
	if net.Abs().GT(capM) {
		t.Fatalf("crowd unbounded: |net|=%s must be ≤ M_cap=%s", net.Abs(), capM)
	}
	if !net.IsNegative() {
		t.Fatalf("fraud verdict must be negative, got %s", net)
	}
}

// paperShapeAdd must be monotonic, capped, and Sybil-damped (two halves add to
// less than the cap, never past it).
func TestPaperShapeAddCapped(t *testing.T) {
	cap := ld("10")
	// Adding a huge x saturates toward, but never reaches, the cap.
	got := paperShapeAdd(ld("0"), ld("1000"), cap)
	if got.GT(cap) {
		t.Fatalf("paperShapeAdd breached cap: %s > %s", got, cap)
	}
	// Diminishing returns: agg=9 (near cap), adding 9 more stays under cap.
	got2 := paperShapeAdd(ld("9"), ld("9"), cap)
	if got2.GT(cap) {
		t.Fatalf("paperShapeAdd breached cap on second add: %s > %s", got2, cap)
	}
	if got2.LTE(ld("9")) {
		t.Fatalf("paperShapeAdd not monotonic: %s <= 9", got2)
	}
}
