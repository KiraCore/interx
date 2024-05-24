package metamask

import (
	"net/http"

	"github.com/KiraCore/interx/common"
	interxtypes "github.com/KiraCore/interx/types"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

func GetAccountInfo(params EthGetTransactionCount, gwCosmosmux *runtime.ServeMux, r *http.Request) (uint64, uint64, error) {
	bech32Addr, err := hex2bech32(params.From, TypeKiraAddr)
	if err != nil {
		return 0, 0, err
	}

	accountNumber, sequence := common.GetAccountNumberSequence(gwCosmosmux, r, bech32Addr)

	return accountNumber, sequence, nil
}

func GetBalance(params EthGetBalanceParams, gwCosmosmux *runtime.ServeMux, r *http.Request) []interxtypes.Coin {
	bech32Addr, err := hex2bech32(params.From, TypeKiraAddr)
	if err != nil {
		return nil
	}

	balances := common.GetAccountBalances(gwCosmosmux, r.Clone(r.Context()), bech32Addr)

	return balances
}
