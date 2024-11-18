package tasks

import (
	"strings"

	jsonrpc2 "github.com/KeisukeYamashita/go-jsonrpc"
	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/gateway/bitcoin"
	"github.com/KiraCore/interx/global"
	"github.com/btcsuite/btcd/btcjson"
)

func SyncQueueAddressBalances(conf *config.BitcoinConfig) {
	// Get suitable wallet to sync balances
	walletAddress, isScanning := bitcoin.GetBestWallet(conf.BTC_WALLETS, conf)
	if len(walletAddress) == 0 {
		return
	}

	clientWallet := jsonrpc2.NewRPCClient(conf.RPC + "/wallet/" + walletAddress)
	if conf.RPC_CRED != "" {
		rpcInfo := strings.Split(conf.RPC_CRED, ":")
		clientWallet.SetBasicAuth(rpcInfo[0], rpcInfo[1])
	}

	// Scan the blockchain to sync addresses balances in the queue
	bitcoin.ScanBlockchain(clientWallet, walletAddress, conf, isScanning)
}

func SyncBitcoinWallets() {
	isSupportedChain, conf := bitcoin.GetChainConfig("testnet")
	if !isSupportedChain {
		return
	}

	// Initialize rpc client
	client := jsonrpc2.NewRPCClient(conf.RPC)
	rpcInfo := []string{}
	if conf.RPC_CRED != "" {
		rpcInfo = strings.Split(conf.RPC_CRED, ":")
		client.SetBasicAuth(rpcInfo[0], rpcInfo[1])
	} else {
		return
	}

	// Get all available wallets linked to the node, iterate interx wallets,
	// and for those which are not involved in available wallets, create new ones in the node.
	listwallets := []string{}
	err := bitcoin.GetResult(client, "listwallets", &listwallets)
	if err != nil {
		return
	}

	for _, walletAddress := range conf.BTC_WALLETS {
		walletExist := bitcoin.Contains(listwallets, walletAddress)

		// Create wallet
		createWallet := btcjson.CreateWalletResult{}
		err = bitcoin.GetResult(client, "createwallet", &createWallet, walletAddress, true, true, "", false, false, true)

		// Load wallet
		loadWallet := new(interface{})
		bitcoin.GetResult(client, "loadwallet", &loadWallet, walletAddress, true)

		// Get all addresses attached to the wallet, and fill the global wallet=>address map with those addresses.
		if walletExist {
			// Initialize wallet rpc client
			clientWallet := jsonrpc2.NewRPCClient(conf.RPC + "/wallet/" + walletAddress)
			clientWallet.SetBasicAuth(rpcInfo[0], rpcInfo[1])

			listaddresses := []btcjson.ListReceivedByAddressResult{}
			err = bitcoin.GetResult(clientWallet, "listreceivedbyaddress", &listaddresses, 0, true, true)
			if err != nil {
				return
			}

			for _, _address := range listaddresses {
				global.AddressToWallet[_address.Address] = walletAddress
			}
		}
	}

	// Append interx watch addresses to the address queue to sync at first.
	global.AddressQueue = []global.QueueItem{}
	for _, watchAddress := range conf.BTC_WATCH_ADDRESSES {
		global.AddressQueue = append(global.AddressQueue, global.QueueItem{
			Address:       watchAddress,
			IsWhiteListed: true,
		})
	}

	if len(global.AddressQueue) > 0 {
		SyncQueueAddressBalances(conf)
	}
}
