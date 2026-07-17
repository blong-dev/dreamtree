#!/usr/bin/env bash
#
# upgrade-1 REHEARSAL (docs/specs/upgrade-1.md, DT-20): the full governed
# in-place upgrade loop on a throwaway devnet, starting from the PRE-rider
# binary so the migration path genuinely exercises:
#
#   old binary devnet -> MsgSoftwareUpgrade(upgrade-1, h) -> vote -> halt at h
#   -> swap binary -> handler applies -> verify riders R2/R3/R4/R5.
#
# Usage: OLD_BIN=... NEW_BIN=... scripts/upgrade1-rehearsal.sh
set -euo pipefail

OLD_BIN=${OLD_BIN:?path to pre-upgrade dreamtreed}
NEW_BIN=${NEW_BIN:?path to post-upgrade dreamtreed}
HOME_DIR=${DT_R_HOME:-${TMPDIR:-/tmp}/dt-up1-rehearsal}
CHAIN=dreamtree-up1-r
KR=(--keyring-backend test --home "$HOME_DIR")
TX=(--chain-id "$CHAIN" "${KR[@]}" --gas 3000000 --yes -o json)
Q=(--home "$HOME_DIR" -o json)

ok()  { printf '\033[1;32mPASS\033[0m %s\n' "$*"; }
die() { printf '\033[1;31mFAIL\033[0m %s\n' "$*"; exit 1; }
BIN=""
wait_tx() { local h=$1 i out; for i in $(seq 1 30); do if out=$("$BIN" q tx "$h" "${Q[@]}" 2>/dev/null); then [ "$(echo "$out"|jq -r .code)" = "0" ] || die "tx $h: $(echo "$out"|jq -r .raw_log)"; return; fi; sleep 1; done; die "tx $h never landed"; }
send() { local out h; out=$("$BIN" tx "$@" "${TX[@]}"); h=$(echo "$out"|jq -r .txhash); wait_tx "$h" >&2; echo "$h"; }
height() { "$BIN" status "${Q[@]}" 2>/dev/null | jq -r '.sync_info.latest_block_height' 2>/dev/null || echo 0; }

# ---- genesis on the OLD binary --------------------------------------------
BIN=$OLD_BIN
rm -rf "$HOME_DIR"; mkdir -p "$HOME_DIR"
"$BIN" init r --chain-id "$CHAIN" --default-denom uphoton --home "$HOME_DIR" >/dev/null 2>&1
"$BIN" keys add alice "${KR[@]}" >/dev/null 2>&1
"$BIN" keys add bob "${KR[@]}" >/dev/null 2>&1
ALICE=$("$BIN" keys show alice -a "${KR[@]}")
BOB=$("$BIN" keys show bob -a "${KR[@]}")
"$BIN" genesis add-genesis-account alice 1000000000uphoton "${KR[@]}" >/dev/null
"$BIN" genesis add-genesis-account bob 50000000uphoton "${KR[@]}" >/dev/null
"$BIN" genesis gentx alice 500000000uphoton --chain-id "$CHAIN" "${KR[@]}" >/dev/null 2>&1
"$BIN" genesis collect-gentxs --home "$HOME_DIR" >/dev/null 2>&1
python3 - "$HOME_DIR/config/genesis.json" <<'PY'
import json, sys
g = sys.argv[1]; d = json.load(open(g))
gp = d['app_state']['gov']['params']
gp['min_deposit'] = [{"denom": "uphoton", "amount": "1"}]
gp['expedited_min_deposit'] = [{"denom": "uphoton", "amount": "1"}]
gp['max_deposit_period'] = "10s"; gp['voting_period'] = "6s"; gp['expedited_voting_period'] = "5s"
json.dump(d, open(g, 'w'), indent=1)
PY
sed -i 's/^timeout_commit = .*/timeout_commit = "1s"/' "$HOME_DIR/config/config.toml"

"$BIN" start --home "$HOME_DIR" --minimum-gas-prices 0uphoton >"$HOME_DIR/node-old.log" 2>&1 &
NODE=$!; trap 'kill $NODE 2>/dev/null || true' EXIT
for i in $(seq 1 40); do [ "$(height)" -ge 1 ] 2>/dev/null && break; sleep 1; done
[ "$(height)" -ge 1 ] || die "old-binary devnet never produced a block"
ok "old-binary devnet up (h=$(height))"

# Old world: everyone has baseline reputation without any grant.
R_OLD=$("$BIN" q reputation reputation "$BOB" science "${Q[@]}" 2>/dev/null | jq -r '.reputation // .r // empty' || true)
echo "  old-world R(bob)=${R_OLD:-<n/a>} (baseline-for-all)"

# Pre-upgrade seeds alice will sell post-upgrade (producer = committer).
ROOT=$(head -c 32 /dev/urandom | sha256sum | cut -d" " -f1)
send seeds commit-batch "$ROOT" 2 2 record --from alice >/dev/null
ok "pre-upgrade: alice committed 2 record seeds (old world works)"

# ---- schedule upgrade-1 ----------------------------------------------------
GOV=$("$BIN" q auth module-account gov "${Q[@]}" | jq -r '.account.value.address // .account.base_account.address')
H=$(height); TARGET=$((H + 15))
cat > "$HOME_DIR/up1.json" <<EOF
{"messages":[{"@type":"/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade","authority":"$GOV","plan":{"name":"upgrade-1","height":"$TARGET","info":"upgrade-1 rehearsal"}}],
 "metadata":"upgrade-1","deposit":"1uphoton","title":"upgrade-1","summary":"riders R2-R5","expedited":false}
EOF
send gov submit-proposal "$HOME_DIR/up1.json" --from alice >/dev/null
PID=$("$BIN" q gov proposals "${Q[@]}" | jq -r '.proposals[-1].id')
send gov vote "$PID" yes --from alice >/dev/null
for i in $(seq 1 20); do
  ST=$("$BIN" q gov proposal "$PID" "${Q[@]}" | jq -r '.proposal.status')
  [ "$ST" = "PROPOSAL_STATUS_PASSED" ] && break; sleep 1
done
[ "$ST" = "PROPOSAL_STATUS_PASSED" ] || die "proposal status: $ST"
PLAN=$("$BIN" q upgrade plan "${Q[@]}" 2>/dev/null | jq -r '.plan.name' || echo "")
[ "$PLAN" = "upgrade-1" ] || die "plan not scheduled: '$PLAN'"
ok "upgrade-1 scheduled at height $TARGET (proposal $PID passed)"

# ---- halt ------------------------------------------------------------------
# CometBFT does NOT exit at the upgrade height: FinalizeBlock panics, consensus
# halts (CONSENSUS FAILURE) and the process hangs. The operator stops it — on
# m3 that's `systemctl stop dreamtree` after the journal shows UPGRADE NEEDED.
for i in $(seq 1 60); do grep -q 'UPGRADE "upgrade-1" NEEDED' "$HOME_DIR/node-old.log" && break; sleep 1; done
grep -q 'UPGRADE "upgrade-1" NEEDED' "$HOME_DIR/node-old.log" || die "old binary never hit the upgrade height"
kill $NODE 2>/dev/null || true
for i in $(seq 1 30); do kill -0 $NODE 2>/dev/null || break; sleep 1; done
kill -0 $NODE 2>/dev/null && { kill -9 $NODE 2>/dev/null || true; sleep 1; }
ok "old binary halted consensus: UPGRADE \"upgrade-1\" NEEDED at $TARGET (operator stop)"

# ---- swap + apply ----------------------------------------------------------
BIN=$NEW_BIN
"$BIN" start --home "$HOME_DIR" --minimum-gas-prices 0uphoton >"$HOME_DIR/node-new.log" 2>&1 &
NODE=$!
for i in $(seq 1 60); do [ "$(height)" -gt "$TARGET" ] 2>/dev/null && break; sleep 1; done
[ "$(height)" -gt "$TARGET" ] || die "new binary did not resume past $TARGET"
grep -q 'applying upgrade "upgrade-1"' "$HOME_DIR/node-new.log" || die "handler never ran"
ok "new binary applied upgrade-1 and resumed (h=$(height))"

# ---- rider verification ----------------------------------------------------
# R5: migration filled e_cap_mult on stored params.
ECM=$("$BIN" q reputation params "${Q[@]}" | jq -r '.params.e_cap_mult')
[ "$ECM" = "2.0" ] || die "e_cap_mult not migrated: '$ECM'"
ok "R5: e_cap_mult=2.0 on migrated params"

# R2: reputation starts at zero; a governed grant confers baseline.
R_PRE=$("$BIN" q reputation reputation "$BOB" science "${Q[@]}" | jq -r '.reputation // .r')
cat > "$HOME_DIR/verify.json" <<EOF
{"messages":[{"@type":"/dreamtree.reputation.v1.MsgSetVerified","authority":"$GOV","address":"$BOB","verified":true}],
 "metadata":"verify bob","deposit":"1uphoton","title":"verify bob","summary":"grant baseline","expedited":false}
EOF
send gov submit-proposal "$HOME_DIR/verify.json" --from alice >/dev/null
PID2=$("$BIN" q gov proposals "${Q[@]}" | jq -r '.proposals[-1].id')
send gov vote "$PID2" yes --from alice >/dev/null
for i in $(seq 1 20); do
  ST2=$("$BIN" q gov proposal "$PID2" "${Q[@]}" | jq -r '.proposal.status')
  [ "$ST2" = "PROPOSAL_STATUS_PASSED" ] && break; sleep 1
done
[ "$ST2" = "PROPOSAL_STATUS_PASSED" ] || die "MsgSetVerified proposal: $ST2"
R_POST=$("$BIN" q reputation reputation "$BOB" science "${Q[@]}" | jq -r '.reputation // .r')
echo "  R(bob): pre-grant=$R_PRE post-grant=$R_POST"
case "$R_PRE" in 0|0.0|0.000000|"0.000000000000000000") : ;; *) die "R2: pre-grant reputation not zero: $R_PRE";; esac
case "$R_POST" in 1|1.0|1.000000|"1.000000000000000000") : ;; *) die "R2: post-grant reputation not baseline: $R_POST";; esac
ok "R2: standing 0 -> governed MsgSetVerified -> baseline 1.0 (proposal $PID2)"

# R4: bob buys alice's pre-upgrade seed at the CONSTANT price (1 photon/seed/day).
B0=$("$BIN" q bank balance "$BOB" uphoton "${Q[@]}" | jq -r '.balance.amount')
send licenses purchase 1 --from bob >/dev/null
B1=$("$BIN" q bank balance "$BOB" uphoton "${Q[@]}" | jq -r '.balance.amount')
[ $((B0 - B1)) -eq 1000000 ] || die "purchase cost $((B0-B1)) uphoton, want 1000000 (1 photon x 1 seed x 1 day)"
ok "R4: purchase of 1 seed cost exactly 1 photon (constant price, no price table)"

# R3: a novel (never-whitelisted) kind mints.
SUP0=$("$BIN" q bank total-supply-of uphoton "${Q[@]}" | jq -r '.amount.amount')
ROOT3=$(head -c 32 /dev/urandom | sha256sum | cut -d" " -f1)
send seeds commit-batch "$ROOT3" 3 3 novel.kind --from alice >/dev/null
SUP1=$("$BIN" q bank total-supply-of uphoton "${Q[@]}" | jq -r '.amount.amount')
[ $((SUP1 - SUP0)) -eq 3000000 ] || die "novel kind minted $((SUP1-SUP0)) uphoton, want 3000000"
ok "R3: novel kind minted 3 photons (all kinds mint)"

echo
ok "upgrade-1 rehearsal COMPLETE: halt -> swap -> migrate -> riders live"
