package evm

import (
	"net/http"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/log"
	"github.com/gorilla/mux"

	// "github.com/powerman/rpc-codec/jsonrpc2"
	jsonrpc2 "github.com/KeisukeYamashita/go-jsonrpc"
)

// RegisterEVMTransactionRoutes registers query status of EVM chains.
func RegisterEVMTransactionRoutes(r *mux.Router, rpcAddr string) {
	r.HandleFunc(config.QueryEVMTransaction, QueryEVMTransactionRequest(rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryEVMTransaction, "This is an API to query transactions.", true)
}

func queryEVMTransactionFromNode(nodeInfo config.EVMNodeConfig, transactionHash string, receipt bool) (interface{}, interface{}, int) {
	client := jsonrpc2.NewRPCClient(nodeInfo.RPC + "/" + nodeInfo.RPCToken)
	if nodeInfo.RPCSecret != "" {
		client.SetBasicAuth(nodeInfo.RPCToken, nodeInfo.RPCSecret)
	}

	response := new(interface{})

	if !receipt {
		data, err := client.Call("eth_getTransactionByHash", transactionHash)
		if err != nil {
			return common.ServeError(0, "failed to get transaction by hash", err.Error(), http.StatusInternalServerError)
		}
		err = data.GetObject(response)
		if err != nil {
			return common.ServeError(0, "failed to get transaction by hash", err.Error(), http.StatusInternalServerError)
		}
	} else {
		data, err := client.Call("eth_getTransactionReceipt", transactionHash)
		if err != nil {
			return common.ServeError(0, "failed to get transaction receipt", err.Error(), http.StatusInternalServerError)
		}
		err = data.GetObject(response)
		if err != nil {
			return common.ServeError(0, "failed to get transaction receipt", err.Error(), http.StatusInternalServerError)
		}
	}

	return response, nil, http.StatusOK
}

func queryEVMTransactionRequestHandle(r *http.Request, chain string, transactionHash string) (interface{}, interface{}, int) {
	_ = r.ParseForm()
	receipt := r.FormValue("receipt") == "true"

	isSupportedChain, chainConfig := GetChainConfig(chain)
	if !isSupportedChain {
		return common.ServeError(0, "", "unsupported chain", http.StatusBadRequest)
	}
	res, err, statusCode := queryEVMTransactionFromNode(chainConfig.QuickNode, transactionHash, receipt)
	if err == nil {
		return res, err, statusCode
	}

	res, err, statusCode = queryEVMTransactionFromNode(chainConfig.Infura, transactionHash, receipt)
	if err == nil {
		return res, err, statusCode
	}

	return queryEVMTransactionFromNode(chainConfig.Pokt, transactionHash, receipt)
}

// QueryEVMTransactionRequest is a function to query transaction of EVM chains.
func QueryEVMTransactionRequest(rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		queries := mux.Vars(r)
		chain := queries["chain"]
		transactionHash := queries["hash"]
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		log.CustomLogger().Info("[query-evm-transaction] Entering transaction query: ", chain)

		if !common.RPCMethods["GET"][config.QueryEVMTransaction].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryEVMTransaction].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					log.CustomLogger().Info("[query-evm-transaction] Returning from the cache: ", chain)
					return
				}
			}

			response.Response, response.Error, statusCode = queryEVMTransactionRequestHandle(r, chain, transactionHash)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryEVMTransaction].CachingEnabled)
	}
}
