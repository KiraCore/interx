package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/KiraCore/interx/types"
	sekaiapp "github.com/KiraCore/sekai/app"
	sekaiappparams "github.com/KiraCore/sekai/app/params"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/sr25519"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	bytesize "github.com/inhies/go-bytesize"
	"github.com/tyler-smith/go-bip39"
)

var (
	// Config is a configuration for interx
	Config = InterxConfig{}
	// EncodingCg is a configuration for Amino Encoding
	EncodingCg = sekaiapp.MakeEncodingConfig()
)

func parseSizeString(size string) int64 {
	b, _ := bytesize.Parse(size)
	return int64(b)
}

func mnemonicFromFile(filename string) string {
	if len(filename) == 0 {
		return ""
	}

	mnemonic, err := ioutil.ReadFile(filename)

	if err != nil {
		panic(err)
	}

	return string(mnemonic)
}

func TrimMnemonic(mnemonic string) string {
	return strings.Join(strings.Fields(mnemonic), " ")
}

// LoadMnemonic is a function to extract mnemonic
func LoadMnemonic(mnemonic string) string {
	if bip39.IsMnemonicValid(mnemonic) {
		return TrimMnemonic(mnemonic)
	}

	return TrimMnemonic(mnemonicFromFile(mnemonic))
}

// serveGRPC is a function to serve GRPC
func serveGRPC(r *http.Request, gwCosmosmux *runtime.ServeMux) (interface{}, interface{}, int) {
	recorder := httptest.NewRecorder()
	gwCosmosmux.ServeHTTP(recorder, r)
	resp := recorder.Result()

	result := new(interface{})
	if json.NewDecoder(resp.Body).Decode(result) == nil {
		if resp.StatusCode == http.StatusOK {
			return result, nil, resp.StatusCode
		}

		return nil, result, resp.StatusCode
	}

	return nil, nil, resp.StatusCode
}

// LoadAddressAndDenom is a function to load addresses and migrate config using custom bech32 and denom prefixes
func LoadAddressAndDenom(configFilePath string, gwCosmosmux *runtime.ServeMux, rpcAddr string, gatewayAddr string) {
	request, _ := http.NewRequest("GET", "http://"+gatewayAddr+"/kira/gov/custom_prefixes", nil)
	response, failure, _ := serveGRPC(request, gwCosmosmux)

	if response == nil {
		panic(failure)
	}

	byteData, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}
	result := map[string]string{}
	err = json.Unmarshal(byteData, &result)
	if err != nil {
		panic(err)
	}

	bech32Prefix := result["bech32Prefix"]
	defaultDenom := result["defaultDenom"]
	sekaiappparams.DefaultDenom = defaultDenom
	sekaiappparams.AccountAddressPrefix = bech32Prefix
	sekaiappparams.AccountPubKeyPrefix = bech32Prefix + "pub"
	sekaiappparams.ValidatorAddressPrefix = bech32Prefix + "valoper"
	sekaiappparams.ValidatorPubKeyPrefix = bech32Prefix + "valoperpub"
	sekaiappparams.ConsNodeAddressPrefix = bech32Prefix + "valcons"
	sekaiappparams.ConsNodePubKeyPrefix = bech32Prefix + "valconspub"

	sekaiappparams.SetConfig()
	file, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		fmt.Println("Invalid configuration: {}", err)
		panic(err)
	}

	configFromFile := InterxConfigFromFile{}

	err = json.Unmarshal([]byte(file), &configFromFile)
	if err != nil {
		fmt.Println("Invalid configuration: {}", err)
		panic(err)
	}

	//=============== interx address ===============
	Config.Mnemonic = LoadMnemonic(configFromFile.MnemonicFile)
	if !bip39.IsMnemonicValid(Config.Mnemonic) {
		fmt.Println("Invalid Interx Mnemonic: ", Config.Mnemonic)
		panic("Invalid Interx Mnemonic")
	}

	seed, err := bip39.NewSeedWithErrorChecking(Config.Mnemonic, "")
	if err != nil {
		panic(err)
	}
	master, ch := hd.ComputeMastersFromSeed(seed)
	priv, err := hd.DerivePrivateKeyForPath(master, ch, "44'/118'/0'/0/0")
	if err != nil {
		panic(err)
	}

	Config.PrivKey = &secp256k1.PrivKey{Key: priv}
	Config.PubKey = Config.PrivKey.PubKey()
	Config.Address = sdk.MustBech32ifyAddressBytes(sdk.GetConfig().GetBech32AccountAddrPrefix(), Config.PubKey.Address())

	// Display mnemonic and keys
	fmt.Println("Interx Address   : ", Config.Address)
	fmt.Println("Interx Public Key: ", Config.PubKey.String())

	//=============== faucet address ===============

	amount, found := configFromFile.Faucet.FaucetAmounts["ukex"]
	if found {
		configFromFile.Faucet.FaucetAmounts[defaultDenom] = amount
		delete(configFromFile.Faucet.FaucetAmounts, "ukex")
	}

	amount, found = configFromFile.Faucet.FaucetMinimumAmounts["ukex"]
	if found {
		configFromFile.Faucet.FaucetMinimumAmounts[defaultDenom] = amount
		delete(configFromFile.Faucet.FaucetMinimumAmounts, "ukex")
	}

	amount, found = configFromFile.Faucet.FeeAmounts["ukex"]
	if found {
		configFromFile.Faucet.FeeAmounts[defaultDenom] = amount
		delete(configFromFile.Faucet.FeeAmounts, "ukex")
	}

	for denom, coinStr := range configFromFile.Faucet.FeeAmounts {
		configFromFile.Faucet.FeeAmounts[denom] = strings.ReplaceAll(coinStr, "ukex", defaultDenom)
	}

	// Faucet Configuration
	Config.Faucet = FaucetConfig{
		Mnemonic:             LoadMnemonic(configFromFile.Faucet.MnemonicFile),
		FaucetAmounts:        configFromFile.Faucet.FaucetAmounts,
		FaucetMinimumAmounts: configFromFile.Faucet.FaucetMinimumAmounts,
		FeeAmounts:           configFromFile.Faucet.FeeAmounts,
		TimeLimit:            configFromFile.Faucet.TimeLimit,
	}

	if !bip39.IsMnemonicValid(Config.Faucet.Mnemonic) {
		fmt.Println("Invalid Faucet Mnemonic: ", Config.Faucet.Mnemonic)
		panic("Invalid Faucet Mnemonic")
	}

	seed, err = bip39.NewSeedWithErrorChecking(Config.Faucet.Mnemonic, "")
	if err != nil {
		panic(err)
	}
	master, ch = hd.ComputeMastersFromSeed(seed)
	priv, err = hd.DerivePrivateKeyForPath(master, ch, "44'/118'/0'/0/0")
	if err != nil {
		panic(err)
	}

	Config.Faucet.PrivKey = &secp256k1.PrivKey{Key: priv}
	Config.Faucet.PubKey = Config.Faucet.PrivKey.PubKey()
	Config.Faucet.Address = sdk.MustBech32ifyAddressBytes(sdk.GetConfig().GetBech32AccountAddrPrefix(), Config.Faucet.PubKey.Address())

	// Display mnemonic and keys
	fmt.Println("Faucet Address   : ", Config.Faucet.Address)
	fmt.Println("Faucet Public Key: ", Config.Faucet.PubKey.String())

	// Faucet configurations
	fmt.Println("Interx Faucet FaucetAmounts       : ", Config.Faucet.FaucetAmounts)
	fmt.Println("Interx Faucet FaucetMinimumAmounts: ", Config.Faucet.FaucetMinimumAmounts)
	fmt.Println("Interx Faucet FeeAmounts          : ", Config.Faucet.FeeAmounts)
	fmt.Println("Interx Faucet TimeLimit           : ", Config.Faucet.TimeLimit)

	// save denom changes to config
	bytes, err := json.MarshalIndent(&configFromFile, "", "  ")
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(configFilePath, bytes, 0644)
	if err != nil {
		panic(err)
	}
}

// LoadConfig is a function to load interx configurations from a given file
func LoadConfig(configFilePath string) {
	Config = InterxConfig{}

	file, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		fmt.Println("Invalid configuration: {}", err)
		panic(err)
	}

	configFromFile := InterxConfigFromFile{}

	err = json.Unmarshal([]byte(file), &configFromFile)
	if err != nil {
		fmt.Println("Invalid configuration: {}", err)
		panic(err)
	}

	// Interx Main Configuration
	Config.InterxVersion = InterxVersion
	Config.ServeHTTPS = configFromFile.ServeHTTPS
	Config.GRPC = configFromFile.GRPC
	Config.RPC = configFromFile.RPC
	Config.PORT = configFromFile.PORT

	Config.Node = configFromFile.Node

	fmt.Println("Interx Version: ", Config.InterxVersion)
	fmt.Println("Interx GRPC: ", Config.GRPC)
	fmt.Println("Interx RPC : ", Config.RPC)
	fmt.Println("Interx PORT: ", Config.PORT)

	Config.AddrBooks = strings.Split(configFromFile.AddrBooks, ",")
	Config.NodeKey, err = p2p.LoadOrGenNodeKey(configFromFile.NodeKey)
	if err != nil {
		panic(err)
	}

	Config.TxModes = strings.Split(configFromFile.TxModes, ",")
	if len(Config.TxModes) == 0 {
		Config.TxModes = strings.Split("sync,async,block", ",")
	}

	Config.NodeDiscovery = configFromFile.NodeDiscovery

	Config.Block.StatusSync = configFromFile.Block.StatusSync
	Config.Block.HaltedAvgBlockTimes = configFromFile.Block.HaltedAvgBlockTimes

	Config.Cache.CacheDir = configFromFile.Cache.CacheDir
	Config.Cache.MaxCacheSize = parseSizeString(configFromFile.Cache.MaxCacheSize)
	Config.Cache.CachingDuration = configFromFile.Cache.CachingDuration
	Config.Cache.DownloadFileSizeLimitation = parseSizeString(configFromFile.Cache.DownloadFileSizeLimitation)

	// Display cache configurations
	fmt.Println("Interx Block StatusSync                : ", Config.Block.StatusSync)
	fmt.Println("Halted Avg Block Times                 : ", Config.Block.HaltedAvgBlockTimes)

	fmt.Println("Interx Cache CacheDir                  : ", Config.Cache.CacheDir)
	fmt.Println("Interx Cache MaxCacheSize              : ", Config.Cache.MaxCacheSize)
	fmt.Println("Interx Cache CachingDuration           : ", Config.Cache.CachingDuration)
	fmt.Println("Interx Cache DownloadFileSizeLimitation: ", Config.Cache.DownloadFileSizeLimitation)

	// RPC Configuration
	Config.RPCMethods = getRPCSettings()

	if _, err := os.Stat(Config.Cache.CacheDir); os.IsNotExist(err) {
		err1 := os.Mkdir(Config.Cache.CacheDir, os.ModePerm)
		if err1 != nil {
			panic(err1)
		}
	}
	if _, err := os.Stat(Config.Cache.CacheDir + "/reference/"); os.IsNotExist(err) {
		err1 := os.Mkdir(Config.Cache.CacheDir+"/reference/", os.ModePerm)
		if err1 != nil {
			panic(err1)
		}
	}
	if _, err := os.Stat(Config.Cache.CacheDir + "/response/"); os.IsNotExist(err) {
		err1 := os.Mkdir(Config.Cache.CacheDir+"/response/", os.ModePerm)
		if err1 != nil {
			panic(err1)
		}
	}
	if _, err := os.Stat(Config.Cache.CacheDir + "/db/"); os.IsNotExist(err) {
		err1 := os.Mkdir(Config.Cache.CacheDir+"/db/", os.ModePerm)
		if err1 != nil {
			panic(err1)
		}
	}
	if _, err := os.Stat(GetReferenceCacheDir() + "/genesis.json"); !os.IsNotExist(err) {
		err1 := os.Remove(GetReferenceCacheDir() + "/genesis.json")
		if err1 != nil {
			panic(err1)
		}
	}

	if _, err := os.Stat(GetDbCacheDir() + "/token-aliases.json"); !os.IsNotExist(err) {
		err1 := os.Remove(GetDbCacheDir() + "/token-aliases.json")
		if err1 != nil {
			panic(err1)
		}
	}

	Config.Evm = configFromFile.Evm
	Config.Bitcoin = configFromFile.Bitcoin
}

// GenPrivKey is a function to generate a privKey
func GenPrivKey() crypto.PrivKey {
	return sr25519.GenPrivKey()
}

// GetReferenceCacheDir is a function to get reference directory
func GetReferenceCacheDir() string {
	return Config.Cache.CacheDir + "/reference"
}

// GetResponseCacheDir is a function to get reference directory
func GetResponseCacheDir() string {
	return Config.Cache.CacheDir + "/response"
}

// GetDbCacheDir is a function to get db directory
func GetDbCacheDir() string {
	return Config.Cache.CacheDir + "/db"
}

func LoadAddressBooks() []types.AddrBookJSON {
	addrBooks := make([]types.AddrBookJSON, 0)
	for _, addrFile := range Config.AddrBooks {
		file, _ := ioutil.ReadFile(addrFile)

		book := types.AddrBookJSON{}

		err := json.Unmarshal([]byte(file), &book)

		if err != nil {
			fmt.Println("Failed to load addrBook: ", addrFile)
		}

		addrBooks = append(addrBooks, book)
	}

	return addrBooks
}

func LoadUniqueIPAddresses() []string {
	ipAddresses := make([]string, 0)

	flag := make(map[string]bool)
	for _, addrFile := range Config.AddrBooks {
		file, _ := ioutil.ReadFile(addrFile)

		book := types.AddrBookJSON{}

		err := json.Unmarshal([]byte(file), &book)

		if err != nil {
			fmt.Println("Failed to load addrBook: ", addrFile)
		}

		for _, addr := range book.Addrs {
			if _, ok := flag[addr.Addr.IP]; !ok {
				ipAddresses = append(ipAddresses, addr.Addr.IP)
			}
			flag[addr.Addr.IP] = true
		}
	}

	return ipAddresses
}

func SnapshotPath() string {
	return GetReferenceCacheDir() + "/snapshot.tar"
}
