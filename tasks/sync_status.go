package tasks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/database"
	"github.com/KiraCore/interx/global"
	"github.com/KiraCore/interx/log"
)

func getStatus(rpcAddr string, isLog bool) {

	log.CustomLogger().Info("[getStatus] Fetching node status.")

	url := fmt.Sprintf("%s/block", rpcAddr)
	resp, err := http.Get(url)
	if err != nil {
		log.CustomLogger().Error("[getStatus] Unable to connect to node.",
			"url", url,
			"error", err,
		)
		return
	}
	defer resp.Body.Close()

	type RPCTempResponse struct {
		Jsonrpc string `json:"jsonrpc"`
		ID      int    `json:"id"`
		Result  struct {
			Block struct {
				Header struct {
					Chainid string `json:"chain_id"`
					Height  string `json:"height"`
					Time    string `json:"time"`
				} `json:"header"`
			} `json:"block"`
		} `json:"result"`
		Error interface{} `json:"error"`
	}

	result := new(RPCTempResponse)
	if json.NewDecoder(resp.Body).Decode(result) != nil {
		log.CustomLogger().Error("[getStatus] Unexpected response while decoding JSON.",
			"url", url,
			"error", err,
		)
		return
	}

	log.CustomLogger().Info("[getStatus] Successfully fetched node status.",
		"chainId", result.Result.Block.Header.Chainid,
		"blockHeight", result.Result.Block.Header.Height,
		"blockTime", result.Result.Block.Header.Time,
	)

	global.Mutex.Lock()
	common.NodeStatus.Chainid = result.Result.Block.Header.Chainid
	common.NodeStatus.Block, _ = strconv.ParseInt(result.Result.Block.Header.Height, 10, 64)
	common.NodeStatus.Blocktime = result.Result.Block.Header.Time
	global.Mutex.Unlock()

	if isLog {
		log.CustomLogger().Info("[getStatus] Detected new block.",
			"blockHeight", common.NodeStatus.Block,
			"blockTime", common.NodeStatus.Blocktime,
		)
	}

	// save block height/time
	blockTime, _ := time.Parse(time.RFC3339, result.Result.Block.Header.Time)
	common.BlockTimes[common.NodeStatus.Block] = blockTime.Unix()
	database.AddBlockTime(common.NodeStatus.Block, blockTime.Unix())
	database.AddBlockNanoTime(common.NodeStatus.Block, blockTime.UnixNano())
	common.AddNewBlock(common.NodeStatus.Block, blockTime.UnixNano())

	log.CustomLogger().Info("Finished 'getStatus' request.")
}

// SyncStatus is a function for syncing sekaid status.
func SyncStatus(rpcAddr string, isLog bool) {

	log.CustomLogger().Info("Starting 'SyncStatus' request...")

	common.LoadAllBlocks()

	for {
		getStatus(rpcAddr, isLog)

		if isLog {

			log.CustomLogger().Info("[CalcSnapshotChecksum] Syncing node status.",
				"Chain_id", common.NodeStatus.Chainid,
				"Block", common.NodeStatus.Block,
				"Block_time", common.NodeStatus.Blocktime,
			)
		}

		time.Sleep(time.Duration(config.Config.Block.StatusSync) * time.Second)

		log.CustomLogger().Info("Finished 'SyncStatus' request.")
	}
}
