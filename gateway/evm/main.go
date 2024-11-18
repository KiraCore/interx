package evm

import (
	"github.com/KiraCore/interx/config"
	"github.com/gorilla/mux"
)

// RegisterRequest is a function to register requests.
func RegisterRequest(router *mux.Router, rpcAddr string) {
	RegisterEVMStatusRoutes(router, rpcAddr)
	RegisterEVMBlockRoutes(router, rpcAddr)
	RegisterEVMTransactionRoutes(router, rpcAddr)
	RegisterEVMTransferRoutes(router, rpcAddr)
	RegisterEVMAccountsRoutes(router, rpcAddr)
	RegisterEVMBalancesRoutes(router, rpcAddr)
	RegisterEVMAbiRoutes(router, rpcAddr)
	RegisterEVMContractRoutes(router, rpcAddr)
	RegisterEVMFaucetRoutes(router, rpcAddr)
}

func GetChainConfig(chain string) (bool, *config.EVMConfig) {
	if conf, ok := config.Config.Evm[chain]; ok {
		return true, &conf
	}

	return false, nil
}
