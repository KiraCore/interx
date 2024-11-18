#!/usr/bin/env bash
# To run test locally: make network-start && ./scripts/test-local/account-balances.sh
set -e
set -x
. /etc/profile

TEST_NAME="EVM-BALANCE-QUERY" && timerStart $TEST_NAME
echoInfo "INFO: $TEST_NAME - Integration Test - START"

INTERX_GATEWAY="127.0.0.1:11000"

RESULT_FROM_INTERX=$(curl --fail 127.0.0.1:11000/api/goerli/balances/0x91b72503fe82a380ac2c98542c2439c6d832b341\?tokens\=0x326c977e6efc84e512bb9c30f76e30c160ed06fb | jq '.balances' | jq ".[1]" || echo "error")
RESULT_SUM_FROM_INTERX="$(echo $RESULT_FROM_INTERX | jq '.amount')"

[[ '"20000000000000000000"' != "$RESULT_SUM_FROM_INTERX" ]] && echoErr "ERROR: Expected contract address" && exit 1

echoInfo "INFO: $TEST_NAME - Integration Test - END, elapsed: $(prettyTime $(timerSpan $TEST_NAME))"