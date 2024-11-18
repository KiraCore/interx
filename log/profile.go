package log

import (
	"net/http"
	_ "net/http/pprof"
)

// Setup HTTP server for pprof
func StartPProfServer(port string) {
	go func() {
		logger.Infof("Starting pprof server on :%s", port)
		log.Println(http.ListenAndServe(":"+port, nil))
	}()
}
