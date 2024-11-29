package data

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/log"
	"github.com/KiraCore/interx/types"
	"github.com/KiraCore/interx/types/rosetta"
	"github.com/KiraCore/interx/types/rosetta/dataapi"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// RegisterNetworkRoutes registers network API routers.
func RegisterNetworkRoutes(r *mux.Router, gwCosmosmux *runtime.ServeMux, rpcAddr string) {
	r.HandleFunc(config.QueryRosettaNetworkList, QueryNetworkListRequest(gwCosmosmux, rpcAddr)).Methods("POST")
	r.HandleFunc(config.QueryRosettaNetworkOptions, QueryNetworkOptionsRequest(gwCosmosmux, rpcAddr)).Methods("POST")
	r.HandleFunc(config.QueryRosettaNetworkStatus, QueryNetworkStatusRequest(gwCosmosmux, rpcAddr)).Methods("POST")

	common.AddRPCMethod("POST", config.QueryRosettaNetworkList, "This is an API to query network list.", true)
	common.AddRPCMethod("POST", config.QueryRosettaNetworkOptions, "This is an API to query network options.", true)
	common.AddRPCMethod("POST", config.QueryRosettaNetworkStatus, "This is an API to query network status.", true)
}

func queryNetworkListHandler(r *http.Request, request types.InterxRequest, rpcAddr string, gwCosmosmux *runtime.ServeMux) (interface{}, interface{}, int) {
	var req dataapi.NetworkListRequest

	err := json.Unmarshal(request.Params, &req)
	if err != nil {
		log.CustomLogger().Error("[rosetta-query-networklist] Failed to decode the request: ", err)
		return common.RosettaServeError(0, "failed to unmarshal", err.Error(), http.StatusBadRequest)
	}

	var response dataapi.NetworkListResponse

	success, failure, status := common.MakeTendermintRPCRequest(rpcAddr, "/status", "")

	if success != nil {
		type TempResponse struct {
			NodeInfo struct {
				Network string `json:"network"`
			} `json:"node_info"`
		}
		result := TempResponse{}

		byteData, err := json.Marshal(success)
		if err != nil {
			log.CustomLogger().Error("[rosetta-query-networklist] Invalid response format", err)
			return common.RosettaServeError(0, "", err.Error(), http.StatusInternalServerError)
		}

		err = json.Unmarshal(byteData, &result)
		if err != nil {
			log.CustomLogger().Error("[rosetta-query-networklist] Invalid response format", err)
			return common.RosettaServeError(0, "", err.Error(), http.StatusInternalServerError)
		}

		response.NetworkIdentifiers = make([]rosetta.NetworkIdentifier, 0)
		response.NetworkIdentifiers = append(response.NetworkIdentifiers, rosetta.NetworkIdentifier{
			Blockchain: "sekaid",
			Network:    result.NodeInfo.Network,
		})

		return response, nil, http.StatusOK
	}

	return nil, failure, status
}

// QueryNetworkListRequest is a function to query network list.
func QueryNetworkListRequest(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		log.CustomLogger().Info("[rosetta-query-networklist] Entering network list query")

		if !common.RPCMethods["POST"][config.QueryRosettaNetworkList].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["POST"][config.QueryRosettaNetworkList].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					log.CustomLogger().Info("[rosetta-query-networklist] Returning from the cache")
					return
				}
			}

			response.Response, response.Error, statusCode = queryNetworkListHandler(r, request, rpcAddr, gwCosmosmux)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["POST"][config.QueryRosettaNetworkList].CachingEnabled)
	}
}

func queryNetworkOptionsHandler(r *http.Request, request types.InterxRequest, rpcAddr string, gwCosmosmux *runtime.ServeMux) (interface{}, interface{}, int) {
	var response dataapi.NetworkOptionsResponse

	response.Version = rosetta.Version{
		NodeVersion:       config.Config.SekaiVersion,
		MiddlewareVersion: config.Config.InterxVersion,
	}

	return response, nil, http.StatusOK
}

// QueryNetworkOptionsRequest is a function to query network options.
func QueryNetworkOptionsRequest(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		log.CustomLogger().Info("[rosetta-query-networkoptions] Entering network list query")

		if !common.RPCMethods["POST"][config.QueryRosettaNetworkOptions].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["POST"][config.QueryRosettaNetworkOptions].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					log.CustomLogger().Info("[rosetta-query-networkoptions] Returning from the cache")
					return
				}
			}

			response.Response, response.Error, statusCode = queryNetworkOptionsHandler(r, request, rpcAddr, gwCosmosmux)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["POST"][config.QueryRosettaNetworkList].CachingEnabled)
	}
}

func queryNetworkStatusHandler(r *http.Request, request types.InterxRequest, rpcAddr string, gwCosmosmux *runtime.ServeMux) (interface{}, interface{}, int) {
	var req dataapi.NetworkStatusRequest

	err := json.Unmarshal(request.Params, &req)
	if err != nil {
		log.CustomLogger().Error("[rosetta-query-networkstatus] Failed to decode the request: ", err)
		return common.RosettaServeError(0, "failed to unmarshal", err.Error(), http.StatusBadRequest)
	}

	var response dataapi.NetworkStatusResponse

	success, failure, status := common.MakeTendermintRPCRequest(rpcAddr, "/status", "")

	if success != nil {
		type TempResponse struct {
			NodeInfo struct {
				ID string `json:"id"`
			} `json:"node_info"`
			SyncInfo struct {
				LatestBlockHash     string `json:"latest_block_hash"`
				LatestAppHash       string `json:"latest_app_hash"`
				LatestBlockHeight   int64  `json:"latest_block_height,string"`
				LatestBlockTime     string `json:"latest_block_time"`
				EarliestBlockHash   string `json:"earliest_block_hash"`
				EarliestAppHash     string `json:"earliest_app_hash"`
				EarliestBlockHeight int64  `json:"earliest_block_height,string"`
				EarliestBlockTime   string `json:"earliest_block_time"`
				CatchingUp          bool   `json:"catching_up"`
			} `json:"sync_info"`
		}

		result := TempResponse{}

		byteData, err := json.Marshal(success)
		if err != nil {
			log.CustomLogger().Error("[rosetta-query-networkstatus] Invalid response format", err)
			return common.RosettaServeError(0, "", err.Error(), http.StatusInternalServerError)
		}

		err = json.Unmarshal(byteData, &result)
		if err != nil {
			log.CustomLogger().Error("[rosetta-query-networkstatus] Invalid response format", err)
			return common.RosettaServeError(0, "", err.Error(), http.StatusInternalServerError)
		}

		response.CurrentBlockIdentifier = rosetta.BlockIdentifier{
			Index: result.SyncInfo.LatestBlockHeight,
			Hash:  result.SyncInfo.LatestBlockHash,
		}

		currentBlockTimestamp, err := time.Parse(time.RFC3339, result.SyncInfo.LatestBlockTime)
		if err == nil {
			response.CurrentBlockTimestamp = currentBlockTimestamp.Unix()
		} else {
			response.CurrentBlockTimestamp = 0
		}

		response.GenesisBlockIdentifier = rosetta.BlockIdentifier{
			Index: result.SyncInfo.EarliestBlockHeight,
			Hash:  result.SyncInfo.EarliestBlockHash,
		}
		response.OldestBlockIdentifier = response.GenesisBlockIdentifier

		response.SyncStatus.Synced = !result.SyncInfo.CatchingUp

		response.Peers = make([]rosetta.Peer, 0)
		response.Peers = append(response.Peers, rosetta.Peer{
			PeerID: result.NodeInfo.ID,
		})

		return response, nil, http.StatusOK
	}

	return nil, failure, status
}

// QueryNetworkStatusRequest is a function to query network status.
func QueryNetworkStatusRequest(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		log.CustomLogger().Info("[rosetta-query-networkstatus] Entering network list query")

		if !common.RPCMethods["POST"][config.QueryRosettaNetworkStatus].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["POST"][config.QueryRosettaNetworkStatus].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					log.CustomLogger().Info("[rosetta-query-networkstatus] Returning from the cache")
					return
				}
			}

			response.Response, response.Error, statusCode = queryNetworkStatusHandler(r, request, rpcAddr, gwCosmosmux)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["POST"][config.QueryRosettaNetworkList].CachingEnabled)
	}
}
