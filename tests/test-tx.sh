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
CO_HOME=./data/ibc0/n0/simappcli
TRAIN_HOME=./data/ibc1/n0/simappcli
HOTEL_HOME=./data/ibc2/n0/simappcli
WAIT_NEW_BLOCK=3s

ACC0=n0

# Get contract ops from each nodes
${NODE_CLI} query --home ${CO_HOME} contract call --node ${TRAIN_NODE} --from ${ACC0} --keyring-backend=test train reserve 0x00000001 --chain-id ${TRAIN_CHAIN} --save ./data/train.json
${NODE_CLI} query --home ${CO_HOME} contract call --node ${HOTEL_NODE} --from ${ACC0} --keyring-backend=test hotel reserve 0x00000002 --chain-id ${HOTEL_CHAIN} --save ./data/hotel.json

SRC01_CHAN=$(${RELAYER_CMD} paths show path01 --json | jq -r '.chains.src."channel-id"')
SRC01_PORT=$(${RELAYER_CMD} paths show path01 --json | jq -r '.chains.src."port-id"')
DST01_CHAN=$(${RELAYER_CMD} paths show path01 --json | jq -r '.chains.dst."channel-id"')
DST01_PORT=$(${RELAYER_CMD} paths show path01 --json | jq -r '.chains.dst."port-id"')

SRC02_CHAN=$(${RELAYER_CMD} paths show path02 --json | jq -r '.chains.src."channel-id"')
SRC02_PORT=$(${RELAYER_CMD} paths show path02 --json | jq -r '.chains.src."port-id"')
DST02_CHAN=$(${RELAYER_CMD} paths show path02 --json | jq -r '.chains.dst."channel-id"')
DST02_PORT=$(${RELAYER_CMD} paths show path02 --json | jq -r '.chains.dst."port-id"')

RELAYER0=$(${NODE_CLI} --home ${CO_HOME} --keyring-backend=test keys show ${ACC0} -a)
RELAYER1=$(${NODE_CLI} --home ${TRAIN_HOME} --keyring-backend=test keys show ${ACC0} -a)
RELAYER2=$(${NODE_CLI} --home ${HOTEL_HOME} --keyring-backend=test keys show ${ACC0} -a)

echo "==> (Re)starting the relayer"
PID=$(pgrep rly || echo "")
if [[ $PID != "" ]]; then
	pkill rly
fi

${RELAYER_CMD} start path01 &
${RELAYER_CMD} start path02 &

### Broadcast MsgInitiate
LATEST_HEIGHT=$(${NODE_CLI} --home ${CO_HOME} status | jq -r '.sync_info.latest_block_height')
TX_ID=$(${NODE_CLI} tx --home ./data/ibc0/n0/simappcli cross create --from ${ACC0} --keyring-backend=test --chain-id ${CO_CHAIN} --yes \
    --contract ./data/train.json --channel ${SRC01_CHAN}:${SRC01_PORT} \
    --contract ./data/hotel.json --channel ${SRC02_CHAN}:${SRC02_PORT} \
    $((${LATEST_HEIGHT}+100)) 0 | jq -r '.data')
###

sleep 20

### Ensure coordinator status is done
STATUS=$(${NODE_CLI} query --home ${CO_HOME} cross coordinator ${TX_ID} | jq -r '.completed')
if [ ${STATUS} = "true" ]; then
  echo "completed!"
else
  echo "failed"
  exit 1
fi
###

echo "==> Stopping the relayer"
pkill rly
