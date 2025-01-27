package metamask

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/KiraCore/interx/common"
)

// Attribute is a struct that represents an attribute in the JSON data
type Attribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Event is a struct that represents an event in the JSON data
type Event struct {
	Type       string      `json:"type"`
	Attributes []Attribute `json:"attributes"`
}

// Msg is a struct that represents a message in the JSON data
type LogInfo struct {
	MsgIndex int     `json:"msg_index"`
	Events   []Event `json:"events"`
}

// map[msgIndex][event_type][attribute_key]
type LogInfoForMap struct {
	LogInfo map[int]map[string]map[string]string
}

type Log struct {
	logForMap    map[int]LogInfoForMap
	logForString string
}

type TxInfo struct {
	Hash     string `json:"hash"`
	Height   string `json:"height"`
	Index    int    `json:"index"`
	TxResult struct {
		Code      int    `json:"code"`
		Data      string `json:"data"`
		Log       string `json:"log"`
		Info      string `json:"info"`
		GasWanted string `json:"gas_wanted"`
		GasUsed   string `json:"gas_used"`
		Events    []struct {
			Type       string `json:"type"`
			Attributes []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
				Index bool   `json:"index"`
			} `json:"attributes"`
		} `json:"events"`
		Codespace string `json:"codespace"`
	} `json:"tx_result"`
	Tx string `json:"tx"`
}

func GetTxInfo(txHash string, rpcAddr string) (TxInfo, Log, interface{}) {
	responseData, err, statusCode := queryTxByHash(rpcAddr, txHash)
	logResult := Log{}

	if err != nil {
		return TxInfo{}, logResult, err
	}

	if statusCode != 200 {
		return TxInfo{}, logResult, fmt.Errorf("request faield, status code - %d", statusCode)
	}

	jsonData, err := json.Marshal(responseData)
	if err != nil {
		return TxInfo{}, logResult, err
	}

	response := TxInfo{}
	err = json.Unmarshal(jsonData, &response)
	if err != nil {
		return TxInfo{}, logResult, err
	}

	var logInfos []LogInfo

	// Unmarshal the JSON data to the msgs slice
	err = json.Unmarshal([]byte(response.TxResult.Log), &logInfos)
	logResult.logForString = response.TxResult.Log
	if err != nil {
		return response, logResult, nil
	}

	logInfosForMap := map[int]LogInfoForMap{}
	for i, logInfo := range logInfos {
		logInfosForMap[i] = LogInfoForMap{LogInfo: make(map[int]map[string]map[string]string)}
		for _, event := range logInfo.Events {
			if logInfosForMap[i].LogInfo[logInfo.MsgIndex] == nil {
				logInfosForMap[i].LogInfo[logInfo.MsgIndex] = make(map[string]map[string]string)
			}
			if logInfosForMap[i].LogInfo[logInfo.MsgIndex][event.Type] == nil {
				logInfosForMap[i].LogInfo[logInfo.MsgIndex][event.Type] = make(map[string]string)
			}
			for _, attribute := range event.Attributes {
				logInfosForMap[i].LogInfo[logInfo.MsgIndex][event.Type][attribute.Key] = attribute.Value
			}
		}
	}

	logResult.logForMap = logInfosForMap
	return response, logResult, nil
}

func queryTxByHash(rpcAddr string, hash string) (interface{}, interface{}, int) {
	if !strings.HasPrefix(hash, "0x") {
		hash = "0x" + hash
	}
	return common.MakeTendermintRPCRequest(rpcAddr, "/tx", fmt.Sprintf("hash=%s", hash))
}
