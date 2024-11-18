package kira

import (
	"encoding/json"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/tasks"
	govTypes "github.com/KiraCore/interx/types/kira/gov"
	sekaitypes "github.com/KiraCore/sekai/types"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// RegisterKiraGovProposalRoutes registers kira gov proposal query routers.
func RegisterKiraGovProposalRoutes(r *mux.Router, gwCosmosmux *runtime.ServeMux, rpcAddr string) {
	r.HandleFunc(config.QueryProposals, QueryProposalsRequest(gwCosmosmux, rpcAddr)).Methods("GET")
	r.HandleFunc(config.QueryProposal, QueryProposalRequest(gwCosmosmux, rpcAddr)).Methods("GET")
	r.HandleFunc(config.QueryVoters, QueryVotersRequest(gwCosmosmux, rpcAddr)).Methods("GET")
	r.HandleFunc(config.QueryVotes, QueryVotesRequest(gwCosmosmux, rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryProposals, "This is an API to query all proposals.", true)
	common.AddRPCMethod("GET", config.QueryProposal, "This is an API to query a proposal by a given id.", true)
	common.AddRPCMethod("GET", config.QueryVoters, "This is an API to query voters by a given proposal_id.", true)
	common.AddRPCMethod("GET", config.QueryVotes, "This is an API to query votes by a given proposal_id.", true)
}

func queryProposalsHandler(r *http.Request, gwCosmosmux *runtime.ServeMux) (interface{}, interface{}, int) {
	var (
		proposer  string   = ""
		dateStart int      = -1
		dateEnd   int      = -1
		sortBy    string   = "dateDESC"
		types     []string = []string{}
		statuses  []string = []string{}
		voter     string
		offset    int = -1
		limit     int = -1
		err       error
	)

	//------------ Proposer ------------
	proposer = r.FormValue("proposer")

	//------------ Voter ------------
	voter = r.FormValue("voter")

	//------------ Sort ------------
	sortParam := r.FormValue("sort")
	if sortParam == "dateASC" || sortParam == "dateDESC" {
		sortBy = sortParam
	}

	//------------ Offset ------------
	if offsetStr := r.FormValue("offset"); offsetStr != "" {
		if offset, err = strconv.Atoi(offsetStr); err != nil {
			common.GetLogger().Error("[query-proposals] Failed to parse parameter 'offset': ", err)
			return common.ServeError(0, "failed to parse parameter 'offset'", err.Error(), http.StatusBadRequest)
		}
	}

	//------------ Limit ------------
	if limitStr := r.FormValue("limit"); limitStr != "" {
		if limit, err = strconv.Atoi(limitStr); err != nil {
			common.GetLogger().Error("[query-proposals] Failed to parse parameter 'limit': ", err)
			return common.ServeError(0, "failed to parse parameter 'limit'", err.Error(), http.StatusBadRequest)
		}

		if limit < 1 || limit > 100 {
			common.GetLogger().Error("[query-proposals] Invalid 'limit' range: ", limit)
			return common.ServeError(0, "'limit' should be 1 ~ 100", "", http.StatusBadRequest)
		}
	}

	//------------ Type ------------
	typeFlags := make(map[string]bool)
	for _, propType := range sekaitypes.AllProposalTypes {
		typeFlags[propType] = true
	}

	proposalTypesParam := r.FormValue("types")
	proposalTypesArray := strings.Split(proposalTypesParam, ",")
	for _, txType := range proposalTypesArray {
		if typeFlags[txType] {
			types = append(types, txType)
		}
	}

	//------------ Status ------------
	statusesParam := r.FormValue("status")
	statusesArray := strings.Split(statusesParam, ",")

	for _, sts := range statusesArray {
		if voteResult, found := govTypes.VoteResult[sts]; found {
			statuses = append(statuses, voteResult)
		}
	}

	//------------ Timestamps ------------
	if dateStStr := r.FormValue("dateStart"); dateStStr != "" {
		if dateStart, err = strconv.Atoi(dateStStr); err != nil {
			layout := "01/02/2006 3:04:05 PM"
			t, err1 := time.Parse(layout, dateStStr+" 12:00:00 AM")
			if err1 != nil {
				common.GetLogger().Error("[query-proposals] Failed to parse parameter 'dateStart': ", err1)
				return common.ServeError(0, "failed to parse parameter 'dateStart'", err1.Error(), http.StatusBadRequest)
			}

			dateStart = int(t.Unix())
		}
	}

	if dateEdStr := r.FormValue("dateEnd"); dateEdStr != "" {
		if dateEnd, err = strconv.Atoi(dateEdStr); err != nil {
			layout := "01/02/2006 3:04:05 PM"
			t, err1 := time.Parse(layout, dateEdStr+" 12:00:00 AM")
			if err1 != nil {
				common.GetLogger().Error("[query-proposals] Failed to parse parameter 'dateEnd': ", err1)
				return common.ServeError(0, "failed to parse parameter 'dateEnd'", err.Error(), http.StatusBadRequest)
			}

			dateEnd = int(t.Unix())
		}
	}

	//------------ Filter proposals by filtering options & Pagination ------------
	voterMap := map[string]bool{}

	voterLimit := sekaitypes.PageIterationLimit - 1
	voterOffset := 0
	for {
		reqPath := strings.Replace(r.URL.Path, "/api/kira/gov", "/kira/gov", -1)
		proposalsQueryRequest, _ := http.NewRequest("GET", reqPath+"?voter="+voter+"&pagination.offset="+strconv.Itoa(voterOffset)+"&pagination.limit="+strconv.Itoa(voterLimit), nil)

		proposalsQueryResponse, failure, statusCode := common.ServeGRPC(proposalsQueryRequest, gwCosmosmux)

		if proposalsQueryResponse == nil {
			return proposalsQueryResponse, failure, statusCode
		}

		byteData, err := json.Marshal(proposalsQueryResponse)
		if err != nil {
			return common.ServeError(0, "failed to parse proposals response", err.Error(), http.StatusBadRequest)
		}

		subResult := govTypes.ProposalsResponse{}
		err = json.Unmarshal(byteData, &subResult)
		if err != nil {
			return common.ServeError(0, "failed to parse proposals response", err.Error(), http.StatusBadRequest)
		}

		if len(subResult.Proposals) == 0 {
			break
		}

		for _, prop := range subResult.Proposals {
			voterMap[prop.ProposalID] = true
		}

		voterOffset += voterLimit
	}

	propResults := []govTypes.Proposal{}
	for _, proposal := range tasks.ProposalsMap {
		if !voterMap[proposal.ProposalID] {
			continue
		}

		if len(statuses) > 0 && !common.Include(statuses, proposal.Result) {
			continue
		}

		if len(types) > 0 && !common.Include(types, proposal.Type) {
			continue
		}

		if proposer != "" && proposal.Proposer != proposer {
			continue
		}

		if dateStart != -1 && proposal.Timestamp < dateStart {
			continue
		}

		if dateEnd != -1 && proposal.Timestamp > dateEnd {
			continue
		}

		propResults = append(propResults, proposal)
	}

	// sort proposals
	if sortBy == "dateASC" {
		sort.Slice(propResults, func(i, j int) bool {
			return propResults[i].Timestamp < propResults[j].Timestamp
		})
	} else {
		sort.Slice(propResults, func(i, j int) bool {
			return propResults[i].Timestamp > propResults[j].Timestamp
		})
	}

	totalCount := len(propResults)

	// pagination
	if limit == -1 {
		limit = 30
	}
	if offset == -1 {
		offset = 0
	}

	if offset > totalCount {
		offset = totalCount
	}

	propResults = propResults[offset:int(math.Min(float64(offset+limit), float64(len(propResults))))]

	//------------ Remove unnecessary fields ------------
	for idx, prop := range propResults {
		prop.Hash = ""
		prop.Timestamp = 0
		prop.BlockHeight = 0
		prop.Type = ""
		prop.Proposer = ""
		propResults[idx] = prop
	}

	res := govTypes.PropsResponse{
		TotalCount: totalCount,
		Proposals:  propResults,
	}
	return res, nil, http.StatusOK
}

// QueryProposalsRequest is a function to query all proposals.
func QueryProposalsRequest(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		common.GetLogger().Info("[query-proposals] Entering proposals query")

		if !common.RPCMethods["GET"][config.QueryProposals].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryProposals].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					common.GetLogger().Info("[query-proposals] Returning from the cache")
					return
				}
			}

			response.Response, response.Error, statusCode = queryProposalsHandler(r, gwCosmosmux)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryProposals].CachingEnabled)
	}
}

func queryProposalHandler(r *http.Request, gwCosmosmux *runtime.ServeMux, proposalID string, _ string) (interface{}, interface{}, int) {
	r.URL.Path = strings.Replace(r.URL.Path, "/api/kira/gov", "/kira/gov", -1)
	success, failure, status := common.ServeGRPC(r, gwCosmosmux)

	if success != nil {
		// query proposal by id
		result := make(map[string]interface{})
		byteData, err := json.Marshal(success)
		if err != nil {
			common.GetLogger().Error("[query-proposal] Invalid response format", err)
			return common.ServeError(0, "", err.Error(), http.StatusInternalServerError)
		}

		err = json.Unmarshal(byteData, &result)
		if err != nil {
			common.GetLogger().Error("[query-proposal] Invalid response format", err)
			return common.ServeError(0, "", err.Error(), http.StatusInternalServerError)
		}
		propResult := tasks.ProposalsMap[proposalID]
		propResult.Hash = ""
		propResult.Timestamp = 0
		propResult.BlockHeight = 0
		propResult.Type = ""
		propResult.Proposer = ""

		result["proposal"] = propResult
		success = result
	}
	return success, failure, status
}

// QueryProposalRequest is a function to query a proposal by a given proposal_id.
func QueryProposalRequest(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		queries := mux.Vars(r)
		proposalID := queries["proposal_id"]
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		common.GetLogger().Info("[query-proposal] Entering proposal query by proposal_id: ", proposalID)

		if !common.RPCMethods["GET"][config.QueryProposal].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryProposal].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					common.GetLogger().Info("[query-proposal] Returning from the cache")
					return
				}
			}

			response.Response, response.Error, statusCode = queryProposalHandler(r, gwCosmosmux, proposalID, rpcAddr)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryProposal].CachingEnabled)
	}
}

func queryVotersHandler(r *http.Request, gwCosmosmux *runtime.ServeMux) (interface{}, interface{}, int) {
	r.URL.Path = strings.Replace(r.URL.Path, "/api/kira/gov", "/kira/gov", -1)
	success, failure, statusCode := common.ServeGRPC(r, gwCosmosmux)

	if success != nil {
		voters, err := common.QueryVotersFromGrpcResult(success)
		if err != nil {
			common.GetLogger().Error("[query-voters] Invalid response format: ", err)
			return common.ServeError(0, "", err.Error(), http.StatusInternalServerError)
		}

		success = voters
	}

	return success, failure, statusCode
}

// QueryVotersRequest is a function to voters by a given proposal_id.
func QueryVotersRequest(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		queries := mux.Vars(r)
		proposalID := queries["proposal_id"]
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		common.GetLogger().Info("[query-voters] Entering proposal query by proposal_id: ", proposalID)

		if !common.RPCMethods["GET"][config.QueryVoters].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryVoters].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					common.GetLogger().Info("[query-voters] Returning from the cache")
					return
				}
			}

			response.Response, response.Error, statusCode = queryVotersHandler(r, gwCosmosmux)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryVoters].CachingEnabled)
	}
}

func queryVotesHandler(r *http.Request, gwCosmosmux *runtime.ServeMux) (interface{}, interface{}, int) {
	r.URL.Path = strings.Replace(r.URL.Path, "/api/kira/gov", "/kira/gov", -1)
	success, failure, statusCode := common.ServeGRPC(r, gwCosmosmux)

	if success != nil {
		votes, err := common.QueryVotesFromGrpcResult(success)
		if err != nil {
			common.GetLogger().Error("[query-votes] Invalid response format: ", err)
			return common.ServeError(0, "", err.Error(), http.StatusInternalServerError)
		}

		success = votes
	}

	return success, failure, statusCode
}

// QueryVotesRequest is a function to votes by a given proposal_id.
func QueryVotesRequest(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		queries := mux.Vars(r)
		proposalID := queries["proposal_id"]
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		common.GetLogger().Info("[query-votes] Entering proposal query by proposal_id: ", proposalID)

		if !common.RPCMethods["GET"][config.QueryVotes].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryVotes].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					common.GetLogger().Info("[query-votes] Returning from the cache")
					return
				}
			}

			response.Response, response.Error, statusCode = queryVotesHandler(r, gwCosmosmux)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryVotes].CachingEnabled)
	}
}
