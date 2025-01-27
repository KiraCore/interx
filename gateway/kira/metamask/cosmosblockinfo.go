package metamask

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/KiraCore/interx/common"
	types "github.com/cometbft/cometbft/proto/tendermint/types"
)

const (
	BlockQueryByNumber = iota
	BLockQueryByHash
)

type PartSetHeader struct {
	Total uint32 `json:"total,omitempty"`
	Hash  []byte `json:"hash,omitempty"`
}

type BlockID struct {
	Hash          string        `json:"hash,omitempty"`
	PartSetHeader PartSetHeader `json:"part_set_header"`
}

type Consensus struct {
	Block uint64 `json:"block,string,omitempty"`
	App   uint64 `json:"app,string,omitempty"`
}

type Header struct {
	// basic block info
	Version Consensus `json:"version"`
	ChainID string    `json:"chain_id,omitempty"`
	Height  int64     `json:"height,string,omitempty"`
	Time    time.Time `json:"time"`
	// prev block info
	LastBlockId types.BlockID `json:"last_block_id"`
	// hashes of block data
	LastCommitHash string `json:"last_commit_hash,omitempty"`
	DataHash       string `json:"data_hash,omitempty"`
	// hashes from the app output from the prev block
	ValidatorsHash     string `json:"validators_hash,omitempty"`
	NextValidatorsHash string `json:"next_validators_hash,omitempty"`
	ConsensusHash      string `json:"consensus_hash,omitempty"`
	AppHash            string `json:"app_hash,omitempty"`
	LastResultsHash    string `json:"last_results_hash,omitempty"`
	// consensus info
	EvidenceHash string `json:"evidence_hash,omitempty"`
	// proposer_address is the original block proposer address, formatted as a Bech32 string.
	// In Tendermint, this type is `bytes`, but in the SDK, we convert it to a Bech32 string
	// for better UX.
	ProposerAddress string `json:"proposer_address,omitempty"`
}

type Commit struct {
	Height     int64             `json:"height,string,omitempty"`
	Round      int32             `json:"round,omitempty"`
	BlockID    BlockID           `json:"block_id"`
	Signatures []types.CommitSig `json:"signatures"`
}

type Data struct {
	// Txs that will be applied by state @ block.Height+1.
	// NOTE: not all txs here are valid.  We're just agreeing on the order first.
	// This means that block.AppHash does not include these txs.
	Txs []string `protobuf:"bytes,1,rep,name=txs,proto3" json:"txs,omitempty"`
}

type Block struct {
	Header     Header             `json:"header"`
	Data       Data               `json:"data"`
	Evidence   types.EvidenceList `json:"evidence"`
	LastCommit Commit             `json:"last_commit,omitempty"`
}

type CosmosBlockInfo struct {
	BlockId BlockID `json:"block_id,omitempty"`
	// Since: cosmos-sdk 0.47
	SdkBlock Block `json:"block,omitempty"`
}

func GetBlockNumber(rpcAddr string) (int, error) {
	sentryStatus := common.GetKiraStatus((rpcAddr))
	currentHeight, err := strconv.Atoi(sentryStatus.SyncInfo.LatestBlockHeight)
	return currentHeight, err
}

func GetBlockByNumberOrHash(blockParam string, rpcAddr string, queryType int) (CosmosBlockInfo, []string, interface{}) {
	var responseData, blockErr interface{}
	var statusCode int
	if queryType == BlockQueryByNumber {
		var blockNum int64
		var err error
		if strings.Contains(blockParam, "0x") {
			blockNum, err = hex2int64(blockParam)
		} else {
			blockNum, err = strconv.ParseInt(blockParam, 10, 64)
		}
		if err != nil {
			return CosmosBlockInfo{}, nil, err
		}

		responseData, blockErr, statusCode = queryBlockByHeight(rpcAddr, strconv.Itoa(int(blockNum)))
	} else if queryType == BLockQueryByHash {
		if !strings.Contains(blockParam, "0x") {
			blockParam = "0x" + blockParam
		}
		responseData, blockErr, statusCode = queryBlockByHash(rpcAddr, blockParam)
	}
	if blockErr != nil {
		return CosmosBlockInfo{}, nil, blockErr
	}

	if statusCode != 200 {
		return CosmosBlockInfo{}, nil, fmt.Errorf("request faield, status code - %d", statusCode)
	}

	jsonData, err := json.Marshal(responseData)
	if err != nil {
		return CosmosBlockInfo{}, nil, err
	}

	response := CosmosBlockInfo{}
	err = json.Unmarshal(jsonData, &response)
	if err != nil {
		return CosmosBlockInfo{}, nil, err
	}

	txhashes := []string{}
	txs := response.SdkBlock.Data.Txs
	for _, txStr := range txs {
		txBz, err := base64.StdEncoding.DecodeString(txStr)
		if err != nil {
			return CosmosBlockInfo{}, nil, err
		}
		converted := []byte(txBz)
		hasher := sha256.New()
		hasher.Write(converted)
		txhashes = append(txhashes, "0x"+hex.EncodeToString(hasher.Sum(nil)))
	}

	return response, txhashes, nil
}

func queryBlockByHeight(rpcAddr string, height string) (interface{}, interface{}, int) {
	success, err, statusCode := common.MakeTendermintRPCRequest(rpcAddr, "/block", fmt.Sprintf("height=%s", height))

	return success, err, statusCode
}

func queryBlockByHash(rpcAddr string, height string) (interface{}, interface{}, int) {
	success, err, statusCode := common.MakeTendermintRPCRequest(rpcAddr, "/block_by_hash", fmt.Sprintf("hash=%s", height))

	return success, err, statusCode
}
