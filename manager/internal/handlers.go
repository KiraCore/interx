package internal

import (
	"encoding/json"

	"go.uber.org/zap"

	"github.com/saiset-co/sai-interx-manager/logger"
	"github.com/saiset-co/sai-interx-manager/p2p"
	"github.com/saiset-co/sai-service/service"
)

func (is *InternalService) NewHandler() service.Handler {
	return service.Handler{
		"metrics": service.HandlerElement{
			Name:        "Metrics",
			Description: "Test endpoint for the balancer",
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

				return result, 200, nil
			},
			Middlewares: []service.Middleware{
				is.server.MetricsCollector().CreateMetricsMiddleware("metrics"),
				is.server.LoadBalancer().CreateLoadBalancerMiddleware("metrics"),
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

				return result, 200, nil
			},
			Middlewares: []service.Middleware{
				is.server.MetricsCollector().CreateMetricsMiddleware("metrics"),
				is.server.LoadBalancer().CreateLoadBalancerMiddleware("metrics"),
			},
		},
		"rosetta": service.HandlerElement{
			Name:        "RosettaAPI",
			Description: "Proxy api endpoint for Rosetta",
			Function: func(data, meta interface{}) (interface{}, int, error) {
				return nil, 500, nil
			},
			Middlewares: []service.Middleware{
				is.server.MetricsCollector().CreateMetricsMiddleware("metrics"),
				is.server.LoadBalancer().CreateLoadBalancerMiddleware("metrics"),
			},
		},
		"bitcoin": service.HandlerElement{
			Name:        "RosettaAPI",
			Description: "Proxy api endpoint for Rosetta",
			Function: func(data, meta interface{}) (interface{}, int, error) {
				return nil, 500, nil
			},
			Middlewares: []service.Middleware{
				is.server.MetricsCollector().CreateMetricsMiddleware("metrics"),
				is.server.LoadBalancer().CreateLoadBalancerMiddleware("metrics"),
			},
		},
		"default": service.HandlerElement{
			Name:        "DefaultAPI",
			Description: "Proxy default api endpoints",
			Function: func(data, meta interface{}) (interface{}, int, error) {
				return nil, 0, nil
			},
			Middlewares: []service.Middleware{
				is.server.MetricsCollector().CreateMetricsMiddleware("metrics"),
				is.server.LoadBalancer().CreateLoadBalancerMiddleware("metrics"),
			},
		},
	}
}
