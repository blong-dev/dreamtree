#!/usr/bin/env bash
#
# Build the fresh-genesis launch state for the dreamtree chain (chain-id:
# dreamtree — no suffix). Produces a validated genesis.json + node config;
# DOES NOT start the node (you review and start deliberately). Idempotent:
# refuses to clobber an initialized home.
#
# Tokenomics (docs/specs/seed-atom-conformance.md): the photon is the chain's
# native asset. Base denom uphoton (1 photon = 10^6 uphoton); supply at genesis
# = the anchored corpus (one photon per distinct atom, exported by
# export-genesis-corpus.py), minted to dt-as-first-storer, who bonds the
# validator with it. dtvp does not exist. No slashing modules are wired, so
# nothing can burn bonded photons — the photons = seeds peg holds structurally.
#
# Required env:
#   DT_TREASURY   dream1… address of the treasury key (storer-reward + marketplace recipient)
#   CORPUS_JSON   path to corpus.json from export-genesis-corpus.py
# Optional env (with defaults):
#   CHAIN_ID          (dreamtree)
#   DT_VALIDATOR_KEY  keyring name of the validator operator key (dt-validator) — MUST be in the keyring
#   KEYRING           keyring backend (test) — use `os`/`file` for a real validator key
#   HOME_DIR          node home ($HOME/.dreamtree) — a NEW dir, does not touch any existing chain
#   BOND_FRACTION     fraction of genesis photons the validator bonds (0.70)
set -euo pipefail

BIN=${DREAMTREED_BIN:-$(command -v dreamtreed || echo "$HOME/go/bin/dreamtreed")}
CHAIN_ID=${CHAIN_ID:-dreamtree}
DT_VALIDATOR_KEY=${DT_VALIDATOR_KEY:-dt-validator}
KEYRING=${KEYRING:-test}
HOME_DIR=${HOME_DIR:-$HOME/.dreamtree}
BOND_FRACTION=${BOND_FRACTION:-0.70}
: "${DT_TREASURY:?set DT_TREASURY to the dt-treasury dream1… address}"
: "${CORPUS_JSON:?set CORPUS_JSON to the corpus.json from export-genesis-corpus.py}"
[ -f "$CORPUS_JSON" ] || { echo "CORPUS_JSON not found: $CORPUS_JSON"; exit 1; }

KR=(--keyring-backend "$KEYRING" --home "$HOME_DIR")
die(){ printf '\033[1;31mFAIL\033[0m %s\n' "$*"; exit 1; }
ok(){ printf '\033[1;32mok\033[0m %s\n' "$*"; }

# Guard: the validator key must already exist in the keyring on THIS host.
VAL_ADDR=$("$BIN" keys show "$DT_VALIDATOR_KEY" -a "${KR[@]}" 2>/dev/null) \
  || die "validator key '$DT_VALIDATOR_KEY' not in the '$KEYRING' keyring at $HOME_DIR — import it (keys add $DT_VALIDATOR_KEY --recover) first"
case "$DT_TREASURY" in dream1*) ;; *) die "DT_TREASURY '$DT_TREASURY' is not a dream1… address" ;; esac
ok "validator $DT_VALIDATOR_KEY = $VAL_ADDR"
ok "treasury          = $DT_TREASURY"

# Corpus arithmetic: supply (uphoton) and bond size.
read -r MINTED UPHOTON BOND < <(python3 - "$CORPUS_JSON" "$BOND_FRACTION" <<'PY'
import json, sys
c = json.load(open(sys.argv[1]))
minted = int(c["minted"])
up = minted * 1_000_000
bond = int(up * float(sys.argv[2]))
print(minted, up, bond)
PY
)
[ "$MINTED" -gt 0 ] || die "corpus carries zero photons — nothing to bond"
ok "corpus: $MINTED photons ($UPHOTON uphoton); validator bonds $BOND uphoton (~${BOND_FRACTION})"

# The keyring lives inside HOME_DIR, so the dir may already exist — only refuse
# if a chain is already initialized here (don't clobber a live/prepared chain).
[ -e "$HOME_DIR/config/genesis.json" ] && die "$HOME_DIR already holds an initialized chain (config/genesis.json) — move it aside first"
"$BIN" init dreamtree --chain-id "$CHAIN_ID" --default-denom uphoton --home "$HOME_DIR" >/dev/null 2>&1
ok "initialized $CHAIN_ID at $HOME_DIR"

# The validator (dt-as-first-storer) receives the entire genesis corpus supply.
"$BIN" genesis add-genesis-account "$DT_VALIDATOR_KEY" "${UPHOTON}uphoton" "${KR[@]}" >/dev/null
ok "genesis account funded with corpus supply"

# Patch app_state BEFORE gentx: corpus batches into x/seeds, minted count into
# x/photons, recipients, gov deposits in uphoton, photon denom metadata.
python3 - "$HOME_DIR/config/genesis.json" "$CORPUS_JSON" "$DT_TREASURY" <<'PY'
import json, sys
g, corpus_path, treasury = sys.argv[1], sys.argv[2], sys.argv[3]
d = json.load(open(g))
c = json.load(open(corpus_path))

# x/seeds: the anchored corpus (leaf model).
d['app_state']['seeds']['batches'] = c['batches']
d['app_state']['seeds']['next_id'] = c['next_id']
d['app_state']['seeds']['next_batch_id'] = c['next_batch_id']

# x/photons: minted counts photons (= distinct atoms); recipients.
d['app_state']['photons']['minted'] = c['minted']
d['app_state']['photons']['params']['storer_reward_recipient'] = treasury
d['app_state']['licenses']['params']['treasury_recipient'] = treasury

# gov deposits in uphoton (10,000 / 50,000 photons).
gp = d['app_state']['gov']['params']
gp['min_deposit'] = [{"denom": "uphoton", "amount": "10000000000"}]
gp['expedited_min_deposit'] = [{"denom": "uphoton", "amount": "50000000000"}]

# bank: photon display metadata (base uphoton, exponent 6).
d['app_state']['bank']['denom_metadata'] = [{
    "description": "The dreamtree photon — the meter-pegged native asset (photons = seeds = distinct atoms).",
    "denom_units": [
        {"denom": "uphoton", "exponent": 0, "aliases": ["microphoton"]},
        {"denom": "photon", "exponent": 6, "aliases": []},
    ],
    "base": "uphoton", "display": "photon", "name": "Photon", "symbol": "PHOTON",
}]

# sanity: dtvp must not exist anywhere in genesis.
assert 'dtvp' not in json.dumps(d), "dtvp found in genesis — the fork must not survive"
json.dump(d, open(g, 'w'), indent=1)
print("patched: %d corpus batches, minted=%s, storer=%s treasury=%s gov-denom=uphoton"
      % (len(c['batches']), c['minted'], treasury, treasury))
PY

# Bond the validator with corpus photons.
"$BIN" genesis gentx "$DT_VALIDATOR_KEY" "${BOND}uphoton" --chain-id "$CHAIN_ID" "${KR[@]}" >/dev/null 2>&1
"$BIN" genesis collect-gentxs --home "$HOME_DIR" >/dev/null 2>&1
ok "validator bonded ${BOND}uphoton"

"$BIN" genesis validate-genesis --home "$HOME_DIR" >/dev/null 2>&1 || die "genesis failed validation"
ok "genesis validated"

echo
echo "=== dreamtree genesis ready (node NOT started) ==="
echo "  chain-id:   $CHAIN_ID"
echo "  home:       $HOME_DIR"
echo "  validator:  $VAL_ADDR (bonded ${BOND}/${UPHOTON} uphoton)"
echo "  treasury:   $DT_TREASURY (storer-reward + marketplace)"
echo "  photons:    $MINTED (= corpus atoms; peg holds from block 0)"
echo "  genesis:    $HOME_DIR/config/genesis.json"
echo
echo "Review it, then start when ready:"
echo "  dreamtreed start --home $HOME_DIR --minimum-gas-prices 0uphoton"
