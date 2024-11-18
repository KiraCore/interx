package kira

import (
	"net/http"
	"strings"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// RegisterKiraUbiRoutes registers kira ubi query routers.
func RegisterKiraUbiRoutes(r *mux.Router, gwCosmosmux *runtime.ServeMux, rpcAddr string) {
	r.HandleFunc(config.QueryUBIRecords, QueryUBIRecordsRequest(gwCosmosmux, rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryUBIRecords, "This is an API to query ubi records.", true)
}

func QueryUBIRecordsHandler(r *http.Request, gwCosmosmux *runtime.ServeMux) (interface{}, interface{}, int) {
	queries := r.URL.Query()
	name := queries["name"]

	if len(name) == 1 {
		r.URL.RawQuery = ""
		r.URL.Path = strings.Replace(r.URL.Path, "/api/kira/ubi-records", "/kira/ubi/ubi_record/"+name[0], -1)
	} else {
		r.URL.RawQuery = ""
		r.URL.Path = strings.Replace(r.URL.Path, "/api/kira/ubi-records", "/kira/ubi/ubi_records", -1)
	}

	return common.ServeGRPC(r, gwCosmosmux)
}

// QueryUBIRecordsRequest is a function to query list of all ubi records.
func QueryUBIRecordsRequest(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		common.GetLogger().Info("[query-ubi-records] Entering upgrade plan query")

		if !common.RPCMethods["GET"][config.QueryUBIRecords].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryUBIRecords].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					common.GetLogger().Info("[query-ubi-records] Returning from the cache")
					return
				}
			}

			response.Response, response.Error, statusCode = QueryUBIRecordsHandler(r, gwCosmosmux)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryUBIRecords].CachingEnabled)
	}
}
