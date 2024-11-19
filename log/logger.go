package log

import (
	"fmt"
	"os"
	"strconv"
	"time"

	cosmosLog "cosmossdk.io/log"
)

func CustomLogger() cosmosLog.Logger {

	printLogs, err := strconv.ParseBool(os.Getenv("PrintLogs"))
	if err != nil {
		fmt.Println("[CustomLogger] Error parsing PrintLogs environment variable:", err)
	}

	if !printLogs {
		return cosmosLog.NewNopLogger()
	}

	// Initialize a new `cosmosLog` logger instance
	logger := cosmosLog.NewLogger(os.Stderr)

	logger = logger.With(
		"timestamp", time.Now().UTC().Format(time.RFC3339),
	)

	return logger
}
