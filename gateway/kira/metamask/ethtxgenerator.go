package metamask

import (
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// LogsBloom is a struct that represents a logsBloom field
type LogsBloom struct {
	m    int    // size of the bit array
	bits []bool // bit array
}

// NewLogsBloom creates a new logsBloom field with the given size
func NewLogsBloom(m int) *LogsBloom {
	return &LogsBloom{
		m:    m,
		bits: make([]bool, m),
	}
}

// hash is a helper function that hashes a byte slice using keccak256 algorithm
func hash(b []byte) []byte {
	return crypto.Keccak256(b)
}

// Add adds an element to the logsBloom field
func (lb *LogsBloom) Add(b []byte) {
	h := hash(b)
	for i := 0; i < 3; i++ {
		// extract 11-bit segments from the first, third, and fifth 16-bit words of the hash
		// https://github.com/ethereum/wiki/wiki/Design-Rationale#bloom-filter
		seg := new(big.Int).SetBytes(h[i*2 : i*2+2])
		seg.Rsh(seg, uint(4-i))
		seg.And(seg, big.NewInt(2047))
		lb.bits[seg.Int64()] = true
	}
}

// Contains checks if an element is in the logsBloom field
func (lb *LogsBloom) Contains(b []byte) bool {
	h := hash(b)
	for i := 0; i < 3; i++ {
		seg := new(big.Int).SetBytes(h[i*2 : i*2+2])
		seg.Rsh(seg, uint(4-i))
		seg.And(seg, big.NewInt(2047))
		if !lb.bits[seg.Int64()] {
			return false
		}
	}
	return true
}

// Or performs a bitwise OR operation with another logsBloom field
func (lb *LogsBloom) Or(other *LogsBloom) {
	for i := 0; i < lb.m; i++ {
		lb.bits[i] = lb.bits[i] || other.bits[i]
	}
}

// CreateLogsBloomFromLogs creates a logsBloom field from a slice of ethereum tx logs
func CreateLogsBloomFromLogs(logs []*types.Log, m int) *LogsBloom {
	lb := NewLogsBloom(m)
	for _, log := range logs {
		// add the topics and data to the logsBloom field
		for _, topic := range log.Topics {
			lb.Add(topic.Bytes())
		}
		lb.Add(log.Data)
	}
	return lb
}

func boolsToBytes(bools []bool) []byte {
	bi := new(big.Int) // create a new big.Int
	for i, b := range bools {
		if b {
			bi.SetBit(bi, i, 1) // set the ith bit to 1
		}
	}
	return bi.Bytes() // return the byte representation
}

// func convertCosmosLog2EvmLog(logInfos []LogInfo) []types.Log {

// 	var logs []types.Log

// 	types.LegacyTx

// 	for _, logInfo := range logInfos {
// 		log := &types.Log{
// 			Address
// 			Topics: []common.Hash{
// 				common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
// 				common.HexToHash("0x000000000000000000000000" + bech322hex(logInfo.Events[0].Attributes[1])),
// 				common.HexToHash("0x000000000000000000000000"),
// 			},
// 			Data: common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000000"),
// 		}

// 		logs = append(logs, *log)
// 	}
// }

func GetLogsBloom(logs []*types.Log) []byte {
	// create a logsBloom field from the log with 2048 bits
	lb := CreateLogsBloomFromLogs(logs, 2048)

	return boolsToBytes(lb.bits)
}
