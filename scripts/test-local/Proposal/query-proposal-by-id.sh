#!/usr/bin/env bash
# To run test locally: make network-start && ./scripts/test-local/account-balances.sh
set -e
set -x
. /etc/profile

TEST_NAME="PROPOSAL-QUERY-BY-ID" && timerStart $TEST_NAME
echoInfo "INFO: $TEST_NAME - Integration Test - START"

VALIDATOR_ADDRESS=$(showAddress validator)

PROPOSAL_RESULT=$(setPoorNetworkMessages validator asdf)
sleep 5

INTERX_GATEWAY="127.0.0.1:11000"
PROPOSAL_ID_INTERX=$(curl --fail "$INTERX_GATEWAY/api/kira/gov/proposals" | jq '.proposals[0].proposal_id' | tr -d '"' || exit 1)
RESULT_INTERX=$(curl --fail $INTERX_GATEWAY/api/kira/gov/proposals/$PROPOSAL_ID_INTERX || exit 1)
echo $RESULT_INTERX
DESCRIPTION_INTERX=$(echo $RESULT_INTERX | jq '.proposal.content.messages[0]')
QUORUM_INTERX=$(echo $RESULT_INTERX | jq '.proposal.quorum' | tr -d '"')
QUORUM_INTERX=$(printf "%.12f000000" "$(echo "$QUORUM_INTERX" | bc)")

DESCRIPTION_CLI=$(showProposal $PROPOSAL_ID_INTERX | jq '.content.messages[0]')
QUORUM_CLI=$(showNetworkProperties | jq '.properties.vote_quorum' | tr -d '"' || exit 1)

[[ $DESCRIPTION_INTERX != $DESCRIPTION_CLI ]] && echoErr "ERROR: Expected proposal description to be '$DESCRIPTION_CLI', but got '$DESCRIPTION_INTERX'" && exit 1
[[ $QUORUM_INTERX != $QUORUM_CLI ]] && echoErr "ERROR: Expected proposal quorum to be '$QUORUM_CLI', but got '$QUORUM_INTERX'" && exit 1

echoInfo "INFO: $TEST_NAME - Integration Test - END, elapsed: $(prettyTime $(timerSpan $TEST_NAME))"