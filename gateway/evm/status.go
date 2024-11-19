package evm

import (
	"math/big"
	"net/http"
	"strconv"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/log"
	"github.com/gorilla/mux"

	// "github.com/powerman/rpc-codec/jsonrpc2"
	jsonrpc2 "github.com/KeisukeYamashita/go-jsonrpc"
)

// RegisterEVMStatusRoutes registers query status of EVM chains.
func RegisterEVMStatusRoutes(r *mux.Router, rpcAddr string) {
	r.HandleFunc(config.QueryEVMStatus, QueryEVMStatusRequest(rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryEVMStatus, "This is an API to query status.", true)
}

type EVMStatus struct {
	NodeInfo struct {
		Network    uint64 `json:"network"`
		RPCAddress string `json:"rpc_address"`
		Version    struct {
			Net      string `json:"net"`
			Web3     string `json:"web3"`
			Protocol string `json:"protocol"`
		} `json:"version"`
	} `json:"node_info"`
	SyncInfo struct {
		CatchingUp          bool   `json:"catching_up"`
		EarliestBlockHash   string `json:"earliest_block_hash"`
		EarliestBlockHeight uint64 `json:"earliest_block_height"`
		EarliestBlockTime   uint64 `json:"earliest_block_time"`
		LatestBlockHash     string `json:"latest_block_hash"`
		LatestBlockHeight   uint64 `json:"latest_block_height"`
		LatestBlockTime     uint64 `json:"latest_block_time"`
	} `json:"sync_info"`
	GasPrice string `json:"gas_price"`
}

func queryEVMStatusFromNode(nodeInfo config.EVMNodeConfig) (interface{}, interface{}, int) {
	client := jsonrpc2.NewRPCClient(nodeInfo.RPC + "/" + nodeInfo.RPCToken)
	if nodeInfo.RPCSecret != "" {
		client.SetBasicAuth(nodeInfo.RPCToken, nodeInfo.RPCSecret)
	}

	response := EVMStatus{}

	response.NodeInfo.RPCAddress = nodeInfo.RPC

	data, err := client.Call("eth_chainId")
	if err != nil {
		return common.ServeError(0, "failed to get chain id", err.Error(), http.StatusInternalServerError)
	}
	chainId, err := data.GetString()
	if err != nil {
		return common.ServeError(0, "failed to get chain id", err.Error(), http.StatusInternalServerError)
	}
	response.NodeInfo.Network, _ = strconv.ParseUint((chainId)[2:], 16, 64)

	data, err = client.Call("web3_clientVersion")
	if err != nil {
		return common.ServeError(0, "failed to get client version", err.Error(), http.StatusInternalServerError)
	}
	clientVersion, err := data.GetString()
	if err != nil {
		clientVersion = ""
	}
	response.NodeInfo.Version.Web3 = clientVersion

	data, err = client.Call("net_version")
	if err != nil {
		return common.ServeError(0, "failed to get net version", err.Error(), http.StatusInternalServerError)
	}
	netVersion, err := data.GetString()
	if err != nil {
		netVersion = ""
	}
	response.NodeInfo.Version.Net = netVersion

	data, err = client.Call("eth_protocolVersion")
	if err != nil {
		return common.ServeError(0, "failed to get protocol version", err.Error(), http.StatusInternalServerError)
	}
	protocolVersion, err := data.GetString()
	if err != nil {
		protocolVersion = ""
	}
	response.NodeInfo.Version.Protocol = protocolVersion

	data, err = client.Call("eth_syncing")
	if err != nil {
		return common.ServeError(0, "failed to get sync status", err.Error(), http.StatusInternalServerError)
	}
	isSyncing, err := data.GetBool()
	if err != nil {
		isSyncing = true
	}
	response.SyncInfo.CatchingUp = isSyncing

	latestBlock := new(struct {
		Hash      string `json:"hash"`
		Number    string `json:"number"`
		Timestamp string `json:"timestamp"`
	})
	data, err = client.Call("eth_getBlockByNumber", "latest", true)
	if err != nil {
		return common.ServeError(0, "failed to get latest block ", err.Error(), http.StatusInternalServerError)
	}
	err = data.GetObject(latestBlock)
	if err != nil {
		return common.ServeError(0, "failed to get latest block ", err.Error(), http.StatusInternalServerError)
	}
	response.SyncInfo.LatestBlockHash = latestBlock.Hash
	response.SyncInfo.LatestBlockHeight, _ = strconv.ParseUint((latestBlock.Number)[2:], 16, 64)
	response.SyncInfo.LatestBlockTime, _ = strconv.ParseUint((latestBlock.Timestamp)[2:], 16, 64)

	earliestBlock := new(struct {
		Hash      string `json:"hash"`
		Number    string `json:"number"`
		Timestamp string `json:"timestamp"`
	})
	data, err = client.Call("eth_getBlockByNumber", "earliest", true)
	if err != nil {
		return common.ServeError(0, "failed to get earliest block ", err.Error(), http.StatusInternalServerError)
	}
	err = data.GetObject(earliestBlock)
	if err != nil {
		return common.ServeError(0, "failed to get earliest block ", err.Error(), http.StatusInternalServerError)
	}
	response.SyncInfo.EarliestBlockHash = earliestBlock.Hash
	response.SyncInfo.EarliestBlockHeight, _ = strconv.ParseUint((earliestBlock.Number)[2:], 16, 64)
	response.SyncInfo.EarliestBlockTime, _ = strconv.ParseUint((earliestBlock.Timestamp)[2:], 16, 64)

	data, err = client.Call("eth_gasPrice")
	if err != nil {
		return common.ServeError(0, "failed to get gas price ", err.Error(), http.StatusInternalServerError)
	}
	gasPrice, err := data.GetString()
	if err != nil {
		return common.ServeError(0, "failed to get gas price ", err.Error(), http.StatusInternalServerError)
	}
	gasPriceBig := *new(big.Int)
	gasPriceBig.SetString((gasPrice)[2:], 16)
	response.GasPrice = gasPriceBig.String()

	return response, nil, http.StatusOK
}

func queryEVMStatusHandle(r *http.Request, chain string) (interface{}, interface{}, int) {

	isSupportedChain, chainConfig := GetChainConfig(chain)
	if !isSupportedChain {
		return common.ServeError(0, "", "unsupported chain", http.StatusBadRequest)
	}

	res, err, statusCode := queryEVMStatusFromNode(chainConfig.QuickNode)
	if err == nil {
		return res, err, statusCode
	}

	res, err, statusCode = queryEVMStatusFromNode(chainConfig.Infura)
	if err == nil {
		return res, err, statusCode
	}

	return queryEVMStatusFromNode(chainConfig.Pokt)
}

// QueryEVMStatusRequest is a function to query status of EVM chains.
func QueryEVMStatusRequest(rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		queries := mux.Vars(r)
		chain := queries["chain"]
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		log.CustomLogger().Info("[query-evm-status] Entering status query: ", chain)

		if !common.RPCMethods["GET"][config.QueryEVMStatus].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryEVMStatus].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					log.CustomLogger().Info("[query-evm-status] Returning from the cache: ", chain)
					return
				}
			}

			response.Response, response.Error, statusCode = queryEVMStatusHandle(r, chain)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryEVMStatus].CachingEnabled)
	}
}
