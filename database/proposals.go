package database

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/global"
	"github.com/KiraCore/interx/log"
	govTypes "github.com/KiraCore/interx/types/kira/gov"
)

// GetProposals is a function to get user proposals from cache
func GetProposals() ([]govTypes.CachedProposal, error) {
	filePath := fmt.Sprintf("%s/proposals/proposals", config.GetDbCacheDir())
	data := []govTypes.CachedProposal{}

	txs, err := ioutil.ReadFile(filePath)
	if err != nil {
		return []govTypes.CachedProposal{}, err
	}

	err = json.Unmarshal([]byte(txs), &data)

	if err != nil {
		return []govTypes.CachedProposal{}, err
	}

	return data, nil
}

// Return the last block number among the cached proposals
func GetLastBlockFetchedForProposals() int64 {
	data, err := GetProposals()

	if err != nil {
		return 0
	}

	if len(data) == 0 {
		return 0
	}

	lastTx := data[0]
	return int64(lastTx.BlockHeight)
}

// SaveProposals is a function to save user proposals to cache
func SaveProposals(propsData []govTypes.CachedProposal) error {
	cachedData, _ := GetProposals()

	// Append new txs to the cached txs array
	if len(cachedData) > 0 {
		propsData = append(propsData, cachedData...)
	}

	data, err := json.Marshal(propsData)
	if err != nil {
		return err
	}

	folderPath := fmt.Sprintf("%s/proposals", config.GetDbCacheDir())
	filePath := fmt.Sprintf("%s/proposals", folderPath)

	global.Mutex.Lock()
	err = os.MkdirAll(folderPath, os.ModePerm)
	if err != nil {
		global.Mutex.Unlock()

		log.CustomLogger().Error("[SaveProposals] Unable to create a folder",
			"folder path", folderPath,
			"error", err,
		)

		return err
	}

	err = ioutil.WriteFile(filePath, data, 0644)
	global.Mutex.Unlock()

	if err != nil {
		fmt.Println("[SaveProposals] Unable to save response", "file path", filePath,
			"error", err,
		)
	}

	return err
}
