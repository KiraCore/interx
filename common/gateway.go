package common

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/database"
	"github.com/KiraCore/interx/log"
	"github.com/KiraCore/interx/types"
	"github.com/KiraCore/interx/types/rosetta"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	keyssecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/go-bip39"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

//TODO: Some of endpoints do not work.

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

	request := types.InterxRequest{}

	request.Method = r.Method
	request.Endpoint = r.URL.String()
	request.Params, _ = ioutil.ReadAll(r.Body)

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

	return base64.StdEncoding.EncodeToString([]byte(signature)), responseHash
}

// SearchCache is a function to search response in cache
func SearchCache(request types.InterxRequest, response *types.ProxyResponse) (bool, interface{}, interface{}, int) {

	log.CustomLogger().Info("Starting `SearchCache` function...")

	requestHash := GetBlake2bHash(request)

	log.CustomLogger().Info("`SearchCache` created request hash", "requestHash", requestHash)

	result, expired, err := GetCache(requestHash)

	if err != nil {
		log.CustomLogger().Error("[SearchCache][GetCache] Failed to find data from the cache",
			"error", err,
		)
		return false, nil, nil, -1
	}

	baseDir := config.GetResponseCacheDir()
	filePath := fmt.Sprintf("%s/%s", baseDir, requestHash)

	if expired {
		defer os.Remove(filePath)
		log.CustomLogger().Info("`SearchCache` Query data expired and removed from cache.",
			"filePath", filePath,
		)
	}

	if IsCacheExpired(result) {
		log.CustomLogger().Info("SearchCache: Cache data not found or marked as expired during IsCacheExpired query.")
		return false, nil, nil, -1
	}

	log.CustomLogger().Info("Finished `SearchCache` request...")

	return true, result.Response.Response, result.Response.Error, result.Status
}

// WrapResponse is a function to wrap response
func WrapResponse(w http.ResponseWriter, request types.InterxRequest, response types.ProxyResponse, statusCode int, saveToCache bool) {
	if statusCode == 0 {
		statusCode = 503 // Service Unavailable Error
	}

	log.CustomLogger().Info("Starting `WrapResponse` request...",
		"status_code", statusCode,
		"save_to_cache", saveToCache,
	)

	if saveToCache {

		log.CustomLogger().Info(" `WrapResponse` adding data to the cache started...",
			"endpoint", request.Endpoint,
		)

		chainIDHash := GetBlake2bHash(response.Chainid)
		endpointHash := GetBlake2bHash(request.Endpoint)
		requestHash := GetBlake2bHash(request)
		if conf, ok := RPCMethods[request.Method][request.Endpoint]; ok {
			err := AddCache(chainIDHash, endpointHash, requestHash, types.InterxResponse{
				Response:           response,
				Status:             statusCode,
				CacheTime:          time.Now().UTC(),
				CacheDuration:      conf.CacheDuration,
				CacheBlockDuration: conf.CacheBlockDuration,
			})
			if err != nil {
				log.CustomLogger().Error("[WrapResponse][AddCache] Failed to save data into the cache.",
					"error", err,
				)
			}
			log.CustomLogger().Info("`WrapResponse` adding data to the cache successfully done.",
				"QueryName", request.Endpoint,
				"ResponsechainId", response.Chainid,
				"Status", statusCode,
				"CacheTime", time.Now().UTC(),
				"CacheDuration", conf.CacheDuration,
				"CacheBlockDuration", conf.CacheBlockDuration,
			)
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
				log.CustomLogger().Error("[WrapResponse][Write] Failed to writes the data to the response.",
					"error", err,
				)
			}
			return
		}

		encoded, _ := conventionalMarshaller{response.Response}.MarshalAndConvert(request.Endpoint)
		_, err := w.Write(encoded)
		if err != nil {
			log.CustomLogger().Error("[WrapResponse][Write] Failed to writes the data to the response.",
				"error", err,
			)
		}
	} else {
		w.WriteHeader(statusCode)

		if response.Error == nil {
			response.Error = "[WrapResponse] service not available"
		}

		encoded, _ := conventionalMarshaller{response.Error}.MarshalAndConvert(request.Endpoint)
		_, err := w.Write(encoded)
		if err != nil {
			log.CustomLogger().Error("[WrapResponse][Write] Failed to writes the data to the response.",
				"error", err,
			)
		}
	}

	log.CustomLogger().Info("Finished `WrapResponse` request...")
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

func GetAccountNumber(address string) (uint64, uint64, error) {

	log.CustomLogger().Info("Starting `GetAccountNumber` request...",
		"address", address,
	)

	grpcConn, _ := grpc.Dial(config.DefaultGrpc, grpc.WithInsecure())
	defer grpcConn.Close()

	queryClient := authtypes.NewQueryClient(grpcConn)
	accountAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		log.CustomLogger().Error("[GetAccountNumber][AccAddressFromBech32] failed to parse address",
			"error", err,
		)
		return 0, 0, err
	}

	// Query the account
	accountRes, err := queryClient.Account(context.Background(), &authtypes.QueryAccountRequest{Address: accountAddr.String()})
	if err != nil {
		log.CustomLogger().Error("[GetAccountNumber][Account] failed to query account",
			"error", err,
		)
		return 0, 0, err
	}

	// Extract account information
	var account authtypes.BaseAccount
	err = account.Unmarshal(accountRes.Account.Value)
	if err != nil {
		log.CustomLogger().Error("[GetAccountNumber][Unmarshal] failed to unmarshal account",
			"error", err,
		)
		return 0, 0, err
	}

	log.CustomLogger().Info("`GetAccountNumber` finished successfully",
		"Account Number", account.AccountNumber,
		"Sequence", account.Sequence,
	)
	return account.AccountNumber, account.Sequence, nil
}

// ConvertStringToCoins takes a string and returns sdk.Coins
func ConvertStringToCoins(input string) (sdk.Coins, error) {
	// Use ParseCoinsNormalized to convert the string into sdk.Coins
	coins, err := sdk.ParseCoinsNormalized(input)
	if err != nil {
		return nil, fmt.Errorf("failed to parse coins: %v", err)
	}

	return coins, nil
}

func CreateTransaction(sender string, reciever string, token string, txBuilder client.TxBuilder) error {
	faucetAmount, _ := strconv.ParseInt(config.Config.Faucet.FaucetAmounts[token], 10, 64)
	feeAmount, _ := ConvertStringToCoins(config.Config.Faucet.FeeAmounts[token])
	msg := banktypes.NewMsgSend(sdk.MustAccAddressFromBech32(sender), sdk.MustAccAddressFromBech32(reciever), sdk.NewCoins(sdk.NewInt64Coin(token, faucetAmount)))
	err := txBuilder.SetMsgs(msg)
	if err != nil {
		log.CustomLogger().Error("[CreateTransaction][SetMsgs] failed to create message of transaction",
			"error", err,
		)
		return err
	}
	memo := "send transaction"
	gasLimit := uint64(config.DefaultGasLimit)

	txBuilder.SetGasLimit(gasLimit)
	txBuilder.SetFeeAmount(feeAmount)
	txBuilder.SetMemo(memo)
	txBuilder.SetTimeoutHeight(0)
	return nil
}

func DerivePrivateKeyFromMnemonic() []byte {
	mnemonic := config.Config.Faucet.Mnemonic
	seed, _ := bip39.NewSeedWithErrorChecking(mnemonic, "")
	master, ch := hd.ComputeMastersFromSeed(seed)
	priv, _ := hd.DerivePrivateKeyForPath(master, ch, "44'/118'/0'/0/0")
	return priv
}

func IsEligibleToTransferToken(r *http.Request, gwCosmosmux *runtime.ServeMux, receiver string, token string) bool {
	faucetBalances := GetAccountBalances(gwCosmosmux, r.Clone(r.Context()), config.Config.Faucet.Address)
	var faucetBalance string
	for _, balance := range faucetBalances {
		if balance.Denom == token {
			faucetBalance = balance.Amount
		}
	}

	bigFaucetBalance, ok := new(big.Float).SetString(faucetBalance)
	if !ok {
		log.CustomLogger().Error("[IsEligibleToTransferToken] failed to convert faucet balance.",
			"faucet Balance", faucetBalance,
		)
		return false
	}

	bigFaucetAmount, ok := new(big.Float).SetString(config.Config.Faucet.FaucetAmounts[token])
	if !ok {
		log.CustomLogger().Error("[IsEligibleToTransferToken] failed to convert calim amount.",
			"calim Amount", faucetBalance,
		)
		return false
	}

	log.CustomLogger().Info("[IsEligibleToTransferToken] fetching balance of the accounts is done",
		"faucet balance", bigFaucetBalance,
		"calim amount", bigFaucetAmount,
		"is faucet balance greather than calim amount", bigFaucetAmount.Cmp(bigFaucetBalance) > 0,
	)
	return bigFaucetAmount.Cmp(bigFaucetBalance) > 0
}

func TransferToken(reciever string, token string) (string, error) {

	log.CustomLogger().Info(" Starting `TransferToken` request...")

	txBuilder := config.EncodingCg.TxConfig.NewTxBuilder()
	err := CreateTransaction(config.Config.Faucet.Address, reciever, token, txBuilder)
	if err != nil {
		log.CustomLogger().Error("[TransferToken][CreateTransaction] failed to create transaction.",
			"error", err,
		)
		return "failed to create transaction", err
	}

	dif := &keyssecp256k1.PrivKey{Key: DerivePrivateKeyFromMnemonic()}
	privs := []cryptotypes.PrivKey{dif}
	num, seq, accountErr := GetAccountNumber(config.Config.Faucet.Address)
	if accountErr != nil {
		log.CustomLogger().Error("[TransferToken][GetAccountNumber] failed to get account. Faucet balance is empty. Please add tokens to the faucet before attempting a transfer.",
			"error", err,
		)
		return "faucet balance is empty", err
	}
	accSeqs := []uint64{seq}
	accNums := []uint64{num}

	// First round: we gather all the signer infos. We use the "set empty signature" hack to do that.
	var sigsV2 []signing.SignatureV2
	for i, priv := range privs {
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  config.EncodingCg.TxConfig.SignModeHandler().DefaultMode(),
				Signature: nil,
			},
			Sequence: accSeqs[i],
		}

		sigsV2 = append(sigsV2, sigV2)
	}

	err = txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		log.CustomLogger().Error("[TransferToken][SetSignatures] first round signing failed.",
			"error", err,
		)
		return "first round signing failed", err
	}

	// Second round: all signer infos are set, so each signer can sign.
	sigsV2 = []signing.SignatureV2{}
	for i, priv := range privs {
		signerData := authsigning.SignerData{
			ChainID:       config.DefaultChainId,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		sigV2, sigErr := tx.SignWithPrivKey(
			config.EncodingCg.TxConfig.SignModeHandler().DefaultMode(), signerData,
			txBuilder, priv, config.EncodingCg.TxConfig, accSeqs[i])
		if sigErr != nil {
			log.CustomLogger().Error("[TransferToken][SignWithPrivKey] second round signing failed.",
				"error", sigErr,
			)
			return "", sigErr
		}
		sigsV2 = append(sigsV2, sigV2)
	}

	err = txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		log.CustomLogger().Error("[TransferToken][SetSignatures] failed to setup signature.",
			"error", err,
		)
		return "failed to setup signature", err
	}

	txBytes, err := config.EncodingCg.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		log.CustomLogger().Error("[TransferToken][TxEncoder] failed to encode tx.",
			"error", err,
		)
		return "failed to encode tx", err
	}

	grpcConn, _ := grpc.Dial(
		config.DefaultGrpc,
		grpc.WithInsecure(),
	)
	defer grpcConn.Close()

	// Simulation
	simClient := txtypes.NewServiceClient(grpcConn)
	simResponse, simErr := simClient.Simulate(
		context.Background(),
		&txtypes.SimulateRequest{
			TxBytes: txBytes,
		},
	)
	if simErr != nil {
		log.CustomLogger().Error("[TransferToken][Simulate] failed to run simulation.",
			"error", simErr,
		)
		return "failed to run simulation", simErr
	}

	log.CustomLogger().Info("`TransferToken` Simulation have done successfully.",
		"simulation result", simResponse.GasInfo,
	)

	txClient := txtypes.NewServiceClient(grpcConn)
	grpcRes, breadcastErr := txClient.BroadcastTx(
		context.Background(),
		&txtypes.BroadcastTxRequest{
			Mode:    txtypes.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes,
		},
	)
	if breadcastErr != nil {
		log.CustomLogger().Error("[TransferToken][BroadcastTx] failed to broadcast tx.",
			"error", breadcastErr,
		)
		return "failed to broadcast tx", breadcastErr
	}

	log.CustomLogger().Info("`TransferToken` broadcasting transaction have sent successfully.",
		"transaction hash", grpcRes.TxResponse.TxHash,
	)
	return grpcRes.TxResponse.TxHash, nil
}
