package common

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/database"
	"github.com/KiraCore/interx/log"
	"github.com/KiraCore/interx/types"
	"github.com/KiraCore/interx/types/rosetta"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// Regexp definitions
var gasWantedRemoveRegex = regexp.MustCompile(`\s*\"gas_wanted\" *: *\".*\"(,|)`)
var gasUsedRemoveRegex = regexp.MustCompile(`\s*\"gas_used\" *: *\".*\"(,|)`)

type conventionalMarshaller struct {
	Value interface{}
}

func (c conventionalMarshaller) MarshalAndConvert(endpoint string) ([]byte, error) {
	marshalled, err := json.MarshalIndent(c.Value, "", "  ")

	// if strings.HasPrefix(endpoint, "/api/status") { // status query
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"FaucetAddr\"", "\"faucet_addr\""))
	// }

	// if strings.HasPrefix(endpoint, "/api/cosmos/auth/accounts/") { // accounts query
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"accountNumber\"", "\"account_number\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"pubKey\"", "\"pub_key\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"typeUrl\"", "\"@type\""))
	// }

	// if strings.HasPrefix(endpoint, "/api/cosmos/bank/balances/") { // accounts query
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"nextKey\"", "\"next_key\""))
	// }

	// if strings.HasPrefix(endpoint, "/api/kira/gov/network_properties") { // network properties query
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"enableForeignFeePayments\"", "\"enable_foreign_fee_payments\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"enableTokenBlacklist\"", "\"enable_token_blacklist\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"enableTokenWhitelist\"", "\"enable_token_whitelist\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"inactiveRankDecreasePercent\"", "\"inactive_rank_decrease_percent\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"inflationPeriod\"", "\"inflation_period\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"inflationRate\"", "\"inflation_rate\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"maxDelegators\"", "\"max_delegators\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"maxJailedPercentage\"", "\"max_jailed_percentage\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"maxMischance\"", "\"max_mischance\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"maxSlashingPercentage\"", "\"max_slashing_percentage\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"maxTxFee\"", "\"max_tx_fee\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"minDelegationPushout\"", "\"min_delegation_pushout\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"minIdentityApprovalTip\"", "\"min_identity_approval_tip\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"minProposalEnactmentBlocks\"", "\"min_proposal_enactment_blocks\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"minProposalEndBlocks\"", "\"min_proposal_end_blocks\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"minTxFee\"", "\"min_tx_fee\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"minValidators\"", "\"min_validators\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"minimumProposalEndTime\"", "\"minimum_proposal_end_time\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"mischanceConfidence\"", "\"mischance_confidence\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"mischanceRankDecreaseAmount\"", "\"mischance_rank_decrease_amount\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"poorNetworkMaxBankSend\"", "\"poor_network_max_bank_send\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"proposalEnactmentTime\"", "\"proposal_enactment_time\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"slashingPeriod\"", "\"slashing_period\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"ubiHardcap\"", "\"ubi_hardcap\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"uniqueIdentityKeys\"", "\"unique_identity_keys\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"unjailMaxTime\"", "\"unjail_max_time\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"unstakingPeriod\"", "\"unstaking_period\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"validatorsFeeShare\"", "\"validators_fee_share\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"voteQuorum\"", "\"vote_quorum\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"abstentionRankDecreaseAmount\"", "\"abstention_rank_decrease_amount\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"dappAutoDenounceTime\"", "\"dapp_auto_denounce_time\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"dappBondDuration\"", "\"dapp_bond_duration\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"dappVerifierBond\"", "\"dapp_verifier_bond\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"maxAbstention\"", "\"max_abstention\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"maxAnnualInflation\"", "\"max_annual_inflation\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"maxCollectiveOutputs\"", "\"max_collective_outputs\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"maxDappBond\"", "\"max_dapp_bond\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"maxProposalChecksumSize\"", "\"max_proposal_checksum_size\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"maxProposalDescriptionSize\"", "\"max_proposal_description_size\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"maxProposalPollOptionCount\"", "\"max_proposal_poll_option_count\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"maxProposalPollOptionSize\"", "\"max_proposal_poll_option_size\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"maxProposalReferenceSize\"", "\"max_proposal_reference_size\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"maxProposalTitleSize\"", "\"max_proposal_title_size\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"minCollectiveBond\"", "\"min_collective_bond\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"minCollectiveBondingTime\"", "\"min_collective_bonding_time\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"minCollectiveClaimPeriod\"", "\"min_collective_claim_period\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"minDappBond\"", "\"min_dapp_bond\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"validatorRecoveryBond\"", "\"validator_recovery_bond\""))
	// }

	// if strings.HasPrefix(endpoint, "/api/kira/tokens/rates") { // network properties query
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"feePayments\"", "\"fee_payments\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"feeRate\"", "\"fee_rate\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"stakeCap\"", "\"stake_cap\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"stakeMin\"", "\"stake_min\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"stakeToken\"", "\"stake_token\""))
	// }

	// if strings.HasPrefix(endpoint, "/api/kira/ubi-records") { // network properties query
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"distributionEnd\"", "\"distribution_end\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"distributionLast\"", "\"distribution_last\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"distributionStart\"", "\"distribution_start\""))
	// }

	// if strings.HasPrefix(endpoint, "/api/kira/spending-pools") { // network properties query
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"ownerAccounts\"", "\"owner_accounts\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"ownerRoles\"", "\"owner_roles\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"claimEnd\"", "\"claim_end\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"claimStart\"", "\"claim_start\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"voteEnactment\"", "\"vote_enactment\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"votePeriod\"", "\"vote_period\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"voteQuorum\"", "\"vote_quorum\""))
	// }

	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"defaultParameters\"", "\"default_parameters\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"executionFee\"", "\"execution_fee\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"failureFee\"", "\"failure_fee\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"timeout\"", "\"timeout\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"transactionType\"", "\"transaction_type\""))

	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"inactiveUntil\"", "\"inactive_until\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"lastPresentBlock\"", "\"last_present_block\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"missedBlocksCounter\"", "\"missed_blocks_counter\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"producedBlocksCounter\"", "\"produced_blocks_counter\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"startHeight\"", "\"start_height\""))

	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"enactmentEndTime\"", "\"enactment_end_time\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"execResult\"", "\"exec_result\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"minEnactmentEndBlockHeight\"", "\"min_enactment_end_block_height\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"minVotingEndBlockHeight\"", "\"min_voting_end_block_height\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"proposalId\"", "\"proposal_id\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"submitTime\"", "\"submit_time\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"votingEndTime\"", "\"voting_end_time\""))

	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"maxCustodyBufferSize\"", "\"max_custody_buffer_size\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"maxCustodyTxSize\"", "\"max_custody_tx_size\""))
	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"minCustodyReward\"", "\"min_custody_reward\""))

	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"verifyRecords\"", "\"verify_records\""))

	marshalled = []byte(strings.ReplaceAll(string(marshalled), "\"txTypes\"", "\"tx_types\""))

	second := gasWantedRemoveRegex.ReplaceAll(
		marshalled,
		[]byte(``),
	)

	third := gasUsedRemoveRegex.ReplaceAll(
		second,
		[]byte(``),
	)

	return third, err
}

// GetInterxRequest is a function to get Interx Request
func GetInterxRequest(r *http.Request) types.InterxRequest {

	log.CustomLogger().Info("Starting 'GetInterxRequest' request...",
		"method", r.Method,
		"endpoint", r.URL.String(),
	)

	request := types.InterxRequest{}

	request.Method = r.Method
	request.Endpoint = r.URL.String()
	request.Params, _ = ioutil.ReadAll(r.Body)

	log.CustomLogger().Info("Finished 'GetInterxRequest' request.")

	return request
}

// GetResponseFormat is a function to get response format
func GetResponseFormat(request types.InterxRequest, rpcAddr string) *types.ProxyResponse {
	response := new(types.ProxyResponse)
	response.Timestamp = time.Now().UTC().Unix()
	response.RequestHash = GetBlake2bHash(request)
	response.Chainid = NodeStatus.Chainid
	response.Block = NodeStatus.Block
	response.Blocktime = NodeStatus.Blocktime

	return response
}

// GetResponseSignature is a function to get response signature
func GetResponseSignature(response types.ProxyResponse) (string, string) {

	log.CustomLogger().Info("Starting 'GetResponseSignature' request...")

	// Get Response Hash
	responseHash := GetBlake2bHash(response.Response)

	// Generate json to be signed
	sign := new(types.ResponseSign)
	sign.Chainid = response.Chainid
	sign.Block = response.Block
	sign.Blocktime = response.Blocktime
	sign.Timestamp = response.Timestamp
	sign.Response = responseHash
	signBytes, err := json.Marshal(sign)
	if err != nil {
		log.CustomLogger().Error("[GetResponseSignature] Failed to create signature.",
			"error", err,
		)
		return "", responseHash
	}

	// Get Signature
	signature, err := config.Config.PrivKey.Sign(signBytes)
	if err != nil {
		log.CustomLogger().Error("[GetResponseSignature] Failed to fetch signature.",
			"error", err,
		)
		return "", responseHash
	}

	log.CustomLogger().Info("Finished 'GetResponseSignature' request.")

	return base64.StdEncoding.EncodeToString([]byte(signature)), responseHash
}

// SearchCache is a function to search response in cache
func SearchCache(request types.InterxRequest, response *types.ProxyResponse) (bool, interface{}, interface{}, int) {

	log.CustomLogger().Info("Starting SearchCache")

	chainIDHash := GetBlake2bHash(response.Chainid)
	endpointHash := GetBlake2bHash(request.Endpoint)
	requestHash := GetBlake2bHash(request)

	log.CustomLogger().Info("`SearchCache` Config Path", "chainIDHash", chainIDHash,
		"endpointHash", endpointHash, "requestHash", requestHash,
	)

	result, err := GetCache(chainIDHash, endpointHash, requestHash)

	if err != nil {
		log.CustomLogger().Error("[SearchCache] Failed to run GetCache function.",
			"error", err,
		)
		return false, nil, nil, -1
	}

	if IsCacheExpired(result) {
		return false, nil, nil, -1
	}

	log.CustomLogger().Info("Finished SearchCache")

	return true, result.Response.Response, result.Response.Error, result.Status
}

// WrapResponse is a function to wrap response
func WrapResponse(w http.ResponseWriter, request types.InterxRequest, response types.ProxyResponse, statusCode int, saveToCache bool) {
	if statusCode == 0 {
		statusCode = 503 // Service Unavailable Error
	}
	if saveToCache {
		log.CustomLogger().Info("Starting Wrap Response...")

		chainIDHash := GetBlake2bHash(response.Chainid)
		endpointHash := GetBlake2bHash(request.Endpoint)
		requestHash := GetBlake2bHash(request)
		if conf, ok := RPCMethods[request.Method][request.Endpoint]; ok {
			err := PutCache(chainIDHash, endpointHash, requestHash, types.InterxResponse{
				Response:             response,
				Status:               statusCode,
				CacheTime:            time.Now().UTC(),
				CachingDuration:      conf.CachingDuration,
				CachingBlockDuration: conf.CachingBlockDuration,
			})
			if err != nil {
				log.CustomLogger().Error("[WrapResponse] Failed to save in the cache.",
					"error", err,
				)
			}
		}
	}

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Interx_chain_id", response.Chainid)
	w.Header().Add("Interx_block", strconv.FormatInt(response.Block, 10))
	w.Header().Add("Interx_blocktime", response.Blocktime)
	w.Header().Add("Interx_timestamp", strconv.FormatInt(response.Timestamp, 10))
	w.Header().Add("Interx_request_hash", response.RequestHash)
	if request.Endpoint == config.QueryDataReference {
		reference, err := database.GetReference(string(request.Params))
		if err == nil {
			w.Header().Add("Interx_ref", "/download/"+reference.FilePath)
		}
	}

	if response.Response != nil {
		response.Signature, response.Hash = GetResponseSignature(response)

		w.Header().Add("Interx_signature", response.Signature)
		w.Header().Add("Interx_hash", response.Hash)
		w.WriteHeader(statusCode)

		switch v := response.Response.(type) {
		case string:
			_, err := w.Write([]byte(v))
			if err != nil {
				log.CustomLogger().Error("[WrapResponse] Failed to make a response.",
					"error", err,
				)
			}
			return
		}

		encoded, _ := conventionalMarshaller{response.Response}.MarshalAndConvert(request.Endpoint)
		_, err := w.Write(encoded)
		if err != nil {
			log.CustomLogger().Error("[WrapResponse] Failed to make a response.",
				"error", err,
			)
		}
	} else {
		w.WriteHeader(statusCode)

		if response.Error == nil {
			response.Error = "service not available"
		}

		encoded, _ := conventionalMarshaller{response.Error}.MarshalAndConvert(request.Endpoint)
		_, err := w.Write(encoded)
		if err != nil {
			log.CustomLogger().Error("[WrapResponse] Failed to make a response.",
				"error", err,
			)
		}
	}
}

// ServeGRPC is a function to serve GRPC
func ServeGRPC(r *http.Request, gwCosmosmux *runtime.ServeMux) (interface{}, interface{}, int) {
	recorder := httptest.NewRecorder()
	gwCosmosmux.ServeHTTP(recorder, r)
	resp := recorder.Result()

	result := new(interface{})
	if json.NewDecoder(resp.Body).Decode(result) == nil {
		if resp.StatusCode == http.StatusOK {
			return result, nil, resp.StatusCode
		}

		return nil, result, resp.StatusCode
	}

	return nil, nil, resp.StatusCode
}

// ServeError is a function to server GRPC
func ServeError(code int, data string, message string, statusCode int) (interface{}, interface{}, int) {
	return nil, types.ProxyResponseError{
		Code:    code,
		Data:    data,
		Message: message,
	}, statusCode
}

func RosettaBuildError(code int, message string, description string, retriable bool, details interface{}) rosetta.Error {
	return rosetta.Error{
		Code:        code,
		Message:     message,
		Description: description,
		Retriable:   retriable,
		Details:     details,
	}
}

func RosettaServeError(code int, data string, message string, statusCode int) (interface{}, interface{}, int) {
	return nil, RosettaBuildError(code, message, data, true, nil), statusCode
}
