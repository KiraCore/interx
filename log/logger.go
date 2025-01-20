package log

import (
	"os"
	"time"

	cosmosLog "cosmossdk.io/log"
)

var PrintLoges bool

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

	if !PrintLoges {
		return cosmosLog.NewNopLogger()
	}

	// Initialize a new `cosmosLog` logger instance
	logger := cosmosLog.NewLogger(os.Stderr)

	logger = logger.With(
		"timestamp", time.Now().UTC().Format(time.RFC3339),
	)

	return logger
}
