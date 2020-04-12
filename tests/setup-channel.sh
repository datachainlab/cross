#!/bin/bash
set -e

NODE_NAME=simappcli
RELAYER_CMD="${RELAYER_CLI} --home $(pwd)/.relayer"

# First initialize your configuration for the relayer
${RELAYER_CMD} config init

# Then add the chains and paths that you will need to work with the 
# gaia chains spun up by the two-chains script
${RELAYER_CMD} chains add -f demo/ibc0.json
${RELAYER_CMD} chains add -f demo/ibc1.json
${RELAYER_CMD} chains add -f demo/ibc2.json

# To finalize your config, add a path between the two chains
${RELAYER_CMD} paths add ibc0 ibc1 path01 -f demo/path01.json
${RELAYER_CMD} paths add ibc0 ibc2 path02 -f demo/path02.json

# Now, add the key seeds from each chain to the relayer to give it funds to work with
${RELAYER_CMD} keys restore ibc0 testkey "$(jq -r '.secret' data/ibc0/n0/${NODE_NAME}/key_seed.json)"
${RELAYER_CMD} keys restore ibc1 testkey "$(jq -r '.secret' data/ibc1/n0/${NODE_NAME}/key_seed.json)"
${RELAYER_CMD} keys restore ibc2 testkey "$(jq -r '.secret' data/ibc2/n0/${NODE_NAME}/key_seed.json)"

# Then its time to initialize the relayer's lite clients for each chain
# All data moving forward is validated by these lite clients.
${RELAYER_CMD} lite init ibc0 -f
${RELAYER_CMD} lite init ibc1 -f
${RELAYER_CMD} lite init ibc2 -f

retry() {
   MAX_RETRY=5
   n=0
   until [ $n -ge $MAX_RETRY ]
   do
      "$@" && break
      n=$[$n+1]
      sleep 1s
   done
   if [ $n -ge $MAX_RETRY ]; then
     echo "failed to execute command ${@}" >&2
     exit 1
   fi
}

# Sometimes an execution of ABCI query is unstable when running on Github action, so we retry it
# Now you can connect the two chains with one command:
retry ${RELAYER_CMD} tx full-path path01 -d -o 3s
retry ${RELAYER_CMD} tx full-path path02 -d -o 3s
