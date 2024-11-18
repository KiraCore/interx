package bitcoin

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/gorilla/mux"

	// "github.com/powerman/rpc-codec/jsonrpc2"
	jsonrpc2 "github.com/KeisukeYamashita/go-jsonrpc"
)

type DecodeScriptResult struct {
	Asm    string `json:"asm"`
	Desc   string `json:"desc,omitempty"`
	Type   string `json:"type"`
	P2sh   string `json:"p2sh,omitempty"`
	Segwit struct {
		Asm        string   `json:"asm"`
		Desc       string   `json:"desc,omitempty"`
		Hex        string   `json:"hex,omitempty"`
		Address    string   `json:"address,omitempty"`
		ReqSigs    int64    `json:"reqSigs"`
		Addresses  []string `json:"addresses,omitempty"`
		Type       string   `json:"type"`
		P2shSegwit string   `json:"p2sh-segwit"`
	} `json:"segwit,omitempty"`
}

type DecodeScriptResponse struct {
	Asm       string   `json:"asm"`
	Hex       string   `json:"hex"`
	ReqSigs   string   `json:"reqSigs"`
	Type      string   `json:"type"`
	Addresses []string `json:"addresses"`
	P2sh      string   `json:"p2sh,omitempty"`
}

type CreateRawTransaction struct {
	Inputs   []btcjson.TransactionInput `json:"inputs"`
	Amounts  map[string]float64         `json:"amounts"`
	LockTime int64                      `json:"locktime"`
}

type RawTxResponse struct {
	RawTx string `json:"raw_tx"`
}

type SendTxResult struct {
	Txid string `json:"txid"`
}

// RegisterBtcTransferRoutes registers query status of Bitcoin chains.
func RegisterBtcTransferRoutes(r *mux.Router, rpcAddr string) {
	r.HandleFunc(config.QueryBitcoinTransfer, QueryBtcTransferRequest(rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryBitcoinTransfer, "This is an API to broadcast a signed transaction.", true)
}

func queryBtcTransferRequestHandle(r *http.Request, chain string) (interface{}, interface{}, int) {
	_ = r.ParseForm()

	queries := r.URL.Query()
	decode := queries["decode"][0] == "true"

	isSupportedChain, conf := GetChainConfig(chain)
	if !isSupportedChain {
		return common.ServeError(0, "", "unsupported chain", http.StatusBadRequest)
	}

	client := jsonrpc2.NewRPCClient(conf.RPC)
	if conf.RPC_CRED != "" {
		rpcInfo := strings.Split(conf.RPC_CRED, ":")
		client.SetBasicAuth(rpcInfo[0], rpcInfo[1])
	}

	if decode {
		rawTx := r.FormValue("rawTx")
		if rawTx != "" {
			response := SearchRawTransactionsResult{}
			err := GetResult(client, "decoderawtransaction", &response, rawTx)
			if err != nil {
				return common.ServeError(0, "failed to get decode", err.Error(), http.StatusInternalServerError)
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

			return response, nil, http.StatusOK
		}

		scriptTx := r.FormValue("scriptTx")
		if scriptTx != "" {
			res := DecodeScriptResult{}
			err := GetResult(client, "decodescript", &res, scriptTx)
			if err != nil {
				return common.ServeError(0, "failed to decode", err.Error(), http.StatusInternalServerError)
			}

			response := DecodeScriptResponse{}
			response.Asm = res.Segwit.Asm
			response.Hex = res.Segwit.Hex
			response.ReqSigs = "1"
			response.Type = res.Segwit.Type
			response.P2sh = res.Segwit.P2shSegwit
			response.Addresses = []string{res.Segwit.Address}
			return response, nil, http.StatusOK
		}
	} else {
		rawTx := r.FormValue("rawTx")
		if rawTx != "" {
			res := btcjson.TxRawResult{}
			err := GetResult(client, "sendrawtransaction", &res, rawTx)
			if err != nil {
				return common.ServeError(0, "failed to send", err.Error(), http.StatusInternalServerError)
			}

			response := SendTxResult{}
			response.Txid = res.Txid
			return response, nil, http.StatusOK
		}

		data := r.FormValue("data")
		if data != "" {
			paramData := CreateRawTransaction{}
			err := json.Unmarshal([]byte(data), &paramData)
			if err != nil {
				return common.ServeError(0, "failed to create", err.Error(), http.StatusInternalServerError)
			}

			res := ""
			err = GetResult(client, "createrawtransaction", &res, paramData.Inputs, paramData.Amounts, paramData.LockTime)
			if err != nil {
				return common.ServeError(0, "failed to create", err.Error(), http.StatusInternalServerError)
			}

			response := RawTxResponse{}
			response.RawTx = res
			return response, nil, http.StatusOK
		}
	}

	return nil, nil, http.StatusOK
}

// QueryBtcTransferRequest is a function to query transfer of Bitcoin chains.
func QueryBtcTransferRequest(rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// queries := mux.Vars(r)
		chain := "testnet"
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)
		statusCode := http.StatusOK

		common.GetLogger().Info("[query-btc-transfer] Entering transfer execute: ", chain)

		if !common.RPCMethods["GET"][config.QueryBitcoinTransfer].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryBitcoinTransfer].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					common.GetLogger().Info("[query-btc-transfer] Returning from the cache: ", chain)
					return
				}
			}

			response.Response, response.Error, statusCode = queryBtcTransferRequestHandle(r, chain)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryBitcoinTransfer].CachingEnabled)
	}
}
