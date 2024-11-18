package bitcoin

import (
	"net/http"
	"strings"

	jsonrpc2 "github.com/KeisukeYamashita/go-jsonrpc"
	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/gorilla/mux"
)

// RegisterBitcoinFaucetRoutes registers faucet services.
func RegisterBitcoinFaucetRoutes(r *mux.Router, rpcAddr string) {
	r.HandleFunc(config.QueryBitcoinFaucet, BitcoinFaucetRequest(rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryBitcoinFaucet, "This is an API to claim faucet tokens.", true)
}

type BitcoinFaucetInfo struct {
	Address string  `json:"address"`
	Balance float64 `json:"balance"`
}

func bitcoinFaucetInfo(r *http.Request, chain string) (interface{}, interface{}, int) {
	isSupportedChain, conf := GetChainConfig(chain)
	if !isSupportedChain {
		return common.ServeError(0, "", "unsupported chain", http.StatusBadRequest)
	}

	if len(conf.BTC_FAUCET) == 0 {
		return common.ServeError(0, "", "faucet address is not set", http.StatusInternalServerError)
	}

	result, err, status := queryBtcBalancesRequestHandle(r, chain, conf.BTC_FAUCET)
	if status != http.StatusOK || err != nil {
		return result, err, status
	}

	info := BitcoinFaucetInfo{
		Address: conf.BTC_FAUCET,
		Balance: result.(BalancesResult).Balance.Confirmed,
	}
	return info, err, status
}

func bitcoinFaucetHandle(r *http.Request, chain string, addr string) (interface{}, interface{}, int) {
	isSupportedChain, conf := GetChainConfig(chain)
	if !isSupportedChain {
		return common.ServeError(0, "", "unsupported chain", http.StatusBadRequest)
	}

	client := jsonrpc2.NewRPCClient(conf.RPC)
	if conf.RPC_CRED != "" {
		rpcInfo := strings.Split(conf.RPC_CRED, ":")
		client.SetBasicAuth(rpcInfo[0], rpcInfo[1])
	}

	response := new(interface{})

	return response, nil, http.StatusOK
}

// BitcoinFaucetRequest is a function to claim tokens from faucet account.
func BitcoinFaucetRequest(rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chain := "testnet"
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)
		statusCode := http.StatusOK

		queries := r.URL.Query()
		claimAddr := queries["claim"]

		common.GetLogger().Info("[bitcoin-faucet] Entering faucet request: ", chain)

		if len(claimAddr) != 0 {
			response.Response, response.Error, statusCode = bitcoinFaucetHandle(r, chain, claimAddr[0])
		} else {
			response.Response, response.Error, statusCode = bitcoinFaucetInfo(r, chain)
		}

		common.WrapResponse(w, request, *response, statusCode, false)
	}
}
