package main

import (
	"github.com/KiraCore/sai-eth-interaction/internal"
	saiService "github.com/KiraCore/sai-service/service"
)

func main() {
	svc := saiService.NewService("saiEthInteraction")

	svc.RegisterConfig("config.yml")

	is := internal.InternalService{Context: svc.Context}

	svc.RegisterInitTask(is.Init)

	svc.RegisterHandlers(
		is.NewHandler())

	svc.Start()

}
