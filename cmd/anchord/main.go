// Command anchord is the anchoring seam between off-chain producers (the roots
// wallet, the gnosis knowledge graph) and the dreamtree chain.
//
// Producers speak HTTP, not gRPC, and never hold a signing key. anchord runs on
// the same host as the chain, holds the single anchor key, and turns an HTTP
// POST into a signed MsgCommitSeed. It serializes broadcasts so the account
// sequence never races.
//
// v0 deliberately shells the installed `dreamtreed` binary to broadcast rather
// than embedding the SDK client — one trusted path, easy to audit, adequate for
// the batched anchor volume. The upgrade to a native SDK client is tracked and
// changes nothing about the HTTP contract below.
//
// Each accepted commit mints exactly one photon on-chain, so a duplicated or
// retried POST must never broadcast twice. anchord is therefore idempotent: it
// keys each request (explicit Idempotency-Key header, or an implicit hash of the
// request fields) into a persistent local store, records the broadcast txhash
// BEFORE polling for inclusion, and replays the cached result on retry.
//
// Contract:
//
//	POST /anchor            (Bearer ANCHORD_TOKEN)
//	  { "subject": "...", "commitment": "<hex>", "kind": "...", "source_ref": "..." }
//	  [ Idempotency-Key: <opaque> ]     (optional)
//	  -> 200 { "id": 1, "txhash": "...", "height": 8 }
//	  -> 200 + "X-Idempotent-Replay: true" when the result was replayed, not re-broadcast
//
//	Batch anchoring (the leaf model — docs/specs/seed-atom-conformance.md): add
//	"leaf_count" and "new_count" to the same POST; the commitment is the Merkle
//	root over the batch's leaf ids. new_count leaf-seeds are registered (and
//	new_count photons minted); converged re-observations count only in
//	leaf_count. Response gains the batch fields:
//	  { "subject": "...", "commitment": "<merkle root hex>", "kind": "record",
//	    "source_ref": "reflow:gen:61932", "leaf_count": 190, "new_count": 187 }
//	  -> 200 { "id": <first_id>, "batch_id": 7, "new_count": 187, "txhash": "...", "height": 8 }
//
//	GET  /healthz           -> 200 { "ok": true, "height": <latest> }
package main

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type config struct {
	bin       string
	key       string
	keyring   string
	home      string
	chainID   string
	node      string
	fees      string
	token     string
	addr      string
	state     string
	maxCommit int
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// defaultStatePath resolves where the idempotency store lives when ANCHORD_STATE
// is unset: under ANCHORD_HOME if set, otherwise under ~/.anchord.
func defaultStatePath(home string) string {
	base := home
	if base == "" {
		if h, err := os.UserHomeDir(); err == nil {
			base = filepath.Join(h, ".anchord")
		} else {
			base = "."
		}
	}
	return filepath.Join(base, "anchord-idem.json")
}

func loadConfig() config {
	home := envOr("ANCHORD_HOME", "")
	return config{
		bin:       envOr("ANCHORD_BIN", "dreamtreed"),
		key:       envOr("ANCHORD_KEY", "alice"),
		keyring:   envOr("ANCHORD_KEYRING", "test"),
		home:      home,
		chainID:   envOr("ANCHORD_CHAIN_ID", "dreamtree-devnet-1"),
		node:      envOr("ANCHORD_NODE", "tcp://localhost:26657"),
		fees:      envOr("ANCHORD_FEES", "0uphoton"),
		token:     os.Getenv("ANCHORD_TOKEN"),
		addr:      envOr("ANCHORD_ADDR", ":9110"),
		state:     envOr("ANCHORD_STATE", defaultStatePath(home)),
		maxCommit: 512,
	}
}

var hexRe = regexp.MustCompile(`^[0-9a-fA-F]+$`)

// kindRe bounds the on-chain `kind` positional to a safe charset. Even with the
// `--` argv terminator in place, keeping kind well-formed is defense in depth.
var kindRe = regexp.MustCompile(`^[a-zA-Z0-9._:-]{1,128}$`)

// validKind rejects empty, over-long, out-of-charset, and leading-dash values.
// The leading-dash check matters because `-` is in the charset (e.g. names like
// `a-b`) but a value like `--home` must never be mistaken for a flag.
func validKind(k string) bool {
	if strings.HasPrefix(k, "-") {
		return false
	}
	return kindRe.MatchString(k)
}

// validField is a cheap sanity guard for flag VALUES (subject, source_ref):
// bound the length, reject a leading dash, and reject control characters.
func validField(v string, max int) bool {
	if len(v) > max {
		return false
	}
	if strings.HasPrefix(v, "-") {
		return false
	}
	for _, r := range v {
		if r < 0x20 {
			return false
		}
	}
	return true
}

const maxField = 1024

type anchorReq struct {
	Subject    string `json:"subject"`
	Commitment string `json:"commitment"`
	Kind       string `json:"kind"`
	SourceRef  string `json:"source_ref"`
	// Batch fields (leaf model). Both zero => single-seed path (batch of one).
	LeafCount uint32 `json:"leaf_count,omitempty"`
	NewCount  uint32 `json:"new_count,omitempty"`
}

// isBatch reports whether the request takes the commit-batch path.
func (r anchorReq) isBatch() bool { return r.LeafCount > 0 || r.NewCount > 0 }

type anchorResp struct {
	ID       uint64 `json:"id"` // first (or only) leaf-seed id
	BatchID  uint64 `json:"batch_id,omitempty"`
	NewCount uint32 `json:"new_count,omitempty"`
	TxHash   string `json:"txhash"`
	Height   int64  `json:"height"`
}

// txRunner executes the dreamtreed CLI and returns its combined output. It is an
// injection point: production wires config.run, tests supply a fake so the
// commit path can be exercised without a live chain.
type txRunner func(args ...string) (string, error)

type server struct {
	cfg          config
	run          txRunner
	store        *idemStore
	pollInterval time.Duration
	pollTimeout  time.Duration
	mu           sync.Mutex // serialize broadcasts: one account, one sequence
}

func (s *server) authed(r *http.Request) bool {
	if s.cfg.token == "" {
		return true // no token configured: dev-only, open
	}
	provided := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	// Constant-time compare: don't leak the token via response timing.
	return subtle.ConstantTimeCompare([]byte(provided), []byte(s.cfg.token)) == 1
}

// txArgs are the flags a `tx` subcommand needs (chain-id is tx-only).
func (c config) txArgs() []string {
	args := []string{"--chain-id", c.chainID, "--node", c.node}
	if c.home != "" {
		args = append(args, "--home", c.home)
	}
	return args
}

// queryArgs are the flags a `query`/`status` subcommand accepts (no chain-id).
func (c config) queryArgs() []string {
	args := []string{"--node", c.node}
	if c.home != "" {
		args = append(args, "--home", c.home)
	}
	return args
}

// commitArgs builds the full argv for `tx seeds commit-seed` (single) or
// `tx seeds commit-batch` (leaf model). All flags come first, then a `--`
// terminator, then the positionals. The terminator stops cobra/pflag from
// interpreting a hostile `kind` (or, in the worst case, commitment) as a flag.
func (c config) commitArgs(req anchorReq) []string {
	sub := "commit-seed"
	if req.isBatch() {
		sub = "commit-batch"
	}
	args := []string{"tx", "seeds", sub,
		"--from", c.key, "--keyring-backend", c.keyring,
		"--fees", c.fees, "--broadcast-mode", "sync", "-y", "--output", "json"}
	if req.Subject != "" {
		args = append(args, "--subject", req.Subject)
	}
	if req.SourceRef != "" {
		args = append(args, "--source-ref", req.SourceRef)
	}
	args = append(args, c.txArgs()...)
	// Everything past `--` is positional; values can never be read as flags.
	if req.isBatch() {
		// autocli positional order: merkle-root, leaf-count, new-count, kind.
		args = append(args, "--", req.Commitment,
			strconv.FormatUint(uint64(req.LeafCount), 10),
			strconv.FormatUint(uint64(req.NewCount), 10),
			req.Kind)
	} else {
		args = append(args, "--", req.Commitment, req.Kind)
	}
	return args
}

// run executes dreamtreed and returns STDOUT (what callers JSON-parse).
// stderr is kept separate: the CLI writes warnings there (e.g. the go1.24
// sonic notice), and merging streams contaminated the JSON. On error, stderr
// is folded into the error so failures stay diagnosable.
func (c config) run(args ...string) (string, error) {
	cmd := exec.Command(c.bin, args...)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	err := cmd.Run()
	if err != nil && errBuf.Len() > 0 {
		err = fmt.Errorf("%w: %s", err, strings.TrimSpace(errBuf.String()))
	}
	return out.String(), err
}

type txBroadcast struct {
	TxHash string `json:"txhash"`
	Code   int    `json:"code"`
	RawLog string `json:"raw_log"`
}

type txResult struct {
	Code   int    `json:"code"`
	Height string `json:"height"`
	RawLog string `json:"raw_log"`
	Events []struct {
		Type       string `json:"type"`
		Attributes []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"attributes"`
	} `json:"events"`
}

// idemStatus values for a stored idempotency record.
const (
	statusPending = "pending" // broadcast succeeded; inclusion not yet confirmed
	statusDone    = "done"    // inclusion confirmed; id/height final
)

type idemRecord struct {
	ID       uint64 `json:"id"`
	BatchID  uint64 `json:"batch_id,omitempty"`
	NewCount uint32 `json:"new_count,omitempty"`
	TxHash   string `json:"txhash"`
	Height   int64  `json:"height"`
	Status   string `json:"status"`
}

// idemStore is a tiny persistent map of idem-key -> record. Volume is low
// (batched anchors), so a fsync'd JSON file guarded by a mutex is sufficient and
// avoids a heavyweight embedded DB. An empty path keeps it purely in-memory
// (used by tests).
type idemStore struct {
	path    string
	mu      sync.Mutex
	entries map[string]idemRecord
}

func newIdemStore(path string) (*idemStore, error) {
	s := &idemStore{path: path, entries: map[string]idemRecord{}}
	if path == "" {
		return s, nil
	}
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		if len(bytes.TrimSpace(data)) > 0 {
			if err := json.Unmarshal(data, &s.entries); err != nil {
				return nil, fmt.Errorf("load idem store %s: %w", path, err)
			}
		}
	case os.IsNotExist(err):
		// fresh store
	default:
		return nil, err
	}
	return s, nil
}

func (s *idemStore) get(key string) (idemRecord, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[key]
	return e, ok
}

func (s *idemStore) put(key string, rec idemRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[key] = rec
	return s.persistLocked()
}

// persistLocked writes the whole map atomically: temp file + fsync + rename.
func (s *idemStore) persistLocked() error {
	if s.path == "" {
		return nil // in-memory only
	}
	data, err := json.Marshal(s.entries)
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// deriveIdemKey computes the implicit idempotency key from the request fields.
// Single-seed requests MUST keep the exact legacy derivation: the persistent
// idem store carries pre-batch-era records under it, and changing the key
// would orphan them — a pre-upgrade pending record would then re-broadcast on
// retry and double-mint (the precise failure this store exists to prevent).
// Batch requests are new, so their counts join the key without collision.
func deriveIdemKey(req anchorReq) string {
	base := req.Subject + "\n" + req.Commitment + "\n" + req.Kind + "\n" + req.SourceRef
	if req.isBatch() {
		base += "\n" + strconv.FormatUint(uint64(req.LeafCount), 10) +
			"\n" + strconv.FormatUint(uint64(req.NewCount), 10)
	}
	h := sha256.Sum256([]byte(base))
	return hex.EncodeToString(h[:])
}

// commit turns a validated request into (at most) one on-chain seed. The bool
// return reports whether the result was replayed from cache (true) rather than
// freshly broadcast (false).
func (s *server) commit(req anchorReq, idemKey string) (anchorResp, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Idempotency check under the broadcast mutex.
	if e, ok := s.store.get(idemKey); ok {
		switch e.Status {
		case statusDone:
			// Already confirmed: replay, never broadcast again.
			return anchorResp{ID: e.ID, BatchID: e.BatchID, NewCount: e.NewCount, TxHash: e.TxHash, Height: e.Height}, true, nil
		case statusPending:
			if e.TxHash != "" {
				// The previous attempt broadcast successfully but its inclusion
				// poll timed out. The tx may well have landed. Re-poll — do NOT
				// re-broadcast (that would mint the photons a second time).
				resp, err := s.pollInclusion(e.TxHash)
				if err != nil {
					return anchorResp{}, true, err
				}
				if err := s.store.put(idemKey, idemRecord{ID: resp.ID, BatchID: resp.BatchID, NewCount: resp.NewCount, TxHash: resp.TxHash, Height: resp.Height, Status: statusDone}); err != nil {
					return anchorResp{}, true, err
				}
				return resp, true, nil
			}
		}
	}

	args := s.cfg.commitArgs(req)
	out, err := s.run(args...)
	if err != nil {
		return anchorResp{}, false, fmt.Errorf("broadcast failed: %v: %s", err, out)
	}
	var b txBroadcast
	if err := json.Unmarshal([]byte(out), &b); err != nil {
		return anchorResp{}, false, fmt.Errorf("parse broadcast: %v: %s", err, out)
	}
	if b.Code != 0 {
		return anchorResp{}, false, fmt.Errorf("checktx rejected (code %d): %s", b.Code, b.RawLog)
	}

	// RACE FIX: persist key -> txhash NOW, before polling. CheckTx accepted the
	// tx, so it will (almost certainly) land. If our poll below times out after
	// it lands, a client retry finds this pending record and re-polls instead of
	// broadcasting a duplicate seed.
	if err := s.store.put(idemKey, idemRecord{TxHash: b.TxHash, Status: statusPending}); err != nil {
		return anchorResp{}, false, fmt.Errorf("persist pending idem entry: %w", err)
	}

	resp, err := s.pollInclusion(b.TxHash)
	if err != nil {
		// Leave the pending record in place: a retry will re-poll this txhash.
		return anchorResp{}, false, err
	}
	if err := s.store.put(idemKey, idemRecord{ID: resp.ID, BatchID: resp.BatchID, NewCount: resp.NewCount, TxHash: resp.TxHash, Height: resp.Height, Status: statusDone}); err != nil {
		return anchorResp{}, false, err
	}
	return resp, false, nil
}

// pollInclusion polls `query tx <hash>` until the DeliverTx result is available
// (carrying the assigned seed id) or the deadline passes.
func (s *server) pollInclusion(txhash string) (anchorResp, error) {
	c := s.cfg
	deadline := time.Now().Add(s.pollTimeout)
	for {
		time.Sleep(s.pollInterval)
		qout, qerr := s.run(append([]string{"query", "tx", txhash, "--output", "json"}, c.queryArgs()...)...)
		if qerr == nil {
			var res txResult
			if json.Unmarshal([]byte(qout), &res) == nil {
				if res.Code != 0 {
					return anchorResp{}, fmt.Errorf("delivertx failed (code %d): %s", res.Code, res.RawLog)
				}
				id, batchID, newCount := extractAnchorIDs(res)
				h, _ := strconv.ParseInt(res.Height, 10, 64)
				return anchorResp{ID: id, BatchID: batchID, NewCount: newCount, TxHash: txhash, Height: h}, nil
			}
		}
		if time.Now().After(deadline) {
			return anchorResp{}, fmt.Errorf("timed out waiting for tx %s inclusion", txhash)
		}
	}
}

// extractAnchorIDs reads the batch event (the canonical emission — single
// commits are batches of one) and returns (first_id, batch_id, new_count).
// Falls back to the legacy seed_committed event if only that is present.
func extractAnchorIDs(res txResult) (id, batchID uint64, newCount uint32) {
	for _, ev := range res.Events {
		if ev.Type != "seed_batch_committed" {
			continue
		}
		for _, a := range ev.Attributes {
			switch a.Key {
			case "first_id":
				id, _ = strconv.ParseUint(a.Value, 10, 64)
			case "batch_id":
				batchID, _ = strconv.ParseUint(a.Value, 10, 64)
			case "new_count":
				n, _ := strconv.ParseUint(a.Value, 10, 32)
				newCount = uint32(n)
			}
		}
		return id, batchID, newCount
	}
	for _, ev := range res.Events {
		if ev.Type != "seed_committed" {
			continue
		}
		for _, a := range ev.Attributes {
			if a.Key == "id" {
				id, _ = strconv.ParseUint(a.Value, 10, 64)
				return id, 0, 1
			}
		}
	}
	return 0, 0, 0
}

func (s *server) handleAnchor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.authed(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req anchorReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	req.Commitment = strings.TrimSpace(req.Commitment)
	req.Kind = strings.TrimSpace(req.Kind)
	if req.Commitment == "" || !hexRe.MatchString(req.Commitment) {
		http.Error(w, "commitment must be non-empty hex", http.StatusBadRequest)
		return
	}
	if len(req.Commitment) > s.cfg.maxCommit {
		http.Error(w, "commitment too long", http.StatusBadRequest)
		return
	}
	if req.Kind == "" {
		http.Error(w, "kind is required", http.StatusBadRequest)
		return
	}
	if !validKind(req.Kind) {
		http.Error(w, "kind must match [a-zA-Z0-9._:-]{1,128} and not start with '-'", http.StatusBadRequest)
		return
	}
	if !validField(req.Subject, maxField) {
		http.Error(w, "subject invalid (too long, leading '-', or control chars)", http.StatusBadRequest)
		return
	}
	if !validField(req.SourceRef, maxField) {
		http.Error(w, "source_ref invalid (too long, leading '-', or control chars)", http.StatusBadRequest)
		return
	}
	if req.isBatch() {
		// new_count == 0 is a valid pure-convergence batch (provenance only).
		if req.NewCount > req.LeafCount {
			http.Error(w, "batch counts invalid: need new_count <= leaf_count", http.StatusBadRequest)
			return
		}
	}

	idemKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idemKey == "" {
		idemKey = deriveIdemKey(req)
	}

	resp, replay, err := s.commit(req, idemKey)
	if err != nil {
		log.Printf("anchor error: %v", err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	if replay {
		w.Header().Set("X-Idempotent-Replay", "true")
		log.Printf("anchor replay seed #%d kind=%s subject=%s tx=%s", resp.ID, req.Kind, req.Subject, resp.TxHash)
	} else {
		log.Printf("anchored seed #%d kind=%s subject=%s tx=%s", resp.ID, req.Kind, req.Subject, resp.TxHash)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	out, err := s.run(append([]string{"status", "--output", "json"}, s.cfg.queryArgs()...)...)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": err.Error()})
		return
	}
	var st struct {
		SyncInfo struct {
			LatestBlockHeight string `json:"latest_block_height"`
		} `json:"sync_info"`
	}
	_ = json.Unmarshal([]byte(out), &st)
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "height": st.SyncInfo.LatestBlockHeight})
}

func main() {
	cfg := loadConfig()
	store, err := newIdemStore(cfg.state)
	if err != nil {
		log.Fatalf("idempotency store: %v", err)
	}
	s := &server{
		cfg:          cfg,
		run:          cfg.run,
		store:        store,
		pollInterval: 1500 * time.Millisecond,
		pollTimeout:  20 * time.Second,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/anchor", s.handleAnchor)
	mux.HandleFunc("/healthz", s.handleHealth)

	if cfg.token == "" {
		log.Printf("WARNING: ANCHORD_TOKEN unset — /anchor is open (dev only)")
	}
	log.Printf("anchord listening on %s (chain=%s node=%s key=%s state=%s)", cfg.addr, cfg.chainID, cfg.node, cfg.key, cfg.state)
	srv := &http.Server{
		Addr:         cfg.addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
