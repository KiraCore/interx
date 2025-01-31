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

func getStatus(rpcAddr string) {

	log.CustomLogger().Info("Starting `getStatus` request...",
		"rpc Addr", rpcAddr,
	)

	url := fmt.Sprintf("%s/block", rpcAddr)
	resp, err := http.Get(url)
	if err != nil {
		log.CustomLogger().Error(" [getStatus] Unable to connect to",
			"url", url,
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
		log.CustomLogger().Error(" [getStatus] Unable to connect to",
			"url", url,
		)
		return
	}

	global.Mutex.Lock()
	common.NodeStatus.Chainid = result.Result.Block.Header.Chainid
	common.NodeStatus.Block, _ = strconv.ParseInt(result.Result.Block.Header.Height, 10, 64)
	common.NodeStatus.Blocktime = result.Result.Block.Header.Time
	global.Mutex.Unlock()

	log.CustomLogger().Info("Processed `getStatus` (new block) height",
		"block", common.NodeStatus.Block,
		"time", common.NodeStatus.Blocktime,
	)

	// save block height/time
	blockTime, _ := time.Parse(time.RFC3339, result.Result.Block.Header.Time)
	common.BlockTimes[common.NodeStatus.Block] = blockTime.Unix()
	database.AddBlockTime(common.NodeStatus.Block, blockTime.Unix())
	database.AddBlockNanoTime(common.NodeStatus.Block, blockTime.UnixNano())
	common.AddNewBlock(common.NodeStatus.Block, blockTime.UnixNano())
}

// SyncStatus is a function for syncing sekaid status.
func SyncStatus(rpcAddr string) {
	common.LoadAllBlocks()
	for {
		getStatus(rpcAddr)

		log.CustomLogger().Info("Processed `getStatus` Syncing node status",
			"Chain_id", common.NodeStatus.Chainid,
			"Block", common.NodeStatus.Block,
			"Blocktime", common.NodeStatus.Blocktime,
		)

		time.Sleep(time.Duration(config.Config.Block.StatusSync) * time.Second)
	}
}
