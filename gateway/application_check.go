package gateway

import (
	"encoding/json"

	"github.com/KiraCore/interx/common"
)

func ToString(data interface{}) string {
	out, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	return string(out)
}

func NodeSyncState(rpcAddr string) bool {
	success, _, _ := common.MakeTendermintRPCRequest(rpcAddr, "/status", "")

	if success == nil {
		return false
	}
	type TempResponse struct {
		SyncInfo struct {
			CatchingUp bool `json:"catching_up"`
		} `json:"sync_info"`
	}

	result := TempResponse{}

	byteData, err := json.Marshal(success)
	if err != nil {
		common.GetLogger().Error("[rosetta-query-networkstatus] Invalid response format", err)
		return false
	}

	err = json.Unmarshal(byteData, &result)
	if err != nil {
		common.GetLogger().Error("[rosetta-query-networkstatus] Invalid response format", err)
		return false
	}

	return !result.SyncInfo.CatchingUp
}
