#!/usr/bin/env bash
# To run test locally: make network-start && ./scripts/test-local/account-balances.sh
set -e
set -x
. /etc/profile

TEST_NAME="EVM-ACCOUNT-QUERY" && timerStart $TEST_NAME
echoInfo "INFO: $TEST_NAME - Integration Test - START"

INTERX_GATEWAY="127.0.0.1:11000"

RESULT_FROM_INTERX=$(curl --fail 127.0.0.1:11000/api/goerli/accounts/0x326C977E6efc84E512bB9C30f76E30c160eD06FB | jq '.account' || echo "error")
RESULT_SUM_FROM_INTERX="$(echo $RESULT_FROM_INTERX | jq '."@type"')$(echo $RESULT_FROM_INTERX | jq '.address')$(echo $RESULT_FROM_INTERX | jq '.pending')$(echo $RESULT_FROM_INTERX | jq '.sequence')"

[[ '"contract""0x326C977E6efc84E512bB9C30f76E30c160eD06FB"11' != "$RESULT_SUM_FROM_INTERX" ]] && echoErr "ERROR: Expected contract address" && exit 1

echoInfo "INFO: $TEST_NAME - Integration Test - END, elapsed: $(prettyTime $(timerSpan $TEST_NAME))"