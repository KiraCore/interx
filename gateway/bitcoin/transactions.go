package bitcoin

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/gorilla/mux"

	// "github.com/powerman/rpc-codec/jsonrpc2"
	jsonrpc2 "github.com/KeisukeYamashita/go-jsonrpc"
)

type SearchRawTransactionsResult struct {
	Hex      string `json:"hex,omitempty"`
	Txid     string `json:"txid"`
	Hash     string `json:"hash"`
	Size     uint64 `json:"size"`
	Vsize    uint64 `json:"vsize"`
	Weight   uint64 `json:"weight"`
	Version  int32  `json:"version"`
	LockTime uint32 `json:"locktime"`
	Vin      []struct {
		Coinbase  string `json:"coinbase,omitempty"`
		Txid      string `json:"txid"`
		Vout      uint32 `json:"vout"`
		ScriptSig *struct {
			DecodeScriptResult
			Asm string `json:"asm"`
			Hex string `json:"hex,omitempty"`
		} `json:"scriptSig"`
		Witness  []string         `json:"txinwitness,omitempty"`
		PrevOut  *btcjson.PrevOut `json:"prevOut,omitempty"`
		Sequence uint32           `json:"sequence"`
	} `json:"vin"`
	Vout []struct {
		Value        float64 `json:"value"`
		N            uint32  `json:"n"`
		ScriptPubKey struct {
			Asm       string   `json:"asm"`
			Hex       string   `json:"hex,omitempty"`
			ReqSigs   int32    `json:"reqSigs,omitempty"`
			Type      string   `json:"type"`
			Addresses []string `json:"addresses,omitempty"`
		} `json:"scriptPubKey"`
	} `json:"vout"`
	BlockHash          string `json:"blockhash,omitempty"`
	Confirmations      uint64 `json:"confirmations,omitempty"`
	BlockConfirmations string `json:"blockconfirmations,omitempty"`
	Time               int64  `json:"time,omitempty"`
	Blocktime          int64  `json:"blocktime,omitempty"`
}

// RegisterBtcTransactionRoutes registers query status of bitcoin chains.
func RegisterBtcTransactionRoutes(r *mux.Router, rpcAddr string) {
	r.HandleFunc(config.QueryBitcoinTransaction, QueryBtcTransactionRequest(rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryBitcoinTransaction, "This is an API to query transactions.", true)
}

func queryBtcTransactionRequestHandle(r *http.Request, chain string, transactionHash string) (interface{}, interface{}, int) {
	isSupportedChain, conf := GetChainConfig(chain)
	if !isSupportedChain {
		return common.ServeError(0, "", "unsupported chain", http.StatusBadRequest)
	}

	client := jsonrpc2.NewRPCClient(conf.RPC)
	if conf.RPC_CRED != "" {
		rpcInfo := strings.Split(conf.RPC_CRED, ":")
		client.SetBasicAuth(rpcInfo[0], rpcInfo[1])
	}

	response := SearchRawTransactionsResult{}
	err := GetResult(client, "getrawtransaction", &response, transactionHash, true, nil)
	if err != nil {
		return common.ServeError(0, "failed to get transaction by hash", err.Error(), http.StatusInternalServerError)
	}

	for idx, vin := range response.Vin {
		res := DecodeScriptResult{}
		if vin.ScriptSig != nil {
			err = GetResult(client, "decodescript", &res, vin.ScriptSig.Hex)
			if err != nil {
				return common.ServeError(0, "failed to get transaction by hash", err.Error(), http.StatusInternalServerError)
			}
			response.Vin[idx].ScriptSig.DecodeScriptResult = res
			response.Vin[idx].ScriptSig.Desc = ""
			response.Vin[idx].ScriptSig.Segwit.Desc = ""
			response.Vin[idx].ScriptSig.Segwit.Hex = ""
			response.Vin[idx].ScriptSig.Segwit.ReqSigs = 1
			response.Vin[idx].ScriptSig.Segwit.Addresses = []string{response.Vin[idx].ScriptSig.Segwit.Address}
			response.Vin[idx].ScriptSig.Segwit.Address = ""
			response.Vin[idx].ScriptSig.Hex = ""
		}
	}

	for idx, vout := range response.Vout {
		res := DecodeScriptResult{}
		err = GetResult(client, "decodescript", &res, vout.ScriptPubKey.Hex)
		if err != nil {
			return common.ServeError(0, "failed to get transaction by hash", err.Error(), http.StatusInternalServerError)
		}

		response.Vout[idx].ScriptPubKey.Hex = ""
		response.Vout[idx].ScriptPubKey.ReqSigs = 1
		response.Vout[idx].ScriptPubKey.Addresses = append(response.Vout[idx].ScriptPubKey.Addresses, res.Segwit.Address)
	}

	response.Hex = ""
	if response.Confirmations > conf.BTC_CONFIRMATIONS {
		response.BlockConfirmations = strconv.Itoa(int(conf.BTC_CONFIRMATIONS)) + "+"
	} else {
		response.BlockConfirmations = strconv.Itoa(int(response.Confirmations))
	}
	response.Confirmations = 0
	return response, nil, http.StatusOK
}

// QueryBtcTransactionRequest is a function to query transaction of Bitcoin chains.
func QueryBtcTransactionRequest(rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queries := mux.Vars(r)
		chain := "testnet"
		transactionHash := queries["hash"]
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)
		statusCode := http.StatusOK

		common.GetLogger().Info("[query-btc-transaction] Entering transaction query: ", chain)

		if !common.RPCMethods["GET"][config.QueryBitcoinTransaction].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryBitcoinTransaction].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					common.GetLogger().Info("[query-btc-transaction] Returning from the cache: ", chain)
					return
				}
			}

			response.Response, response.Error, statusCode = queryBtcTransactionRequestHandle(r, chain, transactionHash)
		}

		isSupportedChain, conf := GetChainConfig(chain)
		enableCache := false
		if isSupportedChain {
			enableCache = response.Response.(SearchRawTransactionsResult).BlockConfirmations == strconv.Itoa(int(conf.BTC_CONFIRMATIONS))+"+"
		}
		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryBitcoinTransaction].CachingEnabled && enableCache)
	}
}
