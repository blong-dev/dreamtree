#!/usr/bin/env bash
#
# Proves the leaf model on a LIVE node (seed = atom, DT-18): a batch commit
# registers N leaf-seeds under one Merkle root, mints exactly N photons
# (= uphoton×10^6) to the storer recipient, individual leaves resolve by id,
# pure-convergence batches anchor without allocating or minting, the retired
# batch_root kind is rejected, and the single-commit (roots) path still works.
# Throwaway isolated devnet, photon-native (bond denom uphoton).
set -euo pipefail

BIN=$(command -v dreamtreed || echo "$(go env GOPATH)/bin/dreamtreed")
HOME_DIR=${DT_LEAF_HOME:-/tmp/dt-leaf}
CHAIN=dreamtree-leaf-1
KR=(--keyring-backend test --home "$HOME_DIR")
TX=(--chain-id "$CHAIN" "${KR[@]}" --gas 3000000 --yes -o json)
Q=(--home "$HOME_DIR" -o json)
ROOT="aa11bb22cc33dd44ee55ff66aa11bb22cc33dd44ee55ff66aa11bb22cc33dd44"

ok()  { printf '\033[1;32mPASS\033[0m %s\n' "$*"; }
die() { printf '\033[1;31mFAIL\033[0m %s\n' "$*"; exit 1; }
wait_tx() { local h=$1 i out; for i in $(seq 1 30); do if out=$("$BIN" q tx "$h" "${Q[@]}" 2>/dev/null); then echo "$out"; return; fi; sleep 1; done; die "tx $h never landed"; }
send() { local out h; out=$("$BIN" tx "$@" "${TX[@]}"); h=$(echo "$out"|jq -r .txhash); wait_tx "$h"; }

rm -rf "$HOME_DIR"; mkdir -p "$HOME_DIR"
"$BIN" init leaf --chain-id "$CHAIN" --default-denom uphoton --home "$HOME_DIR" >/dev/null 2>&1
"$BIN" keys add alice "${KR[@]}" >/dev/null 2>&1
ALICE=$("$BIN" keys show alice -a "${KR[@]}")
"$BIN" genesis add-genesis-account alice 1000000000uphoton "${KR[@]}" >/dev/null
"$BIN" genesis gentx alice 500000000uphoton --chain-id "$CHAIN" "${KR[@]}" >/dev/null 2>&1
"$BIN" genesis collect-gentxs --home "$HOME_DIR" >/dev/null 2>&1

# Route the ingestion mint to alice so the photon arithmetic is observable.
python3 - "$HOME_DIR/config/genesis.json" "$ALICE" <<'PY'
import json, sys
g, alice = sys.argv[1], sys.argv[2]
d = json.load(open(g))
d['app_state']['photons']['params']['storer_reward_recipient'] = alice
json.dump(d, open(g, 'w'), indent=1)
PY
sed -i 's/^timeout_commit = .*/timeout_commit = "1s"/' "$HOME_DIR/config/config.toml"

"$BIN" start --home "$HOME_DIR" --minimum-gas-prices 0uphoton >"$HOME_DIR/node.log" 2>&1 &
NODE=$!; trap 'kill $NODE 2>/dev/null || true' EXIT
for i in $(seq 1 40); do h=$("$BIN" status "${Q[@]}" 2>/dev/null|jq -r '.sync_info.latest_block_height' 2>/dev/null||echo 0); [ "${h:-0}" -ge 1 ] 2>/dev/null && break; sleep 1; done
[ "${h:-0}" -ge 1 ] || die "node never produced a block"
ok "photon-native node live (bond denom uphoton)"

BAL0=$("$BIN" q bank balance "$ALICE" uphoton "${Q[@]}" | jq -r '.balance.amount')

# 1. Batch: 5 leaves, 3 new (2 converged) -> seeds 1..3, 3 photons.
OUT=$(send seeds commit-batch "$ROOT" 5 3 record --subject did:web:test --source-ref reflow:gen:1 --from alice)
[ "$(echo "$OUT"|jq -r .code)" = "0" ] || die "commit-batch rejected: $(echo "$OUT"|jq -r .raw_log)"
FIRST=$(echo "$OUT" | jq -r '.events[] | select(.type=="seed_batch_committed") | .attributes[] | select(.key=="first_id") | .value')
NEWC=$(echo "$OUT"  | jq -r '.events[] | select(.type=="seed_batch_committed") | .attributes[] | select(.key=="new_count") | .value')
[ "$FIRST" = "1" ] || die "first_id=$FIRST, expected 1"
[ "$NEWC" = "3" ]  || die "new_count=$NEWC, expected 3"
ok "batch committed: seeds [1,4) under one root (5 leaves, 3 new)"

MINTED=$("$BIN" q photons supply "${Q[@]}" | jq -r '.minted')
[ "$MINTED" = "3" ] || die "photon supply=$MINTED, expected 3 (= new atoms)"
BAL1=$("$BIN" q bank balance "$ALICE" uphoton "${Q[@]}" | jq -r '.balance.amount')
[ "$((BAL1 - BAL0))" = "3000000" ] || die "storer got $((BAL1-BAL0)) uphoton, expected 3000000 (3 photons)"
ok "photons = seeds: 3 minted (3,000,000 uphoton) to the storer"

# 2. Leaves resolve individually; past-range does not.
LEAF2=$("$BIN" q seeds seed 2 "${Q[@]}")
[ "$(echo "$LEAF2"|jq -r '.seed.commitment')" = "$ROOT" ] || die "leaf 2 commitment mismatch"
[ "$(echo "$LEAF2"|jq -r '.seed.leaf_index')" = "1" ] || die "leaf 2 index != 1"
"$BIN" q seeds seed 4 "${Q[@]}" >/dev/null 2>&1 && die "seed 4 resolved — converged leaves must not allocate"
ok "leaf resolution: seed 2 synthesized from its batch; seed 4 correctly absent"

# 3. Pure convergence: anchors, allocates nothing, mints nothing.
OUT=$(send seeds commit-batch "$ROOT" 4 0 record --source-ref reflow:gen:2 --from alice)
[ "$(echo "$OUT"|jq -r .code)" = "0" ] || die "pure-convergence batch rejected: $(echo "$OUT"|jq -r .raw_log)"
MINTED=$("$BIN" q photons supply "${Q[@]}" | jq -r '.minted')
[ "$MINTED" = "3" ] || die "photon supply=$MINTED after convergence, expected still 3"
ok "pure-convergence batch anchored: no ids, no photons (sigma accrues, supply doesn't)"

# 4. The single-commit (roots) path is a batch of one.
OUT=$(send seeds commit-seed "$ROOT" record --subject did:web:roots --from alice)
SID=$(echo "$OUT" | jq -r '.events[] | select(.type=="seed_committed") | .attributes[] | select(.key=="id") | .value')
[ "$SID" = "4" ] || die "single commit id=$SID, expected 4 (next after 1..3)"
MINTED=$("$BIN" q photons supply "${Q[@]}" | jq -r '.minted')
[ "$MINTED" = "4" ] || die "photon supply=$MINTED, expected 4"
ok "single commit (roots path): seed 4, one photon — batch of one"

# 5. The retired aggregate kind is refused.
OUT=$("$BIN" tx seeds commit-batch "$ROOT" 2 2 reflow.batch_root --from alice "${TX[@]}" 2>&1) || true
H=$(echo "$OUT"|jq -r '.txhash // empty' 2>/dev/null)
if [ -n "$H" ]; then
  RES=$(wait_tx "$H"); CODE=$(echo "$RES"|jq -r .code)
  [ "$CODE" != "0" ] || die "batch_root kind was accepted — must be retired"
fi
ok "batch_root kind rejected (kind names the leaf)"

echo; echo "LEAF PROOF COMPLETE — seed = atom on a live photon-native chain."