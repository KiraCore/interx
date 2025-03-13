package types

import (
	"context"
)

type Gateway interface {
	Handle(ctx context.Context, data []byte) (interface{}, error)
	Close()
}

type GatewayFactory interface {
	CreateGateway(gatewayType string) (Gateway, error)
}

type PathMappings struct {
	Pattern     string
	Replacement string
}
