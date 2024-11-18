package bitcoin

import (
	"encoding/json"
	"math"
	"math/big"
	"net/http"
	"strings"

	jsonrpc2 "github.com/KeisukeYamashita/go-jsonrpc"
	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/gorilla/mux"

	"github.com/btcsuite/btcd/btcjson"
)

// RegisterBitcoinStatusRoutes registers query status of Bitcoin chains.
func RegisterBitcoinStatusRoutes(r *mux.Router, rpcAddr string) {
	r.HandleFunc(config.QueryBitcoinStatus, QueryBitcoinStatusRequest(rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryBitcoinStatus, "This is an API to query status.", true)
}

type BitcoinStatus struct {
	NodeInfo struct {
		Network    string `json:"network"`
		RPCAddress string `json:"rpc_address"`
		Version    struct {
			Net      int32  `json:"net"`
			Sub      string `json:"sub"`
			Protocol int32  `json:"protocol"`
		} `json:"version"`
	} `json:"node_info"`
	SyncInfo struct {
		CatchingUp          bool   `json:"catching_up"`
		EarliestBlockHash   string `json:"earliest_block_hash"`
		EarliestBlockHeight int64  `json:"earliest_block_height"`
		EarliestBlockTime   int64  `json:"earliest_block_time"`
		LatestBlockHash     string `json:"latest_block_hash"`
		LatestBlockHeight   int64  `json:"latest_block_height"`
		LatestBlockTime     int64  `json:"latest_block_time"`
	} `json:"sync_info"`
	GasPrice    string `json:"gas_price"`
	GasPriceAvg uint64 `json:"gas_price_avg"`
	GasPriceMin string `json:"gas_price_min"`
	GasPriceInc string `json:"gas_price_inc"`
}

func GetResult(client *jsonrpc2.RPCClient, method string, x interface{}, params ...interface{}) error {
	res := new(interface{})
	data, err := client.Call(method, params...)
	if err != nil {
		return err
	}

	err = data.GetObject(res)
	if err != nil {
		return err
	}

	bz, err := json.Marshal(res)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bz, x)
	if err != nil {
		return err
	}
	return nil
}

func queryBitcoinStatusHandle(r *http.Request, chain string) (interface{}, interface{}, int) {
	isSupportedChain, conf := GetChainConfig(chain)
	if !isSupportedChain {
		return common.ServeError(0, "", "unsupported chain", http.StatusBadRequest)
	}

	client := jsonrpc2.NewRPCClient(conf.RPC)
	if conf.RPC_CRED != "" {
		rpcInfo := strings.Split(conf.RPC_CRED, ":")
		client.SetBasicAuth(rpcInfo[0], rpcInfo[1])
	}

	response := BitcoinStatus{}

	chainInfo := btcjson.GetBlockChainInfoResult{}
	err := GetResult(client, "getblockchaininfo", &chainInfo)
	if err != nil {
		return common.ServeError(0, "failed to get blockchain info ", err.Error(), http.StatusInternalServerError)
	}

	networkInfo := btcjson.GetNetworkInfoResult{}
	err = GetResult(client, "getnetworkinfo", &networkInfo)
	if err != nil {
		return common.ServeError(0, "failed to get network info ", err.Error(), http.StatusInternalServerError)
	}

	blockStats := btcjson.GetBlockStatsResult{}
	err = GetResult(client, "getblockstats", &blockStats, 1)
	if err != nil {
		return common.ServeError(0, "failed to get block stats info ", err.Error(), http.StatusInternalServerError)
	}

	response.NodeInfo.Network = chainInfo.Chain
	response.NodeInfo.RPCAddress = conf.RPC
	response.NodeInfo.Version.Net = networkInfo.Version
	response.NodeInfo.Version.Sub = networkInfo.SubVersion
	response.NodeInfo.Version.Protocol = networkInfo.ProtocolVersion
	response.SyncInfo.CatchingUp = true
	response.SyncInfo.EarliestBlockHash = blockStats.Hash
	response.SyncInfo.EarliestBlockHeight = blockStats.Height
	response.SyncInfo.EarliestBlockTime = blockStats.Time
	response.SyncInfo.LatestBlockHash = chainInfo.BestBlockHash

	err = GetResult(client, "getblockstats", &blockStats, chainInfo.BestBlockHash)
	if err != nil {
		return common.ServeError(0, "failed to block stats info ", err.Error(), http.StatusInternalServerError)
	}

	smartFee := btcjson.EstimateSmartFeeResult{}
	err = GetResult(client, "estimatesmartfee", &smartFee, 6, "CONSERVATIVE")
	if err != nil {
		return common.ServeError(0, "failed to smart fee info ", err.Error(), http.StatusInternalServerError)
	}

	response.SyncInfo.LatestBlockHeight = blockStats.Height
	response.SyncInfo.LatestBlockTime = blockStats.Time

	gasPrice := big.NewFloat(math.Max(*smartFee.FeeRate, networkInfo.RelayFee) * 1e8)
	response.GasPrice = gasPrice.String()

	response.GasPriceAvg = uint64(blockStats.AverageFeeRate)
	response.GasPriceMin = big.NewFloat(math.Max(networkInfo.RelayFee*1e8, float64(blockStats.MinFeeRate))).String()

	gasPriceInc := big.NewFloat(networkInfo.IncrementalFee * 1e8)
	response.GasPriceInc = gasPriceInc.String()

	if response.GasPrice == response.GasPriceMin {
		response.GasPrice = gasPrice.Add(gasPrice, gasPriceInc).String()
	}
	return response, nil, http.StatusOK
}

// QueryBitcoinStatusRequest is a function to query status of Bitcoin chains.
func QueryBitcoinStatusRequest(rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chain := "testnet"
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)
		statusCode := http.StatusOK

		common.GetLogger().Info("[query-bitcoin-status] Entering status query: ", chain)

		if !common.RPCMethods["GET"][config.QueryBitcoinStatus].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryBitcoinStatus].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					common.GetLogger().Info("[query-bitcoin-status] Returning from the cache: ", chain)
					return
				}
			}

			response.Response, response.Error, statusCode = queryBitcoinStatusHandle(r, chain)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryBitcoinStatus].CachingEnabled)
	}
}
