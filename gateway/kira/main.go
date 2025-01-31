package kira

import (
	"github.com/KiraCore/interx/gateway/kira/metamask"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// RegisterRequest is a function to register requests.
func RegisterRequest(router *mux.Router, gwCosmosmux *runtime.ServeMux, rpcAddr string) {
	RegisterKiraGovRoutes(router, gwCosmosmux, rpcAddr)
	RegisterKiraGovProposalRoutes(router, gwCosmosmux, rpcAddr)
	RegisterKiraQueryRoutes(router, gwCosmosmux, rpcAddr)
	RegisterKiraTokensRoutes(router, gwCosmosmux, rpcAddr)
	RegisterKiraUpgradeRoutes(router, gwCosmosmux, rpcAddr)
	RegisterIdentityRegistrarRoutes(router, gwCosmosmux, rpcAddr)
	RegisterKiraGovRoleRoutes(router, gwCosmosmux, rpcAddr)
	RegisterKiraGovPermissionRoutes(router, gwCosmosmux, rpcAddr)
	RegisterKiraSpendingRoutes(router, gwCosmosmux, rpcAddr)
	RegisterKiraUbiRoutes(router, gwCosmosmux, rpcAddr)
	RegisterKiraMultiStakingRoutes(router, gwCosmosmux, rpcAddr)
	metamask.RegisterKiraMetamaskRoutes(router, gwCosmosmux, rpcAddr)
}
