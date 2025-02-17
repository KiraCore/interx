package internal

import (
	"github.com/saiset-co/sai-interx-manager/gateway"
	"github.com/saiset-co/sai-interx-manager/types"
	"github.com/saiset-co/sai-service/service"
	"github.com/spf13/cast"
)

type InternalService struct {
	Context        *service.Context
	gatewayFactory types.GatewayFactory
	storage        types.Storage
}

func (is *InternalService) Init() {
	is.storage = types.NewStorage(
		cast.ToString(is.Context.GetConfig("storage.url", "")),
		cast.ToString(is.Context.GetConfig("storage.token", "")),
	)
	is.gatewayFactory = gateway.NewGatewayFactory(is.Context, is.storage)
}

func (is *InternalService) Process() {

}
