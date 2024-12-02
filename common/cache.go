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
		"chainIDHash", chainIDHash,
		"endpointHash", endpointHash,
		"requestHash", requestHash,
	)

	data, err := json.Marshal(value)
	if err != nil {
		log.CustomLogger().Error("[PutCache] Failed to marshal the response.",
			"error", err,
		)
		return err
	}

	folderPath := fmt.Sprintf("%s/%s/%s", config.GetResponseCacheDir(), chainIDHash, endpointHash)
	log.CustomLogger().Info("Create folder for saving cache result in the specific path.",
		"folder_path_for_saving_cache", folderPath,
	)

	filePath := fmt.Sprintf("%s/%s", folderPath, requestHash)
	log.CustomLogger().Info("Create file for saving cache result in the specific path.",
		"file_path_for_saving_cache", filePath,
	)

	global.Mutex.Lock()
	log.CustomLogger().Info("Mutex is locked for saving cache result.")

	err = os.MkdirAll(folderPath, os.ModePerm)
	if err != nil {

		global.Mutex.Unlock()
		log.CustomLogger().Error("[PutCache][MkdirAll] Failed to create a directory or lack of permissions.",
			"error", err,
			"folder_Path", folderPath,
			"lock_status", "unlocked",
		)

		return err
	}

	err = os.WriteFile(filePath, data, 0644)
	global.Mutex.Unlock()

	if err != nil {
		log.CustomLogger().Error("[PutCache][WriteFile] Failed to writes data to a specified file.",
			"error", err,
			"file_Path", filePath,
			"lock_status", "unlocked",
		)
	}

	log.CustomLogger().Info("Finished 'PutCache' request. File written successfully.")

	return err
}

// GetCache is a function to get value from cache
func GetCache(chainIDHash string, endpointHash string, requestHash string) (types.InterxResponse, error) {

	log.CustomLogger().Info("Starting 'GetCache' request...",
		"chainID_hash", chainIDHash,
		"endpoint_hash", endpointHash,
		"request_hash", requestHash,
	)

	filePath := fmt.Sprintf("%s/%s/%s/%s", config.GetResponseCacheDir(), chainIDHash, endpointHash, requestHash)
	log.CustomLogger().Info("`GetCache` Create file path to fetch the cached data from it.",
		"file_path", filePath,
	)

	response := types.InterxResponse{}

	data, err := os.ReadFile(filePath)
	if err != nil {
		log.CustomLogger().Error("[GetCache][ReadFile] Failed to read data from the specified file.",
			"error", err,
			"file_Path", filePath,
		)
		return response, err
	}

	err = json.Unmarshal([]byte(data), &response)

	log.CustomLogger().Info("Finished 'GetCache' request. Data read successfully done.")

	return response, err
}
