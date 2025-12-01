package main

import (
	"github.com/KiraCore/sai-interx-proxy/internal"
	"github.com/KiraCore/sai-interx-proxy/logger"

	"github.com/KiraCore/sai-service/service"
)

func main() {
	svc := service.NewService("saiInterxProxy")
	is := internal.InternalService{Context: svc.Context}

	svc.RegisterConfig("config.yml")

	logger.Logger = svc.Logger

	is.Init()

	svc.RegisterTasks([]func(){
		is.Process,
	})

	svc.Start()
}
