package kira

import (
	"net/http"
	"strings"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/log"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// RegisterKiraUpgradeRoutes registers kira upgrade query routers.
func RegisterKiraUpgradeRoutes(r *mux.Router, gwCosmosmux *runtime.ServeMux, rpcAddr string) {
	r.HandleFunc(config.QueryCurrentPlan, QueryCurrentPlanRequest(gwCosmosmux, rpcAddr)).Methods("GET")
	r.HandleFunc(config.QueryNextPlan, QueryNextPlanRequest(gwCosmosmux, rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryCurrentPlan, "This is an API to query current upgrade plan.", true)
	common.AddRPCMethod("GET", config.QueryNextPlan, "This is an API to query next upgrade plan.", true)
}

func QueryCurrentPlanHandler(r *http.Request, gwCosmosmux *runtime.ServeMux) (interface{}, interface{}, int) {
	r.URL.Path = strings.Replace(r.URL.Path, "/api/kira/upgrade", "/kira/upgrade", -1)
	return common.ServeGRPC(r, gwCosmosmux)
}

// QueryCurrentPlanRequest is a function to query current upgrade plan.
func QueryCurrentPlanRequest(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		log.CustomLogger().Info("[query-current-upgrade-plan] Entering upgrade plan query")

		if !common.RPCMethods["GET"][config.QueryCurrentPlan].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryCurrentPlan].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					log.CustomLogger().Info("[query-current-upgrade-plan] Returning from the cache")
					return
				}
			}

			response.Response, response.Error, statusCode = QueryCurrentPlanHandler(r, gwCosmosmux)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryCurrentPlan].CachingEnabled)
	}
}

func QueryNextPlanHandler(r *http.Request, gwCosmosmux *runtime.ServeMux) (interface{}, interface{}, int) {
	r.URL.Path = strings.Replace(r.URL.Path, "/api/kira/upgrade", "/kira/upgrade", -1)
	return common.ServeGRPC(r, gwCosmosmux)
}

// QueryNextPlanRequest is a function to query next upgrade plan.
func QueryNextPlanRequest(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		log.CustomLogger().Info("[query-next-upgrade-plan] Entering upgrade plan query")

		if !common.RPCMethods["GET"][config.QueryNextPlan].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryNextPlan].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					log.CustomLogger().Info("[query-next-upgrade-plan] Returning from the cache")
					return
				}
			}

			response.Response, response.Error, statusCode = QueryNextPlanHandler(r, gwCosmosmux)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryNextPlan].CachingEnabled)
	}
}
