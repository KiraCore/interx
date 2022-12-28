package bitcoin

// import (
// 	"net/http"
// 	"strings"

// 	jsonrpc2 "github.com/KeisukeYamashita/go-jsonrpc"
// 	"github.com/KiraCore/interx/common"
// 	"github.com/KiraCore/interx/config"
// 	"github.com/gorilla/mux"
// )

// // RegisterBitcoinFaucetRoutes registers faucet services.
// func RegisterBitcoinFaucetRoutes(r *mux.Router, rpcAddr string) {
// 	r.HandleFunc(config.BitcoinFaucet, BitcoinFaucetRequest(rpcAddr)).Methods("GET")

// 	common.AddRPCMethod("GET", config.BitcoinFaucet, "This is an API to claim faucet tokens.", true)
// }

// type BitcoinFaucetInfo struct {
// 	Address string  `json:"address"`
// 	Balance float64 `json:"balance"`
// }

// func bitcoinFaucetInfo(r *http.Request, chain string, addr string) (interface{}, interface{}, int) {
// 	result, err, status := queryBtcBalancesRequestHandle(r, chain, addr)

// 	if status == http.StatusOK && err == nil {
// 		return
// 	}
// 	return queryBtcBalancesRequestHandle(r, chain, addr)
// }

// func bitcoinFaucetHandle(r *http.Request, chain string) (interface{}, interface{}, int) {
// 	isSupportedChain, conf := GetChainConfig(chain)
// 	if !isSupportedChain {
// 		return common.ServeError(0, "", "unsupported chain", http.StatusBadRequest)
// 	}

// 	client := jsonrpc2.NewRPCClient(conf.RPC)
// 	if conf.RPC_CRED != "" {
// 		rpcInfo := strings.Split(conf.RPC_CRED, ":")
// 		client.SetBasicAuth(rpcInfo[0], rpcInfo[1])
// 	}

// 	response := BitcoinFaucet{}

// 	return response, nil, http.StatusOK
// }

// // BitcoinFaucetRequest is a function to claim tokens from faucet account.
// func BitcoinFaucetRequest(rpcAddr string) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		chain := "testnet"
// 		request := common.GetInterxRequest(r)
// 		response := common.GetResponseFormat(request, rpcAddr)
// 		statusCode := http.StatusOK

// 		queries := r.URL.Query()
// 		claimAddr := queries["claim"]

// 		common.GetLogger().Info("[bitcoin-faucet] Entering faucet request: ", chain)

// 		if len(claimAddr) == 0 {
// 			response.Response, response.Error, statusCode = bitcoinFaucetInfo(r, chain, claimAddr)
// 		} else {
// 			response.Response, response.Error, statusCode = bitcoinFaucetHandle(r, chain)
// 		}

// 		common.WrapResponse(w, request, *response, statusCode, false)
// 	}
// }
