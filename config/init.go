package config

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/KiraCore/interx/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tyler-smith/go-bip39"
)

func getPostMethods() []string {
	return []string{
		PostTransaction,
		EncodeTransaction,
	}
}

var Cache_Duration int64
var Cache_Block_Duration int64

func getRPCSettings() RPCConfig {

	log.CustomLogger().Info("Starting `getRPCSettings` RPC setting from the config.")
	config := RPCConfig{}

	defaultRPCSetting := RPCSetting{
		Disable:            false,
		RateLimit:          0,
		AuthRateLimit:      0,
		CacheDisable:       false,
		CacheDuration:      30,
		CacheBlockDuration: 5,
	}

	config.API = make(map[string]map[string]RPCSetting)
	config.API["GET"] = make(map[string]RPCSetting)
	config.API["POST"] = make(map[string]RPCSetting)

	// endpoints that can change within 1 block time
	defaultRPCSetting.CacheDuration = -1
	defaultRPCSetting.CacheBlockDuration = 1
	config.API["GET"][QueryAccounts] = defaultRPCSetting
	config.API["GET"][QueryTotalSupply] = defaultRPCSetting
	config.API["GET"][QueryBalances] = defaultRPCSetting
	config.API["GET"][QueryDataReferenceKeys] = defaultRPCSetting
	config.API["GET"][QueryDataReference] = defaultRPCSetting
	config.API["GET"][QueryKiraStatus] = defaultRPCSetting
	config.API["GET"][QueryTransactions] = defaultRPCSetting
	config.API["GET"][QueryStatus] = defaultRPCSetting
	config.API["GET"][QueryValidators] = defaultRPCSetting
	config.API["GET"][QueryValidatorInfos] = defaultRPCSetting
	config.API["GET"][QueryRoles] = defaultRPCSetting
	config.API["GET"][QueryRolesByAddress] = defaultRPCSetting
	config.API["GET"][QueryPermissionsByAddress] = defaultRPCSetting
	config.API["GET"][QueryProposals] = defaultRPCSetting
	config.API["GET"][QueryProposal] = defaultRPCSetting
	config.API["GET"][QueryKiraTokensRates] = defaultRPCSetting
	config.API["GET"][QueryVoters] = defaultRPCSetting
	config.API["GET"][QueryVotes] = defaultRPCSetting
	config.API["GET"][QueryKiraTokensRates] = defaultRPCSetting
	config.API["GET"][QueryNetworkProperties] = defaultRPCSetting
	config.API["GET"][QueryExecutionFee] = defaultRPCSetting
	config.API["GET"][QueryExecutionFees] = defaultRPCSetting
	config.API["GET"][QueryRosettaNetworkList] = defaultRPCSetting
	config.API["GET"][QueryRosettaNetworkOptions] = defaultRPCSetting
	config.API["GET"][QueryRosettaNetworkStatus] = defaultRPCSetting
	config.API["GET"][QueryRosettaAccountBalance] = defaultRPCSetting
	config.API["GET"][QueryCurrentPlan] = defaultRPCSetting
	config.API["GET"][QueryNextPlan] = defaultRPCSetting
	config.API["GET"][QueryAddrBook] = defaultRPCSetting
	config.API["GET"][QueryNetInfo] = defaultRPCSetting
	config.API["GET"][QueryIdentityRecord] = defaultRPCSetting
	config.API["GET"][QueryIdentityRecordsByAddress] = defaultRPCSetting
	config.API["GET"][QueryAllIdentityRecords] = defaultRPCSetting
	config.API["GET"][QueryIdentityRecordVerifyRequest] = defaultRPCSetting
	config.API["GET"][QueryIdentityRecordVerifyRequestsByRequester] = defaultRPCSetting
	config.API["GET"][QueryIdentityRecordVerifyRequestsByApprover] = defaultRPCSetting
	config.API["GET"][QueryAllIdentityRecordVerifyRequests] = defaultRPCSetting
	config.API["GET"][QuerySpendingPools] = defaultRPCSetting
	config.API["GET"][QueryStakingPool] = defaultRPCSetting
	config.API["GET"][QuerySpendingPoolProposals] = defaultRPCSetting
	config.API["GET"][QueryUBIRecords] = defaultRPCSetting

	// endpoints that never change
	neverChangeSettings := defaultRPCSetting
	neverChangeSettings.CacheDuration = -1
	neverChangeSettings.CacheBlockDuration = -1
	config.API["GET"][QueryTransactionHash] = neverChangeSettings
	config.API["GET"][QueryBlocks] = neverChangeSettings
	config.API["GET"][QueryBlockByHeightOrHash] = neverChangeSettings
	config.API["GET"][QueryTransactionResult] = neverChangeSettings

	// endpoints that change constantly
	constantChangeSettings := defaultRPCSetting
	constantChangeSettings.CacheDuration = 1
	constantChangeSettings.CacheBlockDuration = -1
	config.API["GET"][QueryConsensus] = constantChangeSettings
	config.API["GET"][QueryPrivP2PList] = constantChangeSettings
	config.API["GET"][QueryPubP2PList] = constantChangeSettings
	config.API["GET"][QueryInterxList] = constantChangeSettings
	config.API["GET"][QuerySnapList] = constantChangeSettings

	config.API["GET"][QueryEVMStatus] = constantChangeSettings

	Cache_Duration = defaultRPCSetting.CacheDuration
	Cache_Block_Duration = defaultRPCSetting.CacheBlockDuration

	for _, item := range getPostMethods() {
		config.API["POST"][item] = defaultRPCSetting
	}

	log.CustomLogger().Info("Finish `getRPCSettings` RPC setting from the config.",
		"defaultRPCSetting", defaultRPCSetting,
		"neverChangeSettings", neverChangeSettings,
		"constantChangeSettings", constantChangeSettings,
	)

	return config
}

func defaultConfig() InterxConfigFromFile {
	configFromFile := InterxConfigFromFile{}

	configFromFile.ServeHTTPS = false
	configFromFile.GRPC = "dns:///0.0.0.0:9090"
	configFromFile.RPC = "http://0.0.0.0:26657"
	configFromFile.PORT = "11000"

	configFromFile.Node.NodeType = "seed"
	configFromFile.Node.SentryNodeID = ""
	configFromFile.Node.SnapshotNodeID = ""
	configFromFile.Node.ValidatorNodeID = ""
	configFromFile.Node.SeedNodeID = ""

	entropy, _ := bip39.NewEntropy(256)
	configFromFile.MnemonicFile, _ = bip39.NewMnemonic(entropy)

	configFromFile.AddrBooks = "addrbook.json"
	configFromFile.NodeKey = "node_key.json"
	configFromFile.TxModes = "sync,async,block"

	configFromFile.Block.StatusSync = 5
	configFromFile.Block.HaltedAvgBlockTimes = 10

	configFromFile.NodeDiscovery.UseHttps = false
	configFromFile.NodeDiscovery.DefaultInterxPort = "11000"
	configFromFile.NodeDiscovery.DefaultTendermintPort = "26657"
	configFromFile.NodeDiscovery.ConnectionTimeout = "3s"

	configFromFile.Cache.CacheDir = "cache"
	configFromFile.Cache.MaxCacheSize = "2GB"
	configFromFile.Cache.CacheDuration = 5
	configFromFile.Cache.DownloadFileSizeLimitation = "10MB"

	configFromFile.Faucet.MnemonicFile = configFromFile.MnemonicFile

	configFromFile.Faucet.FaucetAmounts = make(map[string]string)
	configFromFile.Faucet.FaucetAmounts["stake"] = "100000"
	configFromFile.Faucet.FaucetAmounts["validatortoken"] = "100000"
	configFromFile.Faucet.FaucetAmounts["ukex"] = "100000"
	configFromFile.Faucet.FaucetMinimumAmounts = make(map[string]string)
	configFromFile.Faucet.FaucetMinimumAmounts["stake"] = "100"
	configFromFile.Faucet.FaucetMinimumAmounts["validatortoken"] = "100"
	configFromFile.Faucet.FaucetMinimumAmounts["ukex"] = "100"
	configFromFile.Faucet.FeeAmounts = make(map[string]string)
	configFromFile.Faucet.FeeAmounts["stake"] = "1000ukex"
	configFromFile.Faucet.FeeAmounts["validatortoken"] = "1000ukex"
	configFromFile.Faucet.FeeAmounts["ukex"] = "1000ukex"
	configFromFile.Faucet.TimeLimit = 20

	configFromFile.Evm = make(map[string]EVMConfig)
	for _, item := range SupportedEVMChains {
		evmConfig := EVMConfig{}
		evmConfig.Name = ""
		evmConfig.Infura.RPC = "https://eth-goerli.public.blastapi.io"
		evmConfig.Infura.RPCToken = ""
		evmConfig.Infura.RPCSecret = ""
		evmConfig.QuickNode.RPC = ""
		evmConfig.QuickNode.RPCToken = ""
		evmConfig.Pokt.RPC = ""
		evmConfig.Pokt.RPCToken = ""
		evmConfig.Pokt.RPCSecret = ""
		evmConfig.Etherscan.API = ""
		evmConfig.Etherscan.APIToken = ""
		evmConfig.Faucet.PrivateKey = "0000000000000000000000000000000000000000000000000000000000000000"
		evmConfig.Faucet.FaucetAmounts = make(map[string]string)
		evmConfig.Faucet.FaucetAmounts["0x0000000000000000000000000000000000000000"] = "10000000000000000"
		evmConfig.Faucet.FaucetMinimumAmounts = make(map[string]string)
		evmConfig.Faucet.FaucetMinimumAmounts["0x0000000000000000000000000000000000000000"] = "1000000000000000"
		evmConfig.Faucet.TimeLimit = 20

		configFromFile.Evm[item] = evmConfig
	}

	configFromFile.Bitcoin = make(map[string]BitcoinConfig)
	for _, item := range SupportedBitcoinChains {
		bitcoinConfig := BitcoinConfig{}
		bitcoinConfig.RPC = "http://127.0.0.1:18332"
		bitcoinConfig.RPC_CRED = "admin:1234"
		bitcoinConfig.BTC_CONFIRMATIONS = 100
		bitcoinConfig.BTC_MAX_RESCANS = 4
		bitcoinConfig.BTC_WALLETS = []string{
			"tb1qzg0c8u8hxlcgnvj0sx5g7tq908phkldxv2tmyx",
		}
		if bitcoinConfig.BTC_FAUCET != "" {
			bitcoinConfig.BTC_WATCH_ADDRESSES = []string{bitcoinConfig.BTC_FAUCET}
		}
		bitcoinConfig.BTC_WATCH_REGEX = ""
		configFromFile.Bitcoin[item] = bitcoinConfig
	}

	return configFromFile
}

// InitConfig is a function to load interx configurations from a given file
func InitConfig(
	configFilePath string,
	serveHTTPS bool,
	grpc string,
	rpc string,
	nodeType string,
	sentryNodeId string,
	snapshotNodeId string,
	validatorNodeId string,
	seedNodeId string,
	port string,
	signingMnemonic string,
	syncStatus int64,
	haltedAvgBlockTimes int64,
	cacheDir string,
	maxCacheSize string,
	cacheDuration int64,
	maxDownloadSize string,
	faucetMnemonic string,
	faucetTimeLimit int64,
	faucetAmounts string,
	faucetMinimumAmounts string,
	feeAmounts string,
	addrBooks string,
	txModes string,
	nodeDiscoveryUseHttps bool,
	nodeDiscoveryInterxPort string,
	nodeDiscoveryTendermintPort string,
	nodeDiscoveryTimeout string,
	nodeKey string,
	snapshotInterval uint64,
) {
	configFromFile := defaultConfig()

	configFromFile.ServeHTTPS = serveHTTPS
	configFromFile.GRPC = grpc
	configFromFile.RPC = rpc
	configFromFile.PORT = port

	configFromFile.Node.NodeType = nodeType
	configFromFile.Node.SentryNodeID = sentryNodeId
	configFromFile.Node.SnapshotNodeID = snapshotNodeId
	configFromFile.Node.ValidatorNodeID = validatorNodeId
	configFromFile.Node.SeedNodeID = seedNodeId

	configFromFile.MnemonicFile = signingMnemonic

	configFromFile.AddrBooks = addrBooks
	configFromFile.NodeKey = nodeKey
	configFromFile.TxModes = txModes

	configFromFile.NodeDiscovery.UseHttps = nodeDiscoveryUseHttps
	configFromFile.NodeDiscovery.DefaultInterxPort = nodeDiscoveryInterxPort
	configFromFile.NodeDiscovery.DefaultTendermintPort = nodeDiscoveryTendermintPort
	configFromFile.NodeDiscovery.ConnectionTimeout = nodeDiscoveryTimeout

	configFromFile.Block.StatusSync = syncStatus
	configFromFile.Block.HaltedAvgBlockTimes = haltedAvgBlockTimes

	configFromFile.Cache.CacheDir = cacheDir
	configFromFile.Cache.MaxCacheSize = maxCacheSize
	configFromFile.Cache.CacheDuration = cacheDuration
	configFromFile.Cache.DownloadFileSizeLimitation = maxDownloadSize

	configFromFile.Faucet.MnemonicFile = faucetMnemonic
	configFromFile.Faucet.TimeLimit = faucetTimeLimit

	configFromFile.Faucet.FaucetAmounts = make(map[string]string)
	for _, amount := range strings.Split(faucetAmounts, ",") {
		coin, err := sdk.ParseCoinNormalized(amount)
		if err == nil {
			configFromFile.Faucet.FaucetAmounts[coin.Denom] = coin.Amount.String()
		}
	}

	configFromFile.Faucet.FaucetMinimumAmounts = make(map[string]string)
	for _, amount := range strings.Split(faucetMinimumAmounts, ",") {
		coin, err := sdk.ParseCoinNormalized(amount)
		if err == nil {
			configFromFile.Faucet.FaucetMinimumAmounts[coin.Denom] = coin.Amount.String()
		}
	}

	configFromFile.Faucet.FeeAmounts = make(map[string]string)
	for _, denom_amount := range strings.Split(feeAmounts, ",") {
		denom := strings.Split(denom_amount, " ")[0]
		amount := strings.Split(denom_amount, " ")[1]
		configFromFile.Faucet.FeeAmounts[denom] = amount
	}

	configFromFile.SnapshotInterval = snapshotInterval

	bytes, err := json.MarshalIndent(&configFromFile, "", "  ")
	if err != nil {
		log.CustomLogger().Error("[InitConfig] Failed to marshal config.",
			"error", err,
		)
		panic(err)
	}

	err = ioutil.WriteFile(configFilePath, bytes, 0644)
	if err != nil {
		log.CustomLogger().Error("[InitConfig] Failed to write to file.",
			"config_File_Path", configFilePath,
			"error", err,
		)
		panic(err)
	}
}
