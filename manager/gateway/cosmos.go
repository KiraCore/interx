package gateway

import (
	"context"
	"encoding/json"
	"github.com/saiset-co/sai-interx-manager/logger"
	"go.uber.org/zap"
	"time"

	"github.com/saiset-co/sai-interx-manager/types"
)

type CosmosGateway struct {
	*BaseGateway
	url string
}

var _ types.Gateway = (*CosmosGateway)(nil) // Verify interface implementation

func NewCosmosGateway(url string, retryAttempts int, retryDelay time.Duration, rateLimit int) *CosmosGateway {
	return &CosmosGateway{
		BaseGateway: NewBaseGateway(retryAttempts, retryDelay, rateLimit),
		url:         url,
	}
}

func (g *CosmosGateway) Handle(ctx context.Context, data []byte) (interface{}, error) {
	var req struct {
		Method   string      `json:"method"`
		Data     interface{} `json:"data"`
		Metadata struct {
			Token string `json:"token"`
		} `json:"metadata"`
	}

	if err := json.Unmarshal(data, &req); err != nil {
		logger.Logger.Error("CosmosGateway - Handle", zap.Error(err))
		return nil, err
	}

	return g.retry.Do(func() (interface{}, error) {
		if err := g.rateLimit.Wait(ctx); err != nil {
			logger.Logger.Error("CosmosGateway - Handle", zap.Error(err))
			return nil, err
		}
		return g.makeRequest(ctx, g.url, req)
	})
}
