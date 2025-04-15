package gateway

import (
	"fmt"
	"github.com/saiset-co/sai-interx-manager/logger"
	"go.uber.org/zap"
	"time"

	saiService "github.com/saiset-co/sai-service/service"
	"github.com/spf13/cast"

	"github.com/saiset-co/sai-interx-manager/types"
)

type GatewayFactory struct {
	context *saiService.Context
	storage types.Storage
}

func NewGatewayFactory(context *saiService.Context, storage types.Storage) *GatewayFactory {
	return &GatewayFactory{
		context: context,
		storage: storage,
	}
}

func (f *GatewayFactory) CreateGateway(gatewayType string) (types.Gateway, error) {
	switch gatewayType {
	case "ethereum":
		return NewEthereumGateway(
			f.context,
			cast.ToStringMapString(f.context.GetConfig("ethereum.nodes", map[string]string{})),
			f.storage,
			cast.ToInt(f.context.GetConfig("ethereum.retries", 1)),
			time.Duration(cast.ToInt64(f.context.GetConfig("ethereum.retry_delay", 10))),
			cast.ToInt(f.context.GetConfig("ethereum.rate_limit", 10)),
		)
	case "cosmos":
		return NewCosmosGateway(
			f.context,
			cast.ToString(f.context.GetConfig("cosmos.node.tendermint", "")),
			cast.ToString(f.context.GetConfig("cosmos.node.json_rpc", "")),
			f.storage,
			cast.ToInt(f.context.GetConfig("cosmos.gw_timeout", 3)),
			cast.ToInt(f.context.GetConfig("cosmos.retries", 1)),
			time.Duration(cast.ToInt64(f.context.GetConfig("cosmos.retry_delay", 10))),
			cast.ToInt(f.context.GetConfig("cosmos.rate_limit", 10)),
			cast.ToStringMapBool(f.context.GetConfig("cosmos.tx_modes", map[string]bool{})),
		)
	case "bitcoin":
		return NewBitcoinGateway(
			f.context,
			cast.ToString(f.context.GetConfig("bitcoin.url", "")),
			cast.ToInt(f.context.GetConfig("bitcoin.retries", 1)),
			time.Duration(cast.ToInt64(f.context.GetConfig("bitcoin.retry_delay", 10))),
			cast.ToInt(f.context.GetConfig("bitcoin.rate_limit", 10)),
		)
	case "storage":
		return NewStorageGateway(
			f.context,
			f.storage,
			cast.ToInt(f.context.GetConfig("storage.retries", 1)),
			time.Duration(cast.ToInt64(f.context.GetConfig("storage.retry_delay", 10))),
			cast.ToInt(f.context.GetConfig("storage.rate_limit", 10)),
		)
	default:
		err := fmt.Errorf("unknown gateway type: %s", gatewayType)
		logger.Logger.Error("GatewayFactory - CreateGateway", zap.Error(err))

		return nil, err
	}
}
