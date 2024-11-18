#!/usr/bin/env bash
# To run test locally: make network-start && ./scripts/test-local/account-balances.sh
set -e
set -x
. /etc/profile

TEST_NAME="EVM-TRANSACTION" && timerStart $TEST_NAME
echoInfo "INFO: $TEST_NAME - Integration Test - START"

INTERX_GATEWAY="127.0.0.1:11000"

RESULT_FROM_INTERX=$(curl --fail 127.0.0.1:11000/api/goerli/transactions/0xf951c4d45c902b2665ec50be8eb6b10f2ba8c6ca09f7b32efa6022ded5451d74 || echo "error")
RESULT_SUM_FROM_INTERX="$(echo $RESULT_FROM_INTERX | jq '.blockHash')$(echo $RESULT_FROM_INTERX | jq '.blockNumber')$(echo $RESULT_FROM_INTERX | jq '.from')$(echo $RESULT_FROM_INTERX | jq '.to')$(echo $RESULT_FROM_INTERX | jq '.value')"

[[ '"0x69110555bd59329b2114b17694c70fad148178741a0270b1cd56023630fabee8""0x7805f6""0x91b72503fe82a380ac2c98542c2439c6d832b341""0x626917cf83eb583a67bcfcd65f2076c613c8c59a""0x38d7ea4c68000"' != "$RESULT_SUM_FROM_INTERX" ]] && echoErr "ERROR: Expected contract address" && exit 1

echoInfo "INFO: $TEST_NAME - Integration Test - END, elapsed: $(prettyTime $(timerSpan $TEST_NAME))"