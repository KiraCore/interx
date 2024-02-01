package cosmos

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/types"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// RegisterCosmosBankRoutes registers query routers.
func RegisterCosmosBankRoutes(r *mux.Router, gwCosmosmux *runtime.ServeMux, rpcAddr string) {
	r.HandleFunc(config.QueryTotalSupply, QuerySupplyRequest(gwCosmosmux, rpcAddr)).Methods("GET")
	r.HandleFunc(config.QueryBalances, QueryBalancesRequest(gwCosmosmux, rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryTotalSupply, "This is an API to query total supply.", true)
	common.AddRPCMethod("GET", config.QueryBalances, "This is an API to query balances of an address.", true)
}

func querySupplyHandle(r *http.Request, gwCosmosmux *runtime.ServeMux) (interface{}, interface{}, int) {
	r.URL.Path = strings.Replace(r.URL.Path, "/api/kira", "/cosmos/bank/v1beta1", -1)
	return common.ServeGRPC(r, gwCosmosmux)
}

// QuerySupplyRequest is a function to query total supply.
func QuerySupplyRequest(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		common.GetLogger().Info("[query-supply] Entering total supply query")

		if !common.RPCMethods["GET"][config.QueryTotalSupply].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryTotalSupply].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					common.GetLogger().Info("[query-supply] Returning from the cache")
					return
				}
			}

			response.Response, response.Error, statusCode = querySupplyHandle(r, gwCosmosmux)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryTotalSupply].CachingEnabled)
	}
}

func queryBalancesHandle(r *http.Request, gwCosmosmux *runtime.ServeMux) (interface{}, interface{}, int) {
	params := mux.Vars(r)
	bech32addr := params["address"]

	// request could have tokens list if so, returns those
	tokensParam := r.URL.Query().Get("tokens")
	excludedParam := r.URL.Query().Get("excluded")
	derivedParam := r.URL.Query().Get("derived")

	// use default balances query when no custom params set
	if tokensParam == "" && excludedParam == "" && (derivedParam != "true" && derivedParam != "false") {
		queries := r.URL.Query()
		offset := queries["offset"]
		limit := queries["limit"]
		countTotal := queries["count_total"]
		var events = make([]string, 0, 3)
		if len(offset) == 1 {
			events = append(events, fmt.Sprintf("pagination.offset=%s", offset[0]))
		}
		if len(limit) == 1 {
			events = append(events, fmt.Sprintf("pagination.limit=%s", limit[0]))
		}
		if len(countTotal) == 1 {
			events = append(events, fmt.Sprintf("pagination.count_total=%s", countTotal[0]))
		} else if len(limit) == 1 { // set count_total flag to true as default when limit flag's set
			events = append(events, "pagination.count_total=true")
		}

		r.URL.RawQuery = strings.Join(events, "&")
		r.URL.Path = fmt.Sprintf("/cosmos/bank/v1beta1/balances/%s", bech32addr)
		return common.ServeGRPC(r, gwCosmosmux)
	}

	type Pagination struct {
		NextKey int `json:"next_key"`
		Total   int `json:"total"`
	}

	type BalancesResult struct {
		Balances   []types.Coin `json:"balances"`
		Pagination *Pagination  `json:"pagination,omitempty"`
	}

	allBalances := common.GetAllBalances(gwCosmosmux, r.Clone(r.Context()), bech32addr)

	data := make([]types.Coin, 0)
	if tokensParam != "" {
		tokensList := strings.Split(tokensParam, ",")
		tokensMap := make(map[string]bool)
		for _, token := range tokensList {
			tokensMap[token] = true
		}

		for _, balance := range allBalances {
			if tokensMap[balance.Denom] {
				data = append(data, balance)
			}
		}
		result := BalancesResult{
			Balances: data,
		}

		return result, nil, http.StatusOK
	}

	excludedList := strings.Split(excludedParam, ",")
	excludedMap := make(map[string]bool)
	for _, token := range excludedList {
		excludedMap[token] = true
	}

	filteredData := make([]types.Coin, 0)
	for _, balance := range allBalances {
		denom := balance.Denom
		if excludedMap[denom] {
			continue
		}

		isDerivedDenom := len(denom) > 2 && denom[0] == 'v' && strings.Contains(denom, "/")
		if derivedParam == "true" && !isDerivedDenom {
			continue
		}
		if derivedParam == "false" && isDerivedDenom {
			continue
		}

		filteredData = append(filteredData, balance)
	}

	// if request does not have tokens list, return with pagination
	offsetParam := r.URL.Query().Get("offset")
	offset, err := strconv.Atoi(offsetParam)
	if err != nil {
		offset = 0
	}
	limitParam := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitParam)
	if err != nil {
		limit = 100
	}

	lastIndex := offset + limit
	if lastIndex > len(filteredData) {
		lastIndex = len(filteredData)
	}
	data = filteredData[offset:lastIndex]

	result := BalancesResult{
		Balances: data,
		Pagination: &Pagination{
			NextKey: lastIndex,
			Total:   len(filteredData),
		},
	}

	return result, nil, http.StatusOK
}

// QueryBalancesRequest is a function to query balances.
func QueryBalancesRequest(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		queries := mux.Vars(r)
		bech32addr := queries["address"]
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		common.GetLogger().Info("[query-balances] Entering balances query: ", bech32addr)

		if !common.RPCMethods["GET"][config.QueryBalances].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryBalances].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					common.GetLogger().Info("[query-balances] Returning from the cache: ", bech32addr)
					return
				}
			}

			response.Response, response.Error, statusCode = queryBalancesHandle(r, gwCosmosmux)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryBalances].CachingEnabled)
	}
}
