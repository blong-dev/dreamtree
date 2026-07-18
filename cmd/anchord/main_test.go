package main

import (
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeRunner is an injectable txRunner. It routes on the subcommand: a "tx …"
// invocation is a broadcast, a "query tx …" invocation is an inclusion poll.
// It counts each so tests can assert what did (and did not) happen.
type fakeRunner struct {
	mu          sync.Mutex
	broadcasts  int
	queries     int
	lastTxArgs  []string
	broadcastFn func() (string, error)
	queryFn     func() (string, error)
}

func (f *fakeRunner) run(args ...string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	switch {
	case len(args) >= 2 && args[0] == "query" && args[1] == "tx":
		f.queries++
		if f.queryFn != nil {
			return f.queryFn()
		}
		return "", nil
	case len(args) >= 1 && args[0] == "tx":
		f.broadcasts++
		f.lastTxArgs = append([]string(nil), args...)
		if f.broadcastFn != nil {
			return f.broadcastFn()
		}
		return `{"txhash":"DEADBEEF","code":0}`, nil
	default:
		return "", nil
	}
}

func newTestServer(t *testing.T, f *fakeRunner) *server {
	t.Helper()
	store, err := newIdemStore("") // in-memory
	if err != nil {
		t.Fatalf("newIdemStore: %v", err)
	}
	return &server{
		cfg:          config{bin: "dreamtreed", key: "alice", keyring: "test", chainID: "test-1", node: "tcp://localhost:26657", fees: "0photon", maxCommit: 512},
		run:          f.run,
		store:        store,
		pollInterval: time.Millisecond,
		pollTimeout:  20 * time.Millisecond,
	}
}

func TestValidKind(t *testing.T) {
	accept := []string{"seed", "reflow.batch_root", "a:b-c_d.e", "A1", strings.Repeat("x", 128)}
	for _, k := range accept {
		if !validKind(k) {
			t.Errorf("validKind(%q) = false, want true", k)
		}
	}
	reject := []string{
		"",                       // empty
		"--home",                 // flag injection
		"--keyring-backend=test", // flag injection with '='
		"-x",                     // leading dash
		"foo bar",                // space
		"foo/bar",                // out-of-charset slash
		strings.Repeat("x", 129), // too long
	}
	for _, k := range reject {
		if validKind(k) {
			t.Errorf("validKind(%q) = true, want false", k)
		}
	}
}

func TestValidField(t *testing.T) {
	if !validField("did:web:id.dreamtree.org:tenants:gnosis", maxField) {
		t.Error("expected a normal DID subject to be valid")
	}
	if !validField("", maxField) {
		t.Error("empty field should be valid (optional)")
	}
	if validField("--home", maxField) {
		t.Error("leading-dash field should be rejected")
	}
	if validField("a\nb", maxField) {
		t.Error("control char should be rejected")
	}
	if validField(strings.Repeat("x", maxField+1), maxField) {
		t.Error("over-long field should be rejected")
	}
}

func TestCommitArgsTerminator(t *testing.T) {
	c := config{key: "alice", keyring: "test", fees: "0photon", chainID: "dreamtree-devnet-1", node: "tcp://localhost:26657", home: "/data/dt"}
	req := anchorReq{Subject: "did:web:x", Commitment: "abcd", Kind: "reflow.batch_root", SourceRef: "s3://ref"}
	args := c.commitArgs(req)

	// There must be exactly one `--`, and the two positionals must follow it in
	// order: commitment, kind.
	dash := -1
	for i, a := range args {
		if a == "--" {
			if dash != -1 {
				t.Fatalf("multiple `--` terminators in argv: %v", args)
			}
			dash = i
		}
	}
	if dash == -1 {
		t.Fatalf("no `--` terminator in argv: %v", args)
	}
	tail := args[dash+1:]
	if len(tail) != 2 || tail[0] != req.Commitment || tail[1] != req.Kind {
		t.Fatalf("positionals after `--` = %v, want [%q %q]", tail, req.Commitment, req.Kind)
	}

	// The flag pairs must live BEFORE the terminator.
	head := strings.Join(args[:dash], " ")
	for _, want := range []string{"--subject did:web:x", "--source-ref s3://ref", "--from alice", "--keyring-backend test"} {
		if !strings.Contains(head, want) {
			t.Errorf("expected %q before `--`; head=%q", want, head)
		}
	}
	// Nothing hostile-looking should appear after the terminator besides our two.
	if strings.Contains(strings.Join(tail, " "), "--") {
		t.Errorf("unexpected flag-like token after terminator: %v", tail)
	}
}

func TestIdempotentCacheHitNoSecondRun(t *testing.T) {
	f := &fakeRunner{}
	s := newTestServer(t, f)
	key := "explicit-key-123"
	// Pre-seed a completed entry.
	if err := s.store.put(key, idemRecord{ID: 42, TxHash: "CAFE", Height: 7, Status: statusDone}); err != nil {
		t.Fatalf("seed store: %v", err)
	}

	resp, replay, err := s.commit(anchorReq{Commitment: "abcd", Kind: "seed"}, key)
	if err != nil {
		t.Fatalf("commit: %v", err)
	}
	if !replay {
		t.Error("expected replay=true for a cached done entry")
	}
	if resp.ID != 42 || resp.TxHash != "CAFE" || resp.Height != 7 {
		t.Errorf("cached result not returned: %+v", resp)
	}
	if f.broadcasts != 0 {
		t.Errorf("cache hit broadcast %d times, want 0", f.broadcasts)
	}
	if f.queries != 0 {
		t.Errorf("cache hit polled %d times, want 0", f.queries)
	}
}

func TestBroadcastThenPollTimeoutRetryRepolls(t *testing.T) {
	f := &fakeRunner{}
	// Broadcast always succeeds. Query returns an empty/unparseable body first,
	// so the first commit's poll times out; later we flip it to a confirmed tx.
	confirmed := false
	f.broadcastFn = func() (string, error) { return `{"txhash":"BEEF01","code":0}`, nil }
	f.queryFn = func() (string, error) {
		if confirmed {
			return `{"code":0,"height":"8","events":[{"type":"seed_committed","attributes":[{"key":"id","value":"5"}]}]}`, nil
		}
		return "", nil // never parses -> poll keeps missing until deadline
	}
	s := newTestServer(t, f)
	key := "poll-timeout-key"
	req := anchorReq{Commitment: "abcd", Kind: "seed"}

	// First attempt: broadcast lands, poll times out.
	_, _, err := s.commit(req, key)
	if err == nil {
		t.Fatal("expected timeout error on first attempt")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout error, got: %v", err)
	}
	if f.broadcasts != 1 {
		t.Fatalf("first attempt broadcasts = %d, want 1", f.broadcasts)
	}
	// The pending record must have been persisted with the txhash BEFORE the poll.
	rec, ok := s.store.get(key)
	if !ok || rec.Status != statusPending || rec.TxHash != "BEEF01" {
		t.Fatalf("expected pending record with txhash after broadcast, got %+v ok=%v", rec, ok)
	}

	// Now the tx is confirmed on-chain. Client retries the same request.
	confirmed = true
	resp, replay, err := s.commit(req, key)
	if err != nil {
		t.Fatalf("retry commit: %v", err)
	}
	if !replay {
		t.Error("expected replay=true on retry (no re-broadcast)")
	}
	if f.broadcasts != 1 {
		t.Errorf("retry re-broadcast: broadcasts = %d, want 1 (must NOT broadcast again)", f.broadcasts)
	}
	if resp.ID != 5 || resp.Height != 8 || resp.TxHash != "BEEF01" {
		t.Errorf("retry result = %+v, want id=5 height=8 tx=BEEF01", resp)
	}
	rec, _ = s.store.get(key)
	if rec.Status != statusDone {
		t.Errorf("record status after retry = %q, want done", rec.Status)
	}
}

// TestReplayNotBlockedByInFlightBroadcast proves the audit-4b fix: a cached
// (statusDone) replay returns via the no-broadcast fast path even while s.mu is
// held by an in-flight broadcast+poll. Before the fix, commit acquired s.mu
// unconditionally, so a replay blocked behind a 20s poll.
func TestReplayNotBlockedByInFlightBroadcast(t *testing.T) {
	f := &fakeRunner{}
	s := newTestServer(t, f)
	// Seed a confirmed entry for this idem key.
	const key = "IDEMKEY"
	if err := s.store.put(key, idemRecord{ID: 7, BatchID: 3, NewCount: 1, TxHash: "ABC", Height: 42, Status: statusDone}); err != nil {
		t.Fatalf("seed store: %v", err)
	}
	// Simulate an in-flight broadcast holding the broadcast mutex.
	s.mu.Lock()
	defer s.mu.Unlock()

	done := make(chan anchorResp, 1)
	go func() {
		resp, _, err := s.commit(anchorReq{Commitment: "deadbeef", Kind: "record"}, key)
		if err == nil {
			done <- resp
		}
	}()

	select {
	case resp := <-done:
		if resp.ID != 7 || resp.Height != 42 {
			t.Fatalf("replay returned wrong result: %+v", resp)
		}
		if f.broadcasts != 0 {
			t.Fatalf("replay must not broadcast; got %d broadcasts", f.broadcasts)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("replay blocked behind the held broadcast mutex (regression of audit-4b fix)")
	}
}
