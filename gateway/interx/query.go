package interx

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	functions "github.com/KiraCore/interx/functions"
	"github.com/KiraCore/interx/log"
	"github.com/KiraCore/interx/tasks"
	"github.com/KiraCore/interx/types"
	"github.com/KiraCore/interx/types/kira"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// RegisterInterxQueryRoutes registers query routers.
func RegisterInterxQueryRoutes(r *mux.Router, gwCosmosmux *runtime.ServeMux, rpcAddr string) {
	r.HandleFunc(config.QueryRPCMethods, QueryRPCMethods(rpcAddr)).Methods("GET")
	r.HandleFunc(config.QueryInterxFunctions, QueryInterxFunctions(rpcAddr)).Methods("GET")
	r.HandleFunc(config.QueryStatus, QueryStatusRequest(rpcAddr)).Methods("GET")
	r.HandleFunc(config.QueryAddrBook, QueryAddrBook(rpcAddr)).Methods("GET")
	r.HandleFunc(config.QueryNetInfo, QueryNetInfo(rpcAddr)).Methods("GET")
	r.HandleFunc(config.QueryDashboard, QueryDashboard(rpcAddr, gwCosmosmux)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryInterxFunctions, "This is an API to query interx functions.", true)
	common.AddRPCMethod("GET", config.QueryStatus, "This is an API to query interx status.", true)
	common.AddRPCMethod("GET", config.QueryAddrBook, "This is an API to query address book.", true)
	common.AddRPCMethod("GET", config.QueryNetInfo, "This is an API to query net info.", true)
}

// QueryRPCMethods is a function to query RPC methods.
func QueryRPCMethods(rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.CustomLogger().Info("Starting 'QueryRPCMethods' request...")

		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)
		statusCode := http.StatusOK

		log.CustomLogger().Info("Processed 'QueryRPCMethods' request.",
			"method", request.Method,
			"endpoint", request.Endpoint,
			"params", request.Params,
			"statusCode", statusCode,
			"error", response.Error,
		)

		response.Response = common.RPCMethods

		common.WrapResponse(w, request, *response, statusCode, false)

		log.CustomLogger().Info("Finished 'QueryRPCMethods' request.")
	}
}

func queryInterxFunctionsHandle(_ string) (interface{}, interface{}, int) {
	metadata := functions.GetInterxMetadata()

	return metadata, nil, http.StatusOK
}

// QueryInterxFunctions is a function to list functions and metadata.
func QueryInterxFunctions(rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.CustomLogger().Info("Starting 'QueryInterxFunctions' request...")

		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		response.Response, response.Error, statusCode = queryInterxFunctionsHandle(rpcAddr)

		log.CustomLogger().Info("Processed 'QueryInterxFunctions' request.",
			"method", request.Method,
			"endpoint", request.Endpoint,
			"params", request.Params,
			"statusCode", statusCode,
			"error", response.Error,
		)

		common.WrapResponse(w, request, *response, statusCode, false)

		log.CustomLogger().Info("Finished 'QueryInterxFunctions' request.")
	}
}

func queryStatusHandle(rpcAddr string) (interface{}, interface{}, int) {
	result := types.InterxStatus{}

	// Handle Interx Pubkey
	pubkeyBytes, err := config.EncodingCg.Amino.MarshalJSON(&config.Config.PubKey)

	if err != nil {
		common.GetLogger().Error("[query-status] Failed to marshal interx pubkey", err)
		return common.ServeError(0, "", err.Error(), http.StatusInternalServerError)
	}

	err = json.Unmarshal(pubkeyBytes, &result.InterxInfo.PubKey)
	if err != nil {
		common.GetLogger().Error("[query-status] Failed to add interx pubkey to status response", err)
		return common.ServeError(0, "", err.Error(), http.StatusInternalServerError)
	}

	result.ID = hex.EncodeToString(config.Config.PubKey.Address())

	// Handle Genesis
	genesis, checksum, err := GetGenesisResults(rpcAddr)
	if err != nil {
		common.GetLogger().Error("[query-status] Failed to query genesis ", err)
		return common.ServeError(0, "", err.Error(), http.StatusInternalServerError)
	}

	// Get Kira Status
	sentryStatus := common.GetKiraStatus((rpcAddr))

	if sentryStatus != nil {
		result.NodeInfo = sentryStatus.NodeInfo
		result.SyncInfo = sentryStatus.SyncInfo
		result.ValidatorInfo = sentryStatus.ValidatorInfo

		result.InterxInfo.Moniker = sentryStatus.NodeInfo.Moniker

		result.InterxInfo.LatestBlockHeight = sentryStatus.SyncInfo.LatestBlockHeight
		result.InterxInfo.CatchingUp = sentryStatus.SyncInfo.CatchingUp
	}

	result.InterxInfo.Node = config.Config.Node

	result.InterxInfo.KiraAddr = config.Config.Address
	result.InterxInfo.KiraPubKey = config.Config.PubKey.String()
	result.InterxInfo.FaucetAddr = config.Config.Faucet.Address
	result.InterxInfo.GenesisChecksum = checksum
	result.InterxInfo.ChainID = genesis.ChainID

	result.InterxInfo.InterxVersion = config.Config.InterxVersion
	result.InterxInfo.SekaiVersion = config.Config.SekaiVersion

	return result, nil, http.StatusOK
}

// QueryStatusRequest is a function to query status.
func QueryStatusRequest(rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.CustomLogger().Info("Starting `QueryStatus` request...")

		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		if !common.RPCMethods["GET"][config.QueryStatus].Enabled {

			log.CustomLogger().Error("Query `QueryStatus` is disabled.",
				"method", request.Method,
				"endpoint", request.Endpoint,
				"params", request.Params,
				"error", response.Error,
			)

			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryStatus].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					log.CustomLogger().Info("Cache hit for `QueryStatus` request.",
						"method", request.Method,
						"endpoint", request.Endpoint,
						"params", request.Params,
						"error", response.Error,
					)

					return
				}
			}

			response.Response, response.Error, statusCode = queryStatusHandle(rpcAddr)
		}

		log.CustomLogger().Info("Processed `QueryStatus` request.",
			"method", request.Method,
			"endpoint", request.Endpoint,
			"params", request.Params,
			"error", response.Error,
		)

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryStatus].CachingEnabled)

		log.CustomLogger().Info("Finished `QueryStatus` request.")
	}
}

func queryAddrBookHandler(_ string) (interface{}, interface{}, int) {
	return config.LoadAddressBooks(), nil, http.StatusOK
}

// QueryAddrBook is a function to query address book.
func QueryAddrBook(rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.CustomLogger().Info("Starting 'QueryAddrBook' request...")

		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		response.Response, response.Error, statusCode = queryAddrBookHandler(rpcAddr)

		log.CustomLogger().Info("Processed 'QueryAddrBook' request.",
			"method", request.Method,
			"endpoint", request.Endpoint,
			"params", request.Params,
			"error", response.Error,
		)

		common.WrapResponse(w, request, *response, statusCode, false)

		log.CustomLogger().Info("Finished 'QueryAddrBook' request.")
	}
}

func queryNetInfoHandler(rpcAddr string) (interface{}, interface{}, int) {
	netInfo, err := tasks.QueryNetInfo(rpcAddr)
	if err != nil {
		return common.ServeError(0, "", err.Error(), http.StatusInternalServerError)
	}

	return netInfo, nil, http.StatusOK
}

// QueryNetInfo is a function to query net info.
func QueryNetInfo(rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.CustomLogger().Info("Starting 'QueryNetInfo' request...")

		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		response.Response, response.Error, statusCode = queryNetInfoHandler(rpcAddr)

		log.CustomLogger().Info("Processed 'QueryNetInfo' request.",
			"method", request.Method,
			"endpoint", request.Endpoint,
			"params", request.Params,
			"error", response.Error,
		)

		common.WrapResponse(w, request, *response, statusCode, false)

		log.CustomLogger().Info("Finished 'QueryNetInfo' request.")
	}
}

func queryDashboardHandler(rpcAddr string, r *http.Request, gwCosmosmux *runtime.ServeMux) (interface{}, interface{}, int) {
	res := struct {
		ConsensusHealth       string `json:"consensus_health"`
		CurrentBlockValidator struct {
			Moniker string `json:"moniker"`
			Address string `json:"address"`
		} `json:"current_block_validator"`
		Validators struct {
			ActiveValidators   int `json:"active_validators"`
			PausedValidators   int `json:"paused_validators"`
			InactiveValidators int `json:"inactive_validators"`
			JailedValidators   int `json:"jailed_validators"`
			TotalValidators    int `json:"total_validators"`
			WaitingValidators  int `json:"waiting_validators"`
		} `json:"validators"`
		Blocks struct {
			CurrentHeight       int     `json:"current_height"`
			SinceGenesis        int     `json:"since_genesis"`
			PendingTransactions int     `json:"pending_transactions"`
			CurrentTransactions int     `json:"current_transactions"`
			LatestTime          float64 `json:"latest_time"`
			AverageTime         float64 `json:"average_time"`
		} `json:"blocks"`
		Proposals struct {
			Total      int    `json:"total"`
			Active     int    `json:"active"`
			Enacting   int    `json:"enacting"`
			Finished   int    `json:"finished"`
			Successful int    `json:"successful"`
			Proposers  string `json:"proposers"`
			Voters     string `json:"voters"`
		} `json:"proposals"`
	}{}

	result, failure, status := queryConsensusHandle(r.Clone(r.Context()), gwCosmosmux, rpcAddr)
	if failure != nil {
		return nil, failure, status
	}
	consensus := result.(kira.ConsensusResponse)

	// consensus health
	res.ConsensusHealth = consensus.ConsensusHealth

	// current block validator
	for _, validator := range tasks.AllValidators.Validators {
		if validator.Address == consensus.Proposer {
			res.CurrentBlockValidator.Moniker = validator.Moniker
			res.CurrentBlockValidator.Address = validator.Address
		}
	}

	// validators
	res.Validators = tasks.AllValidators.Status

	// current height
	sentryStatus := common.GetKiraStatus((rpcAddr))
	res.Blocks.CurrentHeight, _ = strconv.Atoi(sentryStatus.SyncInfo.LatestBlockHeight)

	// since genesis
	earliestBlockHeight, _ := strconv.Atoi(sentryStatus.SyncInfo.EarliestBlockHeight)
	res.Blocks.SinceGenesis = res.Blocks.CurrentHeight - earliestBlockHeight

	// pending transactions
	result, failure, status = queryUnconfirmedTransactionsHandler(rpcAddr, r.Clone(r.Context()))
	if failure != nil {
		return nil, failure, status
	}
	unconfirmedTxs := result.(struct {
		Count      int                                  `json:"n_txs"`
		Total      int                                  `json:"total"`
		TotalBytes int64                                `json:"total_bytes"`
		Txs        []types.TransactionUnconfirmedResult `json:"txs"`
	})
	res.Blocks.PendingTransactions = unconfirmedTxs.Total

	// current transactions
	blockchain, err := common.GetBlockchain(rpcAddr)
	if err != nil {
		return nil, nil, http.StatusInternalServerError
	}
	res.Blocks.CurrentTransactions = blockchain.BlockMetas[0].NumTxs

	// latest time
	res.Blocks.LatestTime = blockchain.BlockMetas[0].Header.Time.Sub(
		blockchain.BlockMetas[1].Header.Time,
	).Seconds()

	// average time
	res.Blocks.AverageTime = consensus.AverageBlockTime

	res.Proposals.Total = tasks.AllProposals.Status.TotalProposals
	res.Proposals.Active = tasks.AllProposals.Status.ActiveProposals
	res.Proposals.Enacting = tasks.AllProposals.Status.EnactingProposals
	res.Proposals.Finished = tasks.AllProposals.Status.FinishedProposals
	res.Proposals.Successful = tasks.AllProposals.Status.SuccessfulProposals
	res.Proposals.Proposers = tasks.AllProposals.Users.Proposers
	res.Proposals.Voters = tasks.AllProposals.Users.Voters

	return res, nil, http.StatusOK
}

// QueryDashboard is a function to query dashboard info.
func QueryDashboard(rpcAddr string, gwCosmosmux *runtime.ServeMux) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.CustomLogger().Info("Starting 'QueryDashboard' request...")

		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		response.Response, response.Error, statusCode = queryDashboardHandler(rpcAddr, r, gwCosmosmux)

		log.CustomLogger().Info("Processed 'QueryDashboard' request.",
			"method", request.Method,
			"endpoint", request.Endpoint,
			"params", request.Params,
			"error", response.Error,
		)

		common.WrapResponse(w, request, *response, statusCode, false)

		log.CustomLogger().Info("Finished 'QueryDashboard' request.")
	}
}
