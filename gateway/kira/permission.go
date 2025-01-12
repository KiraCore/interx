package kira

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// RegisterKiraGovPermissionRoutes registers kira gov permissions query routers.
func RegisterKiraGovPermissionRoutes(r *mux.Router, gwCosmosmux *runtime.ServeMux, rpcAddr string) {
	r.HandleFunc(config.QueryPermissionsByAddress, QueryPermissionsByAddressRequest(gwCosmosmux, rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryPermissionsByAddress, "This is an API to query all permissions by address.", true)
}

func queryPermissionsByAddressHandler(r *http.Request, gwCosmosmux *runtime.ServeMux) (interface{}, interface{}, int) {
	queries := mux.Vars(r)
	bech32addr := queries["val_addr"]

	_, err := sdk.AccAddressFromBech32(bech32addr)
	if err != nil {
		log.CustomLogger().Error("[query-account] Invalid bech32addr: ", bech32addr)
		return common.ServeError(0, "", err.Error(), http.StatusBadRequest)
	}

	r.URL.Path = fmt.Sprintf("/api/kira/gov/permissions_by_address/%s", bech32addr)

	r.URL.Path = strings.Replace(r.URL.Path, "/api/kira/gov", "/kira/gov", -1)
	return common.ServeGRPC(r, gwCosmosmux)
}

// QueryPermissionsByAddressRequest is a function to query all permissions by address.
func QueryPermissionsByAddressRequest(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		log.CustomLogger().Info("[query-permissions-by-address] Entering permissions by address query")

		if !common.RPCMethods["GET"][config.QueryPermissionsByAddress].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryPermissionsByAddress].CacheEnabled {

				log.CustomLogger().Info("Starting search cache for `QueryPermissionsByAddressRequest` request...")

				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					log.CustomLogger().Info("[query-permissions-by-address] Returning from the cache")
					return
				}
			}

			response.Response, response.Error, statusCode = queryPermissionsByAddressHandler(r, gwCosmosmux)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryRoles].CacheEnabled)
	}
}
