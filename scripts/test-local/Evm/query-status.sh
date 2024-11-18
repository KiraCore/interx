#!/usr/bin/env bash
# To run test locally: make network-start && ./scripts/test-local/account-balances.sh
set -e
set -x
. /etc/profile

TEST_NAME="EVM-STATUS" && timerStart $TEST_NAME
echoInfo "INFO: $TEST_NAME - Integration Test - START"

INTERX_GATEWAY="127.0.0.1:11000"

RESULT_FROM_INTERX=$(curl --fail 127.0.0.1:11000/api/goerli/status || echo "error")
RESULT_SUM_FROM_INTERX="$(echo $RESULT_FROM_INTERX | jq '.node_info.network')"

[[ '5' != "$RESULT_SUM_FROM_INTERX" ]] && echoErr "ERROR: Expected contract address" && exit 1
[[ 'null' = "$(echo $RESULT_FROM_INTERX | jq '.node_info.rpc_address')" ]] && echoErr "ERROR: Expected contract address" && exit 1
[[ 'null' = "$(echo $RESULT_FROM_INTERX | jq '.node_info.version')" ]] && echoErr "ERROR: Expected contract address" && exit 1
[[ 'null' = "$(echo $RESULT_FROM_INTERX | jq '.sync_info')" ]] && echoErr "ERROR: Expected contract address" && exit 1
[[ 'null' = "$(echo $RESULT_FROM_INTERX | jq '.gas_price')" ]] && echoErr "ERROR: Expected contract address" && exit 1

echoInfo "INFO: $TEST_NAME - Integration Test - END, elapsed: $(prettyTime $(timerSpan $TEST_NAME))"