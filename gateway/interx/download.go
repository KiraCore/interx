package interx

import (
	"net/http"
	"strings"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// RegisterInterxDownloadRoutes registers download routers.
func RegisterInterxDownloadRoutes(r *mux.Router, gwCosmosmux *runtime.ServeMux, rpcAddr string) {
	r.PathPrefix(config.Download).HandlerFunc(DownloadReference()).Methods("GET")
	r.PathPrefix(config.AppDownload).HandlerFunc(DownloadAppBin()).Methods("GET")

	common.AddRPCMethod("GET", config.Download, "This is an API to download files.", true)
}

// DownloadReference is a function to download reference.
func DownloadReference() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := strings.TrimPrefix(r.URL.Path, config.Download+"/")

		common.GetLogger().Info("[download] Entering reference download: ", filename)

		if len(filename) != 0 {
			http.ServeFile(w, r, config.GetReferenceCacheDir()+"/"+filename)
		}
	}
}

// DownloadReference is a function to download reference.
func DownloadAppBin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: download bin files
		config.Config.CachingBin = true
		filename := strings.TrimPrefix(r.URL.Path, config.AppDownload+"/")

		common.GetLogger().Info("[download] Entering reference download: ", filename)

		if len(filename) != 0 {
			http.ServeFile(w, r, config.GetBinCacheDir()+"/"+filename)
		}
		config.Config.CachingBin = false
	}
}
