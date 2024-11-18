package data

import (
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// RegisterRequest is a function to register requests.
func RegisterRequest(router *mux.Router, gwCosmosmux *runtime.ServeMux, rpcAddr string) {
	RegisterNetworkRoutes(router, gwCosmosmux, rpcAddr)
	RegisterAccountRoutes(router, gwCosmosmux, rpcAddr)
}
