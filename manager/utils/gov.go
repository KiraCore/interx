package utils

import (
	"encoding/json"

	"go.uber.org/zap"

	"github.com/saiset-co/sai-interx-manager/logger"
	"github.com/saiset-co/sai-interx-manager/types"
)

func QueryNetworkPropertiesFromGrpcResult(success interface{}) (*types.NetworkPropertiesResponse, error) {
	result := new(types.NetworkPropertiesResponse)

	byteData, err := json.Marshal(success)
	if err != nil {
		logger.Logger.Error("[query-network-properties] Invalid response format", zap.Error(err))
		return nil, err
	}
	err = json.Unmarshal(byteData, &result)
	if err != nil {
		logger.Logger.Error("[query-network-properties] Invalid response format", zap.Error(err))
		return nil, err
	}

	result.Properties.InactiveRankDecreasePercent = ConvertRate(result.Properties.InactiveRankDecreasePercent)
	result.Properties.ValidatorsFeeShare = ConvertRate(result.Properties.ValidatorsFeeShare)
	result.Properties.InflationRate = ConvertRate(result.Properties.InflationRate)
	result.Properties.MaxSlashingPercentage = ConvertRate(result.Properties.MaxSlashingPercentage)
	result.Properties.MaxAnnualInflation = ConvertRate(result.Properties.MaxAnnualInflation)
	result.Properties.DappVerifierBond = ConvertRate(result.Properties.DappVerifierBond)

	return result, nil
}
