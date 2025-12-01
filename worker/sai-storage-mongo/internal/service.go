package internal

import (
	"github.com/KiraCore/sai-service/service"
	"github.com/KiraCore/sai-storage-mongo/mongo"
)

type InternalService struct {
	Name    string
	Context *service.Context
	Client  *mongo.Client
}
