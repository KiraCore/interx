package gateway

import (
	"context"
	"encoding/json"
	"github.com/saiset-co/sai-interx-manager/logger"
	"go.uber.org/zap"
	"strings"
	"time"

	"github.com/saiset-co/sai-interx-manager/types"
)

type CosmosGateway struct {
	*BaseGateway
	url  string
	node string
}

var _ types.Gateway = (*CosmosGateway)(nil)

func NewCosmosGateway(url, node string, retryAttempts int, retryDelay time.Duration, rateLimit int) *CosmosGateway {
	return &CosmosGateway{
		BaseGateway: NewBaseGateway(retryAttempts, retryDelay, rateLimit),
		url:         url,
		node:        node,
	}
}

func (g *CosmosGateway) Handle(ctx context.Context, data []byte) (interface{}, error) {
	var req struct {
		Method  string      `json:"method"`
		Path    string      `json:"path"`
		Payload interface{} `json:"payload"`
	}

	if err := json.Unmarshal(data, &req); err != nil {
		logger.Logger.Error("CosmosGateway - Handle", zap.Error(err))
		return nil, err
	}

	path := strings.Replace(req.Path, "/api/kira/gov", "/kira/gov", -1)
	path = strings.Replace(path, "/api/kira", "/cosmos/bank/v1beta1", -1)

	//Todo Add check can we handle request from oue storage or execute in the interaction microservice
	//return g.retry.Do(func() (interface{}, error) {
	//	if err := g.rateLimit.Wait(ctx); err != nil {
	//		logger.Logger.Error("CosmosGateway - Handle", zap.Error(err))
	//		return nil, err
	//	}
	//	return g.makeSaiRequest(ctx, g.url, req)
	//})

	//If not, execute API proxy to the cosmos node
	return g.retry.Do(func() (interface{}, error) {
		if err := g.rateLimit.Wait(ctx); err != nil {
			logger.Logger.Error("CosmosGateway - Handle", zap.Error(err))
			return nil, err
		}
		return g.makeRequest(ctx, req.Method, g.node+path, req.Payload)
	})
}
