package bitcoin

import (
	"github.com/KiraCore/interx/config"
	"github.com/gorilla/mux"
)

// RegisterRequest is a function to register requests.
func RegisterRequest(router *mux.Router, rpcAddr string) {
	RegisterBitcoinStatusRoutes(router, rpcAddr)
	RegisterBitcoinBlockRoutes(router, rpcAddr)
	RegisterBtcTransactionRoutes(router, rpcAddr)
	RegisterBtcAccountsRoutes(router, rpcAddr)
	RegisterBtcTransferRoutes(router, rpcAddr)
	RegisterBtcBalancesRoutes(router, rpcAddr)
	// RegisterBitcoinFaucetRoutes(router, rpcAddr)
}

func GetChainConfig(chain string) (bool, *config.BitcoinConfig) {
	if conf, ok := config.Config.Bitcoin[chain]; ok {
		return true, &conf
	}

	return false, nil
}
