package internal

import (
	"github.com/saiset-co/sai-service/service"
)

func (is *InternalService) NewHandler() service.Handler {
	return service.Handler{
		"test": service.HandlerElement{
			Name:        "test",
			Description: "Test handler",
			Function: func(data, meta interface{}) (interface{}, int, error) {
				return nil, 0, nil
			},
		},
	}
}
