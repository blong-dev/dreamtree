#!/usr/bin/env bash

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

# configure dreamtreed
$DREAMTREED_BIN config set client chain-id demo
$DREAMTREED_BIN config set client keyring-backend test
$DREAMTREED_BIN keys add alice
$DREAMTREED_BIN keys add bob
$DREAMTREED_BIN init test --chain-id dreamtree-devnet-1 --default-denom photon
# update genesis
$DREAMTREED_BIN genesis add-genesis-account alice 1000000000photon --keyring-backend test
$DREAMTREED_BIN genesis add-genesis-account bob 1000000photon --keyring-backend test
# create default validator
$DREAMTREED_BIN genesis gentx alice 500000000photon --chain-id dreamtree-devnet-1
$DREAMTREED_BIN genesis collect-gentxs
