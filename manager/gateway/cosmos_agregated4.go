package gateway

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	sekaitypes "github.com/KiraCore/sekai/types"
	"go.uber.org/zap"

	"github.com/saiset-co/sai-interx-manager/logger"
	"github.com/saiset-co/sai-interx-manager/types"
	"github.com/saiset-co/sai-interx-manager/utils"
	"github.com/saiset-co/sai-storage-mongo/external/adapter"
)

func (g *CosmosGateway) createTxResponse(tx types.TxResponse, status string, address string) types.TransactionResultResponse {
	txResponse := types.TransactionResultResponse{
		Time:   tx.Timestamp.Unix(),
		Hash:   tx.TxHash,
		Memo:   tx.Tx.Body.Memo,
		Fee:    tx.Tx.AuthInfo.Fee.Amount,
		Status: status,
	}

	if address != "" {
		txResponse.Direction = utils.DetermineDirection(tx, address)
	}

	txResponse.Txs = utils.ParseTxMessages(tx)

	return txResponse
}

func (g *CosmosGateway) getConfirmedTransactionsCount(params url.Values, includeConfirmed, includeFailed bool) (int, error) {
	var response struct {
		Pagination struct {
			Total string `json:"total"`
		} `json:"pagination"`
	}

	if !includeConfirmed && !includeFailed {
		return 0, nil
	}

	params.Set("pagination.offset", "0")
	params.Set("pagination.limit", "1")
	params.Set("pagination.count_total", "true")

	gatewayReq, err := http.NewRequestWithContext(g.context.Context, "GET", "/cosmos/tx/v1beta1/txs", nil)
	if err != nil {
		logger.Logger.Error("[get-confirmed-transactions-count] Create request failed", zap.Error(err))
		return 0, err
	}

	gatewayReq.URL.RawQuery = params.Encode()
	success, err := g.grpcProxy.ServeGRPC(gatewayReq)
	if err != nil {
		logger.Logger.Error("[get-confirmed-transactions-count] Serve request failed", zap.Error(err))
		return 0, err
	}

	if err := json.Unmarshal(success, &response); err != nil {
		logger.Logger.Error("[get-confirmed-transactions-count] Invalid response format", zap.Error(err))
		return 0, err
	}

	totalCount := 0
	if response.Pagination.Total != "" {
		totalCount, _ = strconv.Atoi(response.Pagination.Total)
	}

	return totalCount, nil
}

func (g *CosmosGateway) getUnconfirmedTransactions(queryParams url.Values) ([]types.TxResponse, error) {
	gatewayReq, err := http.NewRequestWithContext(g.context.Context, "GET", "/cosmos/tx/v1beta1/mempool", nil)
	if err != nil {
		logger.Logger.Error("[get-unconfirmed-transactions] Create request failed", zap.Error(err))
		return nil, err
	}
	//Todo: add params encode?
	success, err := g.grpcProxy.ServeGRPC(gatewayReq)
	if err != nil {
		logger.Logger.Error("[get-unconfirmed-transactions] Serve request failed", zap.Error(err))
		return nil, err
	}

	var mempoolResponse types.MempoolResponse

	if err := json.Unmarshal(success, &mempoolResponse); err != nil {
		logger.Logger.Error("[get-unconfirmed-transactions] Invalid response format", zap.Error(err))
		return nil, err
	}

	//Todo: move to request?
	var filteredTxs []types.TxResponse

	address := queryParams.Get("address")

	for _, tx := range mempoolResponse.Txs {
		if address != "" && !utils.TxСontainsAddress(tx, address) {
			continue
		}

		if txTypes := queryParams.Get("types"); txTypes != "" {
			typesList := strings.Split(txTypes, ",")
			if !utils.MatchesTxType(tx, typesList) {
				continue
			}
		}

		filteredTxs = append(filteredTxs, tx)
	}

	return filteredTxs, nil
}

func (g *CosmosGateway) getConfirmedTransactions(params url.Values, offset, limit int, includeConfirmed, includeFailed bool) ([]types.TxResponse, error) {
	var filteredTxs []types.TxResponse
	var response struct {
		Txs []types.TxResponse `json:"txs"`
	}

	if !includeConfirmed && !includeFailed {
		return []types.TxResponse{}, nil
	}

	gatewayReq, err := http.NewRequestWithContext(g.context.Context, "GET", "/cosmos/tx/v1beta1/txs", nil)
	if err != nil {
		return nil, err
	}

	params.Set("pagination.offset", strconv.Itoa(offset))
	params.Set("pagination.limit", strconv.Itoa(limit))
	gatewayReq.URL.RawQuery = params.Encode()
	success, err := g.grpcProxy.ServeGRPC(gatewayReq)
	if err != nil {
		logger.Logger.Error("[get-confirmed-transactions] Serve request failed", zap.Error(err))
		return nil, err
	}

	if err := json.Unmarshal(success, &response); err != nil {
		logger.Logger.Error("[get-confirmed-transactions] Invalid response format", zap.Error(err))
		return nil, err
	}

	//Todo: move to request?
	for _, tx := range response.Txs {
		isSuccess := tx.Code == 0

		if (isSuccess && includeConfirmed) || (!isSuccess && includeFailed) {
			filteredTxs = append(filteredTxs, tx)
		}
	}

	return filteredTxs, nil
}

func (g *CosmosGateway) tokenRates() (interface{}, error) {
	tokenAliasGRPCResponse := types.TokenAliasesGRPCResponse{}

	type TokenRatesResponse struct {
		Data []types.TokenAlias `json:"data"`
	}
	result := TokenRatesResponse{}

	gatewayReq, err := http.NewRequestWithContext(g.context.Context, "GET", "/kira/tokens/infos", nil)
	if err != nil {
		logger.Logger.Error("[query-token-rates] Create request failed", zap.Error(err))
		return nil, err
	}

	respBody, err := g.grpcProxy.ServeGRPC(gatewayReq)
	if err != nil {
		logger.Logger.Error("[query-token-rates] Serve request failed", zap.Error(err))
		return nil, err
	}

	err = json.Unmarshal(respBody, &tokenAliasGRPCResponse)
	if err != nil {
		logger.Logger.Error("[query-token-rates] Invalid response format", zap.Error(err))
		return nil, err
	}

	for index, tokenRate := range tokenAliasGRPCResponse.Data {
		tokenAliasGRPCResponse.Data[index].Data.FeeRate = utils.ConvertRate(tokenRate.Data.FeeRate)
		tokenAliasGRPCResponse.Data[index].Data.StakeCap = utils.ConvertRate(tokenRate.Data.StakeCap)
		tokenAliasGRPCResponse.Data[index].Data.StakeMin = utils.ConvertRate(tokenRate.Data.StakeMin)

		result.Data = append(result.Data, tokenAliasGRPCResponse.Data[index].Data)
	}

	return result, nil
}

func (g *CosmosGateway) customPrefixes() (*types.CustomPrefixesResponse, error) {
	var customPrefixesResponse = new(types.CustomPrefixesResponse)

	gatewayReq, err := http.NewRequestWithContext(g.context.Context, "GET", "/kira/gov/custom_prefixes", nil)
	if err != nil {
		logger.Logger.Error("[query-custom-prefixes] Create request failed", zap.Error(err))
		return nil, err
	}

	respBody, err := g.grpcProxy.ServeGRPC(gatewayReq)
	if err != nil {
		logger.Logger.Error("[query-custom-prefixes] Serve request failed", zap.Error(err))
		return nil, err
	}

	err = json.Unmarshal(respBody, &customPrefixesResponse)
	if err != nil {
		logger.Logger.Error("[query-custom-prefixes] Invalid response format", zap.Error(err))
		return nil, err
	}

	return customPrefixesResponse, nil
}

func (g *CosmosGateway) tokenAliases(req types.InboundRequest) (interface{}, error) {
	tokenAliasGRPCResponse := types.TokenAliasesGRPCResponse{}
	tokenAliasResponse := types.TokenAliasesResponse{}

	type TokenAliasRequest struct {
		Tokens     []string `json:"tokens,omitempty"`
		Limit      int      `json:"limit,string,omitempty"`
		Offset     int      `json:"offset,string,omitempty"`
		CountTotal int      `json:"count_total,string,omitempty"`
	}

	request := TokenAliasRequest{
		Limit:      sekaitypes.PageIterationLimit - 1,
		Offset:     0,
		CountTotal: 0,
	}

	jsonData, err := json.Marshal(req.Payload)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonData, &request)
	if err != nil {
		return nil, err
	}

	gatewayReq, err := http.NewRequestWithContext(g.context.Context, "GET", "/kira/tokens/infos", nil)
	if err != nil {
		logger.Logger.Error("[query-token-aliases] Create request failed", zap.Error(err))
		return nil, err
	}

	respBody, err := g.grpcProxy.ServeGRPC(gatewayReq)
	if err != nil {
		logger.Logger.Error("[query-token-aliases] Serve request failed", zap.Error(err))
		return nil, err
	}

	err = json.Unmarshal(respBody, &tokenAliasGRPCResponse)
	if err != nil {
		logger.Logger.Error("[query-token-aliases] Invalid response format", zap.Error(err))
		return nil, err
	}

	prefixes, err := g.customPrefixes()
	if err != nil {
		logger.Logger.Error("[query-token-aliases] Failed to get custom prefixes", zap.Error(err))
		return nil, err
	}

	for _, alias := range tokenAliasGRPCResponse.Data {
		tokenAliasResponse.Data = append(tokenAliasResponse.Data, alias.Data)
	}

	if request.Limit > 0 {
		total := len(tokenAliasResponse.Data)
		count := int(math.Min(float64(request.Limit), float64(total)))

		if request.CountTotal > 0 {
			tokenAliasResponse.Pagination.Total = total
		}

		from := int(math.Min(float64(request.Offset), float64(total)))
		to := int(math.Min(float64(request.Offset+count), float64(total)))

		tokenAliasResponse.Data = tokenAliasResponse.Data[from:to]
	}

	tokenAliasResponse.Bech32Prefix = prefixes.Bech32Prefix
	tokenAliasResponse.DefaultDenom = prefixes.DefaultDenom

	return tokenAliasResponse, nil
}

func (g *CosmosGateway) proposalsCount() (int, error) {
	var totalCount = 0
	var response struct {
		Pagination struct {
			Total string `json:"total"`
		} `json:"pagination"`
	}

	gatewayReq, err := http.NewRequestWithContext(g.context.Context, "GET", "/kira/gov/proposals", nil)
	if err != nil {
		logger.Logger.Error("[query-proposals-count] Create request failed", zap.Error(err))
		return totalCount, err
	}

	q := gatewayReq.URL.Query()
	q.Add("pagination.offset", "0")
	q.Add("pagination.limit", "1")
	q.Add("pagination.count_total", "true")
	gatewayReq.URL.RawQuery = q.Encode()

	success, err := g.grpcProxy.ServeGRPC(gatewayReq)
	if err != nil {
		logger.Logger.Error("[[query-proposals-count] Serve request failed", zap.Error(err))
		return totalCount, err
	}

	if err := json.Unmarshal(success, &response); err != nil {
		logger.Logger.Error("[query-proposals-count] Invalid response format", zap.Error(err))
		return totalCount, err
	}

	if response.Pagination.Total != "" {
		totalCount, _ = strconv.Atoi(response.Pagination.Total)
	}

	return totalCount, nil
}

func (g *CosmosGateway) getProposals(req types.InboundRequest) (interface{}, error) {
	proposals := new(types.ProposalsResponse)
	limit := sekaitypes.PageIterationLimit - 1
	offset := 0

	for {
		gatewayReq, err := http.NewRequestWithContext(g.context.Context, "GET", "/kira/gov/proposals", nil)
		if err != nil {
			logger.Logger.Error("[query-proposals] Create request failed", zap.Error(err))
			return nil, err
		}

		q := gatewayReq.URL.Query()
		q.Add("pagination.offset", strconv.Itoa(offset))
		q.Add("pagination.limit", strconv.Itoa(limit))
		gatewayReq.URL.RawQuery = q.Encode()

		respBody, err := g.grpcProxy.ServeGRPC(gatewayReq)
		if err != nil {
			logger.Logger.Error("[query-proposals] Serve request failed", zap.Error(err))
			return nil, err
		}

		subResult := new(types.ProposalsResponse)
		err = json.Unmarshal(respBody, subResult)
		if err != nil {
			logger.Logger.Error("[query-proposals] Invalid response format", zap.Error(err))
			return nil, err
		}

		if len(subResult.Proposals) == 0 {
			break
		}

		if afterProposalID, afterOk := req.Payload["afterProposalId"].(string); afterOk {
			for _, proposal := range proposals.Proposals {
				if proposal.ProposalID > afterProposalID {
					proposals.Proposals = append(proposals.Proposals, proposal)
				}
			}
		} else {
			proposals.Proposals = append(proposals.Proposals, subResult.Proposals...)
		}

		offset += limit
	}

	return proposals, nil
}

func (g *CosmosGateway) proposals(req types.InboundRequest) (interface{}, error) {
	var lastId = "0"

	var proposalsResponse = types.ProposalsResponse{
		Pagination: types.Pagination{
			Total: 0,
		},
	}

	var criteria = map[string]interface{}{
		"internal_id": map[string]interface{}{
			"$ne": nil,
		},
	}

	var proposals []types.Proposal

	var sortBy = map[string]interface{}{
		"proposalId": -1,
	}

	cachedTotal, err := g.storage.Read("proposals_cache", criteria, &adapter.Options{Limit: 1, Count: 1, Sort: sortBy}, []string{})
	if err != nil {
		logger.Logger.Error("[query-proposals] Failed to get cached proposals count", zap.Error(err))
		return proposalsResponse, err
	}

	count, err := g.proposalsCount()
	if err != nil {
		logger.Logger.Error("[query-proposals] Failed to count proposals", zap.Error(err))
		return proposalsResponse, err
	}

	if count < 0 {
		return proposalsResponse, nil
	}

	if len(cachedTotal.Result) > 0 {
		if lastIdI, ok := cachedTotal.Result[0]["proposalId"]; ok {
			if lastIdR, ok := lastIdI.(string); ok {
				lastId = lastIdR
			}
		}
	}

	if count > cachedTotal.Count {
		req.Payload["afterProposalId"] = lastId
		newProposals, err := g.getProposals(req)
		if err != nil {
			logger.Logger.Error("[query-proposals] Failed to get new proposals", zap.Error(err))
			return proposalsResponse, err
		}

		_, err = g.storage.Create("proposals_cache", newProposals)
		if err != nil {
			logger.Logger.Error("[query-proposals] Failed to save proposals cache", zap.Error(err))
			return proposalsResponse, err
		}
	}

	request := types.ProposalsRequest{
		Limit:  sekaitypes.PageIterationLimit - 1,
		Offset: 0,
	}

	jsonData, err := json.Marshal(req.Payload)
	if err != nil {
		return proposalsResponse, err
	}

	err = json.Unmarshal(jsonData, &request)
	if err != nil {
		return proposalsResponse, err
	}

	options := &adapter.Options{
		Limit: request.Limit,
		Skip:  request.Offset,
		Count: request.CountTotal,
	}

	if request.SortBy == "dateASC" {
		options.Sort = map[string]interface{}{"timestamp": 1}
	} else {
		options.Sort = map[string]interface{}{"timestamp": -1}
	}

	if request.Proposer != "" {
		criteria["proposer"] = request.Proposer
	}

	if request.DateStart > 0 || request.DateEnd > 0 {
		timeQuery := make(map[string]interface{})
		if request.DateStart > 0 {
			timeQuery["$gte"] = request.DateStart
		}
		if request.DateEnd > 0 {
			timeQuery["$lte"] = request.DateEnd
		}
		criteria["timestamp"] = timeQuery
	}

	if len(request.Types) > 0 {
		criteria["type"] = map[string]interface{}{"$in": request.Types}
	}

	if len(request.Statuses) > 0 {
		criteria["result"] = map[string]interface{}{"$in": request.Statuses}
	}

	if request.Voter != "" {
		criteria["voter"] = request.Voter
	}

	response, err := g.storage.Read("proposals_cache", criteria, options, []string{})
	if err != nil {
		logger.Logger.Error("[query-proposals] Failed to get proposal from cache", zap.Error(err))
		return proposalsResponse, err
	}

	if len(response.Result) == 0 {
		return proposalsResponse, nil
	}

	responseJsonData, err := json.Marshal(response.Result)
	if err != nil {
		return proposalsResponse, err
	}

	err = json.Unmarshal(responseJsonData, &proposals)
	if err != nil {
		return proposalsResponse, err
	}

	proposalsResponse.Proposals = proposals
	proposalsResponse.Pagination.Total = response.Count

	return proposals, nil
}

func (g *CosmosGateway) faucet(req types.InboundRequest) (interface{}, error) {
	type FaucetRequest struct {
		Claim string `json:"claim,omitempty"`
		Token string `json:"token,omitempty"`
	}

	request := FaucetRequest{}

	jsonData, err := json.Marshal(req.Payload)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonData, &request)
	if err != nil {
		return nil, err
	}

	if request.Claim == "" && request.Token == "" {
		balances, err := g.balances(req, g.address)
		if err != nil {
			return nil, err
		}

		var info = types.FaucetAccountInfo{
			Address:  g.address,
			Balances: balances,
		}

		return info, nil
	} else if request.Claim != "" && request.Token != "" {
		return g.processFaucet(req)
	} else {
		err = errors.New("[faucet] claim and token parameters are required")
	}

	return nil, nil
}

func (g *CosmosGateway) processFaucet(req types.InboundRequest) (interface{}, error) {
	return nil, nil
}
