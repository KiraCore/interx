package tasks

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// RunTasks is a function to run threads.
func RunTasks(gwCosmosmux *runtime.ServeMux, rpcAddr string, gatewayAddr string) {
	go SyncStatus(rpcAddr)
	go DataReferenceCheck()
	go NodeDiscover(rpcAddr)
	go SyncValidators(gwCosmosmux, gatewayAddr)
	go SyncProposals(gwCosmosmux, gatewayAddr, rpcAddr)
	go CalcSnapshotChecksum()
	go SyncBitcoinWallets()
}
