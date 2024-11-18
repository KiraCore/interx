package kira

import (
	"net/http"
	"strings"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// RegisterKiraSpendingRoutes registers kira spending query routers.
func RegisterKiraSpendingRoutes(r *mux.Router, gwCosmosmux *runtime.ServeMux, rpcAddr string) {
	r.HandleFunc(config.QuerySpendingPools, QuerySpendingPoolsRequest(gwCosmosmux, rpcAddr)).Methods("GET")
	r.HandleFunc(config.QuerySpendingPoolProposals, QuerySpendingPoolProposalsRequest(gwCosmosmux, rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QuerySpendingPools, "This is an API to query spending pools.", true)
	common.AddRPCMethod("GET", config.QuerySpendingPoolProposals, "This is an API to query list of spending pool proposals by name.", true)
}

func QuerySpendingPoolsHandler(r *http.Request, gwCosmosmux *runtime.ServeMux) (interface{}, interface{}, int) {
	queries := r.URL.Query()
	account := queries["account"]
	name := queries["name"]

	if len(account) == 1 {
		r.URL.RawQuery = ""
		r.URL.Path = strings.Replace(r.URL.Path, "/api/kira/spending-pools", "/kira/spending/pools/"+account[0], -1)
	} else if len(name) == 1 {
		r.URL.RawQuery = ""
		r.URL.Path = strings.Replace(r.URL.Path, "/api/kira/spending-pools", "/kira/spending/pool/"+name[0], -1)
	} else {
		r.URL.RawQuery = ""
		r.URL.Path = strings.Replace(r.URL.Path, "/api/kira/spending-pools", "/kira/spending/pool_names", -1)
	}

	return common.ServeGRPC(r, gwCosmosmux)
}

// QuerySpendingPoolsRequest is a function to query list of all spending pool names.
func QuerySpendingPoolsRequest(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		common.GetLogger().Info("[query-spending-pool-names] Entering upgrade plan query")

		if !common.RPCMethods["GET"][config.QuerySpendingPools].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QuerySpendingPools].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					common.GetLogger().Info("[query-spending-pool-names] Returning from the cache")
					return
				}
			}

			response.Response, response.Error, statusCode = QuerySpendingPoolsHandler(r, gwCosmosmux)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QuerySpendingPools].CachingEnabled)
	}
}

func QuerySpendingPoolProposalHandler(r *http.Request, gwCosmosmux *runtime.ServeMux) (interface{}, interface{}, int) {
	queries := r.URL.Query()
	name := queries["name"]

	if len(name) == 1 {
		r.URL.RawQuery = ""
		r.URL.Path = strings.Replace(r.URL.Path, "/api/kira/spending-pool-proposals", "/kira/spending/pool_proposals/"+name[0], -1)
	}

	return common.ServeGRPC(r, gwCosmosmux)
}

// QuerySpendingPoolProposalsRequest is a function to query list of spending pools proposal by name.
func QuerySpendingPoolProposalsRequest(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		common.GetLogger().Info("[query-spending-pool-proposals] Entering upgrade plan query")

		if !common.RPCMethods["GET"][config.QuerySpendingPoolProposals].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QuerySpendingPoolProposals].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					common.GetLogger().Info("[query-spending-pool-proposals] Returning from the cache")
					return
				}
			}

			response.Response, response.Error, statusCode = QuerySpendingPoolProposalHandler(r, gwCosmosmux)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QuerySpendingPoolProposals].CachingEnabled)
	}
}
