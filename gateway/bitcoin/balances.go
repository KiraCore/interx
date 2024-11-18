package bitcoin

import (
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/gorilla/mux"

	// "github.com/powerman/rpc-codec/jsonrpc2"
	jsonrpc2 "github.com/KeisukeYamashita/go-jsonrpc"
	"github.com/KiraCore/interx/global"
)

// GetWalletInfoResult models the result of the getwalletinfo command.
type GetWalletInfoResult struct {
	btcjson.GetWalletInfoResult
	Format             string `json:"format"`
	Descriptors        bool   `json:"descriptors"`
	Balance            uint64 `json:"balance"`
	UnconfirmedBalance uint64 `json:"unconfirmed_balance"`
	ImmatureBalance    uint64 `json:"immature_balance"`
}

type IncomingTx struct {
	Txid         string  `json:"txid"`
	Category     string  `json:"category"`
	Time         uint64  `json:"time"`
	Block        uint64  `json:"block"`
	Amount       float64 `json:"amount"`
	Vout         uint64  `json:"vout"`
	From         string  `json:"from,omitempty"`
	Unspent      bool    `json:"unspent"`
	ScriptPubKey string  `json:"scriptpubkey,omitempty"`
}

type OutgoingTx struct {
	Txid   string       `json:"txid"`
	Time   uint64       `json:"time"`
	Block  uint64       `json:"block"`
	Amount float64      `json:"amount"`
	Fee    float64      `json:"fee"`
	To     []OutgoingTo `json:"to"`
}

type OutgoingTo struct {
	Address string  `json:"address"`
	Amount  float64 `json:"amount"`
}

type BalancesResult struct {
	Tracking bool   `json:"tracking"`
	Blocks   uint64 `json:"blocks"`
	TxCount  uint64 `json:"txcount"`
	Wallet   struct {
		AvoidReuse  bool     `json:"avoid_reuse"`
		Format      string   `json:"format"`
		Version     uint64   `json:"version"`
		Name        string   `json:"name"`
		Descriptors bool     `json:"descriptors"`
		Addresses   []string `json:"addresses"`
	} `json:"wallet"`
	Balance struct {
		Confirmed   float64 `json:"confirmed"`
		Unconfirmed uint64  `json:"unconfirmed,omitempty"`
		Immature    uint64  `json:"immature,omitempty"`
		Denom       string  `json:"denom"`
		Decimals    uint64  `json:"decimals"`
	} `json:"balance"`
	Scanning struct {
		Isscanning bool    `json:"isscanning"`
		Progress   float64 `json:"progress"`
		Duration   uint64  `json:"duration"`
		Tail       uint64  `json:"tail"`
	} `json:"scanning"`
	Incomming []IncomingTx `json:"incomming"`
	Outgoing  []OutgoingTx `json:"outgoing"`
}

var scansInProgress = 0

// Insert a value in a slice at a given index
// 0 <= index <= len(a)
func Insert(a []global.QueueItem, index int, value global.QueueItem) []global.QueueItem {
	if len(a) == index { // nil or empty slice or after last element
		return append(a, value)
	}
	a = append(a[:index+1], a[index:]...) // index < len(a)
	a[index] = value
	return a
}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// RegisterBtcBalancesRoutes registers query status of bitcoin chains.
func RegisterBtcBalancesRoutes(r *mux.Router, rpcAddr string) {
	r.HandleFunc(config.QueryBitcoinBalances, QueryBtcBalancesRequest(rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryBitcoinBalances, "This is an API to query balances.", true)
}

func GetBestWallet(wallets []string, conf *config.BitcoinConfig) (string, bool) {
	freeWallets := []string{}
	bestWallet := ""
	scanProgress := float64(0)
	for _, walletAddr := range wallets {
		clientWallet := jsonrpc2.NewRPCClient(conf.RPC + "/wallet/" + walletAddr)
		if conf.RPC_CRED != "" {
			rpcInfo := strings.Split(conf.RPC_CRED, ":")
			clientWallet.SetBasicAuth(rpcInfo[0], rpcInfo[1])
		}

		getWalletInfo := GetWalletInfoResult{}
		err := GetResult(clientWallet, "getwalletinfo", &getWalletInfo)
		if err != nil {
			continue
		}

		if getWalletInfo.Scanning.Value == false {
			freeWallets = append(freeWallets, walletAddr)
		} else {
			_scanProgress := getWalletInfo.Scanning.Value.(btcjson.ScanProgress).Progress
			if scanProgress < _scanProgress {
				scanProgress = _scanProgress
				bestWallet = walletAddr
			}
		}
	}

	if bestWallet != "" {
		return bestWallet, true
	}
	if len(freeWallets) == 0 {
		return "", false
	}
	return freeWallets[0], false
}

func AddAddressToQueue(clientWallet *jsonrpc2.RPCClient, address string, conf *config.BitcoinConfig) {
	if Contains(conf.BTC_WATCH_ADDRESSES, address) {
		indexOfLastWhiteListedAddress := 0
		for _, queueItem := range global.AddressQueue {
			if queueItem.IsWhiteListed {
				indexOfLastWhiteListedAddress++
				continue
			}
		}
		global.AddressQueue = Insert(global.AddressQueue, indexOfLastWhiteListedAddress, global.QueueItem{
			Address:       address,
			IsWhiteListed: true,
		})
	} else {
		global.AddressQueue = append(global.AddressQueue, global.QueueItem{
			Address:       address,
			IsWhiteListed: false,
		})
	}
}

func ScanBlockchain(clientWallet *jsonrpc2.RPCClient, walletAddress string, conf *config.BitcoinConfig, isScanning bool) {
	if scansInProgress >= int(conf.BTC_MAX_RESCANS) {
		return
	}
	scansInProgress++
	// Import all address in the queue into the wallet
	getRescanBlockchain := new(interface{})
	for _, queueItem := range global.AddressQueue {
		importAddress := new(interface{})
		GetResult(clientWallet, "importaddress", &importAddress, queueItem.Address, "", false)
		global.AddressToWallet[queueItem.Address] = walletAddress
	}

	// Empty the queue
	global.AddressQueue = []global.QueueItem{}
	if !isScanning {
		GetResult(clientWallet, "rescanblockchain", &getRescanBlockchain, 0)
	}
	scansInProgress--
}

func queryBtcBalancesRequestHandle(r *http.Request, chain string, address string) (interface{}, interface{}, int) {
	isSupportedChain, conf := GetChainConfig(chain)
	if !isSupportedChain {
		return common.ServeError(0, "", "unsupported chain", http.StatusBadRequest)
	}

	walletAddress := global.AddressToWallet[address]
	isScanning := false
	if len(walletAddress) == 0 {
		walletAddress, isScanning = GetBestWallet(conf.BTC_WALLETS, conf)
	}

	if len(walletAddress) == 0 {
		return common.ServeError(0, "", "failed to get suitable wallet for the address", http.StatusInternalServerError)
	}

	// Initializing rpc clients for wallet and address
	client := jsonrpc2.NewRPCClient(conf.RPC)
	if conf.RPC_CRED != "" {
		rpcInfo := strings.Split(conf.RPC_CRED, ":")
		client.SetBasicAuth(rpcInfo[0], rpcInfo[1])
	}
	clientWallet := jsonrpc2.NewRPCClient(conf.RPC + "/wallet/" + walletAddress)
	if conf.RPC_CRED != "" {
		rpcInfo := strings.Split(conf.RPC_CRED, ":")
		clientWallet.SetBasicAuth(rpcInfo[0], rpcInfo[1])
	}

	// Validate user address
	validateAddress := btcjson.ValidateAddressChainResult{}
	err := GetResult(client, "validateaddress", &validateAddress, address)
	if err != nil || !validateAddress.IsValid {
		return common.ServeError(0, "failed to validate address", err.Error(), http.StatusInternalServerError)
	}

	//=========== Start fetching balances ===========
	if len(global.AddressToWallet[address]) == 0 {
		AddAddressToQueue(clientWallet, address, conf)
		go ScanBlockchain(clientWallet, walletAddress, conf, isScanning)
	}

	getWalletInfo := GetWalletInfoResult{}
	err = GetResult(clientWallet, "getwalletinfo", &getWalletInfo)
	if err != nil {
		return common.ServeError(0, "failed to get wallet information", err.Error(), http.StatusInternalServerError)
	}

	getBalancesResult := btcjson.GetBalancesResult{}
	err = GetResult(clientWallet, "getbalances", &getBalancesResult)
	if err != nil {
		return common.ServeError(0, "failed to get wallet balances", err.Error(), http.StatusInternalServerError)
	}

	getListTransactions := []btcjson.ListTransactionsResult{}
	err = GetResult(clientWallet, "listtransactions", &getListTransactions, "*", 2147483647, 0, true)
	if err != nil {
		return common.ServeError(0, "failed to get list transactions", err.Error(), http.StatusInternalServerError)
	}

	getListUnspent := []btcjson.ListUnspentResult{}
	err = GetResult(clientWallet, "listunspent", &getListUnspent, 1, 2147483647, []string{address}, true)
	if err != nil {
		return common.ServeError(0, "failed to get list unspent transactions", err.Error(), http.StatusInternalServerError)
	}
	unSpentTxs := []string{}
	for _, unspent := range getListUnspent {
		unSpentTxs = append(unSpentTxs, unspent.TxID)
	}

	listAddressesGroupings := [][][]interface{}{}
	err = GetResult(clientWallet, "listaddressgroupings", &listAddressesGroupings)
	if err != nil {
		return common.ServeError(0, "failed to list addressesgroupings", err.Error(), http.StatusInternalServerError)
	}
	walletAddresses := []string{}
	addressToBal := map[string]float64{}
	for _, addrJson1 := range listAddressesGroupings {
		for _, addrJson2 := range addrJson1 {
			walletAddresses = append(walletAddresses, fmt.Sprintf("%v", addrJson2[0]))
			addressToBal[addrJson2[0].(string)] = addrJson2[1].(float64)
		}
	}

	// Get known blocks
	chainInfo := btcjson.GetBlockChainInfoResult{}
	err = GetResult(client, "getblockchaininfo", &chainInfo)
	if err != nil {
		return common.ServeError(0, "failed to get total known blocks ", err.Error(), http.StatusInternalServerError)
	}

	balanceResult := BalancesResult{}
	balanceResult.Blocks = uint64(chainInfo.Blocks)
	balanceResult.Tracking = Contains(conf.BTC_WATCH_ADDRESSES, address)
	balanceResult.TxCount = uint64(getWalletInfo.TransactionCount)
	balanceResult.Wallet.AvoidReuse = getWalletInfo.AvoidReuse
	balanceResult.Wallet.Format = getWalletInfo.Format
	balanceResult.Wallet.Version = uint64(getWalletInfo.WalletVersion)
	balanceResult.Wallet.Name = getWalletInfo.WalletName
	balanceResult.Wallet.Descriptors = getWalletInfo.Descriptors
	balanceResult.Wallet.Addresses = walletAddresses

	balanceResult.Balance.Confirmed = addressToBal[address]
	// balanceResult.Balance.Unconfirmed = getWalletInfo.UnconfirmedBalance
	// balanceResult.Balance.Immature = getWalletInfo.ImmatureBalance
	balanceResult.Balance.Denom = "satoshi"
	balanceResult.Balance.Decimals = 8

	isScanning = getWalletInfo.Scanning.Value != false
	balanceResult.Scanning.Isscanning = isScanning
	if isScanning {
		balanceResult.Scanning.Progress = getWalletInfo.Scanning.Value.(btcjson.ScanProgress).Progress
		balanceResult.Scanning.Duration = uint64(getWalletInfo.Scanning.Value.(btcjson.ScanProgress).Duration)
	} else {
		balanceResult.Scanning.Progress = 0
		balanceResult.Scanning.Duration = 0
	}
	balanceResult.Scanning.Tail = 1

	outTransactions := map[string]*OutgoingTx{}
	for _, tx := range getListTransactions {
		if tx.Amount >= 0 {
			if tx.Address == address {
				in := IncomingTx{}
				in.Amount = tx.Amount
				in.Txid = tx.TxID
				in.Category = tx.Category
				in.Time = uint64(tx.Time)
				in.Block = uint64(*tx.BlockHeight)
				in.Vout = uint64(tx.Vout)
				in.Unspent = Contains(unSpentTxs, in.Txid)
				balanceResult.Incomming = append(balanceResult.Incomming, in)
			}
		} else {
			if outTransactions[tx.TxID] != nil {
				out := outTransactions[tx.TxID]
				out.To = append(out.To, OutgoingTo{
					Address: tx.Address,
					Amount:  math.Abs(tx.Amount),
				})
				outTransactions[tx.TxID] = out
			} else {
				out := OutgoingTx{}
				out.Txid = tx.TxID
				out.Time = uint64(tx.Time)
				out.Block = uint64(*tx.BlockHeight)
				out.Amount = math.Abs(tx.Amount)
				out.Fee = math.Abs(*tx.Fee)
				out.To = append(out.To, OutgoingTo{
					Address: tx.Address,
					Amount:  math.Abs(tx.Amount),
				})
				outTransactions[tx.TxID] = &out
			}
		}
	}
	for _, outTx := range outTransactions {
		balanceResult.Outgoing = append(balanceResult.Outgoing, *outTx)
	}

	return balanceResult, nil, http.StatusOK
}

// QueryBtcBalancesRequest is a function to query balances of Bitcoin addresses.
func QueryBtcBalancesRequest(rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queries := mux.Vars(r)
		chain := "testnet"
		address := queries["address"]
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)
		statusCode := http.StatusOK

		common.GetLogger().Info("[query-btc-balances] Entering balances query: ", chain)

		if !common.RPCMethods["GET"][config.QueryBitcoinBalances].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryBitcoinBalances].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					common.GetLogger().Info("[query-btc-balances] Returning from the cache: ", chain)
					return
				}
			}

			response.Response, response.Error, statusCode = queryBtcBalancesRequestHandle(r, chain, address)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryBitcoinBalances].CachingEnabled)
	}
}
