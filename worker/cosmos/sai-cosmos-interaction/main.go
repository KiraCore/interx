package main

import (
	"github.com/cosmos/cosmos-sdk/types"
	saiService "github.com/KiraCore/sai-service/service"
	"github.com/KiraCore/saiCosmosInteraction/internal"
	"github.com/KiraCore/saiCosmosInteraction/logger"
)

var (
	AccountAddressPrefix   = "kira"
	AccountPubKeyPrefix    = "kirapub"
	ValidatorAddressPrefix = "kiravaloper"
	ValidatorPubKeyPrefix  = "kiravaloperpub"
	ConsNodeAddressPrefix  = "kiravalcons"
	ConsNodePubKeyPrefix   = "kiravalconspub"
)

func main() {
	svc := saiService.NewService("saiCosmosInteraction")
	svc.RegisterConfig("config.yml")

	logger.Logger = svc.Logger

	is := internal.InternalService{Context: svc.Context}

	svc.RegisterInitTask(is.Init)

	config := types.GetConfig()
	config.SetBech32PrefixForAccount(AccountAddressPrefix, AccountPubKeyPrefix)
	config.SetBech32PrefixForValidator(ValidatorAddressPrefix, ValidatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(ConsNodeAddressPrefix, ConsNodePubKeyPrefix)
	config.Seal()

	svc.RegisterHandlers(
		is.NewHandler(),
	)

	svc.Start()
}
