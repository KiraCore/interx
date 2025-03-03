package internal

import (
	"time"

	"github.com/spf13/cast"

	"github.com/saiset-co/sai-interx-manager/gateway"
	"github.com/saiset-co/sai-interx-manager/p2p"
	"github.com/saiset-co/sai-interx-manager/p2p/config"
	"github.com/saiset-co/sai-interx-manager/p2p/net"
	"github.com/saiset-co/sai-interx-manager/types"
	"github.com/saiset-co/sai-service/service"
)

type InternalService struct {
	Context        *service.Context
	gatewayFactory types.GatewayFactory
	storage        types.Storage
	server         p2p.Network
}

func (is *InternalService) Init() {
	is.storage = types.NewStorage(
		cast.ToString(is.Context.GetConfig("storage.url", "")),
		cast.ToString(is.Context.GetConfig("storage.token", "")),
	)

	is.gatewayFactory = gateway.NewGatewayFactory(is.Context, is.storage)

	nodeID := cast.ToString(is.Context.GetConfig("p2p.id", ""))
	windowSize := cast.ToInt(is.Context.GetConfig("balancer.window_size", 60))
	threshold := cast.ToFloat64(is.Context.GetConfig("balancer.threshold", 0.2))

	networkConfig := config.NewNetworkConfig(
		config.WithNodeID(p2p.NodeID(nodeID)),
		config.WithListenAddress(cast.ToString(is.Context.GetConfig("p2p.address", "0.0.0.0:9000"))),
		config.WithMaxPeers(cast.ToInt(is.Context.GetConfig("p2p.max_peers", 3))),
		config.WithHTTPPort(cast.ToInt(is.Context.GetConfig("common.http.port", 8080))),
		config.WithMetricsWindowSize(time.Duration(windowSize)*time.Second),
		config.WithLoadBalancerThreshold(threshold),
		config.WithInitialPeers(cast.ToStringSlice(is.Context.GetConfig("p2p.peers", []string{}))),
	)

	server, err := net.NewNetwork(is.Context.Context, networkConfig)
	if err != nil {
		panic(err)
	}

	is.server = server
}

func (is *InternalService) Process() {
	if err := is.server.Start(); err != nil {
		panic(err)
	}
}
