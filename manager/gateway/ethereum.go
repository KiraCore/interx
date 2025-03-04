package gateway

import (
	"context"
	"encoding/json"
	"github.com/saiset-co/sai-interx-manager/logger"
	"go.uber.org/zap"
	"time"

	"github.com/saiset-co/sai-interx-manager/types"
)

type EthereumGateway struct {
	*BaseGateway
	url string
}

var _ types.Gateway = (*EthereumGateway)(nil)

func NewEthereumGateway(url string, retryAttempts int, retryDelay time.Duration, rateLimit int) *EthereumGateway {
	return &EthereumGateway{
		BaseGateway: NewBaseGateway(retryAttempts, retryDelay, rateLimit),
		url:         url,
	}
}

func (g *EthereumGateway) Handle(ctx context.Context, data []byte) (interface{}, error) {
	var req struct {
		Method  string      `json:"method"`
		Path    string      `json:"path"`
		Payload interface{} `json:"payload"`
	}

	if err := json.Unmarshal(data, &req); err != nil {
		logger.Logger.Error("EthereumGateway - Handle", zap.Error(err))
		return nil, err
	}

	return g.retry.Do(func() (interface{}, error) {
		if err := g.rateLimit.Wait(ctx); err != nil {
			logger.Logger.Error("EthereumGateway - Handle", zap.Error(err))
			return nil, err
		}
		return g.makeSaiRequest(ctx, g.url, req)
	})
}
