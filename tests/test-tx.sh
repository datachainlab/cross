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

ACC0=acc0

# Get contract ops from each nodes
${NODE_CLI} query --home ${CO_HOME} contract call --node ${TRAIN_NODE} --from ${ACC0} --keyring-backend=test train reserve 0x00000001 --chain-id ${TRAIN_CHAIN} --save ./data/train.json
${NODE_CLI} query --home ${CO_HOME} contract call --node ${HOTEL_NODE} --from ${ACC0} --keyring-backend=test hotel reserve 0x00000002 --chain-id ${HOTEL_CHAIN} --save ./data/hotel.json

SRC01_CHAN=$(${RELAYER_CMD} paths show path01 --json | jq -r '.src."channel-id"')
SRC01_PORT=$(${RELAYER_CMD} paths show path01 --json | jq -r '.src."port-id"')
DST01_CHAN=$(${RELAYER_CMD} paths show path01 --json | jq -r '.dst."channel-id"')
DST01_PORT=$(${RELAYER_CMD} paths show path01 --json | jq -r '.dst."port-id"')

SRC02_CHAN=$(${RELAYER_CMD} paths show path02 --json | jq -r '.src."channel-id"')
SRC02_PORT=$(${RELAYER_CMD} paths show path02 --json | jq -r '.src."port-id"')
DST02_CHAN=$(${RELAYER_CMD} paths show path02 --json | jq -r '.dst."channel-id"')
DST02_PORT=$(${RELAYER_CMD} paths show path02 --json | jq -r '.dst."port-id"')

RELAYER0=$(${NODE_CLI} --home ${CO_HOME} --keyring-backend=test keys show ${ACC0} -a)
RELAYER1=$(${NODE_CLI} --home ${TRAIN_HOME} --keyring-backend=test keys show ${ACC0} -a)
RELAYER2=$(${NODE_CLI} --home ${HOTEL_HOME} --keyring-backend=test keys show ${ACC0} -a)

### Broadcast MsgInitiate
LATEST_HEIGHT=$(${NODE_CLI} --home ${CO_HOME} status | jq -r '.sync_info.latest_block_height')
${NODE_CLI} tx --home ./data/ibc0/n0/simappcli cross create --from ${ACC0} --keyring-backend=test --chain-id ${CO_CHAIN} --yes \
    --contract ./data/train.json --channel ${SRC01_CHAN}:${SRC01_PORT} \
    --contract ./data/hotel.json --channel ${SRC02_CHAN}:${SRC02_PORT} \
    $((${LATEST_HEIGHT}+100)) 0
###

sleep ${WAIT_NEW_BLOCK}

### TRAIN_CHAIN receives the PacketDataPrepare
CLIENT_ID=$(${RELAYER_CMD} paths show path01 --json | jq -r '.dst["client-id"]')
${RELAYER_CMD} transactions raw update-client ${TRAIN_CHAIN} ${CO_CHAIN} ${CLIENT_ID}
INCLUDED_AT=$(${RELAYER_CMD} query client ${TRAIN_CHAIN} ibczeroclient | jq -r '.client_state.value.LastHeader.SignedHeader.header.height')
${NODE_CLI} tx --home ${CO_HOME} relayer relay \
  --from ${RELAYER0} --keyring-backend=test --chain-id ${CO_CHAIN} --relayer-address=${RELAYER1} --yes \
  ${INCLUDED_AT} ${SRC01_PORT} ${SRC01_CHAN} 1 ${DST01_PORT} ${DST01_CHAN} > packet0.json
${NODE_CLI} tx --home ${TRAIN_HOME} sign ./packet0.json --from ${RELAYER1} --keyring-backend=test --yes > packet0-signed.json
${NODE_CLI} tx --home ${TRAIN_HOME} broadcast ./packet0-signed.json --broadcast-mode=block --from ${RELAYER1} --keyring-backend=test --yes
###

### HOTEL_CHAIN receives the PacketDataPrepare
CLIENT_ID=$(${RELAYER_CMD} paths show path02 --json | jq -r '.dst["client-id"]')
${RELAYER_CMD} transactions raw update-client ${HOTEL_CHAIN} ${CO_CHAIN} ${CLIENT_ID}
INCLUDED_AT=$(${RELAYER_CMD} query client ${HOTEL_CHAIN} ibczeroclient | jq -r '.client_state.value.LastHeader.SignedHeader.header.height')
${NODE_CLI} tx --home ${CO_HOME} relayer relay \
  --from ${RELAYER0} --keyring-backend=test --chain-id ${CO_CHAIN} --relayer-address=${RELAYER2} --yes \
  ${INCLUDED_AT} ${SRC02_PORT} ${SRC02_CHAN} 1 ${DST02_PORT} ${DST02_CHAN} > packet1.json
${NODE_CLI} tx --home ${HOTEL_HOME} sign ./packet1.json --from ${RELAYER2} --keyring-backend=test --yes > packet1-signed.json
${NODE_CLI} tx --home ${HOTEL_HOME} broadcast ./packet1-signed.json --broadcast-mode=block --from ${RELAYER2} --keyring-backend=test --yes
###

sleep ${WAIT_NEW_BLOCK}

### Coordinator receives PacketDataPrepareResult from TRAIN_CHAIN
CLIENT_ID=$(${RELAYER_CMD} paths show path01 --json | jq -r '.src["client-id"]')
${RELAYER_CMD} transactions raw update-client ${CO_CHAIN} ${TRAIN_CHAIN} ${CLIENT_ID}
INCLUDED_AT=$(${RELAYER_CMD} query client ${CO_CHAIN} ${CLIENT_ID} | jq -r '.client_state.value.LastHeader.SignedHeader.header.height')
${NODE_CLI} tx --home ${TRAIN_HOME} relayer relay \
  --from ${RELAYER1} --keyring-backend=test --chain-id ${TRAIN_CHAIN} --relayer-address=${RELAYER0} --yes \
  ${INCLUDED_AT} ${DST01_PORT} ${DST01_CHAN} 1 ${SRC01_PORT} ${SRC01_CHAN} > packet2.json
${NODE_CLI} tx --home ${CO_HOME} sign ./packet2.json --from ${RELAYER0} --keyring-backend=test --yes > packet2-signed.json
${NODE_CLI} tx --home ${CO_HOME} broadcast ./packet2-signed.json --broadcast-mode=block --from ${RELAYER0} --keyring-backend=test --yes
###

### Coordinator receives PacketDataPrepareResult from HOTEL_CHAIN
CLIENT_ID=$(${RELAYER_CMD} paths show path02 --json | jq -r '.src["client-id"]')
${RELAYER_CMD} transactions raw update-client ${CO_CHAIN} ${HOTEL_CHAIN} ${CLIENT_ID}
INCLUDED_AT=$(${RELAYER_CMD} query client ${CO_CHAIN} ${CLIENT_ID} | jq -r '.client_state.value.LastHeader.SignedHeader.header.height')
${NODE_CLI} tx --home ${HOTEL_HOME} relayer relay \
  --from ${RELAYER2} --keyring-backend=test --chain-id ${HOTEL_CHAIN} --relayer-address=${RELAYER0} --yes \
  ${INCLUDED_AT} ${DST02_PORT} ${DST02_CHAN} 1 ${SRC02_PORT} ${SRC02_CHAN} > packet3.json
${NODE_CLI} tx --home ${CO_HOME} sign ./packet3.json --from ${RELAYER0} --keyring-backend=test --yes > packet3-signed.json
${NODE_CLI} tx --home ${CO_HOME} broadcast ./packet3-signed.json --broadcast-mode=block --from ${RELAYER0} --keyring-backend=test --yes
###

sleep ${WAIT_NEW_BLOCK}

### TRAIN_CHAIN receives PacketDataCommit from coordinator
CLIENT_ID=$(${RELAYER_CMD} paths show path01 --json | jq -r '.dst["client-id"]')
${RELAYER_CMD} transactions raw update-client ${TRAIN_CHAIN} ${CO_CHAIN} ${CLIENT_ID}
INCLUDED_AT=$(${RELAYER_CMD} query client ${TRAIN_CHAIN} ibczeroclient | jq -r '.client_state.value.LastHeader.SignedHeader.header.height')
${NODE_CLI} tx --home ${CO_HOME} relayer relay \
  --from ${RELAYER0} --keyring-backend=test --chain-id ${CO_CHAIN} --relayer-address=${RELAYER1} --yes \
  ${INCLUDED_AT} ${SRC01_PORT} ${SRC01_CHAN} 2 ${DST01_PORT} ${DST01_CHAN} > packet4.json
${NODE_CLI} tx --home ${TRAIN_HOME} sign ./packet4.json --from ${RELAYER1} --keyring-backend=test --yes > packet4-signed.json
${NODE_CLI} tx --home ${TRAIN_HOME} broadcast ./packet4-signed.json --broadcast-mode=block --from ${RELAYER1} --keyring-backend=test --yes
###

### HOTEL_CHAIN receives PacketDataCommit from coordinator
CLIENT_ID=$(${RELAYER_CMD} paths show path02 --json | jq -r '.dst["client-id"]')
${RELAYER_CMD} transactions raw update-client ${HOTEL_CHAIN} ${CO_CHAIN} ${CLIENT_ID}
INCLUDED_AT=$(${RELAYER_CMD} query client ${HOTEL_CHAIN} ibczeroclient | jq -r '.client_state.value.LastHeader.SignedHeader.header.height')
${NODE_CLI} tx --home ${CO_HOME} relayer relay \
  --from ${RELAYER0} --keyring-backend=test --chain-id ${CO_CHAIN} --relayer-address=${RELAYER2} --yes \
  ${INCLUDED_AT} ${SRC02_PORT} ${SRC02_CHAN} 2 ${DST02_PORT} ${DST02_CHAN} > packet5.json
${NODE_CLI} tx --home ${HOTEL_HOME} sign ./packet5.json --from ${RELAYER2} --keyring-backend=test --yes > packet5-signed.json
${NODE_CLI} tx --home ${HOTEL_HOME} broadcast ./packet5-signed.json --broadcast-mode=block --from ${RELAYER2} --keyring-backend=test --yes
###

sleep ${WAIT_NEW_BLOCK}

### Coordinator receives PacketDataAckCommit from TRAIN_CHAIN
CLIENT_ID=$(${RELAYER_CMD} paths show path01 --json | jq -r '.src["client-id"]')
${RELAYER_CMD} transactions raw update-client ${CO_CHAIN} ${TRAIN_CHAIN} ${CLIENT_ID}
INCLUDED_AT=$(${RELAYER_CMD} query client ${CO_CHAIN} ${CLIENT_ID} | jq -r '.client_state.value.LastHeader.SignedHeader.header.height')
${NODE_CLI} tx --home ${TRAIN_HOME} relayer relay \
  --from ${RELAYER1} --keyring-backend=test --chain-id ${TRAIN_CHAIN} --relayer-address=${RELAYER0} --yes \
  ${INCLUDED_AT} ${DST01_PORT} ${DST01_CHAN} 2 ${SRC01_PORT} ${SRC01_CHAN} > packet6.json
${NODE_CLI} tx --home ${CO_HOME} sign ./packet6.json --from ${RELAYER0} --keyring-backend=test --yes > packet6-signed.json
${NODE_CLI} tx --home ${CO_HOME} broadcast ./packet6-signed.json --broadcast-mode=block --from ${RELAYER0} --keyring-backend=test --yes
###

### Coordinator receives PacketDataAckCommit from HOTEL_CHAIN
CLIENT_ID=$(${RELAYER_CMD} paths show path02 --json | jq -r '.src["client-id"]')
${RELAYER_CMD} transactions raw update-client ${CO_CHAIN} ${HOTEL_CHAIN} ${CLIENT_ID}
INCLUDED_AT=$(${RELAYER_CMD} query client ${CO_CHAIN} ${CLIENT_ID} | jq -r '.client_state.value.LastHeader.SignedHeader.header.height')
${NODE_CLI} tx --home ${HOTEL_HOME} relayer relay \
  --from ${RELAYER2} --keyring-backend=test --chain-id ${HOTEL_CHAIN} --relayer-address=${RELAYER0} --yes \
  ${INCLUDED_AT} ${DST02_PORT} ${DST02_CHAN} 2 ${SRC02_PORT} ${SRC02_CHAN} > packet7.json
${NODE_CLI} tx --home ${CO_HOME} sign ./packet7.json --from ${RELAYER0} --keyring-backend=test --yes > packet7-signed.json
${NODE_CLI} tx --home ${CO_HOME} broadcast ./packet7-signed.json --broadcast-mode=block --from ${RELAYER0} --keyring-backend=test --yes
###
