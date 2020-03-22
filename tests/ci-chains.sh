#!/bin/bash
set -e

# TODO make this configurable?
RELAYER=/home/jun/src/relayer/build/rly
NODE_NAME=simappcli

# First initialize your configuration for the relayer
${RELAYER} config init

# Then add the chains and paths that you will need to work with the 
# gaia chains spun up by the two-chains script
${RELAYER} chains add -f demo/ibc0.json
${RELAYER} chains add -f demo/ibc1.json

# To finalize your config, add a path between the two chains
${RELAYER} paths add ibc0 ibc1 demo-path -f demo/path.json

# Now, add the key seeds from each chain to the relayer to give it funds to work with
${RELAYER} keys restore ibc0 testkey "$(jq -r '.secret' data/ibc0/n0/${NODE_NAME}/key_seed.json)" -a
${RELAYER} keys restore ibc1 testkey "$(jq -r '.secret' data/ibc1/n0/${NODE_NAME}/key_seed.json)" -a

# Then its time to initialize the relayer's lite clients for each chain
# All data moving forward is validated by these lite clients.
${RELAYER} lite init ibc0 -f
${RELAYER} lite init ibc1 -f

# Now you can connect the two chains with one command:
${RELAYER} tx full-path ibc0 ibc1
