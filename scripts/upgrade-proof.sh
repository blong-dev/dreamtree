#!/usr/bin/env bash
#
# Proves x/upgrade is wired and governable: a gov proposal schedules a software
# upgrade plan (far-future height, so it schedules without halting the devnet).
# propose -> vote -> execute -> `q upgrade plan` shows the plan.
set -euo pipefail

BIN=$(command -v dreamtreed || echo "$(go env GOPATH)/bin/dreamtreed")
HOME_DIR=${DT_UP_HOME:-/tmp/claude-1000/-home-b/99adc05f-ea53-4822-9a7c-49d0a182ddbf/scratchpad/dt-up}
CHAIN=dreamtree-up-1
KR=(--keyring-backend test --home "$HOME_DIR")
TX=(--chain-id "$CHAIN" "${KR[@]}" --gas 3000000 --yes -o json)
Q=(--home "$HOME_DIR" -o json)

ok()  { printf '\033[1;32mPASS\033[0m %s\n' "$*"; }
die() { printf '\033[1;31mFAIL\033[0m %s\n' "$*"; exit 1; }
wait_tx() { local h=$1 i out; for i in $(seq 1 30); do if out=$("$BIN" q tx "$h" "${Q[@]}" 2>/dev/null); then [ "$(echo "$out"|jq -r .code)" = "0" ] || die "tx $h: $(echo "$out"|jq -r .raw_log)"; return; fi; sleep 1; done; die "tx $h never landed"; }
send() { local out h; out=$("$BIN" tx "$@" "${TX[@]}"); h=$(echo "$out"|jq -r .txhash); wait_tx "$h" >&2; echo "$h"; }

rm -rf "$HOME_DIR"; mkdir -p "$HOME_DIR"
"$BIN" init up --chain-id "$CHAIN" --default-denom dtvp --home "$HOME_DIR" >/dev/null 2>&1
"$BIN" keys add alice "${KR[@]}" >/dev/null 2>&1
ALICE=$("$BIN" keys show alice -a "${KR[@]}")
"$BIN" genesis add-genesis-account alice 1000000000dtvp "${KR[@]}" >/dev/null
"$BIN" genesis gentx alice 500000000dtvp --chain-id "$CHAIN" "${KR[@]}" >/dev/null 2>&1
"$BIN" genesis collect-gentxs --home "$HOME_DIR" >/dev/null 2>&1
python3 - "$HOME_DIR/config/genesis.json" <<'PY'
import json, sys
g = sys.argv[1]; d = json.load(open(g))
gp = d['app_state']['gov']['params']
gp['min_deposit'] = [{"denom": "dtvp", "amount": "1"}]
gp['expedited_min_deposit'] = [{"denom": "dtvp", "amount": "1"}]
gp['max_deposit_period'] = "10s"; gp['voting_period'] = "6s"; gp['expedited_voting_period'] = "5s"
json.dump(d, open(g, 'w'), indent=1)
PY
sed -i 's/^timeout_commit = .*/timeout_commit = "1s"/' "$HOME_DIR/config/config.toml"

"$BIN" start --home "$HOME_DIR" --minimum-gas-prices 0dtvp >"$HOME_DIR/node.log" 2>&1 &
NODE=$!; trap 'kill $NODE 2>/dev/null || true' EXIT
for i in $(seq 1 40); do h=$("$BIN" status "${Q[@]}" 2>/dev/null|jq -r '.sync_info.latest_block_height' 2>/dev/null||echo 0); [ "${h:-0}" -ge 1 ] 2>/dev/null && break; sleep 1; done
[ "${h:-0}" -ge 1 ] || die "node never produced a block"

# No plan before governance acts.
PLAN0=$("$BIN" q upgrade plan "${Q[@]}" 2>/dev/null | jq -r '.plan.name // "none"')
ok "upgrade module live; plan before = $PLAN0"

GOV=$("$BIN" q auth module-account gov "${Q[@]}" | jq -r '.account.value.address // .account.base_account.address // .account.address')
TARGET=$(( ${h:-1} + 100000000 ))   # far future — schedule only, never triggers here
cat > "$HOME_DIR/up.json" <<JSON
{
  "messages": [{
    "@type": "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
    "authority": "$GOV",
    "plan": { "name": "v2", "height": "$TARGET", "info": "governance-scheduled upgrade proof" }
  }],
  "metadata": "ipfs://none",
  "deposit": "1dtvp",
  "title": "Schedule upgrade v2",
  "summary": "upgrade proof: gov schedules a software upgrade plan"
}
JSON

send gov submit-proposal "$HOME_DIR/up.json" --from alice >/dev/null
PID=$("$BIN" q gov proposals "${Q[@]}" | jq -r '.proposals[-1].id')
send gov vote "$PID" yes --from alice >/dev/null
ok "proposal $PID submitted + voted; waiting out voting period"

for i in $(seq 1 20); do
  st=$("$BIN" q gov proposal "$PID" "${Q[@]}" | jq -r '.proposal.status')
  [ "$st" = "PROPOSAL_STATUS_PASSED" ] || [ "$st" = "PROPOSAL_STATUS_REJECTED" ] || [ "$st" = "PROPOSAL_STATUS_FAILED" ] && break
  sleep 1
done
[ "$st" = "PROPOSAL_STATUS_PASSED" ] || die "proposal ended $st, expected PASSED"

PLAN=$("$BIN" q upgrade plan "${Q[@]}" | jq -r '.plan.name')
PLAN_H=$("$BIN" q upgrade plan "${Q[@]}" | jq -r '.plan.height')
[ "$PLAN" = "v2" ] || die "no upgrade plan scheduled (got '$PLAN')"
ok "upgrade plan scheduled by governance: name=$PLAN height=$PLAN_H"

echo; echo "UPGRADE PROOF COMPLETE — gov can schedule coordinated upgrades."
