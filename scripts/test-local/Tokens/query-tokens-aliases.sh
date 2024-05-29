#!/usr/bin/env bash
# To run test locally: make network-start && ./scripts/test-local/account-balances.sh
set -e
set -x
. /etc/profile

TEST_NAME="TOKENS-RATES-QUERY" && timerStart $TEST_NAME
echoInfo "INFO: $TEST_NAME - Integration Test - START"

INTERX_GATEWAY="127.0.0.1:11000"
TOKENS_NUMS_INTERX=$(curl --fail "$INTERX_GATEWAY/api/kira/tokens/aliases" | jq '.token_aliases_data | length' || exit 1)
TOKENS_NUMS_CLI=$(showTokenAliases | jq '.data | length')

[[ $TOKENS_NUMS_INTERX != $TOKENS_NUMS_CLI ]] && echoErr "ERROR: Expected number of tokens to be '$TOKENS_NUMS_CLI', but got '$TOKENS_NUMS_INTERX'" && exit 1

echoInfo "INFO: $TEST_NAME - Integration Test - END, elapsed: $(prettyTime $(timerSpan $TEST_NAME))"