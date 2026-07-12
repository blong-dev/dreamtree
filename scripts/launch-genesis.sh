#!/usr/bin/env bash
#
# Build the fresh-genesis launch state for the real dreamtree chain. Produces a
# validated genesis.json + node config; DOES NOT start the node (you review and
# start deliberately). Idempotent: rebuilds a clean home each run.
#
# Strict tokenomics: dtvp bonds validators (permissioning power, no economic
# value); photon starts at supply 0 and mints only per data-seed. dreamtree
# (dt-treasury) is both the per-seed storer-reward recipient and the marketplace
# treasury. Production reputation params (real review windows) — NOT test values.
#
# Required env:
#   DT_TREASURY   dream1… address of the cold treasury key (storer + marketplace recipient)
# Optional env (with defaults):
#   CHAIN_ID          (dreamtree-1)
#   DT_VALIDATOR_KEY  keyring name of the validator operator key (dt-validator) — MUST be in the keyring
#   KEYRING           keyring backend (test) — use `os`/`file` for a real validator key
#   HOME_DIR          node home ($HOME/.dreamtree-mainnet) — a NEW dir, does not touch any existing chain
#   DTVP_TOTAL        genesis dtvp minted to the validator (1000000000)
#   DTVP_BOND         dtvp the validator bonds via gentx (700000000; keep some liquid for gov)
set -euo pipefail

BIN=${DREAMTREED_BIN:-$(command -v dreamtreed || echo "$(go env GOPATH)/bin/dreamtreed")}
CHAIN_ID=${CHAIN_ID:-dreamtree-1}
DT_VALIDATOR_KEY=${DT_VALIDATOR_KEY:-dt-validator}
KEYRING=${KEYRING:-test}
HOME_DIR=${HOME_DIR:-$HOME/.dreamtree-mainnet}
DTVP_TOTAL=${DTVP_TOTAL:-1000000000}
DTVP_BOND=${DTVP_BOND:-700000000}
: "${DT_TREASURY:?set DT_TREASURY to the dt-treasury dream1… address}"

KR=(--keyring-backend "$KEYRING" --home "$HOME_DIR")
die(){ printf '\033[1;31mFAIL\033[0m %s\n' "$*"; exit 1; }
ok(){ printf '\033[1;32mok\033[0m %s\n' "$*"; }

# Guard: the validator key must already exist in the keyring on THIS host.
VAL_ADDR=$("$BIN" keys show "$DT_VALIDATOR_KEY" -a "${KR[@]}" 2>/dev/null) \
  || die "validator key '$DT_VALIDATOR_KEY' not in the '$KEYRING' keyring at $HOME_DIR — import it (keys add $DT_VALIDATOR_KEY --recover) first"
"$BIN" keys parse "$DT_TREASURY" >/dev/null 2>&1 || \
  case "$DT_TREASURY" in dream1*) ;; *) die "DT_TREASURY '$DT_TREASURY' is not a dream1… address" ;; esac
ok "validator $DT_VALIDATOR_KEY = $VAL_ADDR"
ok "treasury          = $DT_TREASURY"

# The keyring lives inside HOME_DIR, so the dir may already exist — only refuse
# if a chain is already initialized here (don't clobber a live/prepared chain).
[ -e "$HOME_DIR/config/genesis.json" ] && die "$HOME_DIR already holds an initialized chain (config/genesis.json) — move it aside first"
"$BIN" init dreamtree --chain-id "$CHAIN_ID" --default-denom dtvp --home "$HOME_DIR" >/dev/null 2>&1
ok "initialized $CHAIN_ID at $HOME_DIR"

# The validator gets dtvp to bond; nobody gets photon (supply starts at 0).
"$BIN" genesis add-genesis-account "$DT_VALIDATOR_KEY" "${DTVP_TOTAL}dtvp" "${KR[@]}" >/dev/null
"$BIN" genesis gentx "$DT_VALIDATOR_KEY" "${DTVP_BOND}dtvp" --chain-id "$CHAIN_ID" "${KR[@]}" >/dev/null 2>&1
"$BIN" genesis collect-gentxs --home "$HOME_DIR" >/dev/null 2>&1
ok "validator bonded ${DTVP_BOND}dtvp"

# Patch app_state: dt-treasury as both recipients; gov deposits in dtvp; photon
# absent. Reputation params stay at production defaults (untouched).
python3 - "$HOME_DIR/config/genesis.json" "$DT_TREASURY" <<'PY'
import json, sys
g, treasury = sys.argv[1], sys.argv[2]
d = json.load(open(g))
d['app_state']['photons']['params']['storer_reward_recipient'] = treasury
d['app_state']['licenses']['params']['treasury_recipient'] = treasury
gp = d['app_state']['gov']['params']
gp['min_deposit'] = [{"denom": "dtvp", "amount": "10000000"}]
gp['expedited_min_deposit'] = [{"denom": "dtvp", "amount": "50000000"}]
# sanity: photon must not appear in any genesis balance (supply starts at 0)
for bal in d['app_state']['bank'].get('balances', []):
    for c in bal.get('coins', []):
        assert c['denom'] != 'photon', "photon must not be in genesis balances"
json.dump(d, open(g, 'w'), indent=1)
print("patched: storer=%s treasury=%s gov-denom=dtvp" % (treasury, treasury))
PY

"$BIN" genesis validate-genesis --home "$HOME_DIR" >/dev/null 2>&1 || die "genesis failed validation"
ok "genesis validated"

echo
echo "=== dreamtree-1 genesis ready (node NOT started) ==="
echo "  chain-id:   $CHAIN_ID"
echo "  home:       $HOME_DIR"
echo "  validator:  $VAL_ADDR (bonded ${DTVP_BOND}/${DTVP_TOTAL} dtvp)"
echo "  treasury:   $DT_TREASURY (storer-reward + marketplace)"
echo "  photon:     supply 0 (mints per seed only)"
echo "  genesis:    $HOME_DIR/config/genesis.json"
echo
echo "Review it, then start when ready:"
echo "  dreamtreed start --home $HOME_DIR --minimum-gas-prices 0dtvp"
