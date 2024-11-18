package tasks

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// RunTasks is a function to run threads.
func RunTasks(gwCosmosmux *runtime.ServeMux, rpcAddr string, gatewayAddr string) {
	go SyncStatus(rpcAddr, false)
	go CacheHeaderCheck(rpcAddr, false)
	go CacheDataCheck(rpcAddr, false)
	go CacheMaxSizeCheck(false)
	go DataReferenceCheck(false)
	go NodeDiscover(rpcAddr, false)
	go SyncValidators(gwCosmosmux, gatewayAddr, false)
	go SyncProposals(gwCosmosmux, gatewayAddr, rpcAddr, false)
	go CalcSnapshotChecksum(false)
	go SyncBitcoinWallets()
}
