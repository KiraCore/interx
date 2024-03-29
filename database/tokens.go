package database

import (
	"encoding/json"
	"io/ioutil"

	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/global"
	"github.com/KiraCore/interx/types"
)

// GetTokenAliases is a function to get all token aliases
func GetTokenAliases() ([]types.TokenAlias, error) {
	filePath := config.GetDbCacheDir() + "/token-aliases.json"

	tokens := []types.TokenAlias{}

	data, err := ioutil.ReadFile(filePath)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(data), &tokens)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

// AddTokenAliases is a function to add token aliases
func AddTokenAliases(tokens []types.TokenAlias) error {
	data, err := json.Marshal(tokens)
	if err != nil {
		return err
	}

	filePath := config.GetDbCacheDir() + "/token-aliases.json"

	global.Mutex.Lock()
	err = ioutil.WriteFile(filePath, data, 0644)
	global.Mutex.Unlock()

	return err
}
