package kira

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/tasks"
	"github.com/KiraCore/interx/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type QueryStakingPoolDelegatorsResponse struct {
	Pool       types.ValidatorPool `json:"pool"`
	Delegators []string            `json:"delegators,omitempty"`
}

type Undelegation struct {
	ID            int `json:"id,omitempty"`
	ValidatorInfo struct {
		Moniker string `json:"moniker,omitempty"`
		Address string `json:"address,omitempty"`
		ValKey  string `json:"valkey,omitempty"`
		Logo    string `json:"logo,omitempty"`
	} `json:"validator_info"`
	Tokens sdk.Coins `json:"tokens"`
	Expiry string    `json:"expiry,omitempty"`
}

// QueryDelegationsResponse is a struct to be used for query delegations response
type QueryUndelegationsResponse struct {
	Undelegations []Undelegation `json:"undelegations"`
	Pagination    struct {
		Total int `json:"total,string,omitempty"`
	} `json:"pagination,omitempty"`
}

type QueryBalancesResponse struct {
	Balances []types.Coin `json:"balances"`
}

func RegisterKiraMultiStakingRoutes(r *mux.Router, gwCosmosmux *runtime.ServeMux, rpcAddr string) {
	r.HandleFunc(config.QueryStakingPool, QueryStakingPoolRequest(gwCosmosmux, rpcAddr)).Methods("GET")
	r.HandleFunc(config.QueryDelegations, QueryDelegationsRequest(gwCosmosmux, rpcAddr)).Methods("GET")
	r.HandleFunc(config.QueryUndelegations, QueryUndelegationsRequest(gwCosmosmux, rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryStakingPool, "This is an API to query staking pool.", true)
	common.AddRPCMethod("GET", config.QueryDelegations, "This is an API to query delegations.", true)
	common.AddRPCMethod("GET", config.QueryUndelegations, "This is an API to query undelegations.", true)
}

func parseCoinString(input string) sdk.Coin {
	denom := ""
	amount := 0
	for _, poolToken := range tasks.PoolTokens {
		if strings.Contains(input, poolToken) {
			pattern := regexp.MustCompile("[^a-zA-Z0-9]+")
			amountStr := strings.ReplaceAll(input, poolToken, "")
			amountStr = pattern.ReplaceAllString(amountStr, "")

			denom = poolToken
			amount, _ = strconv.Atoi(amountStr)
		}
	}
	return sdk.Coin{
		Denom:  denom,
		Amount: sdk.NewIntFromUint64(uint64(amount)),
	}
}

// queryStakingPoolHandler is a function to query staking pool information for a validator
func queryStakingPoolHandler(r *http.Request, gwCosmosmux *runtime.ServeMux) (interface{}, interface{}, int) {
	queries := r.URL.Query()
	account := queries["validatorAddress"]

	if len(account) == 1 {
		valAddr, found := tasks.AddrToValidator[account[0]]
		if found {
			r.URL.RawQuery = ""
			r.URL.Path = strings.Replace(r.URL.Path, "/api/kira/staking-pool", "/kira/multistaking/v1beta1/staking_pool_delegators/"+valAddr, -1)
		}
	}

	success, failure, status := common.ServeGRPC(r, gwCosmosmux)
	if success != nil {
		result := QueryStakingPoolDelegatorsResponse{}

		byteData, err := json.Marshal(success)
		if err != nil {
			common.GetLogger().Error("[query-staking-pool] Invalid response format", err)
			return common.ServeError(0, "", err.Error(), http.StatusInternalServerError)
		}
		err = json.Unmarshal(byteData, &result)
		if err != nil {
			common.GetLogger().Error("[query-staking-pool] Invalid response format", err)
			return common.ServeError(0, "", err.Error(), http.StatusInternalServerError)
		}

		response := types.QueryValidatorPoolResult{}
		response.ID = result.Pool.ID
		response.Slashed = common.ConvertRate(result.Pool.Slashed)
		response.Commission = common.ConvertRate(result.Pool.Commission)

		response.VotingPower = []sdk.Coin{}
		for _, coinStr := range result.Pool.TotalStakingTokens {
			response.VotingPower = append(response.VotingPower, parseCoinString(coinStr))
		}

		response.TotalDelegators = int64(len(result.Delegators))
		response.Tokens = []string{}
		response.Tokens = tasks.PoolTokens
		success = response
	}
	return success, failure, status
}

func QueryStakingPoolRequest(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		common.GetLogger().Info("[query-staking-pool] Entering staking pool query")

		if !common.RPCMethods["GET"][config.QueryStakingPool].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryStakingPool].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					common.GetLogger().Info("[query-staking-pool] Returning from the cache")
					return
				}
			}

			response.Response, response.Error, statusCode = queryStakingPoolHandler(r, gwCosmosmux)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryStakingPool].CachingEnabled)
	}
}

func findValidator(address string) (types.QueryValidator, bool) {
	for _, validator := range tasks.AllValidators.Validators {
		if validator.Valkey == address {
			return validator, true
		}
	}
	return types.QueryValidator{}, false
}

// queryUndelegationsHandler is a function to query all undelegations for a delegator
func queryUndelegationsHandler(r *http.Request, gwCosmosmux *runtime.ServeMux) (interface{}, interface{}, int) {
	queries := r.URL.Query()
	account := queries["undelegatorAddress"]
	offset := queries["offset"]
	limit := queries["limit"]
	countTotal := queries["count_total"]
	response := QueryUndelegationsResponse{}

	if len(account) == 0 {
		common.GetLogger().Error("[query-undelegations] 'undelegator address' is not set")
		return common.ServeError(0, "'delegator address' is not set", "", http.StatusBadRequest)
	}

	var events = make([]string, 0, 1)
	if len(account) == 1 {
		events = append(events, fmt.Sprintf("delegator=%s", account[0]))
	}

	r.URL.RawQuery = strings.Join(events, "&")

	r.URL.Path = strings.Replace(r.URL.Path, "/api/kira/undelegations", "/kira/multistaking/v1beta1/undelegations", -1)
	success, failure, status := common.ServeGRPC(r, gwCosmosmux)
	if success != nil {
		result := types.QueryUndelegationsResult{}

		// parse user balance data and generate delegation responses from pool tokens
		byteData, err := json.Marshal(success)
		if err != nil {
			common.GetLogger().Error("[query-undelegations] Invalid response format", err)
			return common.ServeError(0, "", err.Error(), http.StatusInternalServerError)
		}

		err = json.Unmarshal(byteData, &result)
		if err != nil {
			common.GetLogger().Error("[query-undelegations] Invalid response format", err)
			return common.ServeError(0, "", err.Error(), http.StatusInternalServerError)
		}

		for _, undelegation := range result.Undelegations {
			undelegationData := Undelegation{}
			validator, found := findValidator(undelegation.ValAddress)

			if !found {
				continue
			}

			undelegationData.ID = int(undelegation.ID)

			undelegationData.ValidatorInfo.Address = validator.Address
			undelegationData.ValidatorInfo.Logo = validator.Logo
			undelegationData.ValidatorInfo.Moniker = validator.Moniker
			undelegationData.ValidatorInfo.ValKey = validator.Valkey
			undelegationData.Expiry = undelegation.Expiry

			for _, token := range undelegation.Amount {
				undelegationData.Tokens = append(undelegationData.Tokens, parseCoinString(token))
			}

			response.Undelegations = append(response.Undelegations, undelegationData)
		}

		// apply pagination
		from := 0
		total := len(response.Undelegations)
		count := int(math.Min(float64(50), float64(total)))
		if len(countTotal) == 1 && countTotal[0] == "true" {
			response.Pagination.Total = total
		}
		if len(offset) == 1 {
			from, err = strconv.Atoi(offset[0])
			if err != nil {
				common.GetLogger().Error("[query-undelegation] Failed to parse parameter 'offset': ", err)
				return common.ServeError(0, "failed to parse parameter 'offset'", err.Error(), http.StatusBadRequest)
			}
		}
		if len(limit) == 1 {
			count, err = strconv.Atoi(limit[0])
			if err != nil {
				common.GetLogger().Error("[query-undelegation] Failed to parse parameter 'limit': ", err)
				return common.ServeError(0, "failed to parse parameter 'limit'", err.Error(), http.StatusBadRequest)
			}
		}

		from = int(math.Min(float64(from), float64(total)))
		to := int(math.Min(float64(from+count), float64(total)))
		response.Undelegations = response.Undelegations[from:to]
		success = response
	}

	return success, failure, status
}

func QueryUndelegationsRequest(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		common.GetLogger().Info("[query-undelegations] Entering undelegations query")

		if !common.RPCMethods["GET"][config.QueryUndelegations].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryUndelegations].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					common.GetLogger().Info("[query-undelegations] Returning from the cache")
					return
				}
			}

			response.Response, response.Error, statusCode = queryUndelegationsHandler(r, gwCosmosmux)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryUndelegations].CachingEnabled)
	}
}

// queryDelegationsHandler is a function to query all delegations for a delegator
func queryDelegationsHandler(r *http.Request, gwCosmosmux *runtime.ServeMux) (interface{}, interface{}, int) {
	queries := r.URL.Query()
	account := queries["delegatorAddress"]
	offset := queries["offset"]
	limit := queries["limit"]
	countTotal := queries["count_total"]
	response := types.QueryDelegationsResult{}

	if len(account) == 1 {
		r.URL.Path = strings.Replace(r.URL.Path, "/api/kira/delegations", "/cosmos/bank/v1beta1/balances/"+account[0], -1)
	}

	// fetch pool share tokens for the account
	success, failure, status := common.ServeGRPC(r, gwCosmosmux)
	if success != nil {
		result := QueryBalancesResponse{}

		// parse user balance data and generate delegation responses from pool tokens
		byteData, err := json.Marshal(success)
		if err != nil {
			common.GetLogger().Error("[query-staking-pool] Invalid response format", err)
			return common.ServeError(0, "", err.Error(), http.StatusInternalServerError)
		}
		err = json.Unmarshal(byteData, &result)
		if err != nil {
			common.GetLogger().Error("[query-staking-pool] Invalid response format", err)
			return common.ServeError(0, "", err.Error(), http.StatusInternalServerError)
		}

		for _, balance := range result.Balances {
			delegation := types.Delegation{}
			denomParts := strings.Split(balance.Denom, "/")
			// if denom format is v{N}/XXX,
			if len(denomParts) == 2 && denomParts[0][0] == 'v' {
				// fetch pool id from denom
				poolID, err := strconv.Atoi(denomParts[0][1:])
				if err != nil {
					continue
				}

				// get pool data from id
				pool, found := tasks.AllPools[int64(poolID)]
				if !found {
					continue
				}
				// fill up PoolInfo
				delegation.PoolInfo.ID = pool.ID
				delegation.PoolInfo.Commission = common.ConvertRate(pool.Commission)
				if pool.Enabled {
					delegation.PoolInfo.Status = "ENABLED"
				} else {
					delegation.PoolInfo.Status = "DISABLED"
				}
				delegation.PoolInfo.Tokens = tasks.PoolTokens

				// fill up ValidatorInfo
				validator, found := tasks.PoolToValidator[pool.ID]
				if found {
					delegation.ValidatorInfo.Address = validator.Address
					delegation.ValidatorInfo.ValKey = validator.Valkey
					delegation.ValidatorInfo.Moniker = validator.Moniker
					delegation.ValidatorInfo.Website = validator.Website
					delegation.ValidatorInfo.Logo = validator.Logo
				}
				response.Delegations = append(response.Delegations, delegation)
			}
		}

		// apply pagination
		from := 0
		total := len(response.Delegations)
		count := int(math.Min(float64(50), float64(total)))
		if len(countTotal) == 1 && countTotal[0] == "true" {
			response.Pagination.Total = total
		}
		if len(offset) == 1 {
			from, err = strconv.Atoi(offset[0])
			if err != nil {
				common.GetLogger().Error("[query-staking-pool] Failed to parse parameter 'offset': ", err)
				return common.ServeError(0, "failed to parse parameter 'offset'", err.Error(), http.StatusBadRequest)
			}
		}
		if len(limit) == 1 {
			count, err = strconv.Atoi(limit[0])
			if err != nil {
				common.GetLogger().Error("[query-staking-pool] Failed to parse parameter 'limit': ", err)
				return common.ServeError(0, "failed to parse parameter 'limit'", err.Error(), http.StatusBadRequest)
			}
		}

		from = int(math.Min(float64(from), float64(total)))
		to := int(math.Min(float64(from+count), float64(total)))
		response.Delegations = response.Delegations[from:to]
		success = response
	}

	return success, failure, status
}

func QueryDelegationsRequest(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		common.GetLogger().Info("[query-delegations] Entering delegations query")

		if !common.RPCMethods["GET"][config.QueryDelegations].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryDelegations].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					common.GetLogger().Info("[query-staking-pool] Returning from the cache")
					return
				}
			}

			response.Response, response.Error, statusCode = queryDelegationsHandler(r, gwCosmosmux)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryDelegations].CachingEnabled)
	}
}
