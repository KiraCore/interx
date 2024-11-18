package global

import (
	"sync"
)

type QueueItem struct {
	Address       string
	IsWhiteListed bool
}

// Mutex will be used for Sync
var Mutex = sync.Mutex{}
var AddressQueue = []QueueItem{}
var WalletToAddresses = map[string][]string{}
var AddressToWallet = map[string]string{}
