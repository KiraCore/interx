package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	jsonrpc2 "github.com/KeisukeYamashita/go-jsonrpc"
	"go.uber.org/zap"

	"github.com/saiset-co/sai-interx-manager/logger"
	"github.com/saiset-co/sai-interx-manager/types"
)

type EthereumGateway struct {
	*BaseGateway
	storage    types.Storage
	rpcProxies map[string]*jsonrpc2.RPCClient
}

var _ types.Gateway = (*EthereumGateway)(nil)

func newJsonRPCClients(chains map[string]string) map[string]*jsonrpc2.RPCClient {
	proxies := map[string]*jsonrpc2.RPCClient{}

	for chainId, url := range chains {
		proxies[chainId] = jsonrpc2.NewRPCClient(url)
	}

	return proxies
}

func NewEthereumGateway(chains map[string]string, storage types.Storage, retryAttempts int, retryDelay time.Duration, rateLimit int) (*EthereumGateway, error) {
	return &EthereumGateway{
		BaseGateway: NewBaseGateway(retryAttempts, retryDelay, rateLimit),
		rpcProxies:  newJsonRPCClients(chains),
		storage:     storage,
	}, nil
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

	chainId, method, err := g.convert(req.Path)
	if err != nil {
		logger.Logger.Error("EthereumGateway - Handle", zap.Error(err))
		return nil, err
	}

	if req.Method == "block" {
		return g.retry.Do(func() (interface{}, error) {
			if err := g.rateLimit.Wait(ctx); err != nil {
				logger.Logger.Error("CosmosGateway - Handle - Rate limit exceeded", zap.Error(err))
				return nil, err
			}

			return g.storage.Read("ethereum"+chainId, map[string]interface{}{}, nil, []string{})
		})
	}

	client, ok := g.rpcProxies[chainId]
	if !ok {
		err = errors.New("chain not found")
		logger.Logger.Error("EthereumGateway - Handle", zap.Error(err))
		return nil, err
	}

	return g.retry.Do(func() (interface{}, error) {
		if err := g.rateLimit.Wait(ctx); err != nil {
			logger.Logger.Error("EthereumGateway - Handle", zap.Error(err))
			return nil, err
		}
		return client.Call(method, req.Payload)
	})
}

func (g *EthereumGateway) Close() {

}

func (g *EthereumGateway) convert(originalPath string) (chainId, method string, err error) {
	paths := strings.Split(originalPath, "/")
	if len(paths) < 3 {
		return "", "", err
	}

	return paths[2], paths[3], nil
}
