package interx

import (
	"net/http"
	"time"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/database"
	"github.com/KiraCore/interx/log"
	"github.com/KiraCore/interx/types"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// RegisterInterxFaucetRoutes registers faucet services.
func RegisterInterxFaucetRoutes(r *mux.Router, gwCosmosmux *runtime.ServeMux, rpcAddr string) {
	r.HandleFunc(config.FaucetRequestURL, FaucetRequest(gwCosmosmux, rpcAddr)).Methods("GET")
	common.AddRPCMethod("GET", config.FaucetRequestURL, "This is an API for faucet service.", false)
}

func serveFaucetInfo(r *http.Request, gwCosmosmux *runtime.ServeMux) (interface{}, interface{}, int) {
	faucetInfo := types.FaucetAccountInfo{}
	faucetInfo.Address = config.Config.Faucet.Address
	faucetInfo.Balances = common.GetAccountBalances(gwCosmosmux, r.Clone(r.Context()), config.Config.Faucet.Address)
	return faucetInfo, nil, http.StatusOK
}

/**
 * Error Codes
 * 0 : InternalServerError
 * 1 : Fail to send tokens
 * 100: Invalid address
 * 101: Claim time left
 * 102: Invalid token
 * 103: No need to send tokens
 * 104: Can't send tokens, less than minimum amount
 * 105: Not enough tokens
 */
func serveFaucet(r *http.Request, gwCosmosmux *runtime.ServeMux, _ types.InterxRequest, receiver string, token string) (interface{}, interface{}, int) {

	log.CustomLogger().Info("Starting `serveFaucet` request ...",
		"receiver", receiver,
		"token", token,
	)

	timeLeft := database.GetClaimTimeLeft(receiver)
	if timeLeft > 0 {
		log.CustomLogger().Info("`serveFaucet` checking time left for the receipt address",
			"receiver", receiver,
			"timeLeft", timeLeft,
		)
		txHash := "faucet claim limit exceeded, Please try again later."
		return types.FaucetResponse{Hash: txHash}, nil, http.StatusOK
	}

	result := common.IsEligibleToTransferToken(r, gwCosmosmux, receiver, token)
	if result {
		log.CustomLogger().Info("`serveFaucet` checking eligibility of the faucet account to transfer token",
			"receiver", receiver,
			"faucet has not enough token to claim", result,
		)
		txHash := "faucet has not enough token to claim"
		return types.FaucetResponse{Hash: txHash}, nil, http.StatusOK
	}

	txHash, err := common.TransferToken(receiver, token)
	if err != nil {
		log.CustomLogger().Error("[serveFaucet][TransferToken] failed to transfer token",
			"error", err,
		)
		return types.FaucetResponse{Hash: txHash}, nil, http.StatusNotFound
	}

	log.CustomLogger().Info("`serveFaucet` transfering token successfully done",
		"txHash", txHash,
	)

	// add new claim
	database.AddNewClaim(receiver, time.Now().UTC())

	log.CustomLogger().Info("`serveFaucet` adding new claim to the db successfully done")

	return types.FaucetResponse{Hash: txHash}, nil, http.StatusOK
}

// FaucetRequest is a function to handle faucet service.
func FaucetRequest(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		queries := r.URL.Query()
		claims := queries["claim"]
		tokens := queries["token"]

		if len(claims) == 0 && len(tokens) == 0 {
			log.CustomLogger().Info("`FaucetRequest` Starting faucet info resuest...")
			response.Response, response.Error, statusCode = serveFaucetInfo(r, gwCosmosmux)
		} else if len(claims) == 1 && len(tokens) == 1 {
			log.CustomLogger().Info("`FaucetRequest` Starting transfering token from faucet to receipt account",
				"receipt address", claims[0],
				"token", tokens[0],
			)
			response.Response, response.Error, statusCode = serveFaucet(r, gwCosmosmux, request, claims[0], tokens[0])
		} else {
			log.CustomLogger().Error("[FaucetRequest] Invalid parameters. failed to transfer token to the receipt from faucet",
				"receipt address", claims[0],
				"token", tokens[0],
			)
			response.Response, response.Error, statusCode = common.ServeError(0, "", "invalid query parameters", http.StatusBadRequest)
		}
		common.WrapResponse(w, request, *response, statusCode, false)
	}
}
