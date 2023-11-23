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
	filePath := fmt.Sprintf("%s/transactions/%s", config.GetDbCacheDir(), address)
	if !isWithdraw {
		filePath = filePath + "-inbound"
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

	// Append new txs to the cached txs array
	if cachedData.TotalCount > 0 {
		txsData.Txs = append(txsData.Txs, cachedData.Txs...)
		txsData.TotalCount = txsData.TotalCount + cachedData.TotalCount
	}

	data, err := json.Marshal(txsData)
	if err != nil {
		return err
	}

	folderPath := fmt.Sprintf("%s/transactions", config.GetDbCacheDir())
	filePath := fmt.Sprintf("%s/%s", folderPath, address)
	if !isWithdraw {
		filePath = filePath + "-inbound"
	}

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
		fmt.Println("[cache] Unable to save response: ", filePath)
	}

	return err
}
