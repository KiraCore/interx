#!/usr/bin/env bash
# To run test locally: make network-start && ./scripts/test-local/account-balances.sh
set -e
set -x
. /etc/profile

TEST_NAME="TX-HASH-QUERY" && timerStart $TEST_NAME
echoInfo "INFO: $TEST_NAME - Integration Test - START"

VALIDATOR_ADDRESS=$(showAddress validator)
addAccount testuser4
TESTUSER_ADDRESS=$(showAddress testuser4)

TXRESULT=$(sekaid tx bank send validator $TESTUSER_ADDRESS 5ukex --keyring-backend=test --chain-id=$NETWORK_NAME --fees 100ukex --broadcast-mode=async --output=json --yes --home=$SEKAID_HOME 2> /dev/null || exit 1)
TX_ID=$(echo $TXRESULT | jsonQuickParse "txhash")
sleep 5
TXQUERYRESULT=$(sekaid query tx $TX_ID --chain-id=$NETWORK_NAME --output=json --home=$SEKAID_HOME 2> /dev/null || exit 1)
BLOCK_HEIGHT=$(echo $TXQUERYRESULT | jsonQuickParse "height")

INTERX_GATEWAY="127.0.0.1:11000"
RESULT_FROM_INTERX=$(curl --fail $INTERX_GATEWAY/api/blocks | jq '.block_metas[]' || exit 1)
FLAG="false"
while read -r height ; do
  read -r txs
  [ '"'$BLOCK_HEIGHT'"' == $height ] && [ $txs == '"1"' ] && FLAG="true" && break
done < <(echo "$RESULT_FROM_INTERX" | jq '.header.height, .num_txs')

echo $FLAG

[[ $FLAG == "false" ]] && echoErr "ERROR: Expected block height and tx information were not fetched from the API" && exit 1

echoInfo "INFO: $TEST_NAME - Integration Test - END, elapsed: $(prettyTime $(timerSpan $TEST_NAME))"