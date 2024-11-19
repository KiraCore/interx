package common

import (
	"time"

	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/database"
	"github.com/KiraCore/interx/log"
)

type BlockHeightTime struct {
	Height    int64   `json:"height"`
	BlockTime float64 `json:"timestamp"`
}

var (
	N                 int               = 0
	LatestNBlockTimes []BlockHeightTime = make([]BlockHeightTime, 0)
	BlockTimes        map[int64]int64   = make(map[int64]int64)
)

func GetAverageBlockTime() float64 {
	var total float64 = 0
	for _, block := range LatestNBlockTimes {
		total += block.BlockTime
	}

	if total == 0 {
		return 0
	}

	log.CustomLogger().Info(" `GetAverageBlockTime` Finished request.",
		"LatestNBlockTimes", LatestNBlockTimes,
	)

	return total / float64(len(LatestNBlockTimes))
}

func LoadAllBlocks() {
	blocks := database.GetAllBlocks()
	for _, rawBlock := range blocks {
		blockData := rawBlock.(map[string]interface{})

		BlockTimes[int64(blockData["height"].(float64))] = int64(blockData["timestamp"].(float64))
	}
}

func AddNewBlock(height int64, timestamp int64) {

	if len(LatestNBlockTimes) > 0 && LatestNBlockTimes[len(LatestNBlockTimes)-1].Height >= height {
		// not a new block
		log.CustomLogger().Error("[AddNewBlock] Failed to fetch new block.",
			"height", height,
		)
		return
	}

	var timespan float64 = 0
	if height != 0 {

		prevBlockTimestamp, err := GetBlockNanoTime(config.Config.RPC, height-1)
		if err != nil {
			log.CustomLogger().Error("[AddNewBlock][GetBlockNanoTime] Failed to fetch a block.",
				"height", height-1,
			)
			return
		}

		timespan = (float64(timestamp) - float64(prevBlockTimestamp)) / 1e9

		if len(LatestNBlockTimes) > 0 && timespan >= GetAverageBlockTime()*float64(config.Config.Block.HaltedAvgBlockTimes) {
			// a block just after a halt
			log.CustomLogger().Error("`AddNewBlock` Block added just after a halt.",
				"height", height,
				"timestamp", timespan,
				"average_block_time", GetAverageBlockTime(),
				"halted_threshold", GetAverageBlockTime()*float64(config.Config.Block.HaltedAvgBlockTimes),
			)
			return
		}
	}

	// insert new block
	LatestNBlockTimes = append(LatestNBlockTimes, BlockHeightTime{
		Height:    height,
		BlockTime: timespan,
	})

	if len(LatestNBlockTimes) > N {
		LatestNBlockTimes = LatestNBlockTimes[len(LatestNBlockTimes)-N:]
	}

	log.CustomLogger().Info("Finished 'AddNewBlock' request.")
}

func UpdateN(_N int) {
	if N > _N {
		LatestNBlockTimes = LatestNBlockTimes[N-_N:]
		N = _N
		return
	}

	var current = NodeStatus.Block - 1
	if len(LatestNBlockTimes) > 0 {
		current = LatestNBlockTimes[0].Height - 1
	}

	for N < _N {

		currentBlockTimestamp, err := GetBlockNanoTime(config.Config.RPC, current)
		if err != nil {
			log.CustomLogger().Error("[UpdateN][GetBlockNanoTime] Failed to fetch a block.",
				"height", current,
			)
			return
		}

		prevBlockTimestamp, err := GetBlockNanoTime(config.Config.RPC, current-1)
		if err != nil {
			log.CustomLogger().Error("[UpdateN][GetBlockNanoTime] Failed to fetch a block.",
				"height", current-1,
			)
			return
		}

		// insert new block
		LatestNBlockTimes = append(
			[]BlockHeightTime{
				{
					Height:    current,
					BlockTime: (float64(currentBlockTimestamp) - float64(prevBlockTimestamp)) / 1e9,
				},
			},
			LatestNBlockTimes...,
		)

		N++
		current = current - 1
	}
}

func IsConsensusStopped(validatorCount int) bool {
	blockHeight := NodeStatus.Block
	blockTime, _ := time.Parse(time.RFC3339, NodeStatus.Blocktime)

	if blockHeight <= 1 {
		log.CustomLogger().Error("[IsConsensusStopped] Failed to UpdateN block <= 1.",
			"height", blockHeight,
		)
		return false
	}

	var n int = int(blockHeight - 1)
	if n > validatorCount {
		n = validatorCount
	}

	UpdateN(n)

	if float64(time.Now().UTC().Unix()-blockTime.Unix()) < GetAverageBlockTime()*float64(config.Config.Block.HaltedAvgBlockTimes) {
		return false
	}

	_, err := GetBlockNanoTime(config.Config.RPC, blockHeight+1)

	return err != nil
}
