package interx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/database"
	"github.com/KiraCore/interx/types"
	kiratypes "github.com/KiraCore/sekai/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	tmTypes "github.com/cometbft/cometbft/rpc/core/types"
	tmJsonRPCTypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/exp/slices"
)

type TxsResponse struct {
	Transactions []types.TransactionResponse `json:"transactions"`
	TotalCount   int                         `json:"total_count"`
}

// RegisterInterxTxRoutes registers tx query routers.
func RegisterInterxTxRoutes(r *mux.Router, gwCosmosmux *runtime.ServeMux, rpcAddr string) {
	r.HandleFunc(config.QueryUnconfirmedTxs, QueryUnconfirmedTxs(rpcAddr)).Methods("GET")
	r.HandleFunc(config.QueryTransactions, QueryTransactions(rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryKiraFunctions, "This is an API to query kira functions and metadata.", true)
	common.AddRPCMethod("GET", config.QueryUnconfirmedTxs, "This is an API to query unconfirmed transactions.", true)
	common.AddRPCMethod("GET", config.QueryTransactions, "This is an API to query transactions filtered by various options.", true)
}

// GetTransactionsWithSync is a function to sync user transactions and return result
func GetTransactionsWithSync(rpcAddr string, address string, isOutbound bool) (*tmTypes.ResultTxSearch, error) {
	var page = 1
	var limit = 100
	var limitPages = 100

	if address == "" {
		return &tmTypes.ResultTxSearch{}, nil
	}

	lastBlock := database.GetLastBlockFetched(address, isOutbound)
	totalResult := tmTypes.ResultTxSearch{
		Txs:        []*tmTypes.ResultTx{},
		TotalCount: 0,
	}

	for page < limitPages {
		var events = make([]string, 0, 5)
		if isOutbound {
			events = append(events, fmt.Sprintf("message.sender='%s'", address))
		} else {
			events = append(events, fmt.Sprintf("transfer.recipient='%s'", address))
		}
		events = append(events, fmt.Sprintf("tx.height>%d", lastBlock))

		// search transactions
		endpoint := fmt.Sprintf("%s/tx_search?query=\"%s\"&page=%d&per_page=%d&order_by=\"desc\"", rpcAddr, strings.Join(events, "%20AND%20"), page, limit)
		common.GetLogger().Info("[query-transaction] Entering transaction search: ", endpoint)

		resp, err := http.Get(endpoint)
		if err != nil {
			common.GetLogger().Error("[query-transaction] Unable to connect to ", endpoint)
			return nil, err
		}
		defer resp.Body.Close()

		respBody, _ := ioutil.ReadAll(resp.Body)

		response := new(tmJsonRPCTypes.RPCResponse)

		if err := json.Unmarshal(respBody, response); err != nil {
			common.GetLogger().Error("[query-transaction] Unable to decode response: ", err)
			break
		}

		if response.Error != nil {
			break
		}

		result := new(tmTypes.ResultTxSearch)
		if err := tmjson.Unmarshal(response.Result, result); err != nil {
			common.GetLogger().Error("[query-transaction] Failed to unmarshal result:", err)
			break
		}

		if result.TotalCount == 0 {
			break
		}

		totalResult.Txs = append(totalResult.Txs, result.Txs...)

		if result.TotalCount < limit {
			break
		}
		page++
	}
	totalResult.TotalCount = len(totalResult.Txs)
	err := database.SaveTransactions(address, totalResult, isOutbound)
	if err != nil {
		common.GetLogger().Error("[query-transaction] Failed to save cache result:", err)
	}

	return database.GetTransactions(address, isOutbound)
}

// GetFilteredTransactions is a function to filter transactions by various options
func GetFilteredTransactions(rpcAddr string, address string, txtypes []string, directions []string, dateStart int, dateEnd int, statuses []string, sortBy string) ([]types.TransactionResponse, error) {
	result, _ := searchUnconfirmed(rpcAddr, "100")
	hashToStatusMap := make(map[string]string)
	hashToDirectionMap := make(map[string][]string)
	for _, unconfirmedTx := range result.Txs {
		hashToStatusMap[string(unconfirmedTx.Hash())] = "unconfirmed"
	}

	// filter transactions by date start/end timestamp, msg type, and direction
	// filter by direction
	cachedTxs := tmTypes.ResultTxSearch{
		Txs:        []*tmTypes.ResultTx{},
		TotalCount: 0,
	}
	if len(directions) == 0 {
		directions = []string{"inbound", "outbound"}
	}
	if slices.Contains(directions, "inbound") {
		cachedTxs1, err := GetTransactionsWithSync(rpcAddr, address, false)
		for _, cachedTx := range cachedTxs1.Txs {
			hashToDirectionMap[cachedTx.Hash.String()] = append(hashToDirectionMap[cachedTx.Hash.String()], "inbound")
		}
		if err != nil {
			return nil, err
		}
		cachedTxs.TotalCount += cachedTxs1.TotalCount
		cachedTxs.Txs = append(cachedTxs.Txs, cachedTxs1.Txs...)
	}

	if slices.Contains(directions, "outbound") {
		cachedTxs2, err := GetTransactionsWithSync(rpcAddr, address, true)
		for _, cachedTx := range cachedTxs2.Txs {
			hashToDirectionMap[cachedTx.Hash.String()] = append(hashToDirectionMap[cachedTx.Hash.String()], "outbound")
		}
		if err != nil {
			return nil, err
		}
		cachedTxs.TotalCount += cachedTxs2.TotalCount
		cachedTxs.Txs = append(cachedTxs.Txs, cachedTxs2.Txs...)
	}

	var res []types.TransactionResponse
	for _, cachedTx := range cachedTxs.Txs {
		// Filter by time
		txTime, err := common.GetBlockTime(rpcAddr, cachedTx.Height)
		if err != nil {
			common.GetLogger().Error("[query-transactions] Block not found: ", cachedTx.Height)
			continue
		}

		if (dateStart != -1 && txTime < int64(dateStart)) || (dateEnd != -1 && txTime > int64(dateEnd)) {
			continue
		}

		// Filter by msg
		tx, err := config.EncodingCg.TxConfig.TxDecoder()(cachedTx.Tx)
		if err != nil {
			common.GetLogger().Error("[query-transactions] Failed to decode transaction: ", err)
			continue
		}

		contain := false
		txResponses := []interface{}{}
		for _, msg := range tx.GetMsgs() {
			txType := kiratypes.MsgType(msg)
			if slices.Contains(txtypes, txType) {
				contain = true
				break
			}

			// Append type field to each tx msg
			a := make(map[string]interface{})
			bz, err := json.Marshal(msg)
			if err != nil {
				continue
			}
			err = json.Unmarshal(bz, &a)
			if err != nil {
				continue
			}

			a["type"] = txType
			txResponses = append(txResponses, a)
		}

		if !contain && len(txtypes) != 0 {
			continue
		}

		// Filter by status
		if len(statuses) != 0 && !slices.Contains(statuses, "unconfirmed") && hashToStatusMap[cachedTx.Hash.String()] == "unconfirmed" {
			continue
		}

		if len(statuses) != 0 && !slices.Contains(statuses, "confirmed") && hashToStatusMap[cachedTx.Hash.String()] == "" {
			continue
		}

		hashStatus := "confirmed"
		if hashToStatusMap[cachedTx.Hash.String()] == "unconfirmed" {
			hashStatus = "unconfirmed"
		}
		txResponse := types.TransactionResponse{
			Time:      txTime,
			Status:    hashStatus,
			Direction: hashToDirectionMap[cachedTx.Hash.String()][0],
			Hash:      fmt.Sprintf("0x%X", cachedTx.Hash),
			Txs:       txResponses,
		}
		if len(hashToDirectionMap[cachedTx.Hash.String()]) > 1 {
			hashToDirectionMap[cachedTx.Hash.String()] = hashToDirectionMap[cachedTx.Hash.String()][1:]
		}

		// Get memo and fee amounts
		txResult, err := parseTransaction(rpcAddr, *cachedTx)
		if err == nil {
			txResponse.Fee = txResult.Fees
			txResponse.Memo = txResult.Memo
		}
		res = append(res, txResponse)
	}
	return res, nil
}

// SearchTxHashHandle is a function to query transactions
func SearchTxHashHandle(rpcAddr string, sender string, recipient string, txType string, page int, limit int, txMinHeight int64, txMaxHeight int64, txHash string) (*tmTypes.ResultTxSearch, error) {
	var events = make([]string, 0, 5)

	if sender != "" {
		events = append(events, fmt.Sprintf("transfer.sender='%s'", sender))
	}

	if recipient != "" {
		events = append(events, fmt.Sprintf("transfer.recipient='%s'", recipient))
	}

	if txType != "all" && txType != "" {
		events = append(events, fmt.Sprintf("message.action='%s'", txType))
	}

	if txHash != "" {
		events = append(events, fmt.Sprintf("tx.hash='%s'", txHash))
	}

	if txMinHeight >= 0 {
		events = append(events, fmt.Sprintf("tx.height>=%d", txMinHeight))
	}

	if txMaxHeight >= 0 {
		events = append(events, fmt.Sprintf("tx.height<=%d", txMaxHeight))
	}

	// search transactions
	endpoint := fmt.Sprintf("%s/tx_search?query=\"%s\"&page=%d&&per_page=%d&order_by=\"desc\"", rpcAddr, strings.Join(events, "%20AND%20"), page, limit)
	if page == 0 {
		endpoint = fmt.Sprintf("%s/tx_search?query=\"%s\"&per_page=%d&order_by=\"desc\"", rpcAddr, strings.Join(events, "%20AND%20"), limit)
	}
	common.GetLogger().Info("[query-transaction] Entering transaction search: ", endpoint)

	resp, err := http.Get(endpoint)
	if err != nil {
		common.GetLogger().Error("[query-transaction] Unable to connect to ", endpoint)
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	response := new(tmJsonRPCTypes.RPCResponse)

	if err := json.Unmarshal(respBody, response); err != nil {
		common.GetLogger().Error("[query-transaction] Unable to decode response: ", err)
		return nil, err
	}

	if response.Error != nil {
		common.GetLogger().Error("[query-transaction] Error response:", response.Error.Message)
		return nil, errors.New(response.Error.Message)
	}

	result := new(tmTypes.ResultTxSearch)
	if err := tmjson.Unmarshal(response.Result, result); err != nil {
		common.GetLogger().Error("[query-transaction] Failed to unmarshal result:", err)
		return nil, fmt.Errorf("error unmarshalling result: %w", err)
	}

	return result, nil
}

// Get block height for tx hash from cache or tendermint
func getBlockHeight(rpcAddr string, hash string) (int64, error) {
	endpoint := fmt.Sprintf("%s/tx?hash=%s", rpcAddr, hash)
	common.GetLogger().Info("[query-block] Entering block query: ", endpoint)

	resp, err := http.Get(endpoint)
	if err != nil {
		common.GetLogger().Error("[query-block] Unable to connect to ", endpoint)
		return 0, err
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	response := new(tmJsonRPCTypes.RPCResponse)

	if err := json.Unmarshal(respBody, response); err != nil {
		common.GetLogger().Error("[query-block] Unable to decode response: ", err)
		return 0, err
	}
	if response.Error != nil {
		common.GetLogger().Error("[query-block] Error response:", response.Error.Message)
		return 0, errors.New(response.Error.Message)
	}

	result := new(tmTypes.ResultTx)
	if err := tmjson.Unmarshal(response.Result, result); err != nil {
		common.GetLogger().Error("[query-block] Failed to unmarshal result:", err)
		return 0, fmt.Errorf("error unmarshalling result: %w", err)
	}

	return result.Height, nil
}

func QueryBlockTransactionsHandler(rpcAddr string, r *http.Request) (interface{}, interface{}, int) {
	err := r.ParseForm()
	if err != nil {
		common.GetLogger().Error("[query-transactions] Failed to parse query parameters:", err)
		return common.ServeError(0, "failed to parse query parameters", err.Error(), http.StatusBadRequest)
	}

	var (
		account    string
		txTypes        = []string{}
		directions     = []string{}
		statuses       = []string{}
		dateStart  int = -1
		dateEnd    int = -1
		sortBy     string
		pageSize   int = -1
		page       int = -1
		limit      int = -1
		offset     int = -1
	)

	//------------ Type ------------
	txTypesParam := r.FormValue("type")
	txTypesArray := strings.Split(txTypesParam, ",")
	for _, txType := range txTypesArray {
		if config.MsgTypes[txType] != "" {
			txTypes = append(txTypes, txType)
		}
	}

	//------------ Address ------------
	account = r.FormValue("address")
	if account == "" {
		common.GetLogger().Error("[query-transactions] 'address' is not set")
		return common.ServeError(0, "'address' is not set", "", http.StatusBadRequest)
	}

	//------------ Direction ------------
	directionsParam := r.FormValue("direction")
	directionsArray := strings.Split(directionsParam, ",")
	for _, drt := range directionsArray {
		if drt == "inbound" || drt == "outbound" {
			directions = append(directions, drt)
		}
	}

	//------------ Status ------------
	statusesParam := r.FormValue("status")
	statusesArray := strings.Split(statusesParam, ",")
	for _, sts := range statusesArray {
		if sts == "pending" || sts == "confirmed" || sts == "failed" {
			statuses = append(statuses, sts)
		}
	}

	//------------ Sort ------------
	sortParam := r.FormValue("sort")
	if sortParam == "dateASC" || sortParam == "dateDESC" {
		sortBy = sortParam
	} else {
		sortBy = "dateDESC"
	}

	//------------ Timestamps ------------
	if dateStStr := r.FormValue("dateStart"); dateStStr != "" {
		if dateStart, err = strconv.Atoi(dateStStr); err != nil {
			layout := "01/02/2006 3:04:05 PM"
			t, err1 := time.Parse(layout, dateStStr+" 12:00:00 AM")
			if err1 != nil {
				common.GetLogger().Error("[query-transactions] Failed to parse parameter 'dateStart': ", err1)
				return common.ServeError(0, "failed to parse parameter 'dateStart'", err.Error(), http.StatusBadRequest)
			}

			dateStart = int(t.Unix())
		}
	}

	if dateEdStr := r.FormValue("dateEnd"); dateEdStr != "" {
		if dateEnd, err = strconv.Atoi(dateEdStr); err != nil {
			layout := "01/02/2006 3:04:05 PM"
			t, err1 := time.Parse(layout, dateEdStr+" 12:00:00 AM")
			if err1 != nil {
				common.GetLogger().Error("[query-transactions] Failed to parse parameter 'dateEnd': ", err1)
				return common.ServeError(0, "failed to parse parameter 'dateEnd'", err.Error(), http.StatusBadRequest)
			}

			dateEnd = int(t.Unix())
		}
	}

	//------------ Pagination ------------
	if pageSizeStr := r.FormValue("page_size"); pageSizeStr != "" {
		if pageSize, err = strconv.Atoi(pageSizeStr); err != nil {
			common.GetLogger().Error("[query-transactions] Failed to parse parameter 'page_size': ", err)
			return common.ServeError(0, "failed to parse parameter 'page_size'", err.Error(), http.StatusBadRequest)
		}
		if pageSize < 1 || pageSize > 100 {
			common.GetLogger().Error("[query-transactions] Invalid 'page_size' range: ", pageSize)
			return common.ServeError(0, "'page_size' should be 1 ~ 100", "", http.StatusBadRequest)
		}
	}

	if pageStr := r.FormValue("page"); pageStr != "" {
		if page, err = strconv.Atoi(pageStr); err != nil {
			common.GetLogger().Error("[query-transactions] Failed to parse parameter 'page': ", err)
			return common.ServeError(0, "failed to parse parameter 'page'", err.Error(), http.StatusBadRequest)
		}
	}

	if pageSize > 0 && page == -1 {
		offset = 0
		limit = pageSize
	} else if pageSize == -1 && page > 0 {
		offset = 30 * (page - 1)
		limit = 30
	} else if pageSize > 0 && page > 0 {
		offset = pageSize * (page - 1)
		limit = pageSize
	} else {
		if limitStr := r.FormValue("limit"); limitStr != "" {
			if limit, err = strconv.Atoi(limitStr); err != nil {
				common.GetLogger().Error("[query-transactions] Failed to parse parameter 'limit': ", err)
				return common.ServeError(0, "failed to parse parameter 'limit'", err.Error(), http.StatusBadRequest)
			}

			if limit < 1 || limit > 100 {
				common.GetLogger().Error("[query-transactions] Invalid 'limit' range: ", limit)
				return common.ServeError(0, "'limit' should be 1 ~ 100", "", http.StatusBadRequest)
			}
		}

		if offsetStr := r.FormValue("offset"); offsetStr != "" {
			if offset, err = strconv.Atoi(offsetStr); err != nil {
				common.GetLogger().Error("[query-transactions] Failed to parse parameter 'offset': ", err)
				return common.ServeError(0, "failed to parse parameter 'offset'", err.Error(), http.StatusBadRequest)
			}
		}

		if limit == -1 {
			limit = 30
		}
		if offset == -1 {
			offset = 0
		}
	}

	txResults, err := GetFilteredTransactions(rpcAddr, account, txTypes, directions, dateStart, dateEnd, statuses, sortBy)
	if err != nil {
		return common.ServeError(0, "", err.Error(), http.StatusInternalServerError)
	}

	// sort txResults
	if sortBy == "dateASC" {
		sort.Slice(txResults, func(i, j int) bool {
			return txResults[i].Time < txResults[j].Time
		})
	} else {
		sort.Slice(txResults, func(i, j int) bool {
			return txResults[i].Time > txResults[j].Time
		})
	}

	totalCount := len(txResults)

	// pagination for txResults
	if offset > len(txResults) {
		offset = len(txResults)
	}
	txResults = txResults[offset:int(math.Min(float64(offset+limit), float64(len(txResults))))]

	res := TxsResponse{
		TotalCount:   totalCount,
		Transactions: txResults,
	}

	return res, nil, http.StatusOK
}

// QueryWithdraws is a function to query all transactions.
func QueryTransactions(rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		common.GetLogger().Info("[query-transactions] Entering transactions query")

		if !common.RPCMethods["GET"][config.QueryTransactions].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryTransactions].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					common.GetLogger().Info("[query-transactions] Returning from the cache")
					return
				}
			}

			response.Response, response.Error, statusCode = QueryBlockTransactionsHandler(rpcAddr, r)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryStatus].CachingEnabled)
	}
}

func searchUnconfirmed(rpcAddr string, limit string) (*tmTypes.ResultUnconfirmedTxs, error) {
	endpoint := fmt.Sprintf("%s/unconfirmed_txs?limit=%s", rpcAddr, limit)
	common.GetLogger().Info("[query-unconfirmed-txs] Entering transaction search: ", endpoint)

	resp, err := http.Get(endpoint)
	if err != nil {
		common.GetLogger().Error("[query-unconfirmed-txs] Unable to connect to ", endpoint)
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	response := new(tmJsonRPCTypes.RPCResponse)

	if err := json.Unmarshal(respBody, response); err != nil {
		common.GetLogger().Error("[query-unconfirmed-txs] Unable to decode response: ", err)
		return nil, err
	}

	if response.Error != nil {
		common.GetLogger().Error("[query-unconfirmed-txs] Error response:", response.Error.Message)
		return nil, errors.New(response.Error.Message)
	}

	result := new(tmTypes.ResultUnconfirmedTxs)
	if err := tmjson.Unmarshal(response.Result, result); err != nil {
		common.GetLogger().Error("[query-unconfirmed-txs] Failed to unmarshal result:", err)
		return nil, fmt.Errorf("error unmarshalling result: %w", err)
	}

	return result, nil
}

func queryUnconfirmedTransactionsHandler(rpcAddr string, r *http.Request) (interface{}, interface{}, int) {
	limit := r.FormValue("limit")
	result, err := searchUnconfirmed(rpcAddr, limit)
	if err != nil {
		common.GetLogger().Error("[query-unconfirmed-txs] Failed to query unconfirmed txs: %w ", err)
		return common.ServeError(0, "", err.Error(), http.StatusInternalServerError)
	}

	response := struct {
		Count      int                                  `json:"n_txs"`
		Total      int                                  `json:"total"`
		TotalBytes int64                                `json:"total_bytes"`
		Txs        []types.TransactionUnconfirmedResult `json:"txs"`
	}{}

	response.Count = result.Count
	response.Total = result.Total
	response.TotalBytes = result.TotalBytes
	response.Txs = make([]types.TransactionUnconfirmedResult, 0)

	for _, tx := range result.Txs {
		decodedTx, err := config.EncodingCg.TxConfig.TxDecoder()(tx)
		if err != nil {
			common.GetLogger().Error("[post-unconfirmed-txs] Failed to decode transaction: ", err)
			return common.ServeError(0, "failed to decode signed TX", err.Error(), http.StatusBadRequest)
		}

		txResult, ok := decodedTx.(signing.Tx)
		if !ok {
			common.GetLogger().Error("[post-unconfirmed-txs] Failed to decode transaction")
			return common.ServeError(0, "failed to decode signed TX", err.Error(), http.StatusBadRequest)
		}

		signature, _ := txResult.GetSignaturesV2()

		var msgs []types.TxMsg = make([]types.TxMsg, 0)

		for _, msg := range txResult.GetMsgs() {
			msgs = append(msgs, types.TxMsg{
				Type: kiratypes.MsgType(msg),
				Data: msg,
			})
		}

		response.Txs = append(response.Txs, types.TransactionUnconfirmedResult{
			Msgs:      msgs,
			Fees:      txResult.GetFee(),
			Gas:       txResult.GetGas(),
			Signature: signature,
			Memo:      txResult.GetMemo(),
		})
	}

	return response, nil, http.StatusOK
}

// QueryUnconfirmedTxs is a function to query unconfirmed transactions.
func QueryUnconfirmedTxs(rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		common.GetLogger().Error("[query-unconfirmed-txs] Entering query")

		if !common.RPCMethods["GET"][config.QueryUnconfirmedTxs].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			response.Response, response.Error, statusCode = queryUnconfirmedTransactionsHandler(rpcAddr, r)
		}

		common.WrapResponse(w, request, *response, statusCode, false)
	}
}
