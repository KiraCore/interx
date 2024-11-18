package common

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/global"
	"github.com/KiraCore/interx/log"
	"github.com/KiraCore/interx/types"
)

// PutCache is a function to save value to cache
func PutCache(chainIDHash string, endpointHash string, requestHash string, value types.InterxResponse) error {

	log.CustomLogger().Info("Starting 'PutCache' request...",
		"endpoint", endpointHash,
	)

	data, err := json.Marshal(value)
	if err != nil {
		log.CustomLogger().Error("[PutCache] Failed to marshal the response.",
			"error", err,
		)
		return err
	}

	folderPath := fmt.Sprintf("%s/%s/%s", config.GetResponseCacheDir(), chainIDHash, endpointHash)
	filePath := fmt.Sprintf("%s/%s", folderPath, requestHash)

	global.Mutex.Lock()
	err = os.MkdirAll(folderPath, os.ModePerm)
	if err != nil {

		log.CustomLogger().Error("[PutCache] Failed to create a folder.",
			"error", err,
			"folder_Path", folderPath,
		)

		global.Mutex.Unlock()

		return err
	}

	err = os.WriteFile(filePath, data, 0644)
	global.Mutex.Unlock()

	if err != nil {
		log.CustomLogger().Error("[PutCache] Failed to write data to the named file.",
			"error", err,
			"file_Path", filePath,
		)
	}

	log.CustomLogger().Info("Finished 'PutCache' request.")

	return err
}

// GetCache is a function to get value from cache
func GetCache(chainIDHash string, endpointHash string, requestHash string) (types.InterxResponse, error) {

	log.CustomLogger().Info("Starting 'GetCache' request...",
		"endpoint", endpointHash,
	)

	filePath := fmt.Sprintf("%s/%s/%s/%s", config.GetResponseCacheDir(), chainIDHash, endpointHash, requestHash)

	response := types.InterxResponse{}

	data, err := os.ReadFile(filePath)

	if err != nil {
		log.CustomLogger().Error("[GetCache] Failed to read data from the named file.",
			"error", err,
			"file_Path", filePath,
		)
		return response, err
	}

	err = json.Unmarshal([]byte(data), &response)

	return response, err
}
