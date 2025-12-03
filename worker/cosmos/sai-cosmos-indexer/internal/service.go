package internal

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	sekaiapp "github.com/KiraCore/sekai/app"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"

	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cast"

	saiService "github.com/KiraCore/sai-service/service"
	"github.com/KiraCore/sai-storage-mongo/external/adapter"
	"github.com/KiraCore/saiCosmosIndexer/internal/model"
	"github.com/KiraCore/saiCosmosIndexer/logger"
	"github.com/KiraCore/saiCosmosIndexer/utils"
)

const (
	filePathAddresses   = "./addresses.json"
	filePathLatestBlock = "./latest_handled_block"
)

type IndexData struct {
	Keys   []bson.M `bson:"keys" json:"keys"`
	Unique bool     `bson:"unique" json:"unique"`
}

type InternalService struct {
	mu            *sync.Mutex
	Context       *saiService.Context
	config        model.ServiceConfig
	handleBlocks  bool
	currentBlock  int64
	addresses     map[string]struct{}
	storageConfig model.StorageConfig
	notifier      Notifier
	client        http.Client
}

func (is *InternalService) Init() {
	config := sdk.GetConfig()
	if config.GetBech32AccountAddrPrefix() != "kira" {
		config.SetBech32PrefixForAccount("kira", "kirapub")
		config.SetBech32PrefixForValidator("kiravaloper", "kiravaloperpub")
		config.SetBech32PrefixForConsensusNode("kiravalcons", "kiravalconspub")
		config.Seal()
	}

	is.mu = &sync.Mutex{}
	is.config = model.ServiceConfig{}
	is.client = http.Client{}

	is.addresses = make(map[string]struct{})
	is.config.TxType = cast.ToString(is.Context.GetConfig("tx_type", ""))
	is.config.NodeAddress = cast.ToString(is.Context.GetConfig("node_address", ""))
	is.config.CollectionName = cast.ToString(is.Context.GetConfig("storage.mongo_collection_name", ""))
	is.config.SkipFailedTxs = cast.ToBool(is.Context.GetConfig("skip_failed_tx", false))
	is.config.HandleBlocks = cast.ToBool(is.Context.GetConfig("handle_blocks", false))
	is.storageConfig = model.StorageConfig{
		Token:      cast.ToString(is.Context.GetConfig("storage.token", "")),
		Url:        cast.ToString(is.Context.GetConfig("storage.url", "")),
		Email:      cast.ToString(is.Context.GetConfig("storage.email", "")),
		Password:   cast.ToString(is.Context.GetConfig("storage.password", "")),
		Collection: cast.ToString(is.Context.GetConfig("storage.mongo_collection_name", "")),
	}

	is.notifier = NewNotifier(
		cast.ToString(is.Context.GetConfig("notifier.sender_id", "")),
		cast.ToString(is.Context.GetConfig("notifier.email", "")),
		cast.ToString(is.Context.GetConfig("notifier.password", "")),
		cast.ToString(is.Context.GetConfig("notifier.token", "")),
		cast.ToString(is.Context.GetConfig("notifier.url", "")),
	)

	err := is.loadAddresses()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		logger.Logger.Error("loadAddresses", zap.Error(err))
	}

	fileBytes, err := os.ReadFile(filePathLatestBlock)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			logger.Logger.Error("can't read "+filePathLatestBlock, zap.Error(err))
		}
	} else {
		latestHandledBlock, err := strconv.Atoi(string(fileBytes))
		if err != nil {
			logger.Logger.Error("strconv.Atoi", zap.Error(err))
		}

		is.currentBlock = int64(latestHandledBlock)
	}

	startBlock := cast.ToInt64(is.Context.GetConfig("start_block", 0))
	if is.currentBlock < startBlock {
		is.currentBlock = startBlock
	}

	is.ProcessIndexes()
}

func (is *InternalService) ProcessIndexes() {
	const maxAttempts = 10
	const retryDelay = 5 * time.Second

	var newBlockIndexes = []IndexData{
		{
			Keys:   []bson.M{{"block_id.hash": 1}},
			Unique: false,
		},
	}

	var newTransactionIndexes = []IndexData{
		{
			Keys:   []bson.M{{"hash": 1}},
			Unique: false,
		},
		{
			Keys:   []bson.M{{"height": 1}},
			Unique: false,
		},
		{
			Keys:   []bson.M{{"messages.typeUrl": 1}},
			Unique: false,
		},
		{
			Keys:   []bson.M{{"cr_time": 1}},
			Unique: false,
		},
		{
			Keys:   []bson.M{{"tx_result.code": 1}},
			Unique: false,
		},
		{
			Keys:   []bson.M{{"messages.from_address": 1}},
			Unique: false,
		},
		{
			Keys:   []bson.M{{"messages.to_address": 1}},
			Unique: false,
		},
		{
			Keys:   []bson.M{{"messages.address": 1}},
			Unique: false,
		},
	}

	// Retry loop for block indexes
	var blockIndexes interface{}
	var err error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		blockIndexes, err = is.getIndexes(is.config.CollectionName + "_blocks")
		if err == nil {
			logger.Logger.Info("ProcessIndexes: storage connected", zap.Int("attempt", attempt))
			break
		}
		if attempt == maxAttempts {
			logger.Logger.Error("ProcessIndexes: failed to connect to storage after max attempts", zap.Error(err))
			return
		}
		logger.Logger.Warn("ProcessIndexes: storage not ready, retrying", zap.Int("attempt", attempt), zap.Error(err))
		time.Sleep(retryDelay)
	}

	if !hasRequiredIndexes(blockIndexes, newBlockIndexes) {
		indexResp, err := is.createIndexes(is.config.CollectionName+"_blocks", newBlockIndexes)
		if err != nil {
			logger.Logger.Error("createIndexes > blockIndexes", zap.Error(err))
			return
		}
		logger.Logger.Info("ProcessIndexes: block indexes created", zap.Any("response", indexResp))
	} else {
		logger.Logger.Info("ProcessIndexes: block indexes already exist")
	}

	transactionIndexes, err := is.getIndexes(is.config.CollectionName + "_txs")
	if err != nil {
		logger.Logger.Error("getIndexes > transactionIndexes", zap.Error(err))
		return
	}

	if !hasRequiredIndexes(transactionIndexes, newTransactionIndexes) {
		indexResp, err := is.createIndexes(is.config.CollectionName+"_txs", newTransactionIndexes)
		if err != nil {
			logger.Logger.Error("createIndexes > transactionIndexes", zap.Error(err))
			return
		}
		logger.Logger.Info("ProcessIndexes: transaction indexes created", zap.Any("response", indexResp))
	} else {
		logger.Logger.Info("ProcessIndexes: transaction indexes already exist")
	}
}

func hasRequiredIndexes(data interface{}, required []IndexData) bool {
	resultStr := fmt.Sprintf("%v", data)

	for _, req := range required {
		found := false
		for _, keyMap := range req.Keys {
			for k := range keyMap {
				indexName := k + "_1"
				if strings.Contains(resultStr, indexName) {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func (is *InternalService) Process() {
	sleepDuration := cast.ToDuration(is.Context.GetConfig("sleep_duration", 2))

	for {
		select {
		case <-is.Context.Context.Done():
			logger.Logger.Debug("saiCosmosIndexer loop is done")
			return
		default:
			latestBlock, err := is.getLatestBlock()
			if err != nil {
				logger.Logger.Error("getLatestBlock", zap.Error(err))
				time.Sleep(time.Second * sleepDuration)
				continue
			}

			lb, err := strconv.Atoi(latestBlock.LastHeight)
			if err != nil {
				logger.Logger.Error("getLatestBlock", zap.Error(err))
				time.Sleep(time.Second * sleepDuration)
				continue
			}

			if is.currentBlock >= int64(lb) {
				time.Sleep(time.Second * sleepDuration)
				continue
			}

			err = is.handleBlockTxs()
			if err != nil {
				logger.Logger.Error("handleBlockTxs", zap.Error(err))
				time.Sleep(time.Second * sleepDuration)
				continue
			}

			logger.Logger.Debug("handleBlockTxs processed", zap.Any("block", is.currentBlock))

			is.currentBlock += 1
		}
	}
}

func (is *InternalService) handleBlockTxs() error {
	blockInfo, err := is.getBlockInfo()
	if err != nil {
		logger.Logger.Error("handleBlockTxs", zap.Error(err))
		return err
	}

	if is.config.HandleBlocks {
		err = is.sendBlockToStorage(blockInfo)
		if err != nil {
			logger.Logger.Error("handleBlockTxs", zap.Error(err))
			return err
		}
	}

	blockTxs, err := is.getBlockTxs()
	if err != nil {
		logger.Logger.Error("handleBlockTxs", zap.Error(err))
		return err
	}

	var txArray []model.Tx
	encode := sekaiapp.MakeEncodingConfig()

	for _, txRes := range blockTxs {
		if is.config.SkipFailedTxs && txRes.TxResult.Code != 0 {
			continue
		}

		if len(txRes.TxResult.Events) < 1 {
			continue
		}

		txRes.Timestamp = blockInfo.Block.Header.Time

		txBytes, err := base64.StdEncoding.DecodeString(txRes.Tx)
		if err != nil {
			return err
		}

		tx, err := encode.TxConfig.TxDecoder()(txBytes)
		if err != nil {
			logger.Logger.Error("handleBlockTxs", zap.Error(err))
			continue
		}

		for _, msg := range tx.GetMsgs() {
			var message = model.Message{}

			msgBytes, err := json.Marshal(msg)
			if err != nil {
				logger.Logger.Error("handleBlockTxs", zap.Error(err))
				continue
			}

			err = json.Unmarshal(msgBytes, &message)
			if err != nil {
				logger.Logger.Error("handleBlockTxs", zap.Error(err))
				continue
			}

			message["typeUrl"] = sdk.MsgTypeURL(msg)

			txRes.Messages = append(txRes.Messages, message)
		}

		txArray = append(txArray, txRes)
		go is.sendTxNotification(txRes)
	}

	if len(txArray) > 0 {
		err = is.sendTxsToStorage(txArray)
		if err != nil {
			logger.Logger.Error("handleBlockTxs", zap.Error(err))
			return err
		}
	}

	err = is.rewriteLastHandledBlock(is.currentBlock)

	return err
}

func (is *InternalService) getIndexes(collection string) (interface{}, error) {
	storageRequest := adapter.Request{
		Method: "get_indexes",
		Data: adapter.GetIndexesRequest{
			Collection: collection,
		},
	}

	bodyBytes, err := jsoniter.Marshal(&storageRequest)
	if err != nil {
		return nil, err
	}

	respBytes, err := utils.SaiQuerySender(bytes.NewBuffer(bodyBytes), is.storageConfig.Url, is.storageConfig.Token)
	if err != nil {
		return nil, err
	}

	var resp interface{}
	err = json.Unmarshal(respBytes, &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (is *InternalService) createIndexes(collection string, indexes []IndexData) (interface{}, error) {
	indexesInterface := make([]interface{}, len(indexes))
	for i, idx := range indexes {
		indexesInterface[i] = idx
	}

	storageRequest := adapter.Request{
		Method: "create_indexes",
		Data: adapter.CreateIndexesRequest{
			Collection: collection,
			Data:       indexesInterface,
		},
	}

	bodyBytes, err := jsoniter.Marshal(&storageRequest)
	if err != nil {
		return nil, err
	}

	bytes, err := utils.SaiQuerySender(bytes.NewBuffer(bodyBytes), is.storageConfig.Url, is.storageConfig.Token)
	if err != nil {
		return nil, err
	}

	var resp interface{}
	err = json.Unmarshal(bytes, &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (is *InternalService) sendBlockToStorage(block *model.BlockInfo) error {
	storageRequest := adapter.Request{
		Method: "upsert",
		Data: adapter.UpsertRequest{
			Select: map[string]interface{}{
				"block_id.hash": block.BlockId.Hash,
			},
			Collection: is.storageConfig.Collection + "_blocks",
			Document:   block,
		},
	}

	bodyBytes, err := jsoniter.Marshal(&storageRequest)
	if err != nil {
		return err
	}

	_, err = utils.SaiQuerySender(bytes.NewBuffer(bodyBytes), is.storageConfig.Url, is.storageConfig.Token)

	return err
}

func (is *InternalService) sendTxsToStorage(txs []model.Tx) error {
	for _, tx := range txs {
		storageRequest := adapter.Request{
			Method: "upsert",
			Data: adapter.UpsertRequest{
				Select: map[string]interface{}{
					"hash": tx.Hash,
				},
				Collection: is.storageConfig.Collection + "_txs",
				Document:   tx,
			},
		}

		bodyBytes, err := jsoniter.Marshal(&storageRequest)
		if err != nil {
			return err
		}

		_, err = utils.SaiQuerySender(bytes.NewBuffer(bodyBytes), is.storageConfig.Url, is.storageConfig.Token)
		if err != nil {
			continue
		}
	}

	return nil
}

func (is *InternalService) sendTxNotification(tx interface{}) {
	err := is.notifier.SendTx(tx)
	if err != nil {
		//logger.Logger.Error("is.notifier.SendTx", zap.Error(err))
	}
}
