package gateway

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	types2 "github.com/cometbft/cometbft/types"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	sekaitypes "github.com/KiraCore/sekai/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/cast"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/saiset-co/sai-interx-manager/logger"
	cosmosAuth "github.com/saiset-co/sai-interx-manager/proto-gen/cosmos/auth/v1beta1"
	cosmosBank "github.com/saiset-co/sai-interx-manager/proto-gen/cosmos/bank/v1beta1"
	kiraGov "github.com/saiset-co/sai-interx-manager/proto-gen/kira/gov"
	kiraMultiStaking "github.com/saiset-co/sai-interx-manager/proto-gen/kira/multistaking"
	kiraSlashing "github.com/saiset-co/sai-interx-manager/proto-gen/kira/slashing/v1beta1"
	kiraSpending "github.com/saiset-co/sai-interx-manager/proto-gen/kira/spending"
	kiraStaking "github.com/saiset-co/sai-interx-manager/proto-gen/kira/staking"
	kiraTokens "github.com/saiset-co/sai-interx-manager/proto-gen/kira/tokens"
	kiraUbi "github.com/saiset-co/sai-interx-manager/proto-gen/kira/ubi"
	kiraUpgrades "github.com/saiset-co/sai-interx-manager/proto-gen/kira/upgrade"
	"github.com/saiset-co/sai-interx-manager/types"
	"github.com/saiset-co/sai-service/service"
)

type Proxy struct {
	mux  *runtime.ServeMux
	conn *grpc.ClientConn
}

type CosmosGateway struct {
	*BaseGateway
	storage   types.Storage
	url       string
	grpcProxy *Proxy
	timeout   time.Duration
}

const (
	Active   string = "ACTIVE"
	Inactive string = "INACTIVE"
	Paused   string = "PAUSED"
	Jailed   string = "JAILED"
)

var _ types.Gateway = (*CosmosGateway)(nil)

func newGRPCGatewayProxy(ctx *service.Context, address string) (*Proxy, error) {
	conn, err := grpc.DialContext(
		ctx.Context,
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %v", err)
	}

	mux := runtime.NewServeMux()

	if err := registerHandlers(ctx, mux, conn); err != nil {
		conn.Close()
		logger.Logger.Error("newGRPCGatewayProxy", zap.Error(err))
		return nil, err
	}

	return &Proxy{
		mux:  mux,
		conn: conn,
	}, nil
}

func (p *Proxy) ServeGRPC(r *http.Request) (interface{}, int, error) {
	r.Header.Set("Content-Type", "application/json")

	//Todo: add cache here

	recorder := httptest.NewRecorder()
	p.mux.ServeHTTP(recorder, r)
	resp := recorder.Result()

	defer resp.Body.Close()

	result := new(interface{})
	err := json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		logger.Logger.Error("ServeGRPC", zap.Error(err))
		return nil, resp.StatusCode, err
	}

	return result, resp.StatusCode, nil
}

func registerHandlers(ctx *service.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	if err := cosmosBank.RegisterQueryHandler(ctx.Context, mux, conn); err != nil {
		logger.Logger.Error("registerHandlers", zap.Error(err))
		return err
	}

	if err := cosmosAuth.RegisterQueryHandler(ctx.Context, mux, conn); err != nil {
		logger.Logger.Error("registerHandlers", zap.Error(err))
		return err
	}

	if err := kiraGov.RegisterQueryHandler(ctx.Context, mux, conn); err != nil {
		logger.Logger.Error("registerHandlers", zap.Error(err))
		return err
	}

	if err := kiraStaking.RegisterQueryHandler(ctx.Context, mux, conn); err != nil {
		logger.Logger.Error("registerHandlers", zap.Error(err))
		return err
	}

	if err := kiraMultiStaking.RegisterQueryHandler(ctx.Context, mux, conn); err != nil {
		logger.Logger.Error("registerHandlers", zap.Error(err))
		return err
	}

	if err := kiraSlashing.RegisterQueryHandler(ctx.Context, mux, conn); err != nil {
		logger.Logger.Error("registerHandlers", zap.Error(err))
		return err
	}

	if err := kiraTokens.RegisterQueryHandler(ctx.Context, mux, conn); err != nil {
		logger.Logger.Error("registerHandlers", zap.Error(err))
		return err
	}

	if err := kiraUpgrades.RegisterQueryHandler(ctx.Context, mux, conn); err != nil {
		logger.Logger.Error("registerHandlers", zap.Error(err))
		return err
	}

	if err := kiraSpending.RegisterQueryHandler(ctx.Context, mux, conn); err != nil {
		logger.Logger.Error("registerHandlers", zap.Error(err))
		return err
	}

	if err := kiraUbi.RegisterQueryHandler(ctx.Context, mux, conn); err != nil {
		logger.Logger.Error("registerHandlers", zap.Error(err))
		return err
	}

	return nil
}

func NewCosmosGateway(ctx *service.Context, url, node string, storage types.Storage, timeout, retryAttempts int, retryDelay time.Duration, rateLimit int) (*CosmosGateway, error) {
	proxy, err := newGRPCGatewayProxy(ctx, node)
	if err != nil {
		logger.Logger.Error("NewCosmosGateway", zap.Error(err))
		return nil, err
	}

	return &CosmosGateway{
		BaseGateway: NewBaseGateway(ctx, retryAttempts, retryDelay, rateLimit),
		storage:     storage,
		url:         url,
		grpcProxy:   proxy,
		timeout:     time.Duration(timeout) * time.Second,
	}, nil
}

func (g *CosmosGateway) Handle(data []byte) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()

	g.context.Context = ctx

	var req types.InboundRequest

	if err := json.Unmarshal(data, &req); err != nil {
		logger.Logger.Error("CosmosGateway - Handle - Unmarshal request failed", zap.Error(err))
		return nil, err
	}

	switch req.Path {
	case "/kira/status":
		{
			return g.retry.Do(func() (interface{}, error) {
				if err := g.rateLimit.Wait(g.context.Context); err != nil {
					logger.Logger.Error("EthereumGateway - Handle", zap.Error(err), zap.Any("ctx", g.context.Context))
					return nil, err
				}
				return g.status()
			})
		}
	case "/valopers":
		{
			return g.retry.Do(func() (interface{}, error) {
				if err := g.rateLimit.Wait(g.context.Context); err != nil {
					logger.Logger.Error("EthereumGateway - Handle", zap.Error(err), zap.Any("ctx", g.context.Context))
					return nil, err
				}
				return g.validators(req)
			})
		}
	case "/status":
		{
			return g.retry.Do(func() (interface{}, error) {
				if err := g.rateLimit.Wait(g.context.Context); err != nil {
					logger.Logger.Error("EthereumGateway - Handle", zap.Error(err), zap.Any("ctx", g.context.Context))
					return nil, err
				}
				return g.statusAPI()
			})
		}
	case "/blocks":
		{
			return g.retry.Do(func() (interface{}, error) {
				if err := g.rateLimit.Wait(g.context.Context); err != nil {
					logger.Logger.Error("EthereumGateway - Handle", zap.Error(err), zap.Any("ctx", g.context.Context))
					return nil, err
				}
				return g.blocks(req)
			})
		}
	case "/dashboard":
		{
			return g.retry.Do(func() (interface{}, error) {
				if err := g.rateLimit.Wait(g.context.Context); err != nil {
					logger.Logger.Error("EthereumGateway - Handle", zap.Error(err), zap.Any("ctx", g.context.Context))
					return nil, err
				}
				return g.dashboard()
			})
		}
	}

	return g.retry.Do(func() (interface{}, error) {
		if err := g.rateLimit.Wait(g.context.Context); err != nil {
			logger.Logger.Error("CosmosGateway - Handle - Rate limit exceeded", zap.Error(err))
			return nil, err
		}

		return g.proxy(req)
	})
}

func (g *CosmosGateway) Close() {
	g.grpcProxy.conn.Close()
}

func (g *CosmosGateway) makeTendermintRPCRequest(ctx context.Context, url string, query string) (interface{}, error) {
	endpoint := fmt.Sprintf("%s%s?%s", g.url, url, query)

	//Todo: add cache here

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		logger.Logger.Error("MakeTendermintRPCRequest - [rpc-call] Unable to connect to server", zap.Error(err))
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Logger.Error("MakeTendermintRPCRequest - [rpc-call] Unable to connect to server", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	response := new(types.RPCResponse)
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		logger.Logger.Debug("MakeTendermintRPCRequest - [rpc-call] Unable to decode response", zap.Any("body", resp.Body))
		logger.Logger.Error("MakeTendermintRPCRequest - [rpc-call] Unable to decode response", zap.Error(err))
		return nil, err
	}

	if response.Error.Code != 0 {
		return nil, errors.New(response.Error.Data)
	}

	return response.Result, nil
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
	var done bool
	var err error
	var query = url.Values{}
	var limit = sekaitypes.PageIterationLimit - 1
	var offset = 0
	var page = 0
	var hasTxs = 0
	var order = "asc"

	limitI, limited := req.Payload["pagination.limit"]
	if limited {
		limitS, ok := limitI.(string)
		if !ok {
			err = errors.New("limit wrong format")
			return nil, err
		}

		limit, err = strconv.Atoi(limitS)
		if err != nil {
			return nil, err
		}
	}

	offsetI, skipped := req.Payload["pagination.offset"]
	if skipped {
		offsetS, ok := offsetI.(string)
		if !ok {
			err = errors.New("offset wrong format")
			return nil, err
		}

		offset, err = strconv.Atoi(offsetS)
		if err != nil {
			return nil, err
		}
	}

	hasTxsI, filtered := req.Payload["has_txs"]
	if filtered {
		hasTxsS, ok := hasTxsI.(string)
		if !ok {
			err = errors.New("has_txs wrong format")
			return nil, err
		}

		hasTxs, err = strconv.Atoi(hasTxsS)
		if err != nil {
			return nil, err
		}
	}

	orderI, ordered := req.Payload["order_by"]
	if ordered {
		order, done = orderI.(string)
		if !done {
			err = errors.New("order_by wrong format")
			return nil, err
		}
	}

	if hasTxs == 1 {
		query.Add("query", "\"block.height > 0 AND block.num_tx > 0\"")
	} else {
		query.Add("query", "\"block.height > 0\"")
	}

	if skipped && limited {
		page = offset/limit + 1
	}

	if limit > 0 {
		query.Add("per_page", fmt.Sprintf("%d", limit))
	}

	if page > 0 {
		query.Add("page", fmt.Sprintf("%d", page))
	}

	query.Add("order_by", fmt.Sprintf("\"%s\"", order))

	return g.makeTendermintRPCRequest(g.context.Context, "/block_search", query.Encode())
}

func (g *CosmosGateway) proxy(req types.InboundRequest) (interface{}, error) {
	dataBytes, err := json.Marshal(req.Payload)
	if err != nil {
		logger.Logger.Error("CosmosGateway - Handle - Marshal payload failed", zap.Error(err))
		return nil, err
	}

	gatewayReq, err := http.NewRequestWithContext(g.context.Context, req.Method, req.Path, strings.NewReader(string(dataBytes)))
	if err != nil {
		logger.Logger.Error("CosmosGateway - Handle - Create request failed", zap.Error(err))
		return nil, err
	}

	gatewayReq.Header.Set("Content-Type", "application/json")

	gatewayReq = g.encodeQuery(gatewayReq, req)

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

func (g *CosmosGateway) dashboard() (interface{}, error) {
	allValidators := types.AllValidators{
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

func (g *CosmosGateway) validators(req types.InboundRequest) (*types.ValidatorsResponse, error) {
	validatorsResponse, err := g.allValidators()
	if err != nil {
		logger.Logger.Error("CosmosGateway - validators - allValidators failed", zap.Error(err))
		return nil, err
	}

	return g.filterAndPaginateValidators(validatorsResponse, req.Payload)
}

func (g *CosmosGateway) filterAndPaginateValidators(response *types.ValidatorsResponse, payload map[string]interface{}) (*types.ValidatorsResponse, error) {
	var result = new(types.ValidatorsResponse)
	var filteredValidators []types.QueryValidator

	result.Actors = response.Actors

	var (
		address           string
		valkey            string
		pubkey            string
		moniker           string
		status            string
		offset            int
		limit             = sekaitypes.PageIterationLimit - 1
		proposer          string
		all               bool
		statusOnly        bool
		stakingPoolStatus string
		validatorNodeId   string
		sentryNodeId      string
	)

	if val, ok := payload["all"]; ok {
		allStr := fmt.Sprintf("%v", val)
		all = allStr == "true"
	}

	if all {
		return response, nil
	}

	if val, ok := payload["address"]; ok {
		address = fmt.Sprintf("%v", val)
	}
	if val, ok := payload["valkey"]; ok {
		valkey = fmt.Sprintf("%v", val)
	}
	if val, ok := payload["pubkey"]; ok {
		pubkey = fmt.Sprintf("%v", val)
	}
	if val, ok := payload["moniker"]; ok {
		moniker = fmt.Sprintf("%v", val)
	}
	if val, ok := payload["status"]; ok {
		status = fmt.Sprintf("%v", val)
	}
	if val, ok := payload["offset"]; ok {
		offsetStr := fmt.Sprintf("%v", val)
		offsetVal, err := strconv.Atoi(offsetStr)
		if err == nil {
			offset = offsetVal
		}
	}
	if val, ok := payload["limit"]; ok {
		limitStr := fmt.Sprintf("%v", val)
		limitVal, err := strconv.Atoi(limitStr)
		if err == nil {
			limit = limitVal
		}
	}

	if val, ok := payload["proposer"]; ok {
		proposer = fmt.Sprintf("%v", val)
	}
	if val, ok := payload["status_only"]; ok {
		statusOnlyStr := fmt.Sprintf("%v", val)
		statusOnly = statusOnlyStr == "true"
	}
	if val, ok := payload["staking_pool_status"]; ok {
		stakingPoolStatus = fmt.Sprintf("%v", val)
	}
	if val, ok := payload["validator_node_id"]; ok {
		validatorNodeId = fmt.Sprintf("%v", val)
	}
	if val, ok := payload["sentry_node_id"]; ok {
		sentryNodeId = fmt.Sprintf("%v", val)
	}

	for _, validator := range response.Validators {
		match := true

		if address != "" && validator.Address != address {
			match = false
		}

		if valkey != "" && validator.Valkey != valkey {
			match = false
		}

		if pubkey != "" && validator.Pubkey != pubkey {
			match = false
		}

		if moniker != "" && !strings.Contains(strings.ToLower(validator.Moniker), strings.ToLower(moniker)) {
			match = false
		}

		if status != "" && validator.Status != status {
			match = false
		}

		if proposer != "" && validator.Proposer != proposer {
			match = false
		}

		if stakingPoolStatus != "" && validator.StakingPoolStatus != stakingPoolStatus {
			match = false
		}

		if validatorNodeId != "" && validator.Validator_node_id != validatorNodeId {
			match = false
		}

		if sentryNodeId != "" && validator.Sentry_node_id != sentryNodeId {
			match = false
		}

		if match {
			if statusOnly {
				simplifiedValidator := types.QueryValidator{
					Address: validator.Address,
					Status:  validator.Status,
				}
				filteredValidators = append(filteredValidators, simplifiedValidator)
			} else {
				filteredValidators = append(filteredValidators, validator)
			}
		}
	}

	if offset >= len(filteredValidators) {
		result.Validators = []types.QueryValidator{}
	} else {
		endIndex := offset + limit
		if endIndex > len(filteredValidators) {
			endIndex = len(filteredValidators)
		}

		result.Validators = filteredValidators[offset:endIndex]
	}

	result.Pagination.Total = len(filteredValidators)

	return result, nil
}

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

func (g *CosmosGateway) encodeQuery(gatewayReq *http.Request, req types.InboundRequest) *http.Request {
	if req.Method == http.MethodGet && len(req.Payload) > 0 {
		q := gatewayReq.URL.Query()
		for k, v := range req.Payload {
			switch val := v.(type) {
			case string:
				q.Add(k, val)
			case float64:
				q.Add(k, fmt.Sprintf("%v", val))
			case bool:
				q.Add(k, fmt.Sprintf("%v", val))
			case []interface{}:
				for _, item := range val {
					q.Add(k, fmt.Sprintf("%v", item))
				}
			case map[string]interface{}:
				for subKey, subVal := range val {
					q.Add(k+"."+subKey, fmt.Sprintf("%v", subVal))
				}
			default:
				q.Add(k, fmt.Sprintf("%v", v))
			}
		}
		gatewayReq.URL.RawQuery = q.Encode()
	}

	return gatewayReq
}
