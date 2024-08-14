package tasks

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	layer2types "github.com/KiraCore/sekai/x/layer2/types"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

var (
	ExecutionRegistrar layer2types.ExecutionRegistrar
)

func QueryExecutionRegistrar(gwCosmosmux *runtime.ServeMux, gatewayAddr string, rpcAddr string) error {
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

	ExecutionRegistrar = *result.ExecutionRegistrar

	return nil
}

func SyncExecutionRegistrar(gwCosmosmux *runtime.ServeMux, gatewayAddr string, rpcAddr string, isLog bool) {
	lastBlock := int64(0)
	for {
		if config.Config.AppSetting.AppMock {
			continue
		}

		if common.NodeStatus.Block != lastBlock {
			err := QueryExecutionRegistrar(gwCosmosmux, gatewayAddr, rpcAddr)

			if err != nil && isLog {
				common.GetLogger().Error("[sync-proposals] Failed to query proposals: ", err)
			}

			lastBlock = common.NodeStatus.Block
		}

		// TODO: check if appdata has interx address that is matched our app_pubkey, interx will support this app

		time.Sleep(1 * time.Second)
	}
}
