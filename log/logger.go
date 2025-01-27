package log

import (
	"os"
	"sync"
	"time"

	cosmosLog "cosmossdk.io/log"
)

var (
	PrintLoges bool
	once       sync.Once
	logger     cosmosLog.Logger
)

// InitializeLogger sets up the logging behavior based on the value of printLogs.
func InitializeLogger(printLogs bool) error {
	// Handle logging behavior based on the value of printLogs
	if printLogs {
		PrintLoges = true
	} else {
		PrintLoges = false
	}

	return nil
}

func CustomLogger() cosmosLog.Logger {
	once.Do(func() {
		if PrintLoges {
			// Create and initialize the logger only once
			logger = cosmosLog.NewLogger(os.Stderr).With(
				"timestamp", time.Now().UTC().Format(time.RFC3339),
			)
		} else {
			// Use a no-op logger if logging is disabled
			logger = cosmosLog.NewNopLogger()
		}
	})
	return logger
}
