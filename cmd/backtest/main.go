// backtest — the outcome backtest (measurement-backtest.md Part C, M4).
//
// Off-chain, over an exported log: walk-forward holdout, Spearman rank score,
// calibration buckets, one-lever-at-a-time sensitivity curves. Runs the SAME
// projection code the chain serves (x/attest/projection) over a snapshot
// LogReader — same math, two data sources.
//
// Input: `dreamtreed export` output (reads app_state.attest.attestations), or
// a bare JSON array of attestations.
//
//	go run ./cmd/backtest -log exported-genesis.json [-cutoffs 5] [-min-outcomes 10]
//
// Honesty rules built in (spec §Limitations): reports n and refuses to crown a
// "least wrong" set below -min-outcomes; scores independent-only outcomes
// separately; this measures INTERNAL predictive validity only — never the
// paper's historical-lambda claim.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"sort"

	"github.com/blong-dev/dreamtree/x/attest"
	"github.com/blong-dev/dreamtree/x/attest/projection"
)

// ---- flexible import of exported JSON --------------------------------------

type jsonAtt struct {
	Id             json.Number `json:"id"`
	Attestor       string      `json:"attestor"`
	Subject        string      `json:"subject"`
	ProofType      string      `json:"proof_type"`
	Domain         string      `json:"domain"`
	SpecificityBps json.Number `json:"specificity_bps"`
	OutcomeKind    string      `json:"outcome_kind"`
	TargetId       json.Number `json:"target_id"`
	IssuedAt       json.Number `json:"issued_at"`
	UsedBy         string      `json:"used_by"`
}

func (j jsonAtt) toAttestation() attest.Attestation {
	num := func(n json.Number) uint64 { v, _ := n.Int64(); return uint64(v) }
	i64 := func(n json.Number) int64 { v, _ := n.Int64(); return v }
	return attest.Attestation{
		Id:             num(j.Id),
		Attestor:       j.Attestor,
		Subject:        j.Subject,
		ProofType:      attest.ProofType(attest.ProofType_value[j.ProofType]),
		Domain:         j.Domain,
		SpecificityBps: uint32(num(j.SpecificityBps)),
		OutcomeKind:    attest.OutcomeKind(attest.OutcomeKind_value[j.OutcomeKind]),
		TargetId:       num(j.TargetId),
		IssuedAt:       i64(j.IssuedAt),
		UsedBy:         j.UsedBy,
	}
}

func loadLog(path string) ([]attest.Attestation, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var arr []jsonAtt
	if err := json.Unmarshal(raw, &arr); err == nil {
		return convert(arr), nil
	}
	var genesis struct {
		AppState struct {
			Attest struct {
				Attestations []jsonAtt `json:"attestations"`
			} `json:"attest"`
		} `json:"app_state"`
	}
	if err := json.Unmarshal(raw, &genesis); err != nil {
		return nil, fmt.Errorf("neither an attestation array nor an exported genesis: %w", err)
	}
	return convert(genesis.AppState.Attest.Attestations), nil
}

func convert(arr []jsonAtt) []attest.Attestation {
	out := make([]attest.Attestation, 0, len(arr))
	for _, j := range arr {
		out = append(out, j.toAttestation())
	}
	return out
}

// ---- snapshot LogReader restricted to issued_at <= cutoff ------------------

type snapshot struct {
	atts   []attest.Attestation
	cutoff int64
}

func (s snapshot) in(a attest.Attestation) bool { return a.IssuedAt <= s.cutoff }

func (s snapshot) GetAttestation(id uint64) (attest.Attestation, error) {
	for _, a := range s.atts {
		if a.Id == id && s.in(a) {
			return a, nil
		}
	}
	return attest.Attestation{}, fmt.Errorf("attestation %d not in snapshot", id)
}

func (s snapshot) WalkSubject(subject string, fn func(attest.Attestation) (bool, error)) error {
	for _, a := range s.atts {
		if a.Subject != subject || !s.in(a) {
			continue
		}
		if stop, err := fn(a); stop || err != nil {
			return err
		}
	}
	return nil
}

func (s snapshot) WalkTarget(targetID uint64, fn func(attest.Attestation) (bool, error)) error {
	for _, a := range s.atts {
		if a.TargetId != targetID || !s.in(a) {
			continue
		}
		if stop, err := fn(a); stop || err != nil {
			return err
		}
	}
	return nil
}

type flatRep struct{ r float64 }

func (f flatRep) ReputationOf(string, string) float64 { return f.r }

// ---- ground truth ----------------------------------------------------------

func outcomeSign(k attest.OutcomeKind) float64 {
	switch k {
	case attest.OutcomeKind_OUTCOME_KIND_VALIDATED:
		return 1
	case attest.OutcomeKind_OUTCOME_KIND_REFUTED:
		return -1
	case attest.OutcomeKind_OUTCOME_KIND_PARTIAL:
		return 0.5
	}
	return 0
}

// groundTruth returns y(subject) = signed sum of post-cutoff OUTCOME
// attestations targeting the subject's pre-cutoff attestations, plus the count
// of outcomes and the count from INDEPENDENT attestors (outcome attestor not
// among the subject's own work attestors — the circularity mitigation).
type truth struct {
	y      float64
	yIndep float64
	n      int
	nIndep int
}

func groundTruths(atts []attest.Attestation, cutoff int64) map[string]truth {
	subjectOf := map[uint64]string{} // attestation id -> subject
	workAttestors := map[string]map[string]bool{}
	for _, a := range atts {
		if a.IssuedAt > cutoff || a.ProofType == attest.ProofType_PROOF_TYPE_OUTCOME {
			continue
		}
		subjectOf[a.Id] = a.Subject
		if workAttestors[a.Subject] == nil {
			workAttestors[a.Subject] = map[string]bool{}
		}
		workAttestors[a.Subject][a.Attestor] = true
	}
	out := map[string]truth{}
	for _, o := range atts {
		if o.ProofType != attest.ProofType_PROOF_TYPE_OUTCOME || o.IssuedAt <= cutoff {
			continue
		}
		subj, ok := subjectOf[o.TargetId]
		if !ok {
			continue // outcome on a post-cutoff work: not scoreable at this cutoff
		}
		t := out[subj]
		s := outcomeSign(o.OutcomeKind)
		t.y += s
		t.n++
		if !workAttestors[subj][o.Attestor] {
			t.yIndep += s
			t.nIndep++
		}
		out[subj] = t
	}
	return out
}

// ---- scoring ---------------------------------------------------------------

func ranks(v []float64) []float64 {
	idx := make([]int, len(v))
	for i := range idx {
		idx[i] = i
	}
	sort.Slice(idx, func(a, b int) bool { return v[idx[a]] < v[idx[b]] })
	r := make([]float64, len(v))
	for i := 0; i < len(idx); {
		j := i
		for j < len(idx) && v[idx[j]] == v[idx[i]] {
			j++
		}
		avg := float64(i+j-1)/2 + 1
		for k := i; k < j; k++ {
			r[idx[k]] = avg
		}
		i = j
	}
	return r
}

func spearman(x, y []float64) float64 {
	if len(x) < 3 {
		return math.NaN()
	}
	rx, ry := ranks(x), ranks(y)
	var mx, my float64
	for i := range rx {
		mx += rx[i]
		my += ry[i]
	}
	mx /= float64(len(rx))
	my /= float64(len(ry))
	var num, dx, dy float64
	for i := range rx {
		num += (rx[i] - mx) * (ry[i] - my)
		dx += (rx[i] - mx) * (rx[i] - mx)
		dy += (ry[i] - my) * (ry[i] - my)
	}
	if dx == 0 || dy == 0 {
		return math.NaN()
	}
	return num / math.Sqrt(dx*dy)
}

type cutoffScore struct {
	cutoff     int64
	n          int
	nOutcomes  int
	rho        float64
	rhoIndep   float64
	nIndepOutc int
}

// scoreParams runs the walk-forward for one parameter set.
func scoreParams(atts []attest.Attestation, cutoffs []int64, p attest.Params) []cutoffScore {
	pf := projection.LoadParamsF(p)
	var out []cutoffScore
	for _, c := range cutoffs {
		proj := projection.Projector{Params: pf, Log: snapshot{atts, c}, Rep: flatRep{pf.BaselineKyc}}
		truths := groundTruths(atts, c)
		var vs, ys, ysI []float64
		nOut, nIndep := 0, 0
		var subjects []string
		for s := range truths {
			subjects = append(subjects, s)
		}
		sort.Strings(subjects)
		for _, s := range subjects {
			t := truths[s]
			v, _, err := proj.WorkValue(s, c)
			if err != nil {
				continue
			}
			vs = append(vs, v)
			ys = append(ys, t.y)
			ysI = append(ysI, t.yIndep)
			nOut += t.n
			nIndep += t.nIndep
		}
		out = append(out, cutoffScore{
			cutoff: c, n: len(vs), nOutcomes: nOut, nIndepOutc: nIndep,
			rho: spearman(vs, ys), rhoIndep: spearman(vs, ysI),
		})
	}
	return out
}

func aggregate(scores []cutoffScore) (mean, worst float64, totalOutcomes int) {
	worst = math.Inf(1)
	n := 0
	for _, s := range scores {
		totalOutcomes += s.nOutcomes
		if math.IsNaN(s.rho) {
			continue
		}
		mean += s.rho
		if s.rho < worst {
			worst = s.rho
		}
		n++
	}
	if n == 0 {
		return math.NaN(), math.NaN(), totalOutcomes
	}
	return mean / float64(n), worst, totalOutcomes
}

// ---- sensitivity sweep -----------------------------------------------------

type leverPoint struct {
	Value float64 `json:"value"`
	Rho   float64 `json:"mean_rho"`
	Worst float64 `json:"worst_rho"`
}

func sweepLever(atts []attest.Attestation, cutoffs []int64, name string, values []float64,
	apply func(*attest.Params, float64)) []leverPoint {
	var curve []leverPoint
	for _, v := range values {
		p := attest.DefaultParams()
		apply(&p, v)
		mean, worst, _ := aggregate(scoreParams(atts, cutoffs, p))
		curve = append(curve, leverPoint{Value: v, Rho: mean, Worst: worst})
	}
	return curve
}

// ---- main ------------------------------------------------------------------

func main() {
	logPath := flag.String("log", "", "exported genesis JSON (or bare attestation array)")
	nCutoffs := flag.Int("cutoffs", 5, "number of walk-forward cutoffs")
	minOutcomes := flag.Int("min-outcomes", 10, "minimum total post-cutoff outcomes to crown a least-wrong set")
	jsonOut := flag.Bool("json", false, "emit the full report as JSON")
	flag.Parse()
	if *logPath == "" {
		fmt.Fprintln(os.Stderr, "usage: backtest -log <exported-genesis.json>")
		os.Exit(2)
	}
	atts, err := loadLog(*logPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load:", err)
		os.Exit(1)
	}
	nOutcomes := 0
	var minT, maxT int64 = math.MaxInt64, math.MinInt64
	for _, a := range atts {
		if a.ProofType == attest.ProofType_PROOF_TYPE_OUTCOME {
			nOutcomes++
		}
		if a.IssuedAt < minT {
			minT = a.IssuedAt
		}
		if a.IssuedAt > maxT {
			maxT = a.IssuedAt
		}
	}
	fmt.Printf("log: %d attestations (%d outcomes), issued_at span [%d, %d]\n", len(atts), nOutcomes, minT, maxT)
	if nOutcomes == 0 {
		fmt.Println("\nNO OUTCOMES IN LOG — nothing to score. The backtest is predictive")
		fmt.Println("(pre-cutoff values vs post-cutoff outcomes); it runs for real once")
		fmt.Println("outcome attestations exist. Harness verified, report honestly empty.")
		return
	}

	// Cutoffs: evenly spaced interior points of the issued_at span.
	var cutoffs []int64
	for i := 1; i <= *nCutoffs; i++ {
		cutoffs = append(cutoffs, minT+(maxT-minT)*int64(i)/int64(*nCutoffs+1))
	}

	base := scoreParams(atts, cutoffs, attest.DefaultParams())
	meanR, worstR, totalOut := aggregate(base)
	fmt.Printf("\nbaseline params: mean rho=%.4f worst rho=%.4f (scored outcomes=%d)\n", meanR, worstR, totalOut)
	for _, s := range base {
		fmt.Printf("  cutoff %d: subjects=%d outcomes=%d rho=%.4f rho_indep=%.4f (indep outcomes=%d)\n",
			s.cutoff, s.n, s.nOutcomes, s.rho, s.rhoIndep, s.nIndepOutc)
	}
	if totalOut < *minOutcomes {
		fmt.Printf("\nINSUFFICIENT OUTCOMES (%d < %d): reporting scores but refusing to crown\n", totalOut, *minOutcomes)
		fmt.Println("a least-wrong parameter set — too few outcomes to rank without overfitting noise.")
	}

	// One-lever-at-a-time sensitivity curves (the deliverable).
	curves := map[string][]leverPoint{
		"citation_uplift_lambda": sweepLever(atts, cutoffs, "citation_uplift_lambda",
			[]float64{0, 0.25, 0.5, 1.0, 2.0, 4.0},
			func(p *attest.Params, v float64) { p.CitationUpliftLambda = fmt.Sprintf("%g", v) }),
		"s_max": sweepLever(atts, cutoffs, "s_max",
			[]float64{2, 5, 10, 20, 50},
			func(p *attest.Params, v float64) { p.SMax = fmt.Sprintf("%g", v) }),
		"weight_use": sweepLever(atts, cutoffs, "weight_use",
			[]float64{1000, 2500, 5000, 7500, 10000},
			func(p *attest.Params, v float64) { p.WeightUse = uint32(v) }),
		"lambda_use": sweepLever(atts, cutoffs, "lambda_use",
			[]float64{0, 0.04, 0.08, 0.16, 0.32},
			func(p *attest.Params, v float64) { p.LambdaUse = fmt.Sprintf("%g", v) }),
		"obsolescence_default": sweepLever(atts, cutoffs, "obsolescence_default",
			[]float64{0.3, 1.0, 3.0},
			func(p *attest.Params, v float64) { p.ObsolescenceDefault = fmt.Sprintf("%g", v) }),
	}
	fmt.Println("\nsensitivity curves (mean rho by lever value; flat = insensitive, peak = evidence):")
	var names []string
	for n := range curves {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		fmt.Printf("  %s:\n", n)
		for _, pt := range curves[n] {
			fmt.Printf("    %-8g mean_rho=%.4f worst=%.4f\n", pt.Value, pt.Rho, pt.Worst)
		}
	}

	if *jsonOut {
		report := map[string]any{
			"attestations": len(atts), "outcomes": nOutcomes,
			"baseline":   map[string]any{"mean_rho": meanR, "worst_rho": worstR, "scored_outcomes": totalOut},
			"sufficient": totalOut >= *minOutcomes,
			"curves":     curves,
		}
		b, _ := json.MarshalIndent(report, "", "  ")
		fmt.Println(string(b))
	}
}
