#!/bin/bash
set -e

PREV_DIR=$(pwd)
RELAYER_DIR=$(mktemp -d)

echo "RELAYER_DIR is ${RELAYER_DIR}"

cd ${RELAYER_DIR}
git clone https://github.com/cosmos/relayer.git
cd ./relayer
echo "Building Relayer..."
make build

export RELAYER=${RELAYER_DIR}/relayer/build/rly

cd ${PREV_DIR}

./three-chainz
# wait for all chains to start.
sleep 10
./setup-channel.sh
