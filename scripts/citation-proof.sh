#!/usr/bin/env bash
#
# Proves creation-credit-forward as a value signal: a source work's value is
# inflated by the success of the works built on it. Two identical source works,
# A and C. A is cited (used_by) by a HIGH-value work B; C is cited by a LOW-value
# work. A must end up worth more than C — the only difference is who built on them.
set -euo pipefail

BIN=$(command -v dreamtreed || echo "$(go env GOPATH)/bin/dreamtreed")
HOME_DIR=${DT_CITE_HOME:-/tmp/claude-1000/-home-b/99adc05f-ea53-4822-9a7c-49d0a182ddbf/scratchpad/dt-cite}
CHAIN=dreamtree-cite-1
KR=(--keyring-backend test --home "$HOME_DIR")
TX=(--chain-id "$CHAIN" "${KR[@]}" --gas 3000000 --yes -o json)
Q=(--home "$HOME_DIR" -o json)

ok()  { printf '\033[1;32mPASS\033[0m %s\n' "$*"; }
die() { printf '\033[1;31mFAIL\033[0m %s\n' "$*"; exit 1; }
wait_tx(){ local h=$1 i out; for i in $(seq 1 30); do if out=$("$BIN" q tx "$h" "${Q[@]}" 2>/dev/null); then [ "$(echo "$out"|jq -r .code)" = "0" ] || die "tx $h: $(echo "$out"|jq -r .raw_log)"; return; fi; sleep 1; done; die "tx $h never landed"; }
att(){ local from=$1; shift; local h; h=$("$BIN" tx attest attest "$@" --from "$from" "${TX[@]}"|jq -r .txhash); wait_tx "$h"; }
V(){ "$BIN" q attest work-value "$1" "${Q[@]}" | jq -r '.value'; }

rm -rf "$HOME_DIR"; mkdir -p "$HOME_DIR"
"$BIN" init cite --chain-id "$CHAIN" --default-denom uphoton --home "$HOME_DIR" >/dev/null 2>&1
for k in alice bob carol dave; do "$BIN" keys add "$k" "${KR[@]}" >/dev/null 2>&1; done
"$BIN" genesis add-genesis-account alice 1000000000uphoton "${KR[@]}" >/dev/null
for k in bob carol dave; do "$BIN" genesis add-genesis-account "$k" 1000000uphoton "${KR[@]}" >/dev/null; done
"$BIN" genesis gentx alice 500000000uphoton --chain-id "$CHAIN" "${KR[@]}" >/dev/null 2>&1
"$BIN" genesis collect-gentxs --home "$HOME_DIR" >/dev/null 2>&1
sed -i 's/^timeout_commit = .*/timeout_commit = "1s"/' "$HOME_DIR/config/config.toml"
"$BIN" start --home "$HOME_DIR" --minimum-gas-prices 0uphoton >"$HOME_DIR/node.log" 2>&1 &
NODE=$!; trap 'kill $NODE 2>/dev/null || true' EXIT
for i in $(seq 1 40); do h=$("$BIN" status "${Q[@]}" 2>/dev/null|jq -r '.sync_info.latest_block_height' 2>/dev/null||echo 0); [ "${h:-0}" -ge 1 ] 2>/dev/null && break; sleep 1; done
[ "${h:-0}" -ge 1 ] || die "node never produced a block"

D="sciences/x"
# B = a HIGH-value work: three independent attestors back it.
att alice work:B origin      --domain "$D" --specificity-bps 10000
att bob   work:B rigor       --domain "$D" --specificity-bps 10000
att carol work:B replication --domain "$D" --specificity-bps 10000
# lo = a LOW-value work: nothing attests it (V ~ 0).
VB=$(V work:B); VLO=$(V work:lo)
ok "high-value work B: V=$VB   |   low work: V=$VLO"

# Two identical sources: one origin attestation each.
att alice work:A origin --domain "$D" --specificity-bps 10000
att alice work:C origin --domain "$D" --specificity-bps 10000
A0=$(V work:A); C0=$(V work:C)
[ "$A0" = "$C0" ] || die "sources not identical at baseline: A=$A0 C=$C0"
ok "identical sources A and C at baseline: V=$A0"

# The only difference: A is built on by the HIGH-value B; C by the LOW work.
att dave work:A use --domain "$D" --specificity-bps 10000 --used-by work:B
att dave work:C use --domain "$D" --specificity-bps 10000 --used-by work:lo
A1=$(V work:A); C1=$(V work:C)
ok "after citations: V(A cited-by-B)=$A1   V(C cited-by-lo)=$C1"

awk -v a="$A1" -v c="$C1" 'BEGIN{exit !(a+0 > c+0)}' \
  || die "source cited by a high-value work ($A1) is not worth more than one cited by a low work ($C1)"
awk -v a="$A1" -v a0="$A0" 'BEGIN{exit !(a+0 > a0+0)}' \
  || die "being built on did not raise the source's value"
ok "creation-credit-forward holds: A ($A1) > C ($C1); a source's value rises with the success of what builds on it"

echo; echo "CITATION-VALUE PROOF COMPLETE — success flows to sources as signal, no compensation."
