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

// RegisterKiraGovRoleRoutes registers kira gov roles query routers.
func RegisterKiraGovRoleRoutes(r *mux.Router, gwCosmosmux *runtime.ServeMux, rpcAddr string) {
	r.HandleFunc(config.QueryRoles, QueryRolesRequest(gwCosmosmux, rpcAddr)).Methods("GET")
	r.HandleFunc(config.QueryRolesByAddress, QueryRolesByAddressRequest(gwCosmosmux, rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryRoles, "This is an API to query all role.", true)
	common.AddRPCMethod("GET", config.QueryRolesByAddress, "This is an API to query all role by address.", true)
}

func queryRolesHandler(r *http.Request, gwCosmosmux *runtime.ServeMux) (interface{}, interface{}, int) {
	r.URL.Path = strings.Replace(r.URL.Path, "/api/kira/gov", "/kira/gov", -1)
	return common.ServeGRPC(r, gwCosmosmux)
}

// QueryRolesRequest is a function to query all roles.
func QueryRolesRequest(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		log.CustomLogger().Info("[query-roles] Entering roles query")

		if !common.RPCMethods["GET"][config.QueryRoles].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryRoles].CacheEnabled {

				log.CustomLogger().Info("Starting search cache for `QueryRolesRequest` request...")

				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					log.CustomLogger().Info("[query-roles] Returning from the cache")
					return
				}
			}

			response.Response, response.Error, statusCode = queryRolesHandler(r, gwCosmosmux)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryRoles].CacheEnabled)
	}
}

func queryRolesByAddressHandler(r *http.Request, gwCosmosmux *runtime.ServeMux) (interface{}, interface{}, int) {
	queries := mux.Vars(r)
	bech32addr := queries["val_addr"]

	_, err := sdk.AccAddressFromBech32(bech32addr)
	if err != nil {
		log.CustomLogger().Error("[query-account] Invalid bech32addr: ", bech32addr)
		return common.ServeError(0, "", err.Error(), http.StatusBadRequest)
	}

	r.URL.Path = fmt.Sprintf("/api/kira/gov/roles_by_address/%s", bech32addr)

	r.URL.Path = strings.Replace(r.URL.Path, "/api/kira/gov", "/kira/gov", -1)
	return common.ServeGRPC(r, gwCosmosmux)
}

// QueryRolesByAddressRequest is a function to query all roles by address.
func QueryRolesByAddressRequest(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		log.CustomLogger().Info("[query-roles-by-address] Entering roles by address query")

		if !common.RPCMethods["GET"][config.QueryRolesByAddress].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryRolesByAddress].CacheEnabled {

				log.CustomLogger().Info("Starting search cache for `QueryRolesByAddressRequest` request...")

				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					log.CustomLogger().Info("[query-roles-by-address] Returning from the cache")
					return
				}
			}

			response.Response, response.Error, statusCode = queryRolesByAddressHandler(r, gwCosmosmux)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryRolesByAddress].CacheEnabled)
	}
}
