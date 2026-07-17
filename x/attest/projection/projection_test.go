package projection_test

import (
	"errors"
	"math"
	"testing"

	"github.com/blong-dev/dreamtree/x/attest"
	"github.com/blong-dev/dreamtree/x/attest/projection"
)

var errNotFound = errors.New("not found")

// memLog — the snapshot-backed LogReader the backtest harness uses (M1: same
// math, two data sources).
type memLog struct{ atts []attest.Attestation }

func (m memLog) GetAttestation(id uint64) (attest.Attestation, error) {
	for _, a := range m.atts {
		if a.Id == id {
			return a, nil
		}
	}
	return attest.Attestation{}, errNotFound
}

func (m memLog) WalkSubject(subject string, fn func(attest.Attestation) (bool, error)) error {
	for _, a := range m.atts {
		if a.Subject != subject {
			continue
		}
		stop, err := fn(a)
		if stop || err != nil {
			return err
		}
	}
	return nil
}

func (m memLog) WalkTarget(targetID uint64, fn func(attest.Attestation) (bool, error)) error {
	for _, a := range m.atts {
		if a.TargetId != targetID {
			continue
		}
		stop, err := fn(a)
		if stop || err != nil {
			return err
		}
	}
	return nil
}

type flatRep struct{ r float64 }

func (f flatRep) ReputationOf(string, string) float64 { return f.r }

func proj(atts []attest.Attestation, p attest.Params) projection.Projector {
	pf := projection.LoadParamsF(p)
	return projection.Projector{Params: pf, Log: memLog{atts}, Rep: flatRep{pf.BaselineKyc}}
}

// The devnet-proven numbers from the x/attest build (2026-07-08) reproduce over
// a snapshot log: origin S=1.0 + rigor S=0.8 → V = 1-(1-0.1)(1-0.08) = 0.172.
func TestKnownReadingsReproduce(t *testing.T) {
	atts := []attest.Attestation{
		{Id: 1, Subject: "work-1", Attestor: "a", ProofType: attest.ProofType_PROOF_TYPE_ORIGIN, IssuedAt: 0},
		{Id: 2, Subject: "work-1", Attestor: "b", ProofType: attest.ProofType_PROOF_TYPE_RIGOR, SpecificityBps: 8000, IssuedAt: 0},
	}
	p := proj(atts, attest.DefaultParams())
	v, count, err := p.WorkValue("work-1", 0)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}
	if math.Abs(v-0.172) > 1e-12 {
		t.Fatalf("V = %v, want 0.172", v)
	}
}

// The dial (M3): the same log under override params yields a different reading —
// reprojection is rerunning the reading with different levers.
func TestParamsOverrideChangesReading(t *testing.T) {
	atts := []attest.Attestation{
		{Id: 1, Subject: "work-1", Attestor: "a", ProofType: attest.ProofType_PROOF_TYPE_ORIGIN, IssuedAt: 0},
	}
	base := attest.DefaultParams()
	v1, _, err := proj(atts, base).WorkValue("work-1", 0)
	if err != nil {
		t.Fatal(err)
	}

	override := attest.DefaultParams()
	override.SMax = "5.0" // tighter ceiling → the same attestation is worth a larger share
	v2, _, err := proj(atts, override).WorkValue("work-1", 0)
	if err != nil {
		t.Fatal(err)
	}
	if !(v2 > v1) {
		t.Fatalf("override should raise V: base=%v override=%v", v1, v2)
	}
	if math.Abs(v1-0.1) > 1e-12 || math.Abs(v2-0.2) > 1e-12 {
		t.Fatalf("v1=%v (want 0.1), v2=%v (want 0.2)", v1, v2)
	}
}

// M2: citation_uplift_lambda is a lever now — uplift 0 kills creation-credit-
// forward, the default reproduces the const-era behavior, and empty params
// (pre-upgrade state) fall back to the const-era 1.0.
func TestCitationUpliftLever(t *testing.T) {
	atts := []attest.Attestation{
		// work-B has its own origin (V_base(B) = 0.1)
		{Id: 1, Subject: "work-B", Attestor: "b", ProofType: attest.ProofType_PROOF_TYPE_ORIGIN, IssuedAt: 0},
		// work-A is cited by work-B (USE with used_by)
		{Id: 2, Subject: "work-A", Attestor: "c", ProofType: attest.ProofType_PROOF_TYPE_USE, UsedBy: "work-B", IssuedAt: 0},
	}
	// Default (1.0): share = 0.05 × (1 + 1.0×0.1) = 0.055
	v, _, err := proj(atts, attest.DefaultParams()).WorkValue("work-A", 0)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(v-0.055) > 1e-12 {
		t.Fatalf("default uplift: V = %v, want 0.055", v)
	}
	// Lever to 0: share = 0.05, no uplift.
	off := attest.DefaultParams()
	off.CitationUpliftLambda = "0.0"
	v0, _, err := proj(atts, off).WorkValue("work-A", 0)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(v0-0.05) > 1e-12 {
		t.Fatalf("uplift off: V = %v, want 0.05", v0)
	}
	// Pre-upgrade params (empty field) fall back to the const-era 1.0.
	legacy := attest.DefaultParams()
	legacy.CitationUpliftLambda = ""
	vl, _, err := proj(atts, legacy).WorkValue("work-A", 0)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(vl-0.055) > 1e-12 {
		t.Fatalf("legacy fallback: V = %v, want 0.055", vl)
	}
}
