package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
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
}

var _ types.Gateway = (*CosmosGateway)(nil)

func newGRPCGatewayProxy(ctx context.Context, address string) (*Proxy, error) {
	conn, err := grpc.DialContext(
		ctx,
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
		return nil, err
	}

	return &Proxy{
		mux:  mux,
		conn: conn,
	}, nil
}

func registerHandlers(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	if err := cosmosBank.RegisterQueryHandler(context.Background(), mux, conn); err != nil {
		logger.Logger.Error("registerHandlers", zap.Error(err))
		return err
	}

	if err := cosmosAuth.RegisterQueryHandler(context.Background(), mux, conn); err != nil {
		logger.Logger.Error("registerHandlers", zap.Error(err))
		return err
	}

	if err := kiraGov.RegisterQueryHandler(context.Background(), mux, conn); err != nil {
		logger.Logger.Error("registerHandlers", zap.Error(err))
		return err
	}

	if err := kiraStaking.RegisterQueryHandler(context.Background(), mux, conn); err != nil {
		logger.Logger.Error("registerHandlers", zap.Error(err))
		return err
	}

	if err := kiraMultiStaking.RegisterQueryHandler(context.Background(), mux, conn); err != nil {
		logger.Logger.Error("registerHandlers", zap.Error(err))
		return err
	}

	if err := kiraSlashing.RegisterQueryHandler(context.Background(), mux, conn); err != nil {
		logger.Logger.Error("registerHandlers", zap.Error(err))
		return err
	}

	if err := kiraTokens.RegisterQueryHandler(context.Background(), mux, conn); err != nil {
		logger.Logger.Error("registerHandlers", zap.Error(err))
		return err
	}

	if err := kiraUpgrades.RegisterQueryHandler(context.Background(), mux, conn); err != nil {
		logger.Logger.Error("registerHandlers", zap.Error(err))
		return err
	}

	if err := kiraSpending.RegisterQueryHandler(context.Background(), mux, conn); err != nil {
		logger.Logger.Error("registerHandlers", zap.Error(err))
		return err
	}

	if err := kiraUbi.RegisterQueryHandler(context.Background(), mux, conn); err != nil {
		logger.Logger.Error("registerHandlers", zap.Error(err))
		return err
	}

	return nil
}

func NewCosmosGateway(ctx context.Context, url, node string, storage types.Storage, timeout, retryAttempts int, retryDelay time.Duration, rateLimit int) (*CosmosGateway, error) {
	ct, _ := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)

	proxy, err := newGRPCGatewayProxy(ct, node)
	if err != nil {
		return nil, err
	}

	return &CosmosGateway{
		BaseGateway: NewBaseGateway(retryAttempts, retryDelay, rateLimit),
		storage:     storage,
		url:         url,
		grpcProxy:   proxy,
	}, nil
}

func (g *CosmosGateway) Handle(ctx context.Context, data []byte) (interface{}, error) {
	var req struct {
		Method  string                 `json:"method"`
		Path    string                 `json:"path"`
		Payload map[string]interface{} `json:"payload"`
	}

	if err := json.Unmarshal(data, &req); err != nil {
		logger.Logger.Error("CosmosGateway - Handle - Unmarshal request failed", zap.Error(err))
		return nil, err
	}

	if req.Method == "block" {
		return g.retry.Do(func() (interface{}, error) {
			if err := g.rateLimit.Wait(ctx); err != nil {
				logger.Logger.Error("CosmosGateway - Handle - Rate limit exceeded", zap.Error(err))
				return nil, err
			}

			return g.storage.Read("kira", map[string]interface{}{}, nil, []string{})
		})
	}

	return g.retry.Do(func() (interface{}, error) {
		if err := g.rateLimit.Wait(ctx); err != nil {
			logger.Logger.Error("CosmosGateway - Handle - Rate limit exceeded", zap.Error(err))
			return nil, err
		}

		dataBytes, err := json.Marshal(req.Payload)
		if err != nil {
			logger.Logger.Error("CosmosGateway - Handle - Marshal payload failed", zap.Error(err))
			return nil, err
		}

		gatewayReq, err := http.NewRequestWithContext(ctx, req.Method, req.Path, strings.NewReader(string(dataBytes)))
		if err != nil {
			logger.Logger.Error("CosmosGateway - Handle - Create request failed", zap.Error(err))
			return nil, err
		}

		gatewayReq.Header.Set("Content-Type", "application/json")

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

		recorder := httptest.NewRecorder()
		g.grpcProxy.mux.ServeHTTP(recorder, gatewayReq)
		resp := recorder.Result()

		defer resp.Body.Close()

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Logger.Error("CosmosGateway - Handle - Read response body failed", zap.Error(err))
			return nil, err
		}

		if resp.StatusCode >= 400 {
			errMsg := fmt.Sprintf("gRPC gateway error: status=%d, body=%s", resp.StatusCode, string(respBody))
			logger.Logger.Error("CosmosGateway - Handle - gRPC gateway error response",
				zap.Int("status", resp.StatusCode),
				zap.String("body", string(respBody)))
			return nil, fmt.Errorf(errMsg)
		}

		var result interface{}
		if len(respBody) > 0 {
			if err := json.Unmarshal(respBody, &result); err != nil {
				logger.Logger.Error("CosmosGateway - Handle - Unmarshal response failed", zap.Error(err))
				return nil, err
			}
		}

		return result, nil
	})
}

func (g *CosmosGateway) Close() {
	g.grpcProxy.conn.Close()
}
