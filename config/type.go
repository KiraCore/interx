package config

import (
	"github.com/KiraCore/interx/types"
	"github.com/cometbft/cometbft/p2p"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
)

// FaucetConfig is a struct to be used for Faucet configuration
type FaucetConfig struct {
	Mnemonic             string            `json:"mnemonic"`
	FaucetAmounts        map[string]string `json:"faucet_amounts"`
	FaucetMinimumAmounts map[string]string `json:"faucet_minimum_amounts"`
	FeeAmounts           map[string]string `json:"fee_amounts"`
	TimeLimit            int64             `json:"time_limit"`
	PrivKey              crypto.PrivKey    `json:"privkey"`
	PubKey               crypto.PubKey     `json:"pubkey"`
	Address              string            `json:"address"`
}

// RPCSetting is a struct to be used for endpoint setting
type RPCSetting struct {
	Disable              bool    `json:"disable"`
	RateLimit            float64 `json:"rate_limit"`
	AuthRateLimit        float64 `json:"auth_rate_limit"`
	CachingDisable       bool    `json:"caching_disable"`
	CachingDuration      int64   `json:"caching_duration"`
	CachingBlockDuration int64   `json:"caching_block_duration"`
}

// RPCConfig is a struct to be used for PRC configuration
type RPCConfig struct {
	API map[string]map[string]RPCSetting `json:"API"`
}

type BlockConfig struct {
	StatusSync          int64 `json:"status_sync"`
	HaltedAvgBlockTimes int64 `json:"halted_avg_block_times"`
}

// CacheConfig is a struct to be used for cache configuration
type CacheConfig struct {
	CacheDir                   string `json:"cache_dir"`
	MaxCacheSize               int64  `json:"max_cache_size"`
	CachingDuration            int64  `json:"caching_duration"`
	DownloadFileSizeLimitation int64  `json:"download_file_size_limitation"`
}

type NodeDiscoveryConfig struct {
	UseHttps              bool   `json:"use_https"`
	DefaultInterxPort     string `json:"default_interx_port"`
	DefaultTendermintPort string `json:"default_tendermint_port"`
	ConnectionTimeout     string `json:"connection_timeout"`
}

type EVMNodeConfig struct {
	RPC       string `json:"rpc"`
	RPCToken  string `json:"rpc_token"`
	RPCSecret string `json:"rpc_secret"`
}

type EVMConfig struct {
	Name      string
	Infura    EVMNodeConfig `json:"infura"`
	QuickNode EVMNodeConfig `json:"quick_node"`
	Pokt      EVMNodeConfig `json:"pokt"`
	Etherscan struct {
		API      string `json:"api"`
		APIToken string `json:"api_token"`
	} `json:"etherscan"`
	Faucet struct {
		PrivateKey           string            `json:"private_key"`
		FaucetAmounts        map[string]string `json:"faucet_amounts"`
		FaucetMinimumAmounts map[string]string `json:"faucet_minimum_amounts"`
		TimeLimit            int64             `json:"time_limit"`
	}
}

type BitcoinConfig struct {
	RPC                 string   `json:"rpc"`
	RPC_CRED            string   `json:"rpc_cred"`
	BTC_CONFIRMATIONS   uint64   `json:"btc_confirmations"`
	BTC_MAX_RESCANS     uint64   `json:"btc_max_rescans"`
	BTC_WATCH_ADDRESSES []string `json:"btc_watch_addresses"`
	BTC_WALLETS         []string `json:"btc_wallets"`
	BTC_WATCH_REGEX     string   `json:"btc_watch_regex"`
	BTC_FAUCET          string   `json:"btc_faucet"`
}

// InterxConfig is a struct to be used for interx configuration
type InterxConfig struct {
	InterxVersion string                   `json:"interx_version"`
	SekaiVersion  string                   `json:"sekai_version"`
	ServeHTTPS    bool                     `json:"serve_https"`
	GRPC          string                   `json:"grpc"`
	RPC           string                   `json:"rpc"`
	PORT          string                   `json:"port"`
	Node          types.NodeConfig         `json:"node"`
	Mnemonic      string                   `json:"mnemonic"`
	AddrBooks     []string                 `json:"addrbooks"`
	NodeKey       *p2p.NodeKey             `json:"node_key"`
	TxModes       []string                 `json:"tx_modes"`
	PrivKey       crypto.PrivKey           `json:"privkey"`
	PubKey        crypto.PubKey            `json:"pubkey"`
	Address       string                   `json:"address"`
	NodeDiscovery NodeDiscoveryConfig      `json:"node_discovery"`
	Block         BlockConfig              `json:"block"`
	Cache         CacheConfig              `json:"cache"`
	Faucet        FaucetConfig             `json:"faucet"`
	RPCMethods    RPCConfig                `json:"rpc_methods"`
	Evm           map[string]EVMConfig     `json:"evm"`
	Bitcoin       map[string]BitcoinConfig `json:"bitcoin"`
}

// InterxConfigFromFile is a struct to be used for interx configuration file
type InterxConfigFromFile struct {
	ServeHTTPS    bool                `json:"serve_https"`
	GRPC          string              `json:"grpc"`
	RPC           string              `json:"rpc"`
	PORT          string              `json:"port"`
	Node          types.NodeConfig    `json:"node"`
	MnemonicFile  string              `json:"mnemonic"`
	AddrBooks     string              `json:"addrbooks"`
	NodeKey       string              `json:"node_key"`
	TxModes       string              `json:"tx_modes"`
	Block         BlockConfig         `json:"block"`
	NodeDiscovery NodeDiscoveryConfig `json:"node_discovery"`
	Cache         struct {
		CacheDir                   string `json:"cache_dir"`
		MaxCacheSize               string `json:"max_cache_size"`
		CachingDuration            int64  `json:"caching_duration"`
		DownloadFileSizeLimitation string `json:"download_file_size_limitation"`
	} `json:"cache"`
	Faucet struct {
		MnemonicFile         string            `json:"mnemonic"`
		FaucetAmounts        map[string]string `json:"faucet_amounts"`
		FaucetMinimumAmounts map[string]string `json:"faucet_minimum_amounts"`
		FeeAmounts           map[string]string `json:"fee_amounts"`
		TimeLimit            int64             `json:"time_limit"`
	} `json:"faucet"`
	Evm     map[string]EVMConfig     `json:"evm"`
	Bitcoin map[string]BitcoinConfig `json:"bitcoin"`
}
