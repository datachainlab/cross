#!/bin/bash
set -e

# Ensure jq is installed
if [[ ! -x "$(which jq)" ]]; then
  echo "jq (a tool for parsing json in the command line) is required..."
  echo "https://stedolan.github.io/jq/download/"
  exit 1
fi

# NODE_URL
CO_NODE=tcp://localhost:26657
TRAIN_NODE=tcp://localhost:26557
HOTEL_NODE=tcp://localhost:26457
CO_CHAIN=ibc0
TRAIN_CHAIN=ibc1
HOTEL_CHAIN=ibc2

# TODO Get account address from node
ACC0=cosmos1wl50tfnq9sg9qq07cn88ny8q6zwsqgjhygpucs

# Get contract ops from each node
${NODE_CLI} query contract call --from ${ACC0} train reserve --chain-id ${TRAIN_CHAIN} --save ./train.json
${NODE_CLI} query contract call --from ${ACC0} hotel reserve --chain-id ${HOTEL_CHAIN} --save ./hotel.json

# TODO set some options correctly (timeout, nonce)
# Compose contracts
${NODE_CLI} tx cross create --from ${ACC0} --chain-id ${CO_CHAIN} \
    --contract ./train.json --channel mychan:myport \
    --contract ./hotel.json --channel mychan:myport \
    10 99
