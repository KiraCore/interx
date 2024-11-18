package bitcoin

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/gorilla/mux"

	// "github.com/powerman/rpc-codec/jsonrpc2"
	jsonrpc2 "github.com/KeisukeYamashita/go-jsonrpc"
	"github.com/btcsuite/btcd/btcjson"
)

type BlockResult struct {
	btcjson.GetBlockVerboseResult
	Confirmations      uint64     `json:"confirmations,omitempty"`
	MedianTime         uint64     `json:"mediantime"`
	NTx                uint64     `json:"nTx"`
	Stats              BlockStats `json:"stats"`
	BlockConfirmations string     `json:"blockconfirmations,omitempty"`
}

type BlockStats struct {
	BlockHash          string   `json:"blockhash,omitempty"`
	AvgFee             uint64   `json:"avgfee"`
	AvgFeeRate         uint64   `json:"avgfeerate"`
	AvgTxSize          uint64   `json:"avgtxsize"`
	FeeRatePercentiles []uint64 `json:"feerate_percentiles"`
	Ins                uint64   `json:"ins"`
	MaxFee             uint64   `json:"maxfee"`
	MaxFeeRate         uint64   `json:"maxfeerate"`
	MaxTxSize          uint64   `json:"maxtxsize"`
	MedianFee          uint64   `json:"medianfee"`
	MedianTxSize       uint64   `json:"mediantxsize"`
	MinFee             uint64   `json:"minfee"`
	MinFeeRate         uint64   `json:"minfeerate"`
	MinTxSize          uint64   `json:"mintxsize"`
	Outs               uint64   `json:"outs"`
	Subsidy            uint64   `json:"subsidy"`
	SwtotalSize        uint64   `json:"swtotal_size"`
	SwtotalWeight      uint64   `json:"swtotal_weight"`
	SwTxs              uint64   `json:"swtxs"`
	TotalOut           uint64   `json:"total_out"`
	TotalSize          uint64   `json:"total_size"`
	TotalWeight        uint64   `json:"total_weight"`
	TotalFee           uint64   `json:"total_fee"`
	UtxoIncrease       uint64   `json:"utxo_increase"`
	UtxoSizeInc        uint64   `json:"utxo_size_inc"`
}

// RegisterBitcoinBlockRoutes registers query status of Bitcoin chains.
func RegisterBitcoinBlockRoutes(r *mux.Router, rpcAddr string) {
	r.HandleFunc(config.QueryBitcoinBlock, QueryBitcoinBlockRequest(rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryBitcoinBlock, "This is an API to query blocks.", true)
}

func queryBitcoinBlockRequestHandle(r *http.Request, chain string, blockHeightOrHash string) (interface{}, interface{}, int) {
	isSupportedChain, conf := GetChainConfig(chain)
	if !isSupportedChain {
		return common.ServeError(0, "", "unsupported chain", http.StatusBadRequest)
	}

	client := jsonrpc2.NewRPCClient(conf.RPC)
	if conf.RPC_CRED != "" {
		rpcInfo := strings.Split(conf.RPC_CRED, ":")
		client.SetBasicAuth(rpcInfo[0], rpcInfo[1])
	}

	response1 := BlockResult{}
	response2 := BlockStats{}

	if strings.HasPrefix(blockHeightOrHash, "0x") {
		// get block result
		hash := strings.TrimPrefix(blockHeightOrHash, "0x")
		err := GetResult(client, "getblock", &response1, hash, 1)
		if err != nil {
			return common.ServeError(0, "failed to get block by hash", err.Error(), http.StatusInternalServerError)
		}

		// get block stats
		blockHeight := response1.Height
		err = GetResult(client, "getblockstats", &response2, blockHeight)
		if err != nil {
			return common.ServeError(0, "failed to get block by number", err.Error(), http.StatusInternalServerError)
		}

		response2.BlockHash = ""
		response1.Stats = response2
	} else {
		blockHeight, err := strconv.ParseUint(blockHeightOrHash, 10, 64)
		if err != nil {
			return common.ServeError(0, "failed to parse blockheight", err.Error(), http.StatusInternalServerError)
		}

		// get block stats
		err = GetResult(client, "getblockstats", &response2, blockHeight)
		if err != nil {
			return common.ServeError(0, "failed to get block by number", err.Error(), http.StatusInternalServerError)
		}

		// get block result
		hash := response2.BlockHash
		err = GetResult(client, "getblock", &response1, hash, 1)
		if err != nil {
			return common.ServeError(0, "failed to get block by hash", err.Error(), http.StatusInternalServerError)
		}

		response2.BlockHash = ""
		response1.Stats = response2
	}

	if response1.Confirmations > conf.BTC_CONFIRMATIONS {
		response1.BlockConfirmations = strconv.Itoa(int(conf.BTC_CONFIRMATIONS)) + "+"
	} else {
		response1.BlockConfirmations = strconv.Itoa(int(response1.Confirmations))
	}
	response1.Confirmations = 0

	return response1, nil, http.StatusOK
}

// QueryBitcoinBlockRequest is a function to query block of Bitcoin chains.
func QueryBitcoinBlockRequest(rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queries := mux.Vars(r)
		chain := "testnet"
		blockHeightOrHash := queries["identifier"]
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)
		statusCode := http.StatusOK

		common.GetLogger().Info("[query-bitcoin-block] Entering block query: ", chain)

		if !common.RPCMethods["GET"][config.QueryBitcoinBlock].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryBitcoinBlock].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					common.GetLogger().Info("[query-bitcoin-block] Returning from the cache: ", chain)
					return
				}
			}

			response.Response, response.Error, statusCode = queryBitcoinBlockRequestHandle(r, chain, blockHeightOrHash)
		}

		isSupportedChain, conf := GetChainConfig(chain)
		enableCache := false
		if isSupportedChain {
			enableCache = response.Response.(BlockResult).BlockConfirmations == strconv.Itoa(int(conf.BTC_CONFIRMATIONS))+"+"
		}
		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryBitcoinBlock].CachingEnabled && enableCache)
	}
}
