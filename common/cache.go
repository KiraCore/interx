package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/global"
	"github.com/KiraCore/interx/log"
	"github.com/KiraCore/interx/types"
)

const MAX_CACHE_SIZE int64 = 2 * 1024 * 1024 * 1024 // Convert 2 GB to bytes

// clearFolder deletes all files in the folder.
func clearFolder(folderPath string) {
	files, _ := ioutil.ReadDir(folderPath)

	for _, file := range files {
		filePath := filepath.Join(folderPath, file.Name())
		_ = os.Remove(filePath)
	}
}

// disk usage of path/disk
func DiskUsage(path string) (disk types.DiskStatus, err error) {
	fs := syscall.Statfs_t{}
	err = syscall.Statfs(path, &fs)
	if err != nil {
		log.CustomLogger().Error("[DiskUsage] Failed to get disk usage",
			"error", err,
		)
		return types.DiskStatus{}, err
	}
	disk.All = fs.Blocks * uint64(fs.Bsize)
	disk.Free = fs.Bfree * uint64(fs.Bsize)
	disk.Used = disk.All - disk.Free
	log.CustomLogger().Info("`DiskUsage` Get disk usage",
		"disk", disk,
	)
	return disk, nil
}

func getFolderSize(folderPath string) (int64, error) {
	// Getting filesystem statistics
	disk, err := DiskUsage(folderPath)
	if err != nil {
		log.CustomLogger().Error("[getFolderSize] Failed to get folder size",
			"error", err,
		)
		return 0, err
	}
	log.CustomLogger().Info("`getFolderSize` Get folder size",
		"disk usage", int64(disk.Used),
	)
	return int64(disk.Used), nil
}

// AddCache is a function to save value to cache
func AddCache(chainIDHash string, endpointHash string, requestHash string, value types.InterxResponse) error {
	log.CustomLogger().Info("Starting 'AddCache' request...",
		"CacheBlockDuration", value.CacheBlockDuration,
		"CacheDuration", value.CacheDuration,
		"requestHash", requestHash,
		"cacheTime", value.CacheTime,
	)

	if requestHash == "" {
		return errors.New("[AddCache] One of path components is empty")
	}

	data, err := json.Marshal(value)
	if err != nil {
		log.CustomLogger().Error("[AddCache] Failed to marshal data",
			"error", err,
		)
		return err
	}

	log.CustomLogger().Info("Successfully marshaled data for the 'AddCache' function.",
		"datasize", len(data),
	)

	// Construct paths always /response
	baseDir := config.GetResponseCacheDir()
	filePath := fmt.Sprintf("%s/%s", baseDir, requestHash)

	// Log paths for debugging
	log.CustomLogger().Info("`AddCache` Created file & folde paths for storing the cache data.",
		"basedir", baseDir,
		"filepath", filePath,
	)

	// Acquire mutex
	global.Mutex.Lock()
	defer global.Mutex.Unlock() // Ensure unlock even on error

	size, err := getFolderSize(baseDir)
	if err != nil {
		log.CustomLogger().Error("[AddCache][getFolderSize] incorrect size.",
			"error", err,
		)
		return err
	}

	log.CustomLogger().Info("`AddCache` Folder size",
		"MAX_CACHE_SIZE", MAX_CACHE_SIZE,
		"size", size,
	)

	// Check if the folder size exceeds the limit
	if size > MAX_CACHE_SIZE {
		log.CustomLogger().Info("`AddCache` Folder size exceeds the limit. Clearing older cached data...")
		clearFolder(baseDir)
	}

	// Write data to the file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		log.CustomLogger().Error("[AddCache][WriteFile] Failed to write data to the cache directory.",
			"error", err,
			"filepath", filePath,
		)
		return err
	}

	log.CustomLogger().Info("`AddCache` Finished successfully. Data stored into the cache directory", "filePath", filePath)
	return nil
}

func GetCache(requestHash string) (types.InterxResponse, bool, error) {
	filePath := fmt.Sprintf("%s/%s", config.GetResponseCacheDir(), requestHash)
	log.CustomLogger().Info("Starting 'GetCache' request...", "filepath", filePath)

	response := types.InterxResponse{}

	// Attempt to read the cache file
	data, err := os.ReadFile(filePath)
	if err != nil {
		// Check if the error indicates that the file does not exist (cache miss)
		if os.IsNotExist(err) {
			log.CustomLogger().Info("[GetCache][ReadFile] Cache miss encountered. Proceeding to add query results to the cache.",
				"filePath", filePath,
			)
			// Return a custom error for cache miss
			return response, false, nil
		}

		// Log actual read errors that are not cache misses
		log.CustomLogger().Error("[GetCache][ReadFile] Failed to read data from the specified file.",
			"filePath", filePath,
			"error", err,
		)
		return response, false, err
	}

	// Verify that data is not nil after read (optional, as ReadFile should not return nil data)
	if data == nil {
		log.CustomLogger().Error("[GetCache][ReadFile] No data read from file.",
			"filePath", filePath,
		)
		return response, false, nil
	}

	// Unmarshal the data into the response structure
	err = json.Unmarshal(data, &response)
	if err != nil {
		log.CustomLogger().Error("[GetCache][Unmarshal] Failed to unmarshal cached data.",
			"error", err,
		)
		return response, false, err // Return the unmarshal error
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Println("Response directory does not exist ---filePath--->", filePath)
	}

	expired := cachedDuration(response)

	log.CustomLogger().Info("Finished 'GetCache' request. Data read successfully.")

	return response, expired, nil // Return the successfully read response
}

func cachedDuration(response types.InterxResponse) bool {
	// Current time (or the newer time)
	newerTime := time.Now()
	duration := newerTime.Sub(response.CacheTime)
	var expired bool

	// Check if the duration is less than 5 blocks
	if duration.Seconds() < 5 {
		log.CustomLogger().Info("True: 'GetCache' Duration is less than 5 blocks.")
		expired = false
	} else {
		log.CustomLogger().Info("False: 'GetCache' Duration is 5 blocks or more.")
		expired = true
	}
	return expired
}
