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
TX_ID=$(${NODE_CLI} tx --home ./data/ibc0/n0/simappcli cross create --from ${ACC0} --keyring-backend=test --chain-id ${CO_CHAIN} --yes \
    --contract ./data/train.json --channel ${SRC01_CHAN}:${SRC01_PORT} \
    --contract ./data/hotel.json --channel ${SRC02_CHAN}:${SRC02_PORT} \
    $((${LATEST_HEIGHT}+100)) 0 | jq -r '.data')
###

sleep ${WAIT_NEW_BLOCK}

### TRAIN_CHAIN receives the PacketDataPrepare
CLIENT_ID=$(${RELAYER_CMD} paths show path01 --json | jq -r '.dst["client-id"]')
${RELAYER_CMD} transactions raw update-client ${TRAIN_CHAIN} ${CO_CHAIN} ${CLIENT_ID}
INCLUDED_AT=$(${RELAYER_CMD} query client ${TRAIN_CHAIN} ibczeroclient | jq -r '.client_state.value.LastHeader.SignedHeader.header.height')
${NODE_CLI} tx --home ${CO_HOME} relayer relay \
  --from ${RELAYER0} --keyring-backend=test --chain-id ${CO_CHAIN} --relayer-address=${RELAYER1} --yes \
  ${INCLUDED_AT} ${SRC01_PORT} ${SRC01_CHAN} 1 ${DST01_PORT} ${DST01_CHAN} > ./data/packet0.json
${NODE_CLI} tx --home ${TRAIN_HOME} sign ./data/packet0.json --from ${RELAYER1} --keyring-backend=test --yes > ./data/packet0-signed.json
${NODE_CLI} tx --home ${TRAIN_HOME} broadcast ./data/packet0-signed.json --broadcast-mode=block --from ${RELAYER1} --keyring-backend=test --yes
###

### HOTEL_CHAIN receives the PacketDataPrepare
CLIENT_ID=$(${RELAYER_CMD} paths show path02 --json | jq -r '.dst["client-id"]')
${RELAYER_CMD} transactions raw update-client ${HOTEL_CHAIN} ${CO_CHAIN} ${CLIENT_ID}
INCLUDED_AT=$(${RELAYER_CMD} query client ${HOTEL_CHAIN} ibczeroclient | jq -r '.client_state.value.LastHeader.SignedHeader.header.height')
${NODE_CLI} tx --home ${CO_HOME} relayer relay \
  --from ${RELAYER0} --keyring-backend=test --chain-id ${CO_CHAIN} --relayer-address=${RELAYER2} --yes \
  ${INCLUDED_AT} ${SRC02_PORT} ${SRC02_CHAN} 1 ${DST02_PORT} ${DST02_CHAN} > ./data/packet1.json
${NODE_CLI} tx --home ${HOTEL_HOME} sign ./data/packet1.json --from ${RELAYER2} --keyring-backend=test --yes > ./data/packet1-signed.json
${NODE_CLI} tx --home ${HOTEL_HOME} broadcast ./data/packet1-signed.json --broadcast-mode=block --from ${RELAYER2} --keyring-backend=test --yes
###

sleep ${WAIT_NEW_BLOCK}

### Coordinator receives PacketDataPrepareResult from TRAIN_CHAIN
CLIENT_ID=$(${RELAYER_CMD} paths show path01 --json | jq -r '.src["client-id"]')
${RELAYER_CMD} transactions raw update-client ${CO_CHAIN} ${TRAIN_CHAIN} ${CLIENT_ID}
INCLUDED_AT=$(${RELAYER_CMD} query client ${CO_CHAIN} ${CLIENT_ID} | jq -r '.client_state.value.LastHeader.SignedHeader.header.height')
${NODE_CLI} tx --home ${TRAIN_HOME} relayer relay \
  --from ${RELAYER1} --keyring-backend=test --chain-id ${TRAIN_CHAIN} --relayer-address=${RELAYER0} --yes \
  ${INCLUDED_AT} ${DST01_PORT} ${DST01_CHAN} 1 ${SRC01_PORT} ${SRC01_CHAN} > ./data/packet2.json
${NODE_CLI} tx --home ${CO_HOME} sign ./data/packet2.json --from ${RELAYER0} --keyring-backend=test --yes > ./data/packet2-signed.json
${NODE_CLI} tx --home ${CO_HOME} broadcast ./data/packet2-signed.json --broadcast-mode=block --from ${RELAYER0} --keyring-backend=test --yes
###

### Coordinator receives PacketDataPrepareResult from HOTEL_CHAIN
CLIENT_ID=$(${RELAYER_CMD} paths show path02 --json | jq -r '.src["client-id"]')
${RELAYER_CMD} transactions raw update-client ${CO_CHAIN} ${HOTEL_CHAIN} ${CLIENT_ID}
INCLUDED_AT=$(${RELAYER_CMD} query client ${CO_CHAIN} ${CLIENT_ID} | jq -r '.client_state.value.LastHeader.SignedHeader.header.height')
${NODE_CLI} tx --home ${HOTEL_HOME} relayer relay \
  --from ${RELAYER2} --keyring-backend=test --chain-id ${HOTEL_CHAIN} --relayer-address=${RELAYER0} --yes \
  ${INCLUDED_AT} ${DST02_PORT} ${DST02_CHAN} 1 ${SRC02_PORT} ${SRC02_CHAN} > ./data/packet3.json
${NODE_CLI} tx --home ${CO_HOME} sign ./data/packet3.json --from ${RELAYER0} --keyring-backend=test --yes > ./data/packet3-signed.json
${NODE_CLI} tx --home ${CO_HOME} broadcast ./data/packet3-signed.json --broadcast-mode=block --from ${RELAYER0} --keyring-backend=test --yes
###

sleep ${WAIT_NEW_BLOCK}

### TRAIN_CHAIN receives PacketDataCommit from coordinator
CLIENT_ID=$(${RELAYER_CMD} paths show path01 --json | jq -r '.dst["client-id"]')
${RELAYER_CMD} transactions raw update-client ${TRAIN_CHAIN} ${CO_CHAIN} ${CLIENT_ID}
INCLUDED_AT=$(${RELAYER_CMD} query client ${TRAIN_CHAIN} ibczeroclient | jq -r '.client_state.value.LastHeader.SignedHeader.header.height')
${NODE_CLI} tx --home ${CO_HOME} relayer relay \
  --from ${RELAYER0} --keyring-backend=test --chain-id ${CO_CHAIN} --relayer-address=${RELAYER1} --yes \
  ${INCLUDED_AT} ${SRC01_PORT} ${SRC01_CHAN} 2 ${DST01_PORT} ${DST01_CHAN} > ./data/packet4.json
${NODE_CLI} tx --home ${TRAIN_HOME} sign ./data/packet4.json --from ${RELAYER1} --keyring-backend=test --yes > ./data/packet4-signed.json
${NODE_CLI} tx --home ${TRAIN_HOME} broadcast ./data/packet4-signed.json --broadcast-mode=block --from ${RELAYER1} --keyring-backend=test --yes
###

### HOTEL_CHAIN receives PacketDataCommit from coordinator
CLIENT_ID=$(${RELAYER_CMD} paths show path02 --json | jq -r '.dst["client-id"]')
${RELAYER_CMD} transactions raw update-client ${HOTEL_CHAIN} ${CO_CHAIN} ${CLIENT_ID}
INCLUDED_AT=$(${RELAYER_CMD} query client ${HOTEL_CHAIN} ibczeroclient | jq -r '.client_state.value.LastHeader.SignedHeader.header.height')
${NODE_CLI} tx --home ${CO_HOME} relayer relay \
  --from ${RELAYER0} --keyring-backend=test --chain-id ${CO_CHAIN} --relayer-address=${RELAYER2} --yes \
  ${INCLUDED_AT} ${SRC02_PORT} ${SRC02_CHAN} 2 ${DST02_PORT} ${DST02_CHAN} > ./data/packet5.json
${NODE_CLI} tx --home ${HOTEL_HOME} sign ./data/packet5.json --from ${RELAYER2} --keyring-backend=test --yes > ./data/packet5-signed.json
${NODE_CLI} tx --home ${HOTEL_HOME} broadcast ./data/packet5-signed.json --broadcast-mode=block --from ${RELAYER2} --keyring-backend=test --yes
###

sleep ${WAIT_NEW_BLOCK}

### Coordinator receives PacketDataAckCommit from TRAIN_CHAIN
CLIENT_ID=$(${RELAYER_CMD} paths show path01 --json | jq -r '.src["client-id"]')
${RELAYER_CMD} transactions raw update-client ${CO_CHAIN} ${TRAIN_CHAIN} ${CLIENT_ID}
INCLUDED_AT=$(${RELAYER_CMD} query client ${CO_CHAIN} ${CLIENT_ID} | jq -r '.client_state.value.LastHeader.SignedHeader.header.height')
${NODE_CLI} tx --home ${TRAIN_HOME} relayer relay \
  --from ${RELAYER1} --keyring-backend=test --chain-id ${TRAIN_CHAIN} --relayer-address=${RELAYER0} --yes \
  ${INCLUDED_AT} ${DST01_PORT} ${DST01_CHAN} 2 ${SRC01_PORT} ${SRC01_CHAN} > ./data/packet6.json
${NODE_CLI} tx --home ${CO_HOME} sign ./data/packet6.json --from ${RELAYER0} --keyring-backend=test --yes > ./data/packet6-signed.json
${NODE_CLI} tx --home ${CO_HOME} broadcast ./data/packet6-signed.json --broadcast-mode=block --from ${RELAYER0} --keyring-backend=test --yes
###

### Coordinator receives PacketDataAckCommit from HOTEL_CHAIN
CLIENT_ID=$(${RELAYER_CMD} paths show path02 --json | jq -r '.src["client-id"]')
${RELAYER_CMD} transactions raw update-client ${CO_CHAIN} ${HOTEL_CHAIN} ${CLIENT_ID}
INCLUDED_AT=$(${RELAYER_CMD} query client ${CO_CHAIN} ${CLIENT_ID} | jq -r '.client_state.value.LastHeader.SignedHeader.header.height')
${NODE_CLI} tx --home ${HOTEL_HOME} relayer relay \
  --from ${RELAYER2} --keyring-backend=test --chain-id ${HOTEL_CHAIN} --relayer-address=${RELAYER0} --yes \
  ${INCLUDED_AT} ${DST02_PORT} ${DST02_CHAN} 2 ${SRC02_PORT} ${SRC02_CHAN} > ./data/packet7.json
${NODE_CLI} tx --home ${CO_HOME} sign ./data/packet7.json --from ${RELAYER0} --keyring-backend=test --yes > ./data/packet7-signed.json
${NODE_CLI} tx --home ${CO_HOME} broadcast ./data/packet7-signed.json --broadcast-mode=block --from ${RELAYER0} --keyring-backend=test --yes
###

sleep ${WAIT_NEW_BLOCK}

### Ensure coordinator status is done
STATUS=$(${NODE_CLI} query --home ${CO_HOME} cross coordinator ${TX_ID} | jq -r '.completed')
if [ ${STATUS} = "true" ]; then
  echo "completed!"
else
  echo "failed"
  exit 1
fi
###
