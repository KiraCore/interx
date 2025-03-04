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
	case "eth":
		return NewEthereumGateway(
			cast.ToString(f.context.GetConfig("ethereum.url", "")),
			cast.ToInt(f.context.GetConfig("ethereum.retries", 1)),
			time.Duration(cast.ToInt64(f.context.GetConfig("ethereum.retry_delay", 10))),
			cast.ToInt(f.context.GetConfig("ethereum.rate_limit", 10)),
		), nil
	case "cosmos":
		return NewCosmosGateway(
			cast.ToString(f.context.GetConfig("cosmos.url", "")),
			cast.ToString(f.context.GetConfig("cosmos.node", "")),
			cast.ToInt(f.context.GetConfig("cosmos.retries", 1)),
			time.Duration(cast.ToInt64(f.context.GetConfig("cosmos.retry_delay", 10))),
			cast.ToInt(f.context.GetConfig("cosmos.rate_limit", 10)),
		), nil
	case "bitcoin":
		return NewBitcoinGateway(
			cast.ToString(f.context.GetConfig("bitcoin.url", "")),
			cast.ToInt(f.context.GetConfig("bitcoin.retries", 1)),
			time.Duration(cast.ToInt64(f.context.GetConfig("bitcoin.retry_delay", 10))),
			cast.ToInt(f.context.GetConfig("bitcoin.rate_limit", 10)),
		), nil
	case "storage":
		return NewStorageGateway(
			f.storage,
			cast.ToInt(f.context.GetConfig("storage.retries", 1)),
			time.Duration(cast.ToInt64(f.context.GetConfig("storage.retry_delay", 10))),
			cast.ToInt(f.context.GetConfig("storage.rate_limit", 10)),
		), nil
	default:
		err := fmt.Errorf("unknown gateway type: %s", gatewayType)
		logger.Logger.Error("GatewayFactory - CreateGateway", zap.Error(err))

		return nil, err
	}
}
