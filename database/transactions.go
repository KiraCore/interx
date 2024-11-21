package database

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/global"
	tmTypes "github.com/cometbft/cometbft/rpc/core/types"
)

// GetTransactions is a function to get user transactions from cache
func GetTransactions(address string, isWithdraw bool) (*tmTypes.ResultTxSearch, error) {
	var filePath string

	basePath := fmt.Sprintf("%s/transactions", config.GetDbCacheDir())
	suffix := "-inbound"
	if isWithdraw {
		suffix = ""
	}

	if address == "" {
		filePath = fmt.Sprintf("%s/all-transactions%s", basePath, suffix)
	} else {
		filePath = fmt.Sprintf("%s/%s%s", basePath, address, suffix)
	}

	data := tmTypes.ResultTxSearch{}

	txs, err := ioutil.ReadFile(filePath)
	if err != nil {
		return &tmTypes.ResultTxSearch{}, err
	}

	err = json.Unmarshal([]byte(txs), &data)

	if err != nil {
		return &tmTypes.ResultTxSearch{}, err
	}

	// Return cached inbound or outbound transactions depending on isWithdraw flag
	return &data, nil
}

// Return the last block number among the cached transactions
func GetLastBlockFetched(address string, isWithdraw bool) int64 {
	data, err := GetTransactions(address, isWithdraw)

	if err != nil {
		return 0
	}

	if len(data.Txs) == 0 {
		return 0
	}

	lastTx := data.Txs[0]
	return lastTx.Height
}

// SaveTransactions is a function to save user transactions to cache
func SaveTransactions(address string, txsData tmTypes.ResultTxSearch, isWithdraw bool) error {
	cachedData, _ := GetTransactions(address, isWithdraw)

	if address != "" {
		// Append new txs to the cached txs array
		if cachedData.TotalCount > 0 {
			txsData.Txs = append(txsData.Txs, cachedData.Txs...)
			txsData.TotalCount = txsData.TotalCount + cachedData.TotalCount
		}
	}

	data, err := json.Marshal(txsData)
	if err != nil {
		return err
	}

	folderPath := fmt.Sprintf("%s/transactions", config.GetDbCacheDir())
	fileName := resolveFileName(address, isWithdraw)
	filePath := fmt.Sprintf("%s/%s", folderPath, fileName)

	global.Mutex.Lock()
	err = os.MkdirAll(folderPath, os.ModePerm)
	if err != nil {
		global.Mutex.Unlock()

		fmt.Println("[cache] Unable to create a folder: ", folderPath)
		return err
	}

	err = ioutil.WriteFile(filePath, data, 0644)
	global.Mutex.Unlock()

	if err != nil {
		fmt.Println("[SaveTransactions][cache] Unable to save response: ", filePath)
	}

	return err
}

// Helper function to determine the file name
func resolveFileName(address string, isWithdraw bool) string {
	if address == "" {
		address = "all-transactions"
	}

	if !isWithdraw {
		return fmt.Sprintf("%s-inbound", address)
	}

	return address
}
