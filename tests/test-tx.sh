#!/bin/bash
set -ex

# Ensure jq is installed
if [[ ! -x "$(which jq)" ]]; then
  echo "jq (a tool for parsing json in the command line) is required..."
  echo "https://stedolan.github.io/jq/download/"
  exit 1
fi

RELAYER_CMD="${RELAYER_CLI} --home $(pwd)/.relayer"
NODE_CLI=$(pwd)/../build/simappcli
# NODE_URL
CO_NODE=tcp://localhost:26657
TRAIN_NODE=tcp://localhost:26557
HOTEL_NODE=tcp://localhost:26457
CO_CHAIN=ibc0
TRAIN_CHAIN=ibc1
HOTEL_CHAIN=ibc2

ACC0=acc0

# Get contract ops from each nodes
${NODE_CLI} query --home ./data/ibc0/n0/simappcli contract call --node ${TRAIN_NODE} --from ${ACC0} --keyring-backend=test train reserve 0x00000001 --chain-id ${TRAIN_CHAIN} --save ./data/train.json
${NODE_CLI} query --home ./data/ibc0/n0/simappcli contract call --node ${HOTEL_NODE} --from ${ACC0} --keyring-backend=test hotel reserve 0x00000002 --chain-id ${HOTEL_CHAIN} --save ./data/hotel.json

SOURCE01_CHAN=$(${RELAYER_CMD} paths show path01 --json | jq -r '.src."channel-id"')
SOURCE01_PORT=$(${RELAYER_CMD} paths show path01 --json | jq -r '.src."port-id"')
SOURCE02_CHAN=$(${RELAYER_CMD} paths show path02 --json | jq -r '.src."channel-id"')
SOURCE02_PORT=$(${RELAYER_CMD} paths show path02 --json | jq -r '.src."port-id"')

LATEST_HEIGHT=$(${NODE_CLI} --home ./data/ibc0/n0/simappcli status | jq -r '.sync_info.latest_block_height')

# Compose contracts
${NODE_CLI} tx --home ./data/ibc0/n0/simappcli cross create --from ${ACC0} --keyring-backend=test --chain-id ${CO_CHAIN} --yes \
    --contract ./data/train.json --channel ${SOURCE01_CHAN}:${SOURCE01_PORT} \
    --contract ./data/hotel.json --channel ${SOURCE02_CHAN}:${SOURCE02_PORT} \
    $((${LATEST_HEIGHT}+100)) 0
