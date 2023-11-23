#!/usr/bin/env bash
# To run test locally: make network-start && ./scripts/test-local/account-balances.sh
set -e
set -x
. /etc/profile

TEST_NAME="TXS-QUERY" && timerStart $TEST_NAME
echoInfo "INFO: $TEST_NAME - Integration Test - START"

VALIDATOR_ADDRESS=$(showAddress validator)
addAccount testuser9
TESTUSER_ADDRESS=$(showAddress testuser9)
RESULT=$(sekaid tx bank send validator $TESTUSER_ADDRESS 5ukex --keyring-backend=test --chain-id=$NETWORK_NAME --fees 100ukex --broadcast-mode=async --output=json --yes --home=$SEKAID_HOME 2> /dev/null || exit 1)
TX_HASH=0x$(echo $RESULT | jsonQuickParse "txhash")

sleep 5

INTERX_GATEWAY="127.0.0.1:11000"
RESULT_FROM_INTERX=$(curl --fail "$INTERX_GATEWAY/api/transactions?address=$TESTUSER_ADDRESS&type=send" || exit 1)

RESULT_TOTAL_COUNT=$(echo $RESULT_FROM_INTERX | jq '.total_count' | tr -d '"')
RESULT_TX_HASH=$(echo $RESULT_FROM_INTERX | jq '.transactions[0].hash' | tr -d '"')
RESULT_TX_STATUS=$(echo $RESULT_FROM_INTERX | jq '.transactions[0].status' | tr -d '"')
RESULT_TX_DIRECTION=$(echo $RESULT_FROM_INTERX | jq '.transactions[0].direction' | tr -d '"')

[[ "1" !=  $RESULT_TOTAL_COUNT ]] && echoErr "ERROR: Expected total transactions to be 1, but got '$RESULT_TOTAL_COUNT'" && exit 1
[[ $TX_HASH !=  $RESULT_TX_HASH ]] && echoErr "ERROR: Expected transaction hash to be '$TX_HASH', but got '$RESULT_TX_HASH'" && exit 1
[[ "confirmed" !=  $RESULT_TX_STATUS ]] && echoErr "ERROR: Expected transaction status to be confirmed, but got '$RESULT_TX_STATUS'" && exit 1
[[ "inbound" !=  $RESULT_TX_DIRECTION ]] && echoErr "ERROR: Expected transaction direction to be inbound, but got '$RESULT_TX_DIRECTION'" && exit 1

echoInfo "INFO: $TEST_NAME - Integration Test - END, elapsed: $(prettyTime $(timerSpan $TEST_NAME))"