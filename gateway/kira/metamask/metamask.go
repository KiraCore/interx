package metamask

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type EthMetamaskRequestBody struct {
	ID      interface{}   `json:"id"`
	Method  string        `json:"method"`
	JsonRPC string        `json:"jsonrpc"`
	Params  []interface{} `json:"params"`
	// Add other fields from the request body if needed
}

type EthMetamaskResponseBody struct {
	ID      interface{} `json:"id"`
	JsonRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	// Add other fields from the request body if needed
}

type EthEstimateGas struct {
	From     string `json:"from"`
	Value    string `json:"value"`
	GasPrice string `json:"gasPrice"`
	Data     string `json:"data"`
	To       string `json:"to"`
}

// balance of the account of given address
type EthGetBalanceParams struct {
	From string
	// integer block number, or the string "latest", "earliest" or "pending"

	// defaultBlock parameter:
	// HEX String - an integer block number
	// String "earliest" for the earliest/genesis block
	// String "latest" - for the latest mined block
	// String "safe" - for the latest safe head block
	// String "finalized" - for the latest finalized block
	// String "pending" - for the pending state/transactions
	BlockNum string
}

type EthGetBlockNumber struct {
	BlockNum string
	// If true it returns the full transaction objects, if false only the hashes of the transactions
	Filter bool
}

// number of transactions sent from an address.
type EthGetTransactionCount struct {
	From string
	// integer block number, or the string "latest", "earliest" or "pending"
	BlockNum string
}

func RegisterKiraMetamaskRoutes(r *mux.Router, gwCosmosmux *runtime.ServeMux, rpcAddr string) {
	r.HandleFunc(config.MetamaskEndpoint, MetamaskRequestHandler(gwCosmosmux, rpcAddr)).Methods("POST")
	// r.HandleFunc(config.RegisterAddrEndpoint, AddressRegisterHandler(gwCosmosmux, rpcAddr)).Methods("Get")

	common.AddRPCMethod("POST", config.MetamaskEndpoint, "This is an API to interact with metamask.", true)
	// common.AddRPCMethod("POST", config.MetamaskEndpoint, "This is an API to interact with metamask.", true)
}

func MetamaskRequestHandler(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		byteData, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Unmarshal the request body into a struct
		var requestBody EthMetamaskRequestBody
		err = json.Unmarshal(byteData, &requestBody)
		if err != nil {
			fmt.Println("unmarshal err:", err, byteData)
			http.Error(w, "Failed to parse request body", http.StatusBadRequest)
			return
		}

		fmt.Println("eth method:", requestBody.Method)
		// fmt.Println("eth params:", requestBody.Params)

		response := map[string]interface{}{
			"id":      requestBody.ID,
			"jsonrpc": requestBody.JsonRPC,
		}

		var result interface{}

		switch requestBody.Method {
		case "eth_chainId":
			result = fmt.Sprintf("0x%x", config.DefaultChainID)
		case "eth_getBalance":
			params := EthGetBalanceParams{
				From:     requestBody.Params[0].(string),
				BlockNum: requestBody.Params[1].(string),
			}

			balances := GetBalance(params, gwCosmosmux, r)

			result = "0x0"
			if len(balances) > 0 {
				for _, coin := range balances {
					if coin.Denom == config.DefaultKiraDenom {
						balance := new(big.Int)
						balance.SetString(coin.Amount, 10)
						balance = balance.Mul(balance, big.NewInt(int64(math.Pow10(12))))
						result = fmt.Sprintf("0x%x", balance)
					}
				}
			}
		case "eth_blockNumber":
			currentHeight, err := GetBlockNumber(rpcAddr)

			if err != nil {
				fmt.Println("eth_getBlockByNumber err:", err)
				return
			}

			result = fmt.Sprintf("0x%x", currentHeight)
		case "net_version":
			result = config.DefaultChainID
		case "eth_getBlockByNumber":
			blockNum := requestBody.Params[0].(string)
			response, txs, err := GetBlockByNumberOrHash(blockNum, rpcAddr, BlockQueryByNumber)
			if err != nil {
				fmt.Println("eth_getBlockByNumber err:", err)
				return
			}

			result = map[string]interface{}{
				"extraData": "0x737061726b706f6f6c2d636e2d6e6f64652d3132",
				"gasLimit":  "0x1c9c380",
				"gasUsed":   "0x79ccd3",
				"hash":      "0x" + response.BlockId.Hash,
				"logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
				"miner":     "0x" + response.SdkBlock.Header.ProposerAddress,
				// "mixHash": "0x3d1fdd16f15aeab72e7db1013b9f034ee33641d92f71c0736beab4e67d34c7a7",
				// "nonce":            "0x4db7a1c01d8a8072",
				"number":           "0x" + fmt.Sprintf("%x", response.SdkBlock.Header.Height),
				"parentHash":       "0x" + hex.EncodeToString(response.SdkBlock.Header.LastBlockId.Hash),
				"receiptsRoot":     "0x" + response.SdkBlock.Header.DataHash,
				"size":             "0x41c7",
				"stateRoot":        "0xddc8b0234c2e0cad087c8b389aa7ef01f7d79b2570bccb77ce48648aa61c904d",
				"timestamp":        fmt.Sprintf("0x%x", response.SdkBlock.Header.Time.UnixMilli()),
				"totalDifficulty":  "0x12ac11391a2f3872fcd",
				"transactions":     txs,
				"transactionsRoot": "0x" + response.SdkBlock.Header.DataHash,
				"sha3Uncles":       "0x0000000000000000000000000000000000000000000000000000000000000000",
				"uncles":           []string{},
			}
		case "eth_getBlockByHash":
			blockHash := requestBody.Params[0].(string)

			response, txs, err := GetBlockByNumberOrHash(blockHash, rpcAddr, BLockQueryByHash)
			if err != nil {
				fmt.Println("eth_getBlockByNumber err:", err)
				return
			}

			result = map[string]interface{}{
				"extraData": "0x737061726b706f6f6c2d636e2d6e6f64652d3132",
				"gasLimit":  "0x1c9c380",
				"gasUsed":   "0x79ccd3",
				"hash":      "0x" + response.BlockId.Hash,
				"logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
				"miner":     "0x" + response.SdkBlock.Header.ProposerAddress,
				// "mixHash": "0x3d1fdd16f15aeab72e7db1013b9f034ee33641d92f71c0736beab4e67d34c7a7",
				// "nonce":            "0x4db7a1c01d8a8072",
				"number":           "0x" + fmt.Sprintf("%x", response.SdkBlock.Header.Height),
				"parentHash":       "0x" + hex.EncodeToString(response.SdkBlock.Header.LastBlockId.Hash),
				"receiptsRoot":     "0x" + response.SdkBlock.Header.DataHash,
				"size":             "0x41c7",
				"stateRoot":        "0xddc8b0234c2e0cad087c8b389aa7ef01f7d79b2570bccb77ce48648aa61c904d",
				"timestamp":        fmt.Sprintf("0x%x", response.SdkBlock.Header.Time.UnixMilli()),
				"totalDifficulty":  "0x12ac11391a2f3872fcd",
				"transactions":     txs,
				"transactionsRoot": "0x" + response.SdkBlock.Header.DataHash,
				"sha3Uncles":       "0x0000000000000000000000000000000000000000000000000000000000000000",
				"uncles":           []string{},
			}

			file, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println("getblockbyhash:", string(file))
		case "eth_estimateGas":
			var params EthEstimateGas
			byteData, err := json.Marshal(requestBody.Params[0])
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			err = json.Unmarshal(byteData, &params)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			result = "0x5cec"
		case "eth_gasPrice":
			result = "0x64"
		case "eth_getCode":
			result = ""
		case "eth_getTransactionCount":
			params := EthGetTransactionCount{
				From:     requestBody.Params[0].(string),
				BlockNum: requestBody.Params[1].(string),
			}

			_, sequence, err := GetAccountInfo(params, gwCosmosmux, r)
			if err != nil {
				fmt.Println("eth_getTransactionCount err:", err)
				return
			}

			result = fmt.Sprintf("0x%x", sequence)
		case "eth_sendTransaction":
			// var params EthSendTransaction
			// byteData, err := json.Marshal(requestBody.Params[0])
			// if err != nil {
			// 	fmt.Println("Error:", err)
			// 	return
			// }
			// err = json.Unmarshal(byteData, &params)
			// if err != nil {
			// 	fmt.Println("Error:", err)
			// 	return
			// }
			// fmt.Println("eth_sendTransaction params:", params)
		case "eth_sendRawTransaction":
			fmt.Println("params:", requestBody.Params)
			strData, ok := requestBody.Params[0].(string)
			if !ok {
				fmt.Println("eth_sendRawTransaction err: convert from interface{} to string failed")
				return
			}

			txHash, err := SendTx(strData, gwCosmosmux, r)
			if err != nil {
				fmt.Println("eth_sendRawTransaction err:", err)
				return
			}

			result = "0x" + txHash

			fmt.Println("sendrawtx:", result)
		case "eth_getTransactionReceipt":
			txHash := requestBody.Params[0].(string)
			fmt.Println("eth_getTransactionReceipt", txHash)
			txInfo, logInfos, err := GetTxInfo(txHash, rpcAddr)
			var txStatus, txIndex int
			blockInfo := CosmosBlockInfo{}
			var fromAddr, toAddr string

			if err != nil {
				txStatus = 0
			} else {
				var txhashes []string
				blockInfo, txhashes, err = GetBlockByNumberOrHash(txInfo.Height, rpcAddr, BlockQueryByNumber)
				if err != nil {
					fmt.Println("eth_getTransactionReceipt err:", err)
					return
				}

				if logInfos.logForString != "" {
					fromAddr, err = bech322hex(logInfos.logForMap[0].LogInfo[0]["transfer"]["sender"])
					if err != nil {
						fmt.Println("eth_getTransactionReceipt err:", err)
						return
					}
					toAddr, err = bech322hex(logInfos.logForMap[0].LogInfo[0]["transfer"]["recipient"])
					if err != nil {
						fmt.Println("eth_getTransactionReceipt err:", err)
						return
					}
				}

				txIndex = 0
				for i, blockTxHash := range txhashes {
					if blockTxHash == txHash {
						txIndex = i
					}
				}

				txStatus = 1
			}

			result = map[string]interface{}{
				"blockHash":         "0x" + blockInfo.BlockId.Hash,
				"blockNumber":       fmt.Sprintf("0x%x", blockInfo.SdkBlock.Header.Height),
				"contractAddress":   nil, // string of the address if it was created
				"cumulativeGasUsed": "0x79ccd3",
				"effectiveGasPrice": "0x64",
				"from":              "0x" + fromAddr,
				"gasUsed":           "0x5208",
				"logs":              []string{},
				"logsBloom":         "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
				"status":            fmt.Sprintf("0x%x", txStatus),
				"to":                "0x" + toAddr,
				"transactionHash":   "0x" + txInfo.Hash,
				"transactionIndex":  fmt.Sprintf("0x%x", txIndex),
				"type":              "0x2",
			}

			file, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println("gettxreceipt:", string(file))
		default:
			fmt.Println("unknown method:", requestBody.Method)
		}

		response["result"] = result

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
