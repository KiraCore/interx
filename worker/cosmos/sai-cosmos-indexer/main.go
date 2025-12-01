package main

import (
	saiService "github.com/KiraCore/sai-service/service"
	"github.com/KiraCore/saiCosmosIndexer/internal"
	"github.com/KiraCore/saiCosmosIndexer/logger"
)

func main() {
	svc := saiService.NewService("saiCosmosIndexer")
	is := internal.InternalService{Context: svc.Context}

	svc.RegisterConfig("config.yml")

	logger.Logger = svc.Logger

	svc.RegisterInitTask(is.Init)

	svc.RegisterTasks([]func(){
		is.Process,
	})

	svc.RegisterHandlers(
		is.NewHandler(),
	)

	svc.Start()
}
