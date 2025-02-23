package internal

import (
	"github.com/saiset-co/sai-interx-manager/gateway"
	"github.com/saiset-co/sai-interx-manager/p2p/balancer"
	"github.com/saiset-co/sai-interx-manager/p2p/core"
	"github.com/saiset-co/sai-interx-manager/p2p/metrics"
	"github.com/saiset-co/sai-interx-manager/p2p/network"
	"github.com/saiset-co/sai-interx-manager/types"
	"github.com/saiset-co/sai-service/service"
	"github.com/spf13/cast"
	"time"
)

type InternalService struct {
	Context        *service.Context
	gatewayFactory types.GatewayFactory
	storage        types.Storage
	server         *network.Server
	balancer       *balancer.LoadBalancer
	metrics        *metrics.Collector
}

func (is *InternalService) Init() {
	is.storage = types.NewStorage(
		cast.ToString(is.Context.GetConfig("storage.url", "")),
		cast.ToString(is.Context.GetConfig("storage.token", "")),
	)

	is.gatewayFactory = gateway.NewGatewayFactory(is.Context, is.storage)

	nodeId := core.NodeID(cast.ToString(is.Context.GetConfig("node_id", "")))

	weights := metrics.Weights{
		CPU:     cast.ToFloat64(is.Context.GetConfig("metrics.weights.cpu", 0.3)),
		RPS:     cast.ToFloat64(is.Context.GetConfig("metrics.weights.rps", 0.2)),
		Memory:  cast.ToFloat64(is.Context.GetConfig("metrics.weights.memory", 0.3)),
		Latency: cast.ToFloat64(is.Context.GetConfig("metrics.weights.latency", 0.2)),
	}

	collector := metrics.NewCollector(
		nodeId,
		weights,
		time.Duration(cast.ToInt(is.Context.GetConfig("metrics.weights.latency", 10))),
	)

	p2pServerConfig := network.ServerConfig{
		NodeID:   nodeId,
		Address:  cast.ToString(is.Context.GetConfig("listen_address", "")),
		MaxPeers: cast.ToInt(is.Context.GetConfig("max_peers", "")),
		Metrics:  collector,
	}

	is.server = network.NewServer(is.Context.Context, p2pServerConfig)

	is.balancer = balancer.New(
		nodeId,
		collector,
		cast.ToFloat64(is.Context.GetConfig("load_balancer.threshold", 0.2)),
	)
}

func (is *InternalService) Process() {
	if err := is.server.Start(); err != nil {
		panic(err)
	}
}
