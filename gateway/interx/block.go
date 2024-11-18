package interx

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/log"
	"github.com/KiraCore/interx/types"
	kiratypes "github.com/KiraCore/sekai/types"
	multistaking "github.com/KiraCore/sekai/x/multistaking/types"
	abciTypes "github.com/cometbft/cometbft/abci/types"
	tmTypes "github.com/cometbft/cometbft/rpc/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	distribution "github.com/cosmos/cosmos-sdk/x/distribution/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// RegisterBlockRoutes registers block/transaction query routers.
func RegisterBlockRoutes(r *mux.Router, gwCosmosmux *runtime.ServeMux, rpcAddr string) {
	r.HandleFunc(config.QueryBlocks, QueryBlocksRequest(gwCosmosmux, rpcAddr)).Methods("GET")
	r.HandleFunc(config.QueryBlockByHeightOrHash, QueryBlockByHeightOrHashRequest(gwCosmosmux, rpcAddr)).Methods("GET")
	r.HandleFunc(config.QueryBlockTransactions, QueryBlockTransactionsRequest(gwCosmosmux, rpcAddr)).Methods("GET")
	r.HandleFunc(config.QueryTransactionResult, QueryTransactionResultRequest(gwCosmosmux, rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryBlocks, "This is an API to query blocks by parameters.", true)
	common.AddRPCMethod("GET", config.QueryBlockByHeightOrHash, "This is an API to query block by height", true)
	common.AddRPCMethod("GET", config.QueryBlockTransactions, "This is an API to query block transactions by height", true)
	common.AddRPCMethod("GET", config.QueryTransactionResult, "This is an API to query transaction result by hash", true)
}

func queryBlocksHandle(rpcAddr string, r *http.Request) (interface{}, interface{}, int) {
	_ = r.ParseForm()

	minHeight := r.FormValue("minHeight")
	maxHeight := r.FormValue("maxHeight")

	var events = make([]string, 0, 2)

	if minHeight != "" {
		events = append(events, fmt.Sprintf("minHeight=%s", minHeight))
	}

	if maxHeight != "" {
		events = append(events, fmt.Sprintf("maxHeight=%s", maxHeight))
	}

	// search blocks

	return common.MakeTendermintRPCRequest(rpcAddr, "/blockchain", strings.Join(events, "&"))
}

// QueryBlocksRequest is a function to query Blocks.
func QueryBlocksRequest(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int

		log.CustomLogger().Info("Starting 'QueryBlocksRequest' request...")

		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		if !common.RPCMethods["GET"][config.QueryBlocks].Enabled {

			log.CustomLogger().Error(" `QueryBlocksRequest` is disabled.",
				"method", request.Method,
				"endpoint", request.Endpoint,
			)

			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryBlocks].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					log.CustomLogger().Info("Cache hit for 'QueryBlocksRequest' request.",
						"method", request.Method,
						"endpoint", request.Endpoint,
						"params", request.Params,
						"error", response.Error,
					)

					return
				}
			}

			response.Response, response.Error, statusCode = queryBlocksHandle(rpcAddr, r)
		}

		log.CustomLogger().Info("Processed 'QueryBlocksRequest' request.",
			"method", request.Method,
			"endpoint", request.Endpoint,
			"params", request.Params,
			"error", response.Error,
		)

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryBlocks].CachingEnabled)

		log.CustomLogger().Info("Finished 'QueryBlocksRequest' request.")
	}
}

func queryBlockByHeightOrHashHandle(rpcAddr string, height string) (interface{}, interface{}, int) {
	success, err, statusCode := common.MakeTendermintRPCRequest(rpcAddr, "/block", fmt.Sprintf("height=%s", height))

	if err != nil {
		log.CustomLogger().Error(" `queryBlockByHeightOrHashHandle` failed to execute.",
			"height", height,
			"method", "/block",
			"err", err,
		)
		success, err, statusCode = common.MakeTendermintRPCRequest(rpcAddr, "/block_by_hash", fmt.Sprintf("hash=%s", height))
	}

	return success, err, statusCode
}

// QueryBlockByHeightOrHashRequest is a function to query Blocks.
func QueryBlockByHeightOrHashRequest(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		queries := mux.Vars(r)
		height := queries["height"]
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		log.CustomLogger().Info("Starting `QueryBlockByHeightOrHashRequest` request...")

		if !common.RPCMethods["GET"][config.QueryBlockByHeightOrHash].Enabled {

			log.CustomLogger().Error("Query `QueryBlockByHeightOrHashRequest` is disabled.",
				"method", request.Method,
				"endpoint", request.Endpoint,
				"params", request.Params,
				"error", response.Error,
			)

			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryBlockByHeightOrHash].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					log.CustomLogger().Info("Cache hit for `QueryBlockByHeightOrHashRequest` request.",
						"method", request.Method,
						"endpoint", request.Endpoint,
						"params", request.Params,
						"error", response.Error,
					)

					return
				}
			}

			response.Response, response.Error, statusCode = queryBlockByHeightOrHashHandle(rpcAddr, height)
		}

		log.CustomLogger().Info("Processed `QueryBlockByHeightOrHashRequest` request.",
			"method", request.Method,
			"endpoint", request.Endpoint,
			"params", request.Params,
			"error", response.Error,
		)

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryBlockByHeightOrHash].CachingEnabled)

		log.CustomLogger().Info("Finished `QueryBlockByHeightOrHashRequest` request.")

	}
}

func getTransactionsFromLog(attributes []abciTypes.EventAttribute) []sdk.Coin {
	feeTxs := []sdk.Coin{}

	log.CustomLogger().Info("Starting `getTransactionsFromLog` request...")

	var evMap = make(map[string]string)
	for _, attribute := range attributes {
		key := string(attribute.GetKey())
		value := string(attribute.GetValue())

		if _, ok := evMap[key]; ok {
			coin, err := sdk.ParseCoinNormalized(evMap["amount"])
			if err == nil {
				feeTx := sdk.Coin{
					Amount: coin.Amount,
					Denom:  coin.GetDenom(),
				}
				feeTxs = append(feeTxs, feeTx)
			}

			evMap = make(map[string]string)
		}

		evMap[key] = value
	}

	if _, ok := evMap["amount"]; ok {
		coin, err := sdk.ParseCoinNormalized(evMap["amount"])
		if err == nil {
			feeTx := sdk.Coin{
				Amount: coin.Amount,
				Denom:  coin.GetDenom(),
			}
			feeTxs = append(feeTxs, feeTx)
		}
	}

	log.CustomLogger().Info("Finished `getTransactionsFromLog` request.")

	return feeTxs
}

func parseTransaction(rpcAddr string, transaction tmTypes.ResultTx) (types.TransactionResult, error) {
	txResult := types.TransactionResult{}

	log.CustomLogger().Info("Starting `parseTransaction` request...")

	tx, err := config.EncodingCg.TxConfig.TxDecoder()(transaction.Tx)
	if err != nil {

		log.CustomLogger().Error("Failed to decode transaction.",
			"method", "TxDecoder",
			"transaction", tx,
			"error", err,
			"result", txResult,
		)

		return txResult, err
	}

	txResult.Hash = hex.EncodeToString(transaction.Hash)
	txResult.Status = "Success"
	if transaction.TxResult.Code != 0 {
		txResult.Status = "Failure"
	}

	txResult.BlockHeight = transaction.Height
	txResult.BlockTimestamp, err = common.GetBlockTime(rpcAddr, transaction.Height)
	if err != nil {

		log.CustomLogger().Error("Failed to find block.",
			"method", "GetBlockTime",
			"RPC", rpcAddr,
			"height", transaction.Height,
			"error", err,
		)

	}
	txResult.Confirmation = common.NodeStatus.Block - transaction.Height + 1
	txResult.GasWanted = transaction.TxResult.GetGasWanted()
	txResult.GasUsed = transaction.TxResult.GetGasUsed()

	txSigning, ok := tx.(signing.Tx)
	if ok {
		txResult.Memo = txSigning.GetMemo()
	}

	txResult.Msgs = make([]types.TxMsg, 0)
	for _, msg := range tx.GetMsgs() {
		txResult.Msgs = append(txResult.Msgs, types.TxMsg{
			Type: kiratypes.MsgType(msg),
			Data: msg,
		})
	}

	log.CustomLogger().Info("Signing tx successfully done.",
		"method", "signing.Tx",
		"signed tx", txSigning,
	)

	txResult.Transactions = []types.Transaction{}
	txResult.Fees = []sdk.Coin{}

	logs, err := sdk.ParseABCILogs(transaction.TxResult.GetLog())
	if err != nil {
		return txResult, nil
	}

	for _, event := range transaction.TxResult.Events {
		if event.GetType() == "transfer" {
			txResult.Fees = append(txResult.Fees, getTransactionsFromLog(event.GetAttributes())...)
		}
	}

	for index, msg := range tx.GetMsgs() {
		log := logs[index]
		txType := kiratypes.MsgType(msg)
		transfers := []types.Transaction{}

		var evMap = make(map[string]([]sdk.Attribute))
		for _, event := range log.GetEvents() {
			evMap[event.GetType()] = event.GetAttributes()
		}

		if txType == "send" {
			msgSend := msg.(*bank.MsgSend)

			amounts := []sdk.Coin{}
			for _, coin := range msgSend.Amount {
				amounts = append(amounts, sdk.Coin{
					Denom:  coin.GetDenom(),
					Amount: coin.Amount,
				})
			}

			transfers = append(transfers, types.Transaction{
				Type:    txType,
				From:    msgSend.FromAddress,
				To:      msgSend.ToAddress,
				Amounts: amounts,
			})
		} else if txType == "multisend" {
			msgMultiSend := msg.(*bank.MsgMultiSend)
			inputs := msgMultiSend.GetInputs()
			outputs := msgMultiSend.GetOutputs()
			if len(inputs) == 1 && len(outputs) == 1 {
				input := inputs[0]
				output := outputs[0]
				amounts := []sdk.Coin{}

				for _, coin := range input.Coins {
					amounts = append(amounts, sdk.Coin{
						Denom:  coin.GetDenom(),
						Amount: coin.Amount,
					})
				}

				transfers = append(transfers, types.Transaction{
					Type:    txType,
					From:    input.Address,
					To:      output.Address,
					Amounts: amounts,
				})
			}
		} else if txType == "create_validator" {
			createValidatorMsg := msg.(*staking.MsgCreateValidator)

			transfers = append(transfers, types.Transaction{
				Type: txType,
				From: createValidatorMsg.DelegatorAddress,
				To:   createValidatorMsg.ValidatorAddress,
				Amounts: []sdk.Coin{
					{
						Denom:  createValidatorMsg.Value.Denom,
						Amount: createValidatorMsg.Value.Amount,
					},
				},
			})
		} else if txType == "delegate" {
			delegateMsg := msg.(*multistaking.MsgDelegate)

			transfers = append(transfers, types.Transaction{
				Type:    txType,
				From:    delegateMsg.DelegatorAddress,
				To:      delegateMsg.ValidatorAddress,
				Amounts: delegateMsg.Amounts,
			})
		} else if txType == "begin_redelegate" {
			reDelegateMsg := msg.(*staking.MsgBeginRedelegate)

			transfers = append(transfers, types.Transaction{
				Type: txType,
				From: reDelegateMsg.ValidatorSrcAddress,
				To:   reDelegateMsg.ValidatorDstAddress,
				Amounts: []sdk.Coin{
					{
						Denom:  reDelegateMsg.Amount.Denom,
						Amount: reDelegateMsg.Amount.Amount,
					},
				},
			})
		} else if txType == "begin_unbonding" {
			unDelegateMsg := msg.(*multistaking.MsgUndelegate)

			transfers = append(transfers, types.Transaction{
				Type:    txType,
				From:    unDelegateMsg.ValidatorAddress,
				To:      unDelegateMsg.DelegatorAddress,
				Amounts: unDelegateMsg.Amounts,
			})
		} else if txType == "withdraw_delegator_reward" {
			var coin sdk.Coin
			if v, found := evMap["withdraw_rewards"]; found && len(v) >= 2 {
				if v[0].GetKey() == "amount" {
					coin, _ = sdk.ParseCoinNormalized(v[0].Value)
				} else if v[1].GetKey() == "amount" {
					coin, _ = sdk.ParseCoinNormalized(v[1].Value)
				}
			}

			withdrawDelegatorRewardMsg := msg.(*distribution.MsgWithdrawDelegatorReward)

			transfers = append(transfers, types.Transaction{
				Type: txType,
				From: withdrawDelegatorRewardMsg.ValidatorAddress,
				To:   withdrawDelegatorRewardMsg.DelegatorAddress,
				Amounts: []sdk.Coin{
					{
						Denom:  coin.Denom,
						Amount: coin.Amount,
					},
				},
			})
		} else {
			attributes := []sdk.Attribute{}
			for _, event := range log.GetEvents() {
				if event.GetType() == "transfer" {
					attributes = event.GetAttributes()
				}
			}

			txs := []types.Transaction{}

			var evMap = make(map[string]string)
			for _, attribute := range attributes {
				key := string(attribute.GetKey())
				value := string(attribute.GetValue())

				if _, ok := evMap[key]; ok {
					coin, err := sdk.ParseCoinNormalized(evMap["amount"])
					if err == nil {
						txs = append(txs, types.Transaction{
							Type: txType,
							From: evMap["sender"],
							To:   evMap["recipient"],
							Amounts: []sdk.Coin{
								{
									Denom:  coin.Denom,
									Amount: coin.Amount,
								},
							},
						})
					}

					evMap = make(map[string]string)
				}

				evMap[key] = value
			}

			if _, ok := evMap["amount"]; ok {
				coin, err := sdk.ParseCoinNormalized(evMap["amount"])
				if err == nil {
					txs = append(txs, types.Transaction{
						Type: txType,
						From: evMap["sender"],
						To:   evMap["recipient"],
						Amounts: []sdk.Coin{
							{
								Denom:  coin.Denom,
								Amount: coin.Amount,
							},
						},
					})
				}
			}

			transfers = append(transfers, txs...)
		}

		for _, transfer := range transfers {
			for _, amount := range transfer.Amounts {
				i := 0
				for i < len(txResult.Fees) {
					if txResult.Fees[i].Amount.Equal(amount.Amount) && txResult.Fees[i].Denom == amount.Denom {
						break
					}
					i++
				}

				if i < len(txResult.Fees) {
					txResult.Fees = append(txResult.Fees[:i], txResult.Fees[i+1:]...)
					break
				}
			}
		}

		txResult.Transactions = append(txResult.Transactions, transfers...)

	}

	log.CustomLogger().Info("Finished `parseTransaction` request.")

	return txResult, nil
}

// QueryBlockTransactionsHandle is a function to query transactions of a block.
func QueryBlockTransactionsHandle(rpcAddr string, height string) (interface{}, interface{}, int) {
	blockHeight, _ := strconv.Atoi(height)
	response, err := SearchTxHashHandle(rpcAddr, "", "", "", 0, 0, int64(blockHeight), int64(blockHeight), "")
	if err != nil {
		log.CustomLogger().Error("`QueryBlockTransactionsHandle` failed to execute.",
			"block_height", height,
			"error", err,
			"response_data", response,
		)
		return common.ServeError(0, "transaction query failed", "", http.StatusBadRequest)
	}

	searchResult := types.TransactionSearchResult{}

	searchResult.TotalCount = response.TotalCount
	searchResult.Txs = []types.TransactionResult{}

	for _, transaction := range response.Txs {
		txResult, err := parseTransaction(rpcAddr, *transaction)
		if err != nil {
			return common.ServeError(0, "", err.Error(), http.StatusBadRequest)
		}

		searchResult.Txs = append(searchResult.Txs, txResult)
	}

	return searchResult, nil, http.StatusOK
}

// QueryBlockTransactionsRequest is a function to query transactions of a block.
func QueryBlockTransactionsRequest(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		queries := mux.Vars(r)
		height := queries["height"]

		log.CustomLogger().Info("Starting `QueryBlockTransactionsRequest`.",
			"block_height", height,
		)

		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		log.CustomLogger().Info("Attempting to fetch transactions from block.",
			"block_height", height,
			"method", request.Method,
			"endpoint", request.Endpoint,
			"params", request.Params,
		)

		if !common.RPCMethods["GET"][config.QueryBlockTransactions].Enabled {

			log.CustomLogger().Error("`QueryBlockTransactionsRequest` is disabled.",
				"block_height", height,
			)

			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryBlockTransactions].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					log.CustomLogger().Info("Cache hit for `QueryBlockTransactionsRequest`.",
						"block_height", height,
					)

					return
				}
			}

			response.Response, response.Error, statusCode = QueryBlockTransactionsHandle(rpcAddr, height)
		}

		log.CustomLogger().Info("Fetched transaction from block (no cache used).",
			"block_height", height,
			"transaction_response", response.Response,
			"error_details", response.Error,
		)
		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryBlockTransactions].CachingEnabled)

		log.CustomLogger().Info("Completed `QueryBlockTransactionsRequest`.",
			"block_height", height,
			"status_code", statusCode,
		)
	}
}

// QueryTransactionResultHandle is a function to query transaction by a given hash.
func QueryTransactionResultHandle(rpcAddr string, txHash string) (interface{}, interface{}, int) {
	txHash = strings.TrimPrefix(txHash, "0x")

	response, err := SearchTxHashHandle(rpcAddr, "", "", "", 0, 0, 0, 0, txHash)
	if err != nil {
		log.CustomLogger().Error("`QueryTransactionResultHandle` failed to execute.",
			"tx_Hash", txHash,
			"error", err,
			"response_data", response,
		)
		return common.ServeError(0, "transaction query failed", "", http.StatusBadRequest)
	}

	txResult := types.TransactionResult{}

	for _, transaction := range response.Txs {
		txResult, err = parseTransaction(rpcAddr, *transaction)
		if err != nil {
			log.CustomLogger().Error("`parseTransaction` failed to execute.",
				"tx_Result", txResult,
				"error", err,
				"response_data", response.Txs,
			)
			return common.ServeError(0, "", err.Error(), http.StatusBadRequest)
		}
	}

	return txResult, nil, http.StatusOK
}

// QueryTransactionResultRequest is a function to query transactions by a given hash.
func QueryTransactionResultRequest(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		queries := mux.Vars(r)
		txHash := queries["txHash"]

		log.CustomLogger().Info("Starting `QueryTransactionResultRequest` request...")

		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		log.CustomLogger().Info("Attempting to fetch transactions by hash.",
			"tx_Hash", txHash,
			"method", request.Method,
			"endpoint", request.Endpoint,
			"params", request.Params,
		)

		if !common.RPCMethods["GET"][config.QueryTransactionResult].Enabled {

			log.CustomLogger().Error("`QueryTransactionResultRequest` is disabled.",
				"tx_Hash", txHash,
			)
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryTransactionResult].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					log.CustomLogger().Info("Cache hit for `QueryTransactionResultRequest`.",
						"tx_Hash", txHash,
					)
					return
				}
			}

			response.Response, response.Error, statusCode = QueryTransactionResultHandle(rpcAddr, txHash)
			log.CustomLogger().Info("Fetched transaction by hash (no cache used).",
				"tx_Hash", txHash,
				"transaction_response", response.Response,
				"error_details", response.Error,
			)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryTransactionResult].CachingEnabled)

		log.CustomLogger().Info("Completed `QueryBlockTransactionsRequest`.",
			"tx_Hash", txHash,
			"status_code", statusCode,
		)
	}
}
