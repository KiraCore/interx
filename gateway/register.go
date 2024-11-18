package gateway

import (
	"github.com/KiraCore/interx/gateway/bitcoin"
	"github.com/KiraCore/interx/gateway/cosmos"
	"github.com/KiraCore/interx/gateway/evm"
	"github.com/KiraCore/interx/gateway/interx"
	"github.com/KiraCore/interx/gateway/kira"
	"github.com/KiraCore/interx/gateway/rosetta"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// RegisterRequest is a function to register requests.
func RegisterRequest(router *mux.Router, gwCosmosmux *runtime.ServeMux, rpcAddr string) {
	cosmos.RegisterRequest(router, gwCosmosmux, rpcAddr)
	kira.RegisterRequest(router, gwCosmosmux, rpcAddr)
	interx.RegisterRequest(router, gwCosmosmux, rpcAddr)
	rosetta.RegisterRequest(router, gwCosmosmux, rpcAddr)
	bitcoin.RegisterRequest(router, rpcAddr)
	evm.RegisterRequest(router, rpcAddr)
}
