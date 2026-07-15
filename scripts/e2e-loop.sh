#!/usr/bin/env bash
#
# End-to-end economic loop on a throwaway devnet. Proves the whole chain runs as
# ONE flow, not just module-by-module:
#
#   commit seed -> photon mints to storer -> attest -> outcome validates it ->
#   review window settles -> reputation moves -> marketplace sells access ->
#   producer earns photons.
#
# Isolated home + fast blocks + tiny review window so outcome windows close in
# seconds. Nothing here touches ~/.dreamtreed or any real chain.
set -euo pipefail

BIN=$(command -v dreamtreed || echo "$(go env GOPATH)/bin/dreamtreed")
HOME_DIR=${DT_E2E_HOME:-/tmp/claude-1000/-home-b/99adc05f-ea53-4822-9a7c-49d0a182ddbf/scratchpad/dt-e2e}
CHAIN=dreamtree-e2e-1
KR=(--keyring-backend test --home "$HOME_DIR")
TX=(--chain-id "$CHAIN" "${KR[@]}" --gas 3000000 --yes -o json)
Q=(--home "$HOME_DIR" -o json)

say() { printf '\n\033[1;36m== %s\033[0m\n' "$*"; }
ok()  { printf '\033[1;32mPASS\033[0m %s\n' "$*"; }
die() { printf '\033[1;31mFAIL\033[0m %s\n' "$*"; exit 1; }

# wait_tx <txhash> — block until the tx is in a block; fail on non-zero code.
wait_tx() {
  local h=$1 i out code
  for i in $(seq 1 30); do
    if out=$("$BIN" q tx "$h" "${Q[@]}" 2>/dev/null); then
      code=$(echo "$out" | jq -r '.code')
      [ "$code" = "0" ] || die "tx $h failed code=$code: $(echo "$out" | jq -r '.raw_log')"
      return 0
    fi
    sleep 1
  done
  die "tx $h never landed"
}
send() { # send <args...> -> echoes txhash after inclusion
  local out h
  out=$("$BIN" tx "$@" "${TX[@]}")
  h=$(echo "$out" | jq -r '.txhash')
  wait_tx "$h" >&2
  echo "$h"
}
bal() { "$BIN" q bank balances "$1" "${Q[@]}" | jq -r --arg d "$2" '.balances[]|select(.denom==$d)|.amount' | head -1; }

# ---------------------------------------------------------------------------
say "1. init isolated devnet ($HOME_DIR)"
rm -rf "$HOME_DIR"; mkdir -p "$HOME_DIR"
"$BIN" init e2e --chain-id "$CHAIN" --default-denom uphoton --home "$HOME_DIR" >/dev/null 2>&1
for k in alice bob; do "$BIN" keys add "$k" "${KR[@]}" >/dev/null 2>&1; done
ALICE=$("$BIN" keys show alice -a "${KR[@]}")
BOB=$("$BIN" keys show bob -a "${KR[@]}")
"$BIN" genesis add-genesis-account alice 1000000000uphoton "${KR[@]}" >/dev/null
"$BIN" genesis add-genesis-account bob 1000000uphoton "${KR[@]}" >/dev/null
"$BIN" genesis gentx alice 500000000uphoton --chain-id "$CHAIN" "${KR[@]}" >/dev/null 2>&1
"$BIN" genesis collect-gentxs --home "$HOME_DIR" >/dev/null 2>&1

# Patch genesis: storer reward -> alice; tiny review window; price "record"=1.
GEN="$HOME_DIR/config/genesis.json"
python3 - "$GEN" "$ALICE" <<'PY'
import json, sys
g, alice = sys.argv[1], sys.argv[2]
d = json.load(open(g))
d['app_state']['photons']['params']['storer_reward_recipient'] = alice
rp = d['app_state']['reputation']['params']
rp['review_window_base'] = "0.0005"       # τ ~5s (bet) / ~15s (outcome), not ~12h
d['app_state']['licenses']['type_prices'] = [{"data_type": "record", "price": "1"}]
json.dump(d, open(g, 'w'), indent=1)
PY
# Fast blocks so the loop finishes quickly.
sed -i 's/^timeout_commit = .*/timeout_commit = "1s"/' "$HOME_DIR/config/config.toml"
ok "genesis: storer=alice, review_window_base=0.0005, price[record]=1"

say "2. start node"
"$BIN" start --home "$HOME_DIR" --minimum-gas-prices 0uphoton >"$HOME_DIR/node.log" 2>&1 &
NODE=$!
trap 'kill $NODE 2>/dev/null || true' EXIT
for i in $(seq 1 40); do
  h=$("$BIN" status "${Q[@]}" 2>/dev/null | jq -r '.sync_info.latest_block_height' 2>/dev/null || echo 0)
  [ "${h:-0}" -ge 1 ] 2>/dev/null && break; sleep 1
done
[ "${h:-0}" -ge 1 ] || die "node never produced a block"
ok "node producing blocks (height $h)"

# ---------------------------------------------------------------------------
say "3. alice commits 3 'record' seeds -> 3 photons mint to alice"
for n in 1 2 3; do
  send seeds commit-seed "$(printf 'deadbeef%064x' $n | head -c64)" record \
      --data-type record --subject "did:dream:work$n" --from alice >/dev/null
done
SUPPLY=$("$BIN" q photons supply "${Q[@]}" | jq -r '.minted // .supply // .count')
APHOT=$(bal "$ALICE" photon)
[ "$SUPPLY" = "3" ] || die "photon supply = $SUPPLY, want 3 (peg photons=seeds)"
[ "$APHOT" = "3" ]  || die "alice photon balance = $APHOT, want 3"
ok "photon supply=$SUPPLY, alice balance=$APHOT (mint-on-ingest to storer works)"

say "4. bob attests alice's work (rigor, science/x) -> bet window"
send attest attest "did:dream:work1" rigor --domain science/x --specificity-bps 5000 \
    --from bob >/dev/null
BOB_ATT=$("$BIN" q attest by-attestor "$BOB" "${Q[@]}" | jq -r '.attestations[0].id')
[ -n "$BOB_ATT" ] && [ "$BOB_ATT" != "null" ] || die "bob's attestation id not found"
ok "bob attestation id=$BOB_ATT"

say "5. alice attests an OUTCOME validating bob's attestation -> outcome window"
send attest attest "did:dream:work1" outcome --outcome-kind validated --target-id "$BOB_ATT" \
    --domain science/x --specificity-bps 5000 --from alice >/dev/null
PEND=$("$BIN" q reputation pending "${Q[@]}" | jq -r '.pending|length')
ok "outcome recorded; $PEND event(s) in review window"

say "6. wait for review windows to close (settlement in EndBlock)"
for i in $(seq 1 25); do
  n=$("$BIN" q reputation pending "${Q[@]}" | jq -r '.pending|length')
  [ "$n" = "0" ] && break; sleep 1
done
[ "$n" = "0" ] || die "windows never drained ($n still pending)"
ok "all windows settled"

say "7. bob's reputation moved above baseline (1.0)"
echo "  bob contributions:"; "$BIN" q reputation contributions "$BOB" "${Q[@]}" | jq -c '.contributions[]? | {mag:.magnitude, bucket:.rate_bucket, domain:.domain, src:.source_att_id}'
R=$("$BIN" q reputation reputation "$BOB" science/x "${Q[@]}" | jq -r '.r // .reputation // .value')
awk -v r="$R" 'BEGIN{exit !(r+0 > 1.0)}' || die "bob R=$R, expected > 1.0 (bet + validated outcome)"
ok "bob R(science/x) = $R  (> baseline 1.0 — validated work paid out)"

say "8. marketplace: alice funds bob, bob buys access to seed 1"
send bank send "$ALICE" "$BOB" 2photon --from alice >/dev/null
BOB_P0=$(bal "$BOB" photon); ALICE_P0=$(bal "$ALICE" photon)
send licenses purchase 1 --from bob >/dev/null
ACCESS=$("$BIN" q licenses access "$BOB" 1 "${Q[@]}" | jq -r '.has_access // .access // .held')
BOB_P1=$(bal "$BOB" photon); ALICE_P1=$(bal "$ALICE" photon)
SUP2=$("$BIN" q photons supply "${Q[@]}" | jq -r '.minted // .supply // .count')
[ "$ACCESS" = "true" ] || die "bob does not hold access to seed 1 (got '$ACCESS')"
[ "$BOB_P1" -lt "$BOB_P0" ]     || die "buyer photons did not decrease ($BOB_P0 -> $BOB_P1)"
[ "$ALICE_P1" -gt "$ALICE_P0" ] || die "producer photons did not increase ($ALICE_P0 -> $ALICE_P1)"
[ "$SUP2" = "3" ] || die "photon supply changed on a sale ($SUP2) — marketplace must not mint"
ok "access granted; buyer $BOB_P0->$BOB_P1, producer $ALICE_P0->$ALICE_P1, supply still $SUP2"

say "LOOP COMPLETE — seed->mint->attest->outcome->reputation->sale->income all proven"
