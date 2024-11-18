package bitcoin

import (
	"net/http"
	"strings"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/gorilla/mux"

	// "github.com/powerman/rpc-codec/jsonrpc2"
	jsonrpc2 "github.com/KeisukeYamashita/go-jsonrpc"
)

type AccountScript struct {
	IsScript bool   `json:"isscript,omitempty"`
	PubKey   string `json:"pubkey,omitempty"`
	Asm      string `json:"asm,omitempty"`
	Desc     string `json:"desc,omitempty"`
	Address  string `json:"address,omitempty"`
	Type     string `json:"type,omitempty"`
}

type AccountWitness struct {
	IsWitness bool   `json:"iswitness,omitempty"`
	Version   int32  `json:"version,omitempty"`
	Program   string `json:"program,omitempty"`
}

type AccountResult struct {
	IsValid        bool           `json:"isvalid"`
	Address        string         `json:"address,omitempty"`
	IsScript       *bool          `json:"isscript,omitempty"`
	IsWitness      *bool          `json:"iswitness,omitempty"`
	WitnessVersion *int32         `json:"witness_version,omitempty"`
	WitnessProgram *string        `json:"witness_program,omitempty"`
	ScriptPubKey   string         `json:"scriptPubkey,omitempty"`
	Script         AccountScript  `json:"script,omitempty"`
	Witness        AccountWitness `json:"witness,omitempty"`
}

// RegisterBtcAccountsRoutes registers query status of bitcoin accounts.
func RegisterBtcAccountsRoutes(r *mux.Router, rpcAddr string) {
	r.HandleFunc(config.QueryBitcoinAccounts, QueryBtcAccountsRequest(rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryBitcoinAccounts, "This is an API to query accounts.", true)
}

func queryBtcAccountsRequestHandle(r *http.Request, chain string, address string) (interface{}, interface{}, int) {
	isSupportedChain, conf := GetChainConfig(chain)
	if !isSupportedChain {
		return common.ServeError(0, "", "unsupported chain", http.StatusBadRequest)
	}

	client := jsonrpc2.NewRPCClient(conf.RPC)
	if conf.RPC_CRED != "" {
		rpcInfo := strings.Split(conf.RPC_CRED, ":")
		client.SetBasicAuth(rpcInfo[0], rpcInfo[1])
	}

	response := AccountResult{}
	err := GetResult(client, "validateaddress", &response, address)
	if err != nil {
		return common.ServeError(0, "failed to get accounts by address", err.Error(), http.StatusInternalServerError)
	}

	if *response.IsScript {
		decodeScript := AccountScript{}
		err = GetResult(client, "decodescript", &decodeScript, response.ScriptPubKey)
		if err != nil {
			return common.ServeError(0, "failed to get accounts by address", err.Error(), http.StatusInternalServerError)
		}

		response.Script.IsScript = *response.IsScript
		response.Script.PubKey = response.ScriptPubKey
		response.Script.Asm = decodeScript.Asm
		response.Script.Desc = decodeScript.Desc
		response.Script.Address = decodeScript.Address
		response.Script.Type = decodeScript.Type
		response.Witness = AccountWitness{}
	}

	if *response.IsWitness {
		decodeWitness := AccountScript{}
		err = GetResult(client, "decodescript", &decodeWitness, response.ScriptPubKey)
		if err != nil {
			return common.ServeError(0, "failed to get accounts by address", err.Error(), http.StatusInternalServerError)
		}

		response.Witness.IsWitness = *response.IsWitness
		response.Witness.Version = *response.WitnessVersion
		response.Witness.Program = *response.WitnessProgram
	}

	response.IsScript = nil
	response.IsWitness = nil
	response.WitnessProgram = nil
	response.WitnessVersion = nil
	response.ScriptPubKey = ""
	return response, nil, http.StatusOK
}

// QueryBtcAccountsRequest is a function to query accounts of Bitcoin chains.
func QueryBtcAccountsRequest(rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queries := mux.Vars(r)
		chain := "testnet"
		address := queries["address"]
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)
		statusCode := http.StatusOK

		common.GetLogger().Info("[query-btc-accounts] Entering accounts query: ", chain)

		if !common.RPCMethods["GET"][config.QueryBitcoinAccounts].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryBitcoinAccounts].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					common.GetLogger().Info("[query-btc-accounts] Returning from the cache: ", chain)
					return
				}
			}

			response.Response, response.Error, statusCode = queryBtcAccountsRequestHandle(r, chain, address)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryBitcoinAccounts].CachingEnabled)
	}
}
