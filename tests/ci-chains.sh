#!/bin/bash
set -e

PREV_DIR=$(pwd)
RELAYER_DIR=$(mktemp -d)

echo "RELAYER_DIR is ${RELAYER_DIR}"

cd ${RELAYER_DIR}
git clone https://github.com/datachainlab/relayer
cd ./relayer
git checkout 82e4306b5becd3be31674941f7575e6cb3d8f192
echo "Building Relayer..."
make build

export RELAYER_CLI=${RELAYER_DIR}/relayer/build/rly

cd ${PREV_DIR}

./three-chainz
# wait for all chains to start.
sleep 10
./setup-channel.sh
./test-tx.sh
