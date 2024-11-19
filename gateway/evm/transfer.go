package evm

import (
	"encoding/hex"
	"log"
	"math/big"
	"net/http"
	"strconv"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	interxLog "github.com/KiraCore/interx/log"
	"github.com/gorilla/mux"

	// "github.com/powerman/rpc-codec/jsonrpc2"
	jsonrpc2 "github.com/KeisukeYamashita/go-jsonrpc"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

type EVMTransaction struct {
	Nonce     uint64 `json:"nonce"`
	GasLimit  uint64 `json:"gasLimit"`
	GasPrice  string `json:"gasPrice"`
	GasFeeCap string `json:"gasFeeCap"`
	GasTipCap string `json:"gasTipCap"`
	From      string `json:"from"`
	To        string `json:"to"`
	Value     string `json:"value"`
	Data      string `json:"data"`
}

type EVMTransferQueryResponse struct {
	Gas             uint64         `json:"gas"`
	GasPrice        string         `json:"gasPrice"`
	SequenceLatest  uint64         `json:"sequence_latest"`
	SequencePending uint64         `json:"sequence_pending"`
	DeocdedTx       EVMTransaction `json:"decoded_tx"`
}

// RegisterEVMTransferRoutes registers query status of EVM chains.
func RegisterEVMTransferRoutes(r *mux.Router, rpcAddr string) {
	r.HandleFunc(config.QueryEVMTransfer, QueryEVMTransferRequest(rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryEVMTransfer, "This is an API to broadcast a signed transaction.", true)
}

func queryEVMTransferFromNode(nodeInfo config.EVMNodeConfig, rawTx string, estimate bool, evmTransaction EVMTransaction) (interface{}, interface{}, int) {
	client := jsonrpc2.NewRPCClient(nodeInfo.RPC + "/" + nodeInfo.RPCToken)
	if nodeInfo.RPCSecret != "" {
		client.SetBasicAuth(nodeInfo.RPCToken, nodeInfo.RPCSecret)
	}

	response := new(EVMTransferQueryResponse)
	response.DeocdedTx = evmTransaction
	response.Gas = evmTransaction.GasLimit
	response.GasPrice = evmTransaction.GasPrice

	data, err := client.Call("eth_getTransactionCount", evmTransaction.From, "latest")
	if err != nil {
		return common.ServeError(0, "failed to get latest sequence", err.Error(), http.StatusInternalServerError)
	}
	sequence, err := data.GetString()
	response.SequenceLatest, _ = strconv.ParseUint((sequence)[2:], 16, 64)
	if err != nil {
		return common.ServeError(0, "failed to get latest sequence", err.Error(), http.StatusInternalServerError)
	}

	data, err = client.Call("eth_getTransactionCount", evmTransaction.From, "pending")
	if err != nil {
		return common.ServeError(0, "failed to get pending sequence", err.Error(), http.StatusInternalServerError)
	}
	sequence, err = data.GetString()
	response.SequencePending, _ = strconv.ParseUint((sequence)[2:], 16, 64)
	if err != nil {
		return common.ServeError(0, "failed to get pending sequence", err.Error(), http.StatusInternalServerError)
	}

	if estimate {
		return response, nil, http.StatusOK
	}

	transactionResult := new(interface{})
	data, err = client.Call("eth_sendRawTransaction", rawTx)
	if err != nil {
		return common.ServeError(0, "failed to send transaction", err.Error(), http.StatusInternalServerError)
	}
	if data.Error != nil {
		return common.ServeError(0, "failed to send transaction", data.Error.Error(), http.StatusInternalServerError)
	}
	err = data.GetObject(transactionResult)
	if err != nil {
		return common.ServeError(0, "failed to send transaction", err.Error(), http.StatusInternalServerError)
	}

	return transactionResult, nil, http.StatusOK
}

func queryEVMTransferRequestHandle(r *http.Request, chain string) (interface{}, interface{}, int) {
	_ = r.ParseForm()
	estimate := r.FormValue("estimate") == "true"
	rawTx := r.FormValue("rawTx")

	evmTransaction := new(EVMTransaction)
	{
		rawTxData, err := hex.DecodeString(rawTx[2:])
		if err != nil {
			return common.ServeError(0, "failed to decode raw transaction", err.Error(), http.StatusBadRequest)
		}

		var tx types.Transaction
		err = rlp.DecodeBytes(rawTxData, &tx)
		if err != nil {
			return common.ServeError(0, "failed to decode raw transaction", err.Error(), http.StatusBadRequest)
		}
		msg, err := tx.AsMessage(types.NewEIP155Signer(tx.ChainId()), big.NewInt(0))
		if err != nil {
			log.Fatal(err)
		}

		evmTransaction.Nonce = tx.Nonce()
		evmTransaction.GasLimit = tx.Gas()
		evmTransaction.GasPrice = tx.GasPrice().String()
		evmTransaction.GasFeeCap = tx.GasFeeCap().String()
		evmTransaction.GasTipCap = tx.GasTipCap().String()
		evmTransaction.From = msg.From().String()
		evmTransaction.To = tx.To().String()
		evmTransaction.Value = tx.Value().String()
		evmTransaction.Data = hex.EncodeToString(tx.Data())
	}

	isSupportedChain, chainConfig := GetChainConfig(chain)
	if !isSupportedChain {
		return common.ServeError(0, "", "unsupported chain", http.StatusBadRequest)
	}

	res, err, statusCode := queryEVMTransferFromNode(chainConfig.QuickNode, rawTx, estimate, *evmTransaction)
	if err == nil {
		return res, err, statusCode
	}

	res, err, statusCode = queryEVMTransferFromNode(chainConfig.Infura, rawTx, estimate, *evmTransaction)
	if err == nil {
		return res, err, statusCode
	}

	return queryEVMTransferFromNode(chainConfig.Pokt, rawTx, estimate, *evmTransaction)
}

// QueryEVMTransferRequest is a function to query transfer of EVM chains.
func QueryEVMTransferRequest(rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		queries := mux.Vars(r)
		chain := queries["chain"]
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		interxLog.CustomLogger().Info("[query-evm-transfer] Entering transactions execute: ", chain)

		if !common.RPCMethods["GET"][config.QueryEVMTransfer].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryEVMTransfer].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					interxLog.CustomLogger().Info("[query-evm-transfer] Returning from the cache: ", chain)
					return
				}
			}

			response.Response, response.Error, statusCode = queryEVMTransferRequestHandle(r, chain)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryEVMTransfer].CachingEnabled)
	}
}
