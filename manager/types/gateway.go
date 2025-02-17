package types

import (
	"context"
)

type Gateway interface {
	Handle(ctx context.Context, data []byte) (interface{}, error)
}

type GatewayFactory interface {
	CreateGateway(gatewayType string) (Gateway, error)
}
