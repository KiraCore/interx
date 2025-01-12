package kira

import (
	"net/http"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	functions "github.com/KiraCore/interx/functions"
	"github.com/KiraCore/interx/log"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// RegisterKiraQueryRoutes registers tx query routers.
func RegisterKiraQueryRoutes(r *mux.Router, gwCosmosmux *runtime.ServeMux, rpcAddr string) {
	r.HandleFunc(config.QueryKiraFunctions, QueryKiraFunctions(rpcAddr)).Methods("GET")
	r.HandleFunc(config.QueryKiraStatus, QueryKiraStatusRequest(rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryKiraFunctions, "This is an API to query kira functions and metadata.", true)
	common.AddRPCMethod("GET", config.QueryKiraStatus, "This is an API to query kira status.", true)
}

func queryKiraFunctionsHandle(_ string) (interface{}, interface{}, int) {
	functions := functions.GetKiraFunctions()
	return functions, nil, http.StatusOK
}

// QueryKiraFunctions is a function to list functions and metadata.
func QueryKiraFunctions(rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)
		response.Response, response.Error, statusCode = queryKiraFunctionsHandle(rpcAddr)
		common.WrapResponse(w, request, *response, statusCode, false)
	}
}

// QueryKiraStatusRequest is a function to query kira status.
func QueryKiraStatusRequest(rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		log.CustomLogger().Info("Starting `QueryKiraStatusRequest` request...")

		if !common.RPCMethods["GET"][config.QueryKiraStatus].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryKiraStatus].CacheEnabled {
				log.CustomLogger().Info("Starting search cache for `QueryKiraStatusRequest` request...")

				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)
					log.CustomLogger().Info("`QueryKiraStatusRequest` Returning cached response.")
					return
				}
			}
			response.Response, response.Error, statusCode = common.MakeTendermintRPCRequest(rpcAddr, "/status", "")
		}

		log.CustomLogger().Info("Finished `QueryKiraStatusRequest` request")

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryKiraStatus].CacheEnabled)
	}
}
