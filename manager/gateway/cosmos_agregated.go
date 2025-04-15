package gateway

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	sekaitypes "github.com/KiraCore/sekai/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	types2 "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cast"
	"go.uber.org/zap"

	"github.com/saiset-co/sai-interx-manager/logger"
	"github.com/saiset-co/sai-interx-manager/types"
	"github.com/saiset-co/sai-interx-manager/utils"
)

func (g *CosmosGateway) allValidators() (*types.ValidatorsResponse, error) {
	validators := new(types.ValidatorsResponse)
	limit := sekaitypes.PageIterationLimit - 1
	offset := 0

	for {
		gatewayReq, err := http.NewRequestWithContext(g.context.Context, "GET", "/kira/staking/validators", nil)
		if err != nil {
			logger.Logger.Error("CosmosGateway - validators - Create request failed", zap.Error(err))
			return nil, err
		}

		q := gatewayReq.URL.Query()
		q.Add("pagination.offset", strconv.Itoa(offset))
		q.Add("pagination.limit", strconv.Itoa(limit))
		gatewayReq.URL.RawQuery = q.Encode()

		respBody, _, err := g.grpcProxy.ServeGRPC(gatewayReq)
		if err != nil {
			logger.Logger.Error("CosmosGateway - validators - Serve request failed", zap.Error(err))
			return nil, err
		}

		byteData, err := json.Marshal(respBody)
		if err != nil {
			logger.Logger.Error("CosmosGateway - validators - Marshaling response failed", zap.Error(err))
			return nil, err
		}

		subResult := new(types.ValidatorsResponse)
		err = json.Unmarshal(byteData, subResult)
		if err != nil {
			logger.Logger.Error("CosmosGateway - validators - Unmarshal response failed", zap.Error(err))
			return nil, err
		}

		if len(subResult.Validators) == 0 {
			break
		}

		validators.Actors = subResult.Actors
		validators.Validators = append(validators.Validators, subResult.Validators...)
		offset += limit
	}

	return validators, nil
}

func (g *CosmosGateway) tokens() ([]string, error) {
	tokenRatesResponse := types.TokenRatesResponse{}
	poolTokens := make([]string, 0)

	gatewayReq, err := http.NewRequestWithContext(g.context.Context, "GET", "/kira/tokens/rates", nil)
	if err != nil {
		logger.Logger.Error("CosmosGateway - validators - Create request failed", zap.Error(err))
		return nil, err
	}

	respBody, _, err := g.grpcProxy.ServeGRPC(gatewayReq)
	if err != nil {
		logger.Logger.Error("CosmosGateway - validators - Serve request failed", zap.Error(err))
		return nil, err
	}

	if respBody != nil {
		byteData, err := json.Marshal(respBody)
		if err != nil {
			logger.Logger.Error("CosmosGateway - validators - Marshaling response failed", zap.Error(err))
			return nil, err
		}

		err = json.Unmarshal(byteData, &tokenRatesResponse)
		if err != nil {
			logger.Logger.Error("CosmosGateway - validators - Unmarshal response failed", zap.Error(err))
			return nil, err
		}

		for _, tokenRate := range tokenRatesResponse.Data {
			poolTokens = append(poolTokens, tokenRate.Denom)
		}
	}

	return poolTokens, nil
}

func (g *CosmosGateway) signingInfos() (*types.ValidatorInfoResponse, error) {
	validatorInfosResponse := new(types.ValidatorInfoResponse)
	limit := sekaitypes.PageIterationLimit - 1
	offset := 0

	for {
		gatewayReq, err := http.NewRequestWithContext(g.context.Context, "GET", "/kira/slashing/v1beta1/signing_info", nil)
		if err != nil {
			logger.Logger.Error("CosmosGateway - validators - Create request failed", zap.Error(err))
			return nil, err
		}

		q := gatewayReq.URL.Query()
		q.Add("pagination.offset", strconv.Itoa(offset))
		q.Add("pagination.limit", strconv.Itoa(limit))
		gatewayReq.URL.RawQuery = q.Encode()

		respBody, _, err := g.grpcProxy.ServeGRPC(gatewayReq)
		if err != nil {
			logger.Logger.Error("CosmosGateway - validators - Serve request failed", zap.Error(err))
			return nil, err
		}

		byteData, err := json.Marshal(respBody)
		if err != nil {
			logger.Logger.Error("CosmosGateway - validators - Marshaling response failed", zap.Error(err))
			return nil, err
		}

		subResult := new(types.ValidatorInfoResponse)
		err = json.Unmarshal(byteData, subResult)
		if err != nil {
			logger.Logger.Error("CosmosGateway - validators - Unmarshal response failed", zap.Error(err))
			return nil, err
		}

		if len(subResult.ValValidatorInfos) == 0 {
			break
		}

		validatorInfosResponse.ValValidatorInfos = append(validatorInfosResponse.ValValidatorInfos, subResult.ValValidatorInfos...)
		offset += limit
	}

	return validatorInfosResponse, nil
}

func (g *CosmosGateway) validatorsPool() (*types.AllPools, error) {
	type ValidatorPoolsResponse struct {
		Pools []types.ValidatorPool `json:"pools,omitempty"`
	}

	pools := ValidatorPoolsResponse{}

	allPools := &types.AllPools{
		ValToPool: make(map[string]types.ValidatorPool),
		IdToPool:  make(map[int64]types.ValidatorPool),
	}

	gatewayReq, err := http.NewRequestWithContext(g.context.Context, "GET", "/kira/multistaking/v1beta1/staking_pools", nil)
	if err != nil {
		logger.Logger.Error("CosmosGateway - validators - Create request failed", zap.Error(err))
		return nil, err
	}

	respBody, _, err := g.grpcProxy.ServeGRPC(gatewayReq)
	if err != nil {
		logger.Logger.Error("CosmosGateway - validators - Serve request failed", zap.Error(err))
		return nil, err
	}

	if respBody != nil {
		byteData, err := json.Marshal(respBody)
		if err != nil {
			logger.Logger.Error("CosmosGateway - validators - Marshaling response failed", zap.Error(err))
			return nil, err
		}

		err = json.Unmarshal(byteData, &pools)
		if err != nil {
			logger.Logger.Error("CosmosGateway - validators - Unmarshal response failed", zap.Error(err))
			return nil, err
		}

		for _, pool := range pools.Pools {
			allPools.ValToPool[pool.Validator] = pool
			allPools.IdToPool[pool.ID] = pool
		}
	}

	return allPools, nil
}

func (g *CosmosGateway) dashboard() (*types.AllValidators, error) {
	allValidators := &types.AllValidators{
		AddrToValidator: make(map[string]string),
		PoolToValidator: make(map[int64]types.QueryValidator),
		PoolTokens:      make([]string, 0),
	}

	validatorsData, err := g.allValidators()
	if err != nil {
		logger.Logger.Error("CosmosGateway - dashboard - validators", zap.Error(err))
		return nil, err
	}

	tokens, err := g.tokens()
	if err != nil {
		logger.Logger.Error("CosmosGateway - dashboard - validators", zap.Error(err))
		return nil, err
	}

	signingInfos, err := g.signingInfos()
	if err != nil {
		logger.Logger.Error("CosmosGateway - dashboard - validators", zap.Error(err))
		return nil, err
	}

	validatorsPool, err := g.validatorsPool()
	if err != nil {
		logger.Logger.Error("CosmosGateway - dashboard - validators", zap.Error(err))
		return nil, err
	}

	for index, validator := range validatorsData.Validators {
		pubkeyHexString := validator.Pubkey[14 : len(validator.Pubkey)-1]
		bytes, _ := hex.DecodeString(pubkeyHexString)
		pubkey := ed25519.PubKey{
			Key: bytes,
		}
		address := sdk.ConsAddress(pubkey.Address()).String()
		allValidators.AddrToValidator[validator.Address] = validator.Valkey

		var valSigningInfo types.ValidatorSigningInfo

		for _, signingInfo := range signingInfos.ValValidatorInfos {
			if signingInfo.Address == address {
				valSigningInfo = signingInfo
				break
			}
		}

		for _, record := range validatorsData.Validators[index].Identity {
			if record.Key == "logo" || record.Key == "avatar" {
				validatorsData.Validators[index].Logo = record.Value
			} else if record.Key == "description" {
				validatorsData.Validators[index].Description = record.Value
			} else if record.Key == "website" {
				validatorsData.Validators[index].Website = record.Value
			} else if record.Key == "social" {
				validatorsData.Validators[index].Social = record.Value
			} else if record.Key == "contact" {
				validatorsData.Validators[index].Contact = record.Value
			} else if record.Key == "validator_node_id" {
				validatorsData.Validators[index].Validator_node_id = record.Value
			} else if record.Key == "sentry_node_id" {
				validatorsData.Validators[index].Sentry_node_id = record.Value
			}
		}

		validatorsData.Validators[index].Identity = nil
		validatorsData.Validators[index].StartHeight = valSigningInfo.StartHeight
		validatorsData.Validators[index].InactiveUntil = valSigningInfo.InactiveUntil
		validatorsData.Validators[index].Mischance = valSigningInfo.Mischance
		validatorsData.Validators[index].MischanceConfidence = valSigningInfo.MischanceConfidence
		validatorsData.Validators[index].LastPresentBlock = valSigningInfo.LastPresentBlock
		validatorsData.Validators[index].MissedBlocksCounter = valSigningInfo.MissedBlocksCounter
		validatorsData.Validators[index].ProducedBlocksCounter = valSigningInfo.ProducedBlocksCounter

		pool, found := validatorsPool.ValToPool[validator.Valkey]
		if found {
			validatorsData.Validators[index].StakingPoolId = pool.ID
			if pool.Enabled {
				validatorsData.Validators[index].StakingPoolStatus = "ENABLED"
			} else {
				validatorsData.Validators[index].StakingPoolStatus = "DISABLED"
			}

			allValidators.PoolToValidator[validatorsData.Validators[index].StakingPoolId] = validatorsData.Validators[index]
		}
	}

	sort.Sort(types.QueryValidators(validatorsData.Validators))
	for index := range validatorsData.Validators {
		validatorsData.Validators[index].Top = index + 1
	}

	allValidators.PoolTokens = tokens
	allValidators.Validators = validatorsData.Validators
	allValidators.Waiting = make([]string, 0)

	for _, actor := range validatorsData.Actors {
		isWaiting := true
		for _, validator := range validatorsData.Validators {
			if validator.Address == actor {
				isWaiting = false
				break
			}
		}

		if isWaiting {
			allValidators.Waiting = append(allValidators.Waiting, actor)
		}
	}

	allValidators.Status.TotalValidators = len(validatorsData.Validators)
	allValidators.Status.WaitingValidators = len(allValidators.Waiting)
	allValidators.Status.ActiveValidators = 0
	allValidators.Status.PausedValidators = 0
	allValidators.Status.InactiveValidators = 0
	allValidators.Status.JailedValidators = 0

	for _, validator := range validatorsData.Validators {
		if validator.Status == Active {
			allValidators.Status.ActiveValidators++
		}
		if validator.Status == Inactive {
			allValidators.Status.InactiveValidators++
		}
		if validator.Status == Paused {
			allValidators.Status.PausedValidators++
		}
		if validator.Status == Jailed {
			allValidators.Status.JailedValidators++
		}
	}

	return allValidators, nil
}

func (g *CosmosGateway) txs(req types.InboundRequest) (interface{}, error) {
	type PostTxReq struct {
		Tx   string `json:"tx"`
		Mode string `json:"mode"`
	}

	var request = new(PostTxReq)

	reqParams, err := json.Marshal(req.Payload)
	if err != nil {
		logger.Logger.Error("txs", zap.Error(err))
		return nil, err
	}

	err = json.Unmarshal(reqParams, &request)
	if err != nil {
		logger.Logger.Error("txs", zap.Error(err))
		return nil, err
	}

	if request.Mode != "" {
		if allowed, ok := g.txModes[request.Mode]; !ok || !allowed {
			err = errors.New("[post-transaction] Invalid transaction mode")
			return nil, err
		}
	}

	_url := "/broadcast_tx_sync"

	switch request.Mode {
	case "block":
		_url = "/broadcast_tx_commit"
	case "async":
		_url = "/broadcast_tx_async"
	}

	txBytes, err := base64.StdEncoding.DecodeString(request.Tx)
	if err != nil {
		logger.Logger.Error("txs", zap.Error(err))
		return nil, err
	}

	return g.makeTendermintRPCRequest(g.context.Context, _url, fmt.Sprintf("tx=0x%X", txBytes))
}

func (g *CosmosGateway) validators(req types.InboundRequest) (*types.ValidatorsResponse, error) {
	validatorsResponse, err := g.allValidators()
	if err != nil {
		logger.Logger.Error("CosmosGateway - validators - allValidators failed", zap.Error(err))
		return nil, err
	}

	return g.filterAndPaginateValidators(validatorsResponse, req.Payload)
}

func (g *CosmosGateway) statusAPI() (interface{}, error) {
	result := types.InterxStatus{
		ID: cast.ToString(g.context.GetConfig("p2p.id", "")),
	}

	genesis, err := g.genesis()
	if err != nil {
		logger.Logger.Error("[query-status] Failed to query genesis", zap.Error(err))
		return nil, err
	}

	result.InterxInfo.ChainID = genesis.GenesisDoc.ChainID
	result.InterxInfo.GenesisChecksum = fmt.Sprintf("%x", sha256.Sum256(genesis.GenesisData))

	sentryStatus, err := g.status()
	if err != nil {
		logger.Logger.Error("[query-status] Failed to query status", zap.Error(err))
		return nil, err
	}

	result.NodeInfo = sentryStatus.NodeInfo
	result.SyncInfo = sentryStatus.SyncInfo
	result.ValidatorInfo = sentryStatus.ValidatorInfo
	result.InterxInfo.Moniker = sentryStatus.NodeInfo.Moniker
	result.InterxInfo.LatestBlockHeight = sentryStatus.SyncInfo.LatestBlockHeight
	result.InterxInfo.CatchingUp = sentryStatus.SyncInfo.CatchingUp

	//result.InterxInfo.Node = config.Config.Node
	//result.InterxInfo.KiraAddr = config.Config.Address
	//result.InterxInfo.KiraPubKey = config.Config.PubKey.String()
	//result.InterxInfo.FaucetAddr = config.Config.Faucet.Address
	//result.InterxInfo.InterxVersion = config.Config.InterxVersion
	//result.InterxInfo.SekaiVersion = config.Config.SekaiVersion

	return result, nil
}

func (g *CosmosGateway) status() (*types.KiraStatus, error) {
	success, err := g.makeTendermintRPCRequest(g.context.Context, "/status", "")
	if err != nil {
		logger.Logger.Error("[kira-status] Invalid response format", zap.Error(err))
		return nil, err
	}

	result := new(types.KiraStatus)

	byteData, err := json.Marshal(success)
	if err != nil {
		logger.Logger.Error("[kira-status] Invalid response format", zap.Error(err))
		return nil, err
	}

	err = json.Unmarshal(byteData, result)
	if err != nil {
		logger.Logger.Error("[kira-status] Invalid response format", zap.Error(err))
		return nil, err
	}

	return result, nil
}

func (g *CosmosGateway) genesisChunked(chunk int) (*types.GenesisChunkedResponse, error) {
	data, _ := g.makeTendermintRPCRequest(g.context.Context, "/genesis_chunked", fmt.Sprintf("chunk=%d", chunk))

	genesis := new(types.GenesisChunkedResponse)
	byteData, err := json.Marshal(data)
	if err != nil {
		logger.Logger.Error("CosmosGateway - genesisChunked", zap.Error(err))
		return nil, err
	}

	err = json.Unmarshal(byteData, genesis)
	if err != nil {
		logger.Logger.Error("CosmosGateway - genesisChunked", zap.Error(err))
		return nil, err
	}

	return genesis, nil
}

func (g *CosmosGateway) genesis() (*types.GenesisInfo, error) {
	gInfo := new(types.GenesisInfo)
	gInfo.GenesisDoc = new(types2.GenesisDoc)

	genesisData, err := g.genesisChunked(0)
	if err != nil {
		logger.Logger.Error("CosmosGateway - genesis", zap.Error(err))
		return nil, err
	}

	total, err := strconv.Atoi(genesisData.Total)
	if err != nil {
		logger.Logger.Error("CosmosGateway - genesis", zap.Error(err))
		return nil, err
	}

	if total > 1 {
		for i := 1; i < total; i++ {
			nextData, err := g.genesisChunked(i)
			if err != nil {
				logger.Logger.Error("CosmosGateway - genesis", zap.Error(err))
				return nil, err
			}

			genesisData.Data = append(genesisData.Data, nextData.Data...)
		}
	}

	gInfo.GenesisData = genesisData.Data

	err = tmjson.Unmarshal(genesisData.Data, gInfo.GenesisDoc)
	if err != nil {
		logger.Logger.Error("CosmosGateway - genesis", zap.Error(err))
		return nil, err
	}

	err = gInfo.GenesisDoc.ValidateAndComplete()
	if err != nil {
		logger.Logger.Error("CosmosGateway - genesis", zap.Error(err))
		return nil, err
	}

	return gInfo, nil
}

func (g *CosmosGateway) blocks(req types.InboundRequest) (interface{}, error) {
	type BlocksRequest struct {
		Limit  int    `json:"limit,omitempty"`
		Offset int    `json:"offset,omitempty"`
		HasTxs int    `json:"has_txs,omitempty"`
		Order  string `json:"order_by,omitempty"`
	}

	request := BlocksRequest{
		Limit:  sekaitypes.PageIterationLimit - 1,
		Offset: 0,
		HasTxs: 0,
		Order:  "asc",
	}

	jsonData, err := json.Marshal(req.Payload)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonData, &request)
	if err != nil {
		return nil, err
	}

	var page = 0
	var query = url.Values{}

	if request.HasTxs == 1 {
		query.Add("query", "\"block.height > 0 AND block.num_tx > 0\"")
	} else {
		query.Add("query", "\"block.height > 0\"")
	}

	if request.Limit > 0 {
		page = request.Offset/request.Limit + 1
		query.Add("per_page", fmt.Sprintf("%d", request.Limit))
	}

	if page > 0 {
		query.Add("page", fmt.Sprintf("%d", page))
	}

	query.Add("order_by", fmt.Sprintf("\"%s\"", request.Order))

	return g.makeTendermintRPCRequest(g.context.Context, "/block_search", query.Encode())
}

func (g *CosmosGateway) balances(req types.InboundRequest, accountID string) (interface{}, error) {
	type BalancesRequest struct {
		Limit      int `json:"limit,omitempty"`
		Offset     int `json:"offset,omitempty"`
		CountTotal int `json:"count_total,omitempty"`
	}

	request := BalancesRequest{
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

	gatewayReq, err := http.NewRequestWithContext(g.context.Context, "GET", "/cosmos/bank/v1beta1/balances/"+accountID, nil)
	if err != nil {
		logger.Logger.Error("[query-balances] Create request failed", zap.Error(err))
		return nil, err
	}

	q := gatewayReq.URL.Query()
	q.Add("pagination.offset", strconv.Itoa(request.Offset))
	q.Add("pagination.limit", strconv.Itoa(request.Limit))
	if request.CountTotal > 0 {
		q.Add("pagination.count_total", "true")
	}
	gatewayReq.URL.RawQuery = q.Encode()

	respBody, statusCode, err := g.grpcProxy.ServeGRPC(gatewayReq)

	if statusCode >= 400 {
		errMsg := fmt.Sprintf("gRPC gateway error: status=%d, body=%s", statusCode, respBody)
		logger.Logger.Error("CosmosGateway - Handle - gRPC gateway error response",
			zap.Int("status", statusCode),
			zap.Any("body", respBody))

		return nil, fmt.Errorf(errMsg)
	}

	return respBody, nil
}

func (g *CosmosGateway) delegations(req types.InboundRequest) (interface{}, error) {
	var response = new(types.QueryDelegationsResult)

	type DelegationsRequest struct {
		Account    string `json:"delegatorAddress,omitempty"`
		Limit      int    `json:"limit,omitempty"`
		Offset     int    `json:"offset,omitempty"`
		CountTotal int    `json:"count_total,omitempty"`
	}

	request := DelegationsRequest{
		Account:    "",
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

	if request.Account != "" {
		req.Path = "/cosmos/bank/v1beta1/balances/" + request.Account
	}

	gatewayReq, err := http.NewRequestWithContext(g.context.Context, "GET", req.Path, nil)
	if err != nil {
		logger.Logger.Error("[query-staking-pool] Create request failed", zap.Error(err))
		return nil, err
	}

	success, statusCode, err := g.grpcProxy.ServeGRPC(gatewayReq)

	if statusCode >= 400 {
		errMsg := fmt.Sprintf("gRPC gateway error: status=%d, body=%s", statusCode, success)
		logger.Logger.Error("CosmosGateway - Handle - gRPC gateway error response",
			zap.Int("status", statusCode),
			zap.Any("body", success))

		return nil, fmt.Errorf(errMsg)
	}

	if success != nil {
		allPools, err := g.validatorsPool()
		if err != nil {
			logger.Logger.Error("[query-staking-pool] Error getting validators pool", zap.Error(err))
			return nil, err
		}

		tokens, err := g.tokens()
		if err != nil {
			logger.Logger.Error("[query-staking-pool] Error getting tokens", zap.Error(err))
			return nil, err
		}

		validators, err := g.dashboard()
		if err != nil {
			logger.Logger.Error("[query-staking-pool] Error getting validators", zap.Error(err))
			return nil, err
		}

		result := types.QueryBalancesResponse{}

		// parse user balance data and generate delegation responses from pool tokens
		byteData, err := json.Marshal(success)
		if err != nil {
			logger.Logger.Error("[query-staking-pool] Invalid response format", zap.Error(err))
			return nil, err
		}
		err = json.Unmarshal(byteData, &result)
		if err != nil {
			logger.Logger.Error("[query-staking-pool] Invalid response format", zap.Error(err))
			return nil, err
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
				pool, found := allPools.IdToPool[int64(poolID)]
				if !found {
					continue
				}
				// fill up PoolInfo
				delegation.PoolInfo.ID = pool.ID
				delegation.PoolInfo.Commission = utils.ConvertRate(pool.Commission)
				if pool.Enabled {
					delegation.PoolInfo.Status = "ENABLED"
				} else {
					delegation.PoolInfo.Status = "DISABLED"
				}
				delegation.PoolInfo.Tokens = tokens

				// fill up ValidatorInfo
				validator, found := validators.PoolToValidator[pool.ID]
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

		if request.Limit > 0 {
			// apply pagination
			total := len(response.Delegations)
			count := int(math.Min(float64(request.Limit), float64(total)))

			if request.CountTotal > 0 {
				response.Pagination.Total = total
			}

			from := int(math.Min(float64(request.Offset), float64(total)))
			to := int(math.Min(float64(request.Offset+count), float64(total)))

			response.Delegations = response.Delegations[from:to]
		}

		success = response
	}

	return success, nil
}

func (g *CosmosGateway) identityRecords(address string) (interface{}, error) {
	accAddr, _ := sdk.AccAddressFromBech32(address)

	gatewayReq, err := http.NewRequestWithContext(g.context.Context, "GET", "/kira/gov/identity_records/"+base64.URLEncoding.EncodeToString(accAddr.Bytes()), nil)
	if err != nil {
		logger.Logger.Error("CosmosGateway - identityRecords - Create request failed", zap.Error(err))
		return nil, err
	}

	response, statusCode, err := g.grpcProxy.ServeGRPC(gatewayReq)

	if statusCode >= 400 {
		errMsg := fmt.Sprintf("gRPC gateway error: status=%d, body=%s", statusCode, response)
		logger.Logger.Error("CosmosGateway - Handle - gRPC gateway error response",
			zap.Int("status", statusCode),
			zap.Any("body", response))

		return nil, fmt.Errorf(errMsg)
	}

	return response, nil
}

func (g *CosmosGateway) identityVerifyRequestsByApprover(req types.InboundRequest, approver string) (interface{}, error) {
	type IdentityVerifyRequestsByApproverRequest struct {
		Key        int `json:"key,omitempty"`
		Limit      int `json:"limit,omitempty"`
		Offset     int `json:"offset,omitempty"`
		CountTotal int `json:"count_total,omitempty"`
	}

	request := IdentityVerifyRequestsByApproverRequest{
		Key:        0,
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

	accAddr, _ := sdk.AccAddressFromBech32(approver)
	gatewayReq, err := http.NewRequestWithContext(g.context.Context, "GET", "/kira/gov/identity_verify_requests_by_approver/"+base64.URLEncoding.EncodeToString(accAddr.Bytes()), nil)
	if err != nil {
		logger.Logger.Error("[query-identity-record-verify-requests-by-approver] Create request failed", zap.Error(err))
		return nil, err
	}

	q := gatewayReq.URL.Query()
	q.Add("pagination.offset", strconv.Itoa(request.Offset))
	q.Add("pagination.limit", strconv.Itoa(request.Limit))
	if request.Key > 0 {
		q.Add("pagination.key", strconv.Itoa(request.Key))
	}
	if request.CountTotal > 0 {
		q.Add("pagination.count_total", "true")
	}
	gatewayReq.URL.RawQuery = q.Encode()

	response, statusCode, err := g.grpcProxy.ServeGRPC(gatewayReq)

	if statusCode >= 400 {
		errMsg := fmt.Sprintf("gRPC gateway error: status=%d, body=%s", statusCode, response)
		logger.Logger.Error("CosmosGateway - Handle - gRPC gateway error response",
			zap.Int("status", statusCode),
			zap.Any("body", response))

		return nil, fmt.Errorf(errMsg)
	}

	if response != nil {
		res := types.IdVerifyRequests{}
		bz, err := json.Marshal(response)
		if err != nil {
			logger.Logger.Error("[query-identity-record-verify-requests-by-approver] Invalid response format", zap.Error(err))
			return nil, err
		}

		err = json.Unmarshal(bz, &res)
		if err != nil {
			logger.Logger.Error("[query-identity-record-verify-requests-by-approver] Invalid response format", zap.Error(err))
			return nil, err
		}

		for idx, record := range res.VerifyRecords {
			coin, err := g.parseCoinString(record.Tip)
			if err != nil {
				logger.Logger.Error("[query-identity-record-verify-requests-by-approver] Coin can not be parsed", zap.Error(err))
				return nil, err
			}

			res.VerifyRecords[idx].Tip = coin.String()
		}

		response = res
	}

	return response, nil
}

func (g *CosmosGateway) identityVerifyRequestsByRequester(req types.InboundRequest, requester string) (interface{}, error) {
	type IdentityVerifyRequestsByRequesterRequest struct {
		Key        int `json:"key,omitempty"`
		Limit      int `json:"limit,omitempty"`
		Offset     int `json:"offset,omitempty"`
		CountTotal int `json:"count_total,omitempty"`
	}

	request := IdentityVerifyRequestsByRequesterRequest{
		Key:        0,
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

	accAddr, _ := sdk.AccAddressFromBech32(requester)
	gatewayReq, err := http.NewRequestWithContext(g.context.Context, "GET", "/kira/gov/identity_verify_requests_by_requester/"+base64.URLEncoding.EncodeToString(accAddr.Bytes()), nil)
	if err != nil {
		logger.Logger.Error("[query-identity-record-verify-requests-by-requester] Create request failed", zap.Error(err))
		return nil, err
	}

	q := gatewayReq.URL.Query()
	q.Add("pagination.offset", strconv.Itoa(request.Offset))
	q.Add("pagination.limit", strconv.Itoa(request.Limit))
	if request.Key > 0 {
		q.Add("pagination.key", strconv.Itoa(request.Key))
	}
	if request.CountTotal > 0 {
		q.Add("pagination.count_total", "true")
	}
	gatewayReq.URL.RawQuery = q.Encode()

	response, statusCode, err := g.grpcProxy.ServeGRPC(gatewayReq)

	if statusCode >= 400 {
		errMsg := fmt.Sprintf("gRPC gateway error: status=%d, body=%s", statusCode, response)
		logger.Logger.Error("CosmosGateway - Handle - gRPC gateway error response",
			zap.Int("status", statusCode),
			zap.Any("body", response))

		return nil, fmt.Errorf(errMsg)
	}

	if response != nil {
		res := types.IdVerifyRequests{}
		bz, err := json.Marshal(response)
		if err != nil {
			logger.Logger.Error("[query-identity-record-verify-requests-by-requester] Invalid response format", zap.Error(err))
			return nil, err
		}

		err = json.Unmarshal(bz, &res)
		if err != nil {
			logger.Logger.Error("[query-identity-record-verify-requests-by-requester] Invalid response format", zap.Error(err))
			return nil, err
		}

		for idx, record := range res.VerifyRecords {
			coin, err := g.parseCoinString(record.Tip)
			if err != nil {
				logger.Logger.Error("[query-identity-record-verify-requests-by-approver] Coin can not be parsed", zap.Error(err))
				continue
			}

			res.VerifyRecords[idx].Tip = coin.String()
		}

		response = res
	}

	return response, nil
}

func (g *CosmosGateway) parseCoinString(input string) (*sdk.Coin, error) {
	denom := ""
	amount := 0

	tokens, err := g.tokens()
	if err != nil {
		logger.Logger.Error("CosmosGateway - parseCoinString - Failtd to get tokens", zap.Error(err))
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
		logger.Logger.Error("[execution-fee] - Create request failed", zap.Error(err))
		return nil, err
	}

	response, statusCode, err := g.grpcProxy.ServeGRPC(gatewayReq)

	if statusCode >= 400 {
		errMsg := fmt.Sprintf("gRPC gateway error: status=%d, body=%s", statusCode, response)
		logger.Logger.Error("CosmosGateway - Handle - gRPC gateway error response",
			zap.Int("status", statusCode),
			zap.Any("body", response))

		return nil, fmt.Errorf(errMsg)
	}

	return response, nil
}

func (g *CosmosGateway) networkProperties() (interface{}, error) {
	gatewayReq, err := http.NewRequestWithContext(g.context.Context, "GET", "/kira/gov/network_properties", nil)
	if err != nil {
		logger.Logger.Error("[query-network-properties] Create request failed", zap.Error(err))
		return nil, err
	}

	response, statusCode, err := g.grpcProxy.ServeGRPC(gatewayReq)

	if statusCode >= 400 {
		errMsg := fmt.Sprintf("gRPC gateway error: status=%d, body=%s", statusCode, response)
		logger.Logger.Error("CosmosGateway - Handle - gRPC gateway error response",
			zap.Int("status", statusCode),
			zap.Any("body", response))

		return nil, fmt.Errorf(errMsg)
	}

	if response != nil {
		result, err := utils.QueryNetworkPropertiesFromGrpcResult(response)
		if err != nil {
			logger.Logger.Error("[query-network-properties] Create request failed", zap.Error(err))
			return nil, err
		}

		response = result
	}

	return response, nil
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
		logger.Logger.Error("[query-network-properties] Create request failed", zap.Error(err))
		return nil, err
	}

	response, statusCode, err := g.grpcProxy.ServeGRPC(gatewayReq)

	if statusCode >= 400 {
		errMsg := fmt.Sprintf("gRPC gateway error: status=%d, body=%v, err=%v", statusCode, response, err)
		logger.Logger.Error("CosmosGateway - Handle - gRPC gateway error response",
			zap.Int("status", statusCode),
			zap.Any("body", response),
			zap.Error(err))

		return nil, fmt.Errorf(errMsg)
	}

	if response != nil {
		responseResult := types.QueryStakingPoolDelegatorsResponse{}

		byteData, err := json.Marshal(response)
		if err != nil {
			logger.Logger.Error("[query-staking-pool] Invalid response format", zap.Error(err))
			return nil, err
		}
		err = json.Unmarshal(byteData, &responseResult)
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

		response = newResponse
	}

	return response, nil
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
		Limit      int    `json:"limit,omitempty"`
		Offset     int    `json:"offset,omitempty"`
		CountTotal int    `json:"count_total,omitempty"`
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

	success, statusCode, err := g.grpcProxy.ServeGRPC(gatewayReq)

	if statusCode >= 400 {
		errMsg := fmt.Sprintf("gRPC gateway error: status=%d, body=%v, err=%v", statusCode, response, err)
		logger.Logger.Error("CosmosGateway - Handle - gRPC gateway error response",
			zap.Int("status", statusCode),
			zap.Any("body", response),
			zap.Error(err))

		return nil, fmt.Errorf(errMsg)
	}

	if success != nil {
		validators, err := g.allValidators()
		if err != nil {
			logger.Logger.Error("[query-undelegations] Getting validators failed", zap.Error(err))
			return nil, err
		}

		result := types.QueryUndelegationsResult{}

		// parse user balance data and generate delegation responses from pool tokens
		byteData, err := json.Marshal(success)
		if err != nil {
			logger.Logger.Error("[query-undelegations] Invalid response format", zap.Error(err))
			return nil, err
		}

		err = json.Unmarshal(byteData, &result)
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

		success = response
	}

	return success, nil
}
