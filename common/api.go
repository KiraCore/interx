package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"

	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/database"
	"github.com/KiraCore/interx/global"
	"github.com/KiraCore/interx/log"
	"github.com/KiraCore/interx/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	tmTypes "github.com/cometbft/cometbft/rpc/core/types"
	tmJsonRPCTypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// MakeTendermintRPCRequest is a function to make GET request
func MakeTendermintRPCRequest(rpcAddr string, url string, query string) (interface{}, interface{}, int) {
	endpoint := fmt.Sprintf("%s%s?%s", rpcAddr, url, query)

	log.CustomLogger().Info("Starting 'MakeTendermintRPCRequest' request...",
		"endpoint", endpoint,
	)

	resp, err := http.Get(endpoint)
	if err != nil {
		log.CustomLogger().Error("[MakeTendermintRPCRequest] Failed to connect to endpoint.",
			"endpoint", endpoint,
			"error", err,
		)
		return ServeError(0, "", err.Error(), http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	response := new(types.RPCResponse)
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		log.CustomLogger().Error("[MakeTendermintRPCRequest] Failed to unmarshal the response.",
			"response", resp.StatusCode,
			"error", err,
		)
		return nil, err.Error(), resp.StatusCode
	}

	log.CustomLogger().Info("Finished 'MakeTendermintRPCRequest' request.")

	return response.Result, response.Error, resp.StatusCode
}

// MakeGetRequest is a function to make GET request
func MakeGetRequest(rpcAddr string, url string, query string) (Result interface{}, Error interface{}, StatusCode int) {
	endpoint := fmt.Sprintf("%s%s?%s", rpcAddr, url, query)

	log.CustomLogger().Info("Starting 'MakeGetRequest' request...",
		"endpoint", endpoint,
	)

	resp, err := http.Get(endpoint)
	if err != nil {
		log.CustomLogger().Error("[MakeGetRequest] Failed to connect to endpoint.",
			"endpoint", endpoint,
			"error", err,
		)
		return ServeError(0, "", err.Error(), http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	StatusCode = resp.StatusCode

	err = json.NewDecoder(resp.Body).Decode(&Result)
	if err != nil {
		log.CustomLogger().Error("[MakeGetRequest] Failed to unmarshal the response.",
			"response", resp.StatusCode,
			"error", err,
		)
		Error = err.Error()
	}

	log.CustomLogger().Info("Finished 'MakeGetRequest' request.")

	return Result, Error, StatusCode
}

// DownloadResponseToFile is a function to save GET response as a file
func DownloadResponseToFile(rpcAddr string, url string, query string, filepath string) error {
	endpoint := fmt.Sprintf("%s%s?%s", rpcAddr, url, query)

	log.CustomLogger().Info("Starting 'DownloadResponseToFile' request...",
		"endpoint", endpoint,
	)

	resp, err := http.Get(endpoint)
	if err != nil {
		log.CustomLogger().Error("[DownloadResponseToFile] Failed to connect to endpoint.",
			"endpoint", endpoint,
			"error", err,
		)
		return err
	}
	defer resp.Body.Close()

	fileout, _ := os.Create(filepath)
	defer fileout.Close()

	global.Mutex.Lock()
	_, err = io.Copy(fileout, resp.Body)
	if err != nil {
		log.CustomLogger().Error("[DownloadResponseToFile] Failed to save the response.",
			"response", resp.StatusCode,
			"error", err,
		)
	}

	global.Mutex.Unlock()

	log.CustomLogger().Info("Finished 'DownloadResponseToFile' request.")

	return err
}

// GetAccountBalances is a function to get balances of an address
func GetAccountBalances(gwCosmosmux *runtime.ServeMux, r *http.Request, bech32addr string) []types.Coin {

	log.CustomLogger().Info("Starting 'GetAccountBalances' request...")

	_, err := sdk.AccAddressFromBech32(bech32addr)
	if err != nil {
		log.CustomLogger().Error("[GetAccountBalances] Invalid bech32addr.",
			"bech32addr", bech32addr,
			"error", err,
		)
		return nil
	}

	r.URL.Path = fmt.Sprintf("/cosmos/bank/v1beta1/balances/%s", bech32addr)
	r.URL.RawQuery = ""
	r.Method = "GET"

	recorder := httptest.NewRecorder()
	gwCosmosmux.ServeHTTP(recorder, r)
	resp := recorder.Result()

	type BalancesResponse struct {
		Balances []types.Coin `json:"balances"`
	}

	result := BalancesResponse{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.CustomLogger().Error("[GetAccountBalances] Failed to decode the response.",
			"response", resp.StatusCode,
			"error", err,
		)
	}

	log.CustomLogger().Info("Finished 'GetAccountBalances' request.")

	return result.Balances
}

// GetAccountNumberSequence is a function to get AccountNumber and Sequence
func GetAccountNumberSequence(gwCosmosmux *runtime.ServeMux, r *http.Request, bech32addr string) (uint64, uint64) {

	log.CustomLogger().Info("Starting 'GetAccountNumberSequence' request...")

	_, err := sdk.AccAddressFromBech32(bech32addr)
	if err != nil {
		log.CustomLogger().Error("[GetAccountNumberSequence] Invalid bech32addr.",
			"bech32addr", bech32addr,
			"error", err,
		)
		return 0, 0
	}

	r.URL.Path = fmt.Sprintf("/cosmos/auth/v1beta1/accounts/%s", bech32addr)
	r.URL.RawQuery = ""
	r.Method = "GET"

	// GetLogger().Info("[grpc-call] Entering grpc call: ", r.URL.Path)

	recorder := httptest.NewRecorder()
	gwCosmosmux.ServeHTTP(recorder, r)
	resp := recorder.Result()

	type QueryAccountResponse struct {
		Account struct {
			Address       string `json:"addresss"`
			PubKey        string `json:"pubKey"`
			AccountNumber string `json:"accountNumber"`
			Sequence      string `json:"sequence"`
		} `json:"account"`
	}
	result := QueryAccountResponse{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.CustomLogger().Error("[GetAccountNumberSequence] Failed to decode the response.",
			"response", resp.StatusCode,
			"error", err,
		)
	}

	accountNumber, _ := strconv.ParseInt(result.Account.AccountNumber, 10, 64)
	sequence, _ := strconv.ParseInt(result.Account.Sequence, 10, 64)

	log.CustomLogger().Info("Finished 'GetAccountNumberSequence' request.")

	return uint64(accountNumber), uint64(sequence)
}

// BroadcastTransaction is a function to post transaction, returns txHash
func BroadcastTransaction(rpcAddr string, txBytes []byte) (string, error) {
	endpoint := fmt.Sprintf("%s/broadcast_tx_async?tx=0x%X", rpcAddr, txBytes)

	log.CustomLogger().Info("Starting 'BroadcastTransaction' request...")

	resp, err := http.Get(endpoint)
	if err != nil {
		log.CustomLogger().Error("[BroadcastTransaction] Failed to connect to endpoint.",
			"endpoint", endpoint,
			"error", err,
		)
		return "", err
	}
	defer resp.Body.Close()

	type RPCTempResponse struct {
		Jsonrpc string `json:"jsonrpc"`
		ID      int    `json:"id"`
		Result  struct {
			Height string `json:"height"`
			Hash   string `json:"hash"`
		} `json:"result,omitempty"`
		Error struct {
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}

	result := new(RPCTempResponse)
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		log.CustomLogger().Error("[BroadcastTransaction] Failed to unmarshal the response.",
			"response", resp.StatusCode,
			"error", err,
		)
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		log.CustomLogger().Error("[BroadcastTransaction] Failed to successful response.",
			"resp_StatusCode", resp.StatusCode,
			"http_StatusOK", http.StatusOK,
		)
		return "", errors.New(result.Error.Message)
	}

	log.CustomLogger().Info("Finished 'BroadcastTransaction' request.")

	return result.Result.Hash, nil
}

// GetPermittedTxTypes is a function to get all permitted tx types and function ids
func GetPermittedTxTypes(rpcAddr string, account string) (map[string]string, error) {
	permittedTxTypes := map[string]string{}
	permittedTxTypes["ExampleTx"] = "123"
	return permittedTxTypes, nil
}

// GetBlockTime is a function to get block time
func GetBlockTime(rpcAddr string, height int64) (int64, error) {

	log.CustomLogger().Info("Starting 'GetBlockTime' request...")

	blockTime, found := BlockTimes[height]
	if found {
		return blockTime, nil
	}

	endpoint := fmt.Sprintf("%s/block?height=%d", rpcAddr, height)
	// GetLogger().Info("[rpc-call] Entering rpc call: ", endpoint)

	resp, err := http.Get(endpoint)
	if err != nil {
		log.CustomLogger().Error("[GetBlockTime] Failed to connect to endpoint.",
			"endpoint", endpoint,
			"error", err,
		)
		return 0, fmt.Errorf("block not found: %d", height)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	response := new(tmJsonRPCTypes.RPCResponse)

	if err := json.Unmarshal(respBody, response); err != nil {
		log.CustomLogger().Error("[GetBlockTime] Failed to unmarshal the response.",
			"response", resp.StatusCode,
			"error", err,
		)
		return 0, err
	}

	if response.Error != nil {
		log.CustomLogger().Error("[GetBlockTime] Failed to find a valid block.",
			"height", height,
			"error", response.Error,
		)
		return 0, fmt.Errorf("block not found: %d", height)
	}

	result := new(tmTypes.ResultBlock)
	if err := tmjson.Unmarshal(response.Result, result); err != nil {
		log.CustomLogger().Error("[GetBlockTime] Failed to unmarshal the response.",
			"response", resp.StatusCode,
			"error", err,
		)
		return 0, err
	}

	blockTime = result.Block.Header.Time.Unix()

	// save block time
	BlockTimes[NodeStatus.Block] = blockTime
	database.AddBlockTime(height, blockTime)

	log.CustomLogger().Info("Block time successfully saved in data base.",
		"Block_Time", blockTime,
		"height", height,
	)

	// save block nano time
	database.AddBlockNanoTime(height, result.Block.Header.Time.UnixNano())

	log.CustomLogger().Info("Block nano time successfully saved in data base.",
		"Block_Nano_Time", result.Block.Header.Time.UnixNano(),
		"height", height,
	)

	log.CustomLogger().Info("Finished 'GetBlockTime' request.")

	return blockTime, nil
}

// GetBlockNanoTime is a function to get block nano time
func GetBlockNanoTime(rpcAddr string, height int64) (int64, error) {

	log.CustomLogger().Info("Starting 'GetBlockNanoTime' request...")

	blockTime, err := database.GetBlockNanoTime(height)
	if err == nil {
		return blockTime, nil
	}

	endpoint := fmt.Sprintf("%s/block?height=%d", rpcAddr, height)
	// GetLogger().Info("[rpc-call] Entering rpc call: ", endpoint)

	resp, err := http.Get(endpoint)
	if err != nil {
		log.CustomLogger().Error("[GetBlockNanoTime] Failed to connect to endpoint.",
			"endpoint", endpoint,
			"error", err,
		)
		return 0, fmt.Errorf("block not found: %d", height)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	response := new(tmJsonRPCTypes.RPCResponse)

	if err := json.Unmarshal(respBody, response); err != nil {
		log.CustomLogger().Error("[GetBlockNanoTime] Failed to unmarshal the response.",
			"response", resp.StatusCode,
			"error", err,
		)
		return 0, err
	}

	if response.Error != nil {
		log.CustomLogger().Error("[GetBlockNanoTime] Failed to find a valid block.",
			"height", height,
			"error", response.Error,
		)
		return 0, fmt.Errorf("block not found: %d", height)
	}

	result := new(tmTypes.ResultBlock)
	if err := tmjson.Unmarshal(response.Result, result); err != nil {
		log.CustomLogger().Error("[GetBlockNanoTime] Failed to unmarshal the response.",
			"response", resp.StatusCode,
			"error", err,
		)
		return 0, err
	}

	blockTime = result.Block.Header.Time.UnixNano()

	// save block time
	BlockTimes[height] = result.Block.Header.Time.Unix()
	database.AddBlockTime(height, result.Block.Header.Time.Unix())

	log.CustomLogger().Info("Block time successfully saved in data base.",
		"Block_Time", blockTime,
		"height", height,
	)

	// save block nano time
	database.AddBlockNanoTime(height, blockTime)

	log.CustomLogger().Info("Block nano time successfully saved in data base.",
		"Block_Nano_Time", result.Block.Header.Time.UnixNano(),
		"height", height,
	)

	log.CustomLogger().Info("Finished 'GetBlockNanoTime' request.")

	return blockTime, nil
}

// GetTokenAliases is a function to get token aliases
func GetTokenAliases(gwCosmosmux *runtime.ServeMux, r *http.Request) ([]types.TokenAlias, string, string) {
	// tokens, err := database.GetTokenAliases()
	// if err == nil {
	// 	return tokens
	// }

	log.CustomLogger().Info("Starting 'GetTokenAliases' request...")

	r.URL.Path = strings.Replace(config.QueryKiraTokensAliases, "/api", "", 1)
	r.URL.RawQuery = ""
	r.Method = "GET"

	// GetLogger().Info("[grpc-call] Entering grpc call: ", r.URL.Path)
	recorder := httptest.NewRecorder()
	gwCosmosmux.ServeHTTP(recorder, r)
	resp := recorder.Result()

	type TokenAliasesResponse struct {
		Data         []types.TokenAlias `json:"data"`
		DefaultDenom string             `json:"defaultDenom"`
		Bech32Prefix string             `json:"bech32Prefix"`
	}

	result := TokenAliasesResponse{}

	err := json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.CustomLogger().Error("[GetTokenAliases] Failed to unmarshal the response.",
			"response", resp.StatusCode,
			"error", err,
		)
	}

	// save block time
	err = database.AddTokenAliases(result.Data)
	if err != nil {
		log.CustomLogger().Error("[GetTokenAliases] Failed to save block time.",
			"response", resp.StatusCode,
			"error", err,
		)
	}

	log.CustomLogger().Info("Finished 'GetTokenAliases' request.")

	return result.Data, result.DefaultDenom, result.Bech32Prefix
}

// GetAllBalances is a function to get all balances with full limitation
func GetAllBalances(gwCosmosmux *runtime.ServeMux, r *http.Request, bech32Addr string) []types.Coin {

	log.CustomLogger().Info("Starting 'GetAllBalances' request...")

	r.URL.Path = fmt.Sprintf("/cosmos/bank/v1beta1/balances/%s", bech32Addr)
	r.URL.RawQuery = "pagination.limit=100000"
	r.Method = "GET"

	log.CustomLogger().Info("Proccessing `GetAllBalances` request...",
		"url", r.URL.Path,
	)

	recorder := httptest.NewRecorder()
	gwCosmosmux.ServeHTTP(recorder, r)
	resp := recorder.Result()

	type AllBalancesResponse struct {
		Balances []types.Coin `json:"balances"`
	}

	result := AllBalancesResponse{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.CustomLogger().Error("[GetAllBalances] Failed to unmarshal the response.",
			"response", resp.StatusCode,
			"error", err,
		)
	}

	log.CustomLogger().Info("Finished 'GetAllBalances' request.")

	return result.Balances
}

// GetTokenSupply is a function to get token supply
func GetTokenSupply(gwCosmosmux *runtime.ServeMux, r *http.Request) []types.TokenSupply {

	log.CustomLogger().Info("Starting 'GetTokenSupply' request...")

	r.URL.Path = strings.Replace(config.QueryTotalSupply, "/api/kira", "/cosmos/bank/v1beta1", -1)
	r.URL.RawQuery = ""
	r.Method = "GET"

	log.CustomLogger().Info("Proccessing `GetTokenSupply` `GetTokenSupply` request...",
		"url", r.URL.Path,
	)

	recorder := httptest.NewRecorder()
	gwCosmosmux.ServeHTTP(recorder, r)
	resp := recorder.Result()

	type TokenAliasesResponse struct {
		Supply []types.TokenSupply `json:"supply"`
	}

	result := TokenAliasesResponse{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.CustomLogger().Error("[GetTokenSupply] Failed to unmarshal the response.",
			"response", resp.StatusCode,
			"error", err,
		)
	}

	log.CustomLogger().Info("Finished 'GetTokenSupply' request.")

	return result.Supply
}

func GetKiraStatus(rpcAddr string) *types.KiraStatus {

	log.CustomLogger().Info("Starting 'GetKiraStatus' request...")

	success, _, _ := MakeTendermintRPCRequest(rpcAddr, "/status", "")

	log.CustomLogger().Info("Proccessed `GetKiraStatus` `MakeTendermintRPCRequest` request...",
		"rpc", rpcAddr,
		"query", "/status",
		"success", true,
	)

	if success != nil {
		result := types.KiraStatus{}

		byteData, err := json.Marshal(success)
		if err != nil {
			log.CustomLogger().Error("[GetKiraStatus] Failed to marshal the response.",
				"error", err,
			)
		}

		err = json.Unmarshal(byteData, &result)
		if err != nil {
			log.CustomLogger().Error("[GetKiraStatus] Failed to unmarshal the response.",
				"error", err,
			)
		}

		return &result
	}

	log.CustomLogger().Info("Finished 'GetKiraStatus' request.")

	return nil
}

func GetInterxStatus(interxAddr string) *types.InterxStatus {

	log.CustomLogger().Info("Starting 'GetInterxStatus' request...")

	success, _, _ := MakeGetRequest(interxAddr, "/api/status", "")

	log.CustomLogger().Info("Proccessed `GetInterxStatus` `MakeGetRequest` request...",
		"rpc", interxAddr,
		"query", "/api/status",
		"success", true,
	)

	if success != nil {
		result := types.InterxStatus{}

		byteData, err := json.Marshal(success)
		if err != nil {
			log.CustomLogger().Error("[GetInterxStatus] Failed to marshal the response.",
				"error", err,
			)
			return nil
		}

		err = json.Unmarshal(byteData, &result)
		if err != nil {
			log.CustomLogger().Error("[GetInterxStatus] Failed to unmarshal the response.",
				"error", err,
			)
			return nil
		}

		return &result
	}

	log.CustomLogger().Info("Finished 'GetInterxStatus' request.")

	return nil
}

func GetSnapshotInfo(interxAddr string) *types.SnapShotChecksumResponse {

	log.CustomLogger().Info("Starting 'GetSnapshotInfo' request...")

	success, _, _ := MakeGetRequest(interxAddr, "/api/snapshot_info", "")

	log.CustomLogger().Info("Proccessed `GetSnapshotInfo` `MakeGetRequest` request...",
		"rpc", interxAddr,
		"query", "/api/snapshot_info",
		"success", true,
	)

	if success != nil {
		result := types.SnapShotChecksumResponse{}

		byteData, err := json.Marshal(success)
		if err != nil {
			log.CustomLogger().Error("[GetSnapshotInfo] Failed to marshal the response.",
				"error", err,
			)
			return nil
		}

		err = json.Unmarshal(byteData, &result)
		if err != nil {
			log.CustomLogger().Error("[GetSnapshotInfo] Failed to unmarshal the response.",
				"error", err,
			)
			return nil
		}

		if result.Size == 0 {
			return nil
		}

		return &result
	}

	log.CustomLogger().Info("Finished 'GetSnapshotInfo' request.")

	return nil
}

// GetBlockchain is a function to get block nano time
func GetBlockchain(rpcAddr string) (*tmTypes.ResultBlockchainInfo, error) {

	log.CustomLogger().Info("Starting 'GetBlockchain' request...")

	endpoint := fmt.Sprintf("%s/blockchain", rpcAddr)
	resp, err := http.Get(endpoint)
	if err != nil {
		log.CustomLogger().Error("[GetBlockchain] Failed to connect to endpoint.",
			"endpoint", endpoint,
			"error", err,
		)
		return nil, fmt.Errorf("blockchain query error")
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	response := new(tmJsonRPCTypes.RPCResponse)

	if err := json.Unmarshal(respBody, response); err != nil {
		log.CustomLogger().Error("[GetBlockchain] Failed to unmarshal the response.",
			"error", err,
		)
		return nil, err
	}

	if response.Error != nil {
		log.CustomLogger().Error("[GetBlockchain] Failed to query the blockchain.",
			"error", err,
		)
		return nil, fmt.Errorf("blockchain query error")
	}

	result := new(tmTypes.ResultBlockchainInfo)
	if err := tmjson.Unmarshal(response.Result, result); err != nil {
		log.CustomLogger().Error("[GetBlockchain] Failed to unmarshal the response.",
			"error", err,
		)
		return nil, err
	}

	log.CustomLogger().Info("Finished 'GetBlockchain' request.")

	return result, nil
}
