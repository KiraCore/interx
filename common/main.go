package common

import (
	"time"

	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/log"
	"github.com/KiraCore/interx/types"
)

// RPCMethods is a variable for rpc methods
var RPCMethods = make(map[string]map[string]types.RPCMethod)

// AddRPCMethod is a function to add a RPC method
func AddRPCMethod(method string, url string, description string, canCache bool) {
	newMethod := types.RPCMethod{}
	newMethod.Description = description
	newMethod.Enabled = true
	newMethod.CacheEnabled = true

	if conf, ok := config.Config.RPCMethods.API[method][url]; ok {
		newMethod.Enabled = !conf.Disable
		newMethod.CacheEnabled = !conf.CacheDisable
		newMethod.RateLimit = conf.RateLimit
		newMethod.AuthRateLimit = conf.AuthRateLimit
		newMethod.CacheDuration = conf.CacheDuration
		newMethod.CacheBlockDuration = conf.CacheBlockDuration
	}

	if !canCache {
		newMethod.CacheEnabled = false
	}

	if _, ok := RPCMethods[method]; !ok {
		RPCMethods[method] = map[string]types.RPCMethod{}
	}
	RPCMethods[method][url] = newMethod
}

// NodeStatus is a struct to be used for node status
var NodeStatus struct {
	Chainid   string `json:"chain_id"`
	Block     int64  `json:"block"`
	Blocktime string `json:"block_time"`
}

func IsCacheExpired(result types.InterxResponse) bool {
	log.CustomLogger().Info("Starting `IsCacheExpired` function...",
		"cache_block_duration", result.CacheBlockDuration,
		"cache_duration", result.CacheDuration,
		"response_block", result.Response.Block,
		"node_block", NodeStatus.Block,
		"cache_time", result.CacheTime,
	)

	// Check if the cache is expired based on block duration
	isBlockExpire := false
	switch {
	case result.CacheBlockDuration == 0:
		log.CustomLogger().Info("`CacheBlockDuration` is 0. Block-based cache expiration detected.")
		isBlockExpire = true
	case result.CacheBlockDuration == -1:
		log.CustomLogger().Info("`CacheBlockDuration` is -1. Block-based cache has not expire.")
		isBlockExpire = false
	case result.Response.Block+result.CacheBlockDuration <= NodeStatus.Block:
		log.CustomLogger().Info("`CacheBlockDuration` Block-based cache has expired.",
			"response_block", result.Response.Block,
			"cache_block_duration", result.CacheBlockDuration,
			"current_block", NodeStatus.Block,
		)
		isBlockExpire = true
	default:
		log.CustomLogger().Info("`CacheBlockDuration` Block-based cache has not expired.",
			"response_block", result.Response.Block,
			"cache_block_duration", result.CacheBlockDuration,
			"current_block", NodeStatus.Block,
		)
		isBlockExpire = false
	}

	// Check if the cache is expired based on timestamp duration
	isTimestampExpire := false
	switch {
	case result.CacheDuration == 0:
		log.CustomLogger().Info("`CacheDuration` is 0. Timestamp-based cache expiration detected.")
		isTimestampExpire = true
	case result.CacheDuration == -1:
		log.CustomLogger().Info("`CacheDuration` is -1. Timestamp-based cache has not expire.")
		isTimestampExpire = false
	case result.CacheTime.Add(time.Duration(result.CacheDuration) * time.Second).Before(time.Now().UTC()):
		log.CustomLogger().Info("`CacheDuration` Timestamp-based cache has expired.",
			"cache_time", result.CacheTime,
			"cache_duration_seconds", result.CacheDuration,
			"current_time", time.Now().UTC(),
		)
		isTimestampExpire = true
	default:
		log.CustomLogger().Info("`CacheDuration` Timestamp-based cache has not expired.",
			"cache_time", result.CacheTime,
			"cache_duration_seconds", result.CacheDuration,
			"current_time", time.Now().UTC(),
		)
		isTimestampExpire = false
	}

	// Cache is expired if either block or timestamp condition is met
	isExpired := isBlockExpire || isTimestampExpire
	log.CustomLogger().Info("Completed `IsCacheExpired` function.",
		"is_block_expired", isBlockExpire,
		"is_timestamp_expired", isTimestampExpire,
		"is_expired", isExpired,
	)

	return isExpired
}
