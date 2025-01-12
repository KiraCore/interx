package evm

import (
	"net/http"
	"strconv"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/log"
	"github.com/gorilla/mux"

	// "github.com/powerman/rpc-codec/jsonrpc2"
	jsonrpc2 "github.com/KeisukeYamashita/go-jsonrpc"
)

type EVMAccount struct {
	Type     string `json:"@type"`
	Address  string `json:"address"`
	Pending  uint64 `json:"pending"`
	Sequence uint64 `json:"sequence"`
}

type EVMAccountResponse struct {
	Account      EVMAccount `json:"account"`
	ContractCode string     `json:"contract_code"`
}

// RegisterEVMAccountsRoutes registers query status of EVM chains.
func RegisterEVMAccountsRoutes(r *mux.Router, rpcAddr string) {
	r.HandleFunc(config.QueryEVMAccounts, RegisterEVMAccountsRequest(rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryEVMAccounts, "This is an API to query accounts infomation.", true)
}

func queryEVMAccountsFromNode(nodeInfo config.EVMNodeConfig, address string) (interface{}, interface{}, int) {
	client := jsonrpc2.NewRPCClient(nodeInfo.RPC + "/" + nodeInfo.RPCToken)
	if nodeInfo.RPCSecret != "" {
		client.SetBasicAuth(nodeInfo.RPCToken, nodeInfo.RPCSecret)
	}

	response := new(EVMAccountResponse)

	response.Account.Type = "wallet"
	response.Account.Address = address

	data, err := client.Call("eth_getCode", address, "latest")
	if err == nil {
		response.ContractCode, err = data.GetString()
		if err != nil {
			return common.ServeError(0, "failed to get code", err.Error(), http.StatusInternalServerError)
		}

		response.Account.Type = "contract"
	}

	data, err = client.Call("eth_getTransactionCount", address, "latest")
	if err != nil {
		return common.ServeError(0, "failed to get latest sequence", err.Error(), http.StatusInternalServerError)
	}
	sequence, err := data.GetString()
	response.Account.Sequence, _ = strconv.ParseUint((sequence)[2:], 16, 64)
	if err != nil {
		return common.ServeError(0, "failed to get latest sequence", err.Error(), http.StatusInternalServerError)
	}

	data, err = client.Call("eth_getTransactionCount", address, "pending")
	if err != nil {
		return common.ServeError(0, "failed to get pending sequence", err.Error(), http.StatusInternalServerError)
	}
	sequence, err = data.GetString()
	response.Account.Pending, _ = strconv.ParseUint((sequence)[2:], 16, 64)
	if err != nil {
		return common.ServeError(0, "failed to get pending sequence", err.Error(), http.StatusInternalServerError)
	}

	return response, nil, http.StatusOK
}

func queryEVMAccountsRequestHandle(r *http.Request, chain string, address string) (interface{}, interface{}, int) {
	isSupportedChain, chainConfig := GetChainConfig(chain)
	if !isSupportedChain {
		return common.ServeError(0, "", "unsupported chain", http.StatusBadRequest)
	}

	res, err, statusCode := queryEVMAccountsFromNode(chainConfig.QuickNode, address)
	if err == nil {
		return res, err, statusCode
	}

	res, err, statusCode = queryEVMAccountsFromNode(chainConfig.Infura, address)
	if err == nil {
		return res, err, statusCode
	}

	return queryEVMAccountsFromNode(chainConfig.Pokt, address)
}

// RegisterEVMAccountsRequest is a function to query evm account infomation
func RegisterEVMAccountsRequest(rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		queries := mux.Vars(r)
		chain := queries["chain"]
		address := queries["address"]
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		log.CustomLogger().Info("`RegisterEVMAccountsRequest` Starting EVM Account request...",
			"chain", chain,
		)

		if !common.RPCMethods["GET"][config.QueryEVMAccounts].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryEVMAccounts].CacheEnabled {

				log.CustomLogger().Info("Starting search cache for `RegisterEVMAccountsRequest` request...")

				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					log.CustomLogger().Info("`RegisterEVMAccountsRequest` Returning from the cache...",
						"chain", chain,
					)

					return
				}
			}

			response.Response, response.Error, statusCode = queryEVMAccountsRequestHandle(r, chain, address)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryEVMAccounts].CacheEnabled)
	}
}
