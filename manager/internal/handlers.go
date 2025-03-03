package internal

import (
	"encoding/json"

	"github.com/saiset-co/sai-interx-manager/logger"
	"github.com/saiset-co/sai-interx-manager/p2p"
	"github.com/saiset-co/sai-service/service"

	"go.uber.org/zap"
)

func (is *InternalService) NewHandler() service.Handler {
	return service.Handler{
		"metrics": service.HandlerElement{
			Name:        "metrics",
			Description: "metrics",
			Function: func(data, meta interface{}) (interface{}, int, error) {
				metrics := is.server.MetricsCollector().GetAllNodesMetrics()
				nodeId := is.server.PeerManager().GetPeerId()
				return struct {
					NodeSentReport p2p.NodeID
					Metrics        map[p2p.NodeID]p2p.NodeMetrics
				}{
					NodeSentReport: nodeId,
					Metrics:        metrics,
				}, 200, nil
			},
			Middlewares: []service.Middleware{
				is.server.MetricsCollector().CreateMetricsMiddleware("metrics"),
				is.server.LoadBalancer().CreateLoadBalancerMiddleware("metrics"),
			},
		},
		"ethereum": service.HandlerElement{
			Name:        "EthereumAPI",
			Description: "Proxy api endpoint for an ethereum network",
			Function: func(data, meta interface{}) (interface{}, int, error) {
				gateway, err := is.gatewayFactory.CreateGateway("eth")
				if err != nil {
					logger.Logger.Error("DelEthereumAPI", zap.Error(err))
					return nil, 500, err
				}

				dataBytes, err := json.Marshal(data)
				if err != nil {
					logger.Logger.Error("EthereumAPI", zap.Error(err))
					return nil, 500, err
				}

				result, err := gateway.Handle(is.Context.Context, dataBytes)
				if err != nil {
					logger.Logger.Error("EthereumAPI", zap.Error(err))
					return nil, 500, err
				}

				return result, 0, nil
			},
		},
		"cosmos": service.HandlerElement{
			Name:        "CosmosAPI",
			Description: "Proxy api endpoint for a cosmos network",
			Function: func(data, meta interface{}) (interface{}, int, error) {
				gateway, err := is.gatewayFactory.CreateGateway("cosmos")
				if err != nil {
					logger.Logger.Error("CosmosAPI", zap.Error(err))
					return nil, 500, err
				}

				dataBytes, err := json.Marshal(data)
				if err != nil {
					logger.Logger.Error("CosmosAPI", zap.Error(err))
					return nil, 500, err
				}

				result, err := gateway.Handle(is.Context.Context, dataBytes)
				if err != nil {
					logger.Logger.Error("EthereumAPI", zap.Error(err))
					return nil, 500, err
				}

				return result, 0, nil
			},
		},
		"storage": service.HandlerElement{
			Name:        "StorageAPI",
			Description: "Proxy api endpoint for the storage",
			Function: func(data, meta interface{}) (interface{}, int, error) {
				gateway, err := is.gatewayFactory.CreateGateway("storage")
				if err != nil {
					logger.Logger.Error("CStorageAPI", zap.Error(err))
					return nil, 500, err
				}

				dataBytes, err := json.Marshal(data)
				if err != nil {
					logger.Logger.Error("StorageAPI", zap.Error(err))
					return nil, 500, err
				}

				result, err := gateway.Handle(is.Context.Context, dataBytes)
				if err != nil {
					logger.Logger.Error("EthereumAPI", zap.Error(err))
					return nil, 500, err
				}

				return result, 0, nil
			},
		},
	}
}
