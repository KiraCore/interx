#!/usr/bin/env bash
# To run test locally: make network-start && ./scripts/test-local/account-balances.sh
set -e
set -x
. /etc/profile

TEST_NAME="EVM-BLOCK" && timerStart $TEST_NAME
echoInfo "INFO: $TEST_NAME - Integration Test - START"

INTERX_GATEWAY="127.0.0.1:11000"

RESULT_FROM_INTERX=$(curl --fail 127.0.0.1:11000/api/goerli/blocks/100000 || echo "error")
RESULT_SUM_FROM_INTERX="$(echo $RESULT_FROM_INTERX | jq '.timestamp')$(echo $RESULT_FROM_INTERX | jq '.number')"

[[ '"0x5c6a6ffc""0x186a0"' != "$RESULT_SUM_FROM_INTERX" ]] && echoErr "ERROR: Expected contract address" && exit 1

echoInfo "INFO: $TEST_NAME - Integration Test - END, elapsed: $(prettyTime $(timerSpan $TEST_NAME))"