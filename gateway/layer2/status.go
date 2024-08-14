package layer2

import (
	"net/http"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// RegisterBlockRoutes registers layer2/status query routers.
func RegisterStatusRoutes(r *mux.Router, gwCosmosmux *runtime.ServeMux, rpcAddr string) {
	r.HandleFunc(config.QueryLayer2Status, QueryLayer2StatusRequest(gwCosmosmux, rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryLayer2Status, "This is an API to query layer2 status.", true)
}

func queryLayer2StatusHandle(rpcAddr string, r *http.Request) (interface{}, interface{}, int) {
	queries := mux.Vars(r)
	appName := queries["appName"]
	_ = r.ParseForm()
	return common.Layer2Status[appName], nil, http.StatusOK
}

// QueryLayer2StatusRequest is a function to query layer2 status.
func QueryLayer2StatusRequest(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		common.GetLogger().Info("[query-layer2-status] Entering Layer2 status query")

		if !common.RPCMethods["GET"][config.QueryLayer2Status].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryLayer2Status].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					common.GetLogger().Info("[query-layer2-status] Returning from the cache")
					return
				}
			}

			response.Response, response.Error, statusCode = queryLayer2StatusHandle(rpcAddr, r)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryLayer2Status].CachingEnabled)
	}
}
