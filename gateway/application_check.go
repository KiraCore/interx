package gateway

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	layer2types "github.com/KiraCore/sekai/x/layer2/types"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
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

func CheckApplicationState(gwCosmosmux *runtime.ServeMux, gatewayAddr string) error {
	if config.Config.AppSetting.AppMock {
		return nil
	}

	result := layer2types.QueryExecutionRegistrarResponse{}
	appstateQueryRequest, _ := http.NewRequest("GET", "http://"+gatewayAddr+"/kira/layer2/execution_registrar/"+config.Config.AppSetting.AppName, nil)

	appstateQueryResponse, failure, _ := common.ServeGRPC(appstateQueryRequest, gwCosmosmux)

	if appstateQueryResponse == nil {
		return errors.New(ToString(failure))
	}

	byteData, err := json.Marshal(appstateQueryResponse)
	if err != nil {
		return err
	}

	err = json.Unmarshal(byteData, &result)
	if err != nil {
		return err
	}

	// TODO : check if config.appmode is not same with appmode in result
	// result.Dapp.verifier
	return nil
}
