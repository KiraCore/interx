package main

import (
	"github.com/KiraCore/sai-interx-manager/internal"
	"github.com/KiraCore/sai-interx-manager/logger"

	"github.com/KiraCore/sai-service/service"
)

func main() {
	svc := service.NewService("saiInterxManager")
	is := internal.InternalService{Context: svc.Context}

	svc.RegisterConfig("config.yml")

	logger.Logger = svc.Logger

	is.Init()

	svc.RegisterTasks([]func(){
		is.Process,
	})

	svc.RegisterHandlers(
		is.NewHandler(),
	)

	svc.Start()
}
