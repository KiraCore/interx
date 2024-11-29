package evm

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/log"
	"github.com/gorilla/mux"
	"github.com/holiman/uint256"

	// "github.com/powerman/rpc-codec/jsonrpc2"
	jsonrpc2 "github.com/KeisukeYamashita/go-jsonrpc"
)

type EVMBalance struct {
	Contract string `json:"contract"`
	Amount   string `json:"amount"`
	Symbol   string `json:"symbol"`
	Decimals uint64 `json:"decimals"`
}

type EVMBalancesResponse struct {
	Balances []EVMBalance `json:"balances"`
}

type EVMCall struct {
	To   string `json:"to"`
	Data string `json:"data"`
}

// RegisterEVMBalancesRoutes registers query status of EVM chains.
func RegisterEVMBalancesRoutes(r *mux.Router, rpcAddr string) {
	r.HandleFunc(config.QueryEVMBalances, RegisterEVMBalancesRequest(rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryEVMBalances, "This is an API to query balances infomation.", true)
}

func queryEVMBalancesFromNode(nodeInfo config.EVMNodeConfig, address string, tokens []string) (interface{}, interface{}, int) {
	client := jsonrpc2.NewRPCClient(nodeInfo.RPC + "/" + nodeInfo.RPCToken)
	if nodeInfo.RPCSecret != "" {
		client.SetBasicAuth(nodeInfo.RPCToken, nodeInfo.RPCSecret)
	}

	response := new(EVMBalancesResponse)

	balance := new(EVMBalance)

	// eth balance
	balance.Contract = ""
	balance.Symbol = "ETH"
	balance.Decimals = 18
	data, err := client.Call("eth_getBalance", address, "latest")
	if err != nil {
		return common.ServeError(0, "failed to get eth balance ", err.Error(), http.StatusInternalServerError)
	}
	ethBalance, err := data.GetString()
	if err != nil {
		return common.ServeError(0, "failed to get eth balance ", err.Error(), http.StatusInternalServerError)
	}
	amount := *new(big.Int)
	amount.SetString((ethBalance)[2:], 16)
	balanceAmount, _ := uint256.FromBig(&amount)
	balance.Amount = fmt.Sprintf("%d", balanceAmount)
	response.Balances = append(response.Balances, *balance)

	// token balances
	for _, token := range tokens {
		balance.Contract = token

		call := new(EVMCall)
		call.To = token
		call.Data = "0x95d89b410000000000000000000000000000000000000000000000000000000000000000" // symbol
		data, err = client.Call("eth_call", *call, "latest")
		if err != nil {
			return common.ServeError(0, "failed to get token symbol", err.Error(), http.StatusInternalServerError)
		}
		balance.Symbol, err = data.GetString()
		if err != nil || balance.Symbol == "0x" {
			return common.ServeError(0, "failed to get token symbol", "", http.StatusInternalServerError)
		}
		symbol, err := hex.DecodeString(strings.ReplaceAll(balance.Symbol[2:], "00", ""))
		if err != nil {
			return common.ServeError(0, "failed to decode token symbol", "", http.StatusInternalServerError)
		}

		balance.Symbol = string(symbol[1:])

		call.To = token
		call.Data = "0x313ce5670000000000000000000000000000000000000000000000000000000000000000" // decimals
		data, err = client.Call("eth_call", *call, "latest")
		if err != nil {
			return common.ServeError(0, "failed to get token decimals", err.Error(), http.StatusInternalServerError)
		}
		tokenDecimals, err := data.GetString()
		if err != nil || tokenDecimals == "0x" {
			return common.ServeError(0, "failed to get token decimals", "", http.StatusInternalServerError)
		}
		balance.Decimals, _ = strconv.ParseUint((tokenDecimals)[2:], 16, 64)

		call.To = token
		call.Data = "0x70a08231000000000000000000000000" + address[2:] // balanceOf
		data, err = client.Call("eth_call", *call, "latest")
		if err != nil {
			return common.ServeError(0, "failed to get token balances", err.Error(), http.StatusInternalServerError)
		}
		tokenBalances, err := data.GetString()
		if err != nil {
			return common.ServeError(0, "failed to get token balances", err.Error(), http.StatusInternalServerError)
		}
		if tokenBalances == "0x" {
			balance.Amount = "0"
		} else {
			amount := *new(big.Int)
			amount.SetString((tokenBalances)[2:], 16)
			balanceAmount, _ := uint256.FromBig(&amount)
			balance.Amount = fmt.Sprintf("%d", balanceAmount)
		}
		response.Balances = append(response.Balances, *balance)
	}

	return response, nil, http.StatusOK
}

func queryEVMBalancesRequestHandle(r *http.Request, chain string, address string) (interface{}, interface{}, int) {
	isSupportedChain, chainConfig := GetChainConfig(chain)
	if !isSupportedChain {
		return common.ServeError(0, "", "unsupported chain", http.StatusBadRequest)
	}

	_ = r.ParseForm()
	tokens := new([]string)

	if len(r.FormValue("tokens")) > 0 {
		*tokens = strings.Split(r.FormValue("tokens"), ",")
	}

	res, err, statusCode := queryEVMBalancesFromNode(chainConfig.QuickNode, address, *tokens)
	if err == nil {
		return res, err, statusCode
	}

	res, err, statusCode = queryEVMBalancesFromNode(chainConfig.Infura, address, *tokens)
	if err == nil {
		return res, err, statusCode
	}

	return queryEVMBalancesFromNode(chainConfig.Pokt, address, *tokens)
}

// RegisterEVMBalancesRequest is a function to query evm account balances infomation
func RegisterEVMBalancesRequest(rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		queries := mux.Vars(r)
		chain := queries["chain"]
		address := queries["address"]
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		log.CustomLogger().Info("`RegisterEVMBalancesRequest` Starting EVM balance request...",
			"chain", chain,
		)

		if !common.RPCMethods["GET"][config.QueryEVMBalances].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryEVMBalances].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					log.CustomLogger().Info("`RegisterEVMBalancesRequest` Returning from the cache...",
						"chain", chain,
					)

					return
				}
			}

			response.Response, response.Error, statusCode = queryEVMBalancesRequestHandle(r, chain, address)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryEVMBalances].CachingEnabled)
	}
}
