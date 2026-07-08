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
// Contract:
//
//	POST /anchor            (Bearer ANCHORD_TOKEN)
//	  { "subject": "...", "commitment": "<hex>", "kind": "...", "source_ref": "..." }
//	  -> 200 { "id": 1, "txhash": "...", "height": 8 }
//	GET  /healthz           -> 200 { "ok": true, "height": <latest> }
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
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
	maxCommit int
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func loadConfig() config {
	return config{
		bin:       envOr("ANCHORD_BIN", "dreamtreed"),
		key:       envOr("ANCHORD_KEY", "alice"),
		keyring:   envOr("ANCHORD_KEYRING", "test"),
		home:      envOr("ANCHORD_HOME", ""),
		chainID:   envOr("ANCHORD_CHAIN_ID", "dreamtree-devnet-1"),
		node:      envOr("ANCHORD_NODE", "tcp://localhost:26657"),
		fees:      envOr("ANCHORD_FEES", "0photon"),
		token:     os.Getenv("ANCHORD_TOKEN"),
		addr:      envOr("ANCHORD_ADDR", ":9110"),
		maxCommit: 512,
	}
}

var hexRe = regexp.MustCompile(`^[0-9a-fA-F]+$`)

type anchorReq struct {
	Subject    string `json:"subject"`
	Commitment string `json:"commitment"`
	Kind       string `json:"kind"`
	SourceRef  string `json:"source_ref"`
}

type anchorResp struct {
	ID     uint64 `json:"id"`
	TxHash string `json:"txhash"`
	Height int64  `json:"height"`
}

type server struct {
	cfg config
	mu  sync.Mutex // serialize broadcasts: one account, one sequence
}

func (s *server) authed(r *http.Request) bool {
	if s.cfg.token == "" {
		return true // no token configured: dev-only, open
	}
	h := r.Header.Get("Authorization")
	return strings.TrimPrefix(h, "Bearer ") == s.cfg.token
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

// run executes dreamtreed and returns combined output.
func (c config) run(args ...string) (string, error) {
	cmd := exec.Command(c.bin, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
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

func (s *server) commit(req anchorReq) (anchorResp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c := s.cfg
	args := []string{"tx", "seeds", "commit-seed", req.Commitment, req.Kind,
		"--from", c.key, "--keyring-backend", c.keyring,
		"--fees", c.fees, "--broadcast-mode", "sync", "-y", "--output", "json"}
	if req.Subject != "" {
		args = append(args, "--subject", req.Subject)
	}
	if req.SourceRef != "" {
		args = append(args, "--source-ref", req.SourceRef)
	}
	args = append(args, c.txArgs()...)

	out, err := c.run(args...)
	if err != nil {
		return anchorResp{}, fmt.Errorf("broadcast failed: %v: %s", err, out)
	}
	var b txBroadcast
	if err := json.Unmarshal([]byte(out), &b); err != nil {
		return anchorResp{}, fmt.Errorf("parse broadcast: %v: %s", err, out)
	}
	if b.Code != 0 {
		return anchorResp{}, fmt.Errorf("checktx rejected (code %d): %s", b.Code, b.RawLog)
	}

	// Poll for inclusion (DeliverTx result carries the assigned seed id).
	deadline := time.Now().Add(20 * time.Second)
	for {
		time.Sleep(1500 * time.Millisecond)
		qout, qerr := c.run(append([]string{"query", "tx", b.TxHash, "--output", "json"}, c.queryArgs()...)...)
		if qerr == nil {
			var res txResult
			if json.Unmarshal([]byte(qout), &res) == nil {
				if res.Code != 0 {
					return anchorResp{}, fmt.Errorf("delivertx failed (code %d): %s", res.Code, res.RawLog)
				}
				id := extractSeedID(res)
				h, _ := strconv.ParseInt(res.Height, 10, 64)
				return anchorResp{ID: id, TxHash: b.TxHash, Height: h}, nil
			}
		}
		if time.Now().After(deadline) {
			return anchorResp{}, fmt.Errorf("timed out waiting for tx %s inclusion", b.TxHash)
		}
	}
}

func extractSeedID(res txResult) uint64 {
	for _, ev := range res.Events {
		if ev.Type != "seed_committed" {
			continue
		}
		for _, a := range ev.Attributes {
			if a.Key == "id" {
				id, _ := strconv.ParseUint(a.Value, 10, 64)
				return id
			}
		}
	}
	return 0
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

	resp, err := s.commit(req)
	if err != nil {
		log.Printf("anchor error: %v", err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	log.Printf("anchored seed #%d kind=%s subject=%s tx=%s", resp.ID, req.Kind, req.Subject, resp.TxHash)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	out, err := s.cfg.run(append([]string{"status", "--output", "json"}, s.cfg.queryArgs()...)...)
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
	s := &server{cfg: cfg}
	mux := http.NewServeMux()
	mux.HandleFunc("/anchor", s.handleAnchor)
	mux.HandleFunc("/healthz", s.handleHealth)

	if cfg.token == "" {
		log.Printf("WARNING: ANCHORD_TOKEN unset — /anchor is open (dev only)")
	}
	log.Printf("anchord listening on %s (chain=%s node=%s key=%s)", cfg.addr, cfg.chainID, cfg.node, cfg.key)
	srv := &http.Server{
		Addr:         cfg.addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
