package gateway

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	sekaitypes "github.com/KiraCore/sekai/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/zap"

	"github.com/saiset-co/sai-interx-manager/logger"
	"github.com/saiset-co/sai-interx-manager/types"
	"github.com/saiset-co/sai-interx-manager/utils"
)

func (g *CosmosGateway) txByHash(hash string) (interface{}, error) {
	txHash := strings.TrimPrefix(hash, "0x")

	gatewayReq, err := http.NewRequestWithContext(g.context.Context, "GET", "/cosmos/tx/v1beta1/txs/"+txHash, nil)
	if err != nil {
		logger.Logger.Error("[query-transaction-by-hash] Create request failed", zap.Error(err))
		return nil, err
	}

	grpcBytes, err := g.grpcProxy.ServeGRPC(gatewayReq)
	if err != nil {
		logger.Logger.Error("[query-transaction-by-hash] Serve request failed", zap.Error(err))
		return nil, err
	}

	var result interface{}

	err = json.Unmarshal(grpcBytes, &result)
	if err != nil {
		logger.Logger.Error("[query-transaction-by-hash] Unmarshal response failed", zap.Error(err))
		return nil, err
	}

	return result, nil
}

func (g *CosmosGateway) blockById(req types.InboundRequest, blockID string) (interface{}, error) {
	req.Payload["height"] = blockID
	return g.blocks(req)
}

func (g *CosmosGateway) txByBlock(req types.InboundRequest, blockID string) (interface{}, error) {
	req.Payload["height"] = blockID
	return g.transactions(req)
}

func (g *CosmosGateway) parseCoinString(input string) (*sdk.Coin, error) {
	denom := ""
	amount := 0

	tokens, err := g.tokens()
	if err != nil {
		logger.Logger.Error("[parse-coin-string] Failed to get tokens", zap.Error(err))
		return nil, err
	}

	for _, poolToken := range tokens {
		if strings.Contains(input, poolToken) {
			pattern := regexp.MustCompile("[^a-zA-Z0-9]+")
			amountStr := strings.ReplaceAll(input, poolToken, "")
			amountStr = pattern.ReplaceAllString(amountStr, "")

			denom = poolToken
			amount, _ = strconv.Atoi(amountStr)
		}
	}
	return &sdk.Coin{
		Denom:  denom,
		Amount: sdk.NewIntFromUint64(uint64(amount)),
	}, nil
}

func (g *CosmosGateway) executionFee(req types.InboundRequest) (interface{}, error) {
	type ExecutionFeeRequest struct {
		Message string `json:"message,omitempty"`
	}

	request := ExecutionFeeRequest{}

	jsonData, err := json.Marshal(req.Payload)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonData, &request)
	if err != nil {
		return nil, err
	}

	gatewayReq, err := http.NewRequestWithContext(g.context.Context, "GET", "/kira/gov/execution_fee/"+request.Message, nil)
	if err != nil {
		logger.Logger.Error("[execution-fee] Create request failed", zap.Error(err))
		return nil, err
	}

	grpcBytes, err := g.grpcProxy.ServeGRPC(gatewayReq)
	if err != nil {
		logger.Logger.Error("[execution-fee] Serve request failed", zap.Error(err))
		return nil, err
	}

	var result interface{}

	err = json.Unmarshal(grpcBytes, &result)
	if err != nil {
		logger.Logger.Error("[execution-fee] Invalid response format", zap.Error(err))
		return nil, err
	}

	return result, nil
}

func (g *CosmosGateway) networkProperties() (interface{}, error) {
	gatewayReq, err := http.NewRequestWithContext(g.context.Context, "GET", "/kira/gov/network_properties", nil)
	if err != nil {
		logger.Logger.Error("[query-network-properties] Create request failed", zap.Error(err))
		return nil, err
	}

	response, err := g.grpcProxy.ServeGRPC(gatewayReq)
	if err != nil {
		logger.Logger.Error("[query-network-properties] Serve request failed", zap.Error(err))
		return nil, err
	}

	result, err := utils.QueryNetworkPropertiesFromGrpcResult(response)
	if err != nil {
		logger.Logger.Error("[query-network-properties] Invalid response format", zap.Error(err))
		return nil, err
	}

	return result, nil
}

func (g *CosmosGateway) stakingPool(req types.InboundRequest) (interface{}, error) {
	type StakingPoolRequest struct {
		Account string `json:"validatorAddress,omitempty"`
	}

	request := StakingPoolRequest{}

	jsonData, err := json.Marshal(req.Payload)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonData, &request)
	if err != nil {
		return nil, err
	}

	tokens, err := g.tokens()
	if err != nil {
		logger.Logger.Error("[query-staking-pool] Getting tokens failed", zap.Error(err))
		return nil, err
	}

	validators, err := g.dashboard()
	if err != nil {
		logger.Logger.Error("[query-staking-pool] Getting validators failed", zap.Error(err))
		return nil, err
	}

	if request.Account == "" {
		err = fmt.Errorf("[query-staking-pool] validatorAddress required")
		return nil, err
	}

	valAddr, found := validators.AddrToValidator[request.Account]
	if !found {
		err = fmt.Errorf("[query-staking-pool] validatorAddress not found")
		return nil, err
	}

	gatewayReq, err := http.NewRequestWithContext(g.context.Context, "GET", "/kira/multistaking/v1beta1/staking_pool_delegators/"+valAddr, nil)
	if err != nil {
		logger.Logger.Error("[query-staking-pool] Create request failed", zap.Error(err))
		return nil, err
	}

	response, err := g.grpcProxy.ServeGRPC(gatewayReq)
	if err != nil {
		logger.Logger.Error("[query-staking-pool] Serve request failed", zap.Error(err))
		return nil, err
	}

	responseResult := types.QueryStakingPoolDelegatorsResponse{}

	err = json.Unmarshal(response, &responseResult)
	if err != nil {
		logger.Logger.Error("[query-staking-pool] Invalid response format", zap.Error(err))
		return nil, err
	}

	newResponse := types.QueryValidatorPoolResult{}
	newResponse.ID = responseResult.Pool.ID
	newResponse.Slashed = utils.ConvertRate(responseResult.Pool.Slashed)
	newResponse.Commission = utils.ConvertRate(responseResult.Pool.Commission)

	newResponse.VotingPower = []sdk.Coin{}
	for _, coinStr := range responseResult.Pool.TotalStakingTokens {
		coin, err := g.parseCoinString(coinStr)
		if err != nil {
			logger.Logger.Error("[query-staking-pool] Coin can not be parsed", zap.Error(err))
			continue
		}
		newResponse.VotingPower = append(newResponse.VotingPower, *coin)
	}

	newResponse.TotalDelegators = int64(len(responseResult.Delegators))
	newResponse.Tokens = []string{}
	newResponse.Tokens = tokens

	return newResponse, nil
}

func (g *CosmosGateway) undelegations(req types.InboundRequest) (interface{}, error) {
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

	type QueryUndelegationsResponse struct {
		Undelegations []Undelegation `json:"undelegations"`
		Pagination    struct {
			Total int `json:"total,string,omitempty"`
		} `json:"pagination,omitempty"`
	}

	type UndelegationsRequest struct {
		Account    string `json:"undelegatorAddress,omitempty"`
		Limit      int    `json:"limit,string,omitempty"`
		Offset     int    `json:"offset,string,omitempty"`
		CountTotal int    `json:"count_total,string,omitempty"`
	}

	request := UndelegationsRequest{}
	response := QueryUndelegationsResponse{}

	jsonData, err := json.Marshal(req.Payload)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonData, &request)
	if err != nil {
		return nil, err
	}

	if request.Account == "" {
		err = fmt.Errorf("[query-undelegations] validatorAddress required")
		return nil, err
	}

	gatewayReq, err := http.NewRequestWithContext(g.context.Context, "GET", "/kira/multistaking/v1beta1/undelegations", nil)
	if err != nil {
		logger.Logger.Error("[query-undelegations] Create request failed", zap.Error(err))
		return nil, err
	}

	q := gatewayReq.URL.Query()
	q.Add("delegator", request.Account)
	gatewayReq.URL.RawQuery = q.Encode()

	success, err := g.grpcProxy.ServeGRPC(gatewayReq)
	if err != nil {
		logger.Logger.Error("[query-undelegations] Serve request failed", zap.Error(err))
		return nil, err
	}

	validators, err := g.allValidators()
	if err != nil {
		logger.Logger.Error("[query-undelegations] Getting validators failed", zap.Error(err))
		return nil, err
	}

	result := types.QueryUndelegationsResult{}

	err = json.Unmarshal(success, &result)
	if err != nil {
		logger.Logger.Error("[query-undelegations] Invalid response format", zap.Error(err))
		return nil, err
	}

	for _, undelegation := range result.Undelegations {
		undelegationData := Undelegation{}

		validator := types.QueryValidator{}

		for _, _validator := range validators.Validators {
			if _validator.Valkey == undelegation.ValAddress {
				validator = _validator
			}
		}

		if validator.Address == "" {
			continue
		}

		undelegationData.ID = int(undelegation.ID)

		undelegationData.ValidatorInfo.Address = validator.Address
		undelegationData.ValidatorInfo.Logo = validator.Logo
		undelegationData.ValidatorInfo.Moniker = validator.Moniker
		undelegationData.ValidatorInfo.ValKey = validator.Valkey
		undelegationData.Expiry = undelegation.Expiry

		for _, token := range undelegation.Amount {
			coin, err := g.parseCoinString(token)
			if err != nil {
				logger.Logger.Error("[query-undelegations] Parsing coin failed", zap.Error(err))
				continue
			}
			undelegationData.Tokens = append(undelegationData.Tokens, *coin)
		}

		response.Undelegations = append(response.Undelegations, undelegationData)
	}

	if request.Limit > 0 {
		total := len(response.Undelegations)
		count := int(math.Min(float64(request.Limit), float64(total)))

		if request.CountTotal > 0 {
			response.Pagination.Total = total
		}

		from := int(math.Min(float64(request.Offset), float64(total)))
		to := int(math.Min(float64(request.Offset+count), float64(total)))

		response.Undelegations = response.Undelegations[from:to]
	}

	return response, nil
}

func (g *CosmosGateway) transactions(req types.InboundRequest) (interface{}, error) {
	var result types.TxsResultResponse
	var includeConfirmed = true
	var includeFailed = false
	var includeUnconfirmed = false
	var unconfirmedTxs []types.TxResponse
	var confirmedOffset = 0
	var confirmedLimit = 0

	request := types.QueryTxsParams{
		Offset: 0,
		Limit:  sekaitypes.PageIterationLimit - 1,
	}

	jsonData, err := json.Marshal(req.Payload)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonData, &request)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("pagination.offset", strconv.Itoa(request.Offset))
	params.Add("pagination.limit", strconv.Itoa(request.Limit))
	if request.BlockId != "" {
		params.Add("events", fmt.Sprintf("tx.height='%s'", request.BlockId))
	} else {
		params.Add("events", "tx.minheight='1'")
	}

	if request.Address != "" {
		params.Add("events", fmt.Sprintf("transfer.sender='%s'", request.Address))
		params.Add("events", fmt.Sprintf("transfer.recipient='%s'", request.Address))
	}

	for _, txType := range request.Types {
		params.Add("message.action", txType)
	}

	if request.StartDate > 0 {
		startTime := time.Unix(request.StartDate, 0).UTC()
		params.Add("tx.mintime", startTime.Format(time.RFC3339))
	}

	if request.EndDate > 0 {
		endTime := time.Unix(request.EndDate, 0).UTC()
		params.Add("tx.maxtime", endTime.Format(time.RFC3339))
	}

	if len(request.Statuses) > 0 {
		includeConfirmed = false
		includeFailed = false
		includeUnconfirmed = false

		for _, status := range request.Statuses {
			if status == "success" {
				includeConfirmed = true
			} else if status == "failed" {
				includeFailed = true
			} else if status == "unconfirmed" {
				includeUnconfirmed = true
			}
		}
	}

	if includeConfirmed && !includeFailed {
		params.Add("tx.code", "0")
	} else if !includeConfirmed && includeFailed {
		params.Add("tx.code.gt", "0")
	}

	totalConfirmedCount, err := g.getConfirmedTransactionsCount(params, includeConfirmed, includeFailed)
	if err != nil {
		logger.Logger.Error("[query-transactions] Failed to get total count", zap.Error(err))
		return nil, err
	}

	if includeUnconfirmed {
		memPoolResult, err := g.getUnconfirmedTransactions(params)
		if err == nil {
			unconfirmedTxs = memPoolResult
		} else {
			logger.Logger.Error("[query-transactions] Failed to get unconfirmed txs", zap.Error(err))
		}
	}

	totalTxCount := totalConfirmedCount + len(unconfirmedTxs)

	if totalTxCount == 0 || request.Offset >= totalTxCount {
		return types.TxsResultResponse{
			Transactions: []types.TransactionResultResponse{},
			TotalCount:   totalTxCount,
		}, nil
	}

	var txResponses []types.TransactionResultResponse
	unconfirmedCount := len(unconfirmedTxs)

	if request.Offset < unconfirmedCount && request.Offset+request.Limit <= unconfirmedCount {
		for i := request.Offset; i < request.Offset+request.Limit && i < unconfirmedCount; i++ {
			tx := unconfirmedTxs[i]
			txResponse := g.createTxResponse(tx, "unconfirmed", request.Address)
			txResponses = append(txResponses, txResponse)
		}
	} else {
		if request.Offset < unconfirmedCount {
			for i := request.Offset; i < unconfirmedCount; i++ {
				tx := unconfirmedTxs[i]
				txResponse := g.createTxResponse(tx, "unconfirmed", request.Address)
				txResponses = append(txResponses, txResponse)
			}

			confirmedOffset = 0
			confirmedLimit = request.Limit - len(txResponses)
		} else {
			confirmedOffset = request.Offset - unconfirmedCount
			confirmedLimit = request.Limit
		}

		if confirmedLimit > 0 && (includeConfirmed || includeFailed) {
			confirmedTxs, err := g.getConfirmedTransactions(params, confirmedOffset, confirmedLimit,
				includeConfirmed, includeFailed)
			if err != nil {
				logger.Logger.Error("[query-transactions] Failed to get confirmed txs", zap.Error(err))
				return nil, err
			}

			for _, tx := range confirmedTxs {
				status := "success"
				if tx.Code > 0 {
					status = "failed"
				}

				txResponse := g.createTxResponse(tx, status, request.Address)
				txResponses = append(txResponses, txResponse)
			}
		}
	}

	return result, nil
}
