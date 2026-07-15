#!/usr/bin/env bash
#
# Dev devnet init. Strict tokenomics (2026-07-10): the validator bonds a
# native asset uphoton (micro-photon; 1 photon = 10^6 uphoton). Extra dev supply
# is pre-funded for dev convenience; real photons mint per data-seed (leaf model).
# Photon is never the staking/gas token.

rm -rf $HOME/.dreamtreed
DREAMTREED_BIN=$(which dreamtreed)
if [ -z "$DREAMTREED_BIN" ]; then
    GOBIN=$(go env GOPATH)/bin
    DREAMTREED_BIN=$(which $GOBIN/dreamtreed)
fi

if [ -z "$DREAMTREED_BIN" ]; then
    echo "please verify dreamtreed is installed"
    exit 1
fi

$DREAMTREED_BIN config set client chain-id demo
$DREAMTREED_BIN config set client keyring-backend test
$DREAMTREED_BIN keys add alice
$DREAMTREED_BIN keys add bob
# bond denom = uphoton (the photon IS the native asset — seed-atom conformance).
$DREAMTREED_BIN init test --chain-id dreamtree-devnet-1 --default-denom uphoton
$DREAMTREED_BIN genesis add-genesis-account alice 1000000000uphoton --keyring-backend test
$DREAMTREED_BIN genesis add-genesis-account bob 1000000uphoton --keyring-backend test
$DREAMTREED_BIN genesis gentx alice 500000000uphoton --chain-id dreamtree-devnet-1
$DREAMTREED_BIN genesis collect-gentxs

# route the per-seed ingestion photon to alice (dreamtree stand-in) for dev.
ALICE=$($DREAMTREED_BIN keys show alice -a --keyring-backend test)
GEN=$HOME/.dreamtreed/config/genesis.json
python3 - "$GEN" "$ALICE" <<'PY'
import json, sys
g, alice = sys.argv[1], sys.argv[2]
d = json.load(open(g))
d['app_state']['photons']['params']['storer_reward_recipient'] = alice
# gov deposits must be in uphoton — the default min_deposit denom is "stake",
# which does not exist on this chain, so governance would be unusable.
gp = d['app_state']['gov']['params']
gp['min_deposit'] = [{"denom": "uphoton", "amount": "10000000"}]
gp['expedited_min_deposit'] = [{"denom": "uphoton", "amount": "50000000"}]
json.dump(d, open(g, 'w'), indent=1)
PY
