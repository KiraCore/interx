#!/usr/bin/env bash
# To run test locally: make network-start && ./scripts/test-local/account-balances.sh
set -e
set -x
. /etc/profile

TEST_NAME="TX-HASH-QUERY" && timerStart $TEST_NAME
echoInfo "INFO: $TEST_NAME - Integration Test - START"

VALIDATOR_ADDRESS=$(showAddress validator)
addAccount testuser8
TESTUSER_ADDRESS=$(showAddress testuser8)

TXRESULT=$(sekaid tx bank send validator $TESTUSER_ADDRESS 5ukex --keyring-backend=test --chain-id=$NETWORK_NAME --fees 100ukex --broadcast-mode=async --output=json --yes --home=$SEKAID_HOME 2> /dev/null || exit 1)
TX_ID=$(echo $TXRESULT | jsonQuickParse "txhash")
sleep 5
TXQUERYRESULT=$(sekaid query tx $TX_ID --chain-id=$NETWORK_NAME --output=json --home=$SEKAID_HOME 2> /dev/null || exit 1)
BLOCK_HEIGHT=$(echo $TXQUERYRESULT | jsonQuickParse "height")
echo $TX_ID
echo $BLOCK_HEIGHT

INTERX_GATEWAY="127.0.0.1:11000"
RESULT_FROM_INTERX=$(curl --fail $INTERX_GATEWAY/api/kira/txs/$TX_ID || exit 1)
RESULT_ID=$(echo $RESULT_FROM_INTERX  | jq '.hash')
RESULT_HEIGHT=$(echo $RESULT_FROM_INTERX | jq '.height')

[[ '"'$TX_ID'"' != $RESULT_ID ]] && echoErr "ERROR: Expected tx hash to be '$TX_ID', but got '$RESULT_ID'" && exit 1
[[ '"'$BLOCK_HEIGHT'"' != "$RESULT_HEIGHT" ]] && echoErr "ERROR: Expected tx block height to be '$BLOCK_HEIGHT', but got '$RESULT_HEIGHT'" && exit 1

echoInfo "INFO: $TEST_NAME - Integration Test - END, elapsed: $(prettyTime $(timerSpan $TEST_NAME))"