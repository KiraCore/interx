package evm

import (
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/database"
	"github.com/KiraCore/interx/log"
	"github.com/gorilla/mux"

	// "github.com/powerman/rpc-codec/jsonrpc2"
	jsonrpc2 "github.com/KeisukeYamashita/go-jsonrpc"
	"github.com/ethereum/go-ethereum"
	goEthCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// RegisterEVMFaucetRoutes registers query status of EVM chains.
func RegisterEVMFaucetRoutes(r *mux.Router, rpcAddr string) {
	r.HandleFunc(config.QueryEVMFaucet, RegisterEVMFaucetRequest(rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryEVMFaucet, "This is an API to faucet.", true)
}

func queryEVMFaucetFromNode(chain string, chainConfig *config.EVMConfig, nodeInfo config.EVMNodeConfig, address string, token string) (interface{}, interface{}, int) {
	client := jsonrpc2.NewRPCClient(nodeInfo.RPC + "/" + nodeInfo.RPCToken)
	if nodeInfo.RPCSecret != "" {
		client.SetBasicAuth(nodeInfo.RPCToken, nodeInfo.RPCSecret)
	}

	// check claim limit
	timeLeft := database.GetClaimTimeLeft(chain + address + token)
	if timeLeft > 0 {
		return common.ServeError(101, "", fmt.Sprintf("claim limit: %d second(s) left", timeLeft), http.StatusBadRequest)
	}

	privateKey, _ := crypto.HexToECDSA(chainConfig.Faucet.PrivateKey)
	publicKey := privateKey.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
	faucetAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	faucetAccountBalance := big.NewInt(0)
	Y := big.NewInt(0) // userBalance

	_X, ok := chainConfig.Faucet.FaucetAmounts[token]
	if !ok {
		return common.ServeError(0, "", "unsupported token", http.StatusBadRequest)
	}

	X, ok := new(big.Int).SetString(_X, 10)
	if !ok {
		return common.ServeError(0, "", "parse error", http.StatusBadRequest)
	}

	_M, ok := chainConfig.Faucet.FaucetMinimumAmounts[token]
	if !ok {
		return common.ServeError(0, "", "unsupported token", http.StatusBadRequest)
	}

	M, ok := new(big.Int).SetString(_M, 10)
	if !ok {
		return common.ServeError(0, "", "parse error", http.StatusBadRequest)
	}

	if token == "0x0000000000000000000000000000000000000000" {
		// faucet account eth balance
		data, err := client.Call("eth_getBalance", faucetAddress.String(), "latest")
		if err != nil {
			return common.ServeError(0, "failed to get eth balance of faucet account", err.Error(), http.StatusInternalServerError)
		}
		ethBalanceString, err := data.GetString()
		if err != nil {
			return common.ServeError(0, "failed to get eth balance of faucet account", err.Error(), http.StatusInternalServerError)
		}
		faucetAccountBalance, ok = faucetAccountBalance.SetString((ethBalanceString)[2:], 16)
		if !ok {
			return common.ServeError(0, "", "failed to parse string", http.StatusInternalServerError)
		}

		// user eth balance
		data, err = client.Call("eth_getBalance", address, "latest")
		if err != nil {
			return common.ServeError(0, "failed to get eth balance of user ", err.Error(), http.StatusInternalServerError)
		}
		ethBalanceString, err = data.GetString()
		if err != nil {
			return common.ServeError(0, "failed to get eth balance of user", err.Error(), http.StatusInternalServerError)
		}
		Y, ok = Y.SetString((ethBalanceString)[2:], 16)
		if !ok {
			return common.ServeError(0, "", "failed to parse string", http.StatusInternalServerError)
		}
	} else {

		// faucet account token balance
		call := new(EVMCall)
		call.To = token
		call.Data = "0x70a08231000000000000000000000000" + faucetAddress.String()[2:] // balanceOf
		data, err := client.Call("eth_call", *call, "latest")
		if err != nil {
			return common.ServeError(0, "failed to get token balances", err.Error(), http.StatusInternalServerError)
		}
		tokenBalancesString, err := data.GetString()
		if err != nil {
			return common.ServeError(0, "failed to get token balances", err.Error(), http.StatusInternalServerError)
		}
		if tokenBalancesString != "0x" {
			faucetAccountBalance, ok = faucetAccountBalance.SetString((tokenBalancesString)[2:], 16)
			if !ok {
				return common.ServeError(0, "", "failed to parse string", http.StatusInternalServerError)
			}
		}

		// user token balance
		call.To = token
		call.Data = "0x70a08231000000000000000000000000" + address[2:] // balanceOf
		data, err = client.Call("eth_call", *call, "latest")
		if err != nil {
			return common.ServeError(0, "failed to get token balances", err.Error(), http.StatusInternalServerError)
		}
		tokenBalancesString, err = data.GetString()
		if err != nil {
			return common.ServeError(0, "failed to get token balances", err.Error(), http.StatusInternalServerError)
		}
		if tokenBalancesString != "0x" {
			Y, ok = Y.SetString((tokenBalancesString)[2:], 16)
			if !ok {
				return common.ServeError(0, "", "failed to parse string", http.StatusInternalServerError)
			}
		}
	}

	if Y.Cmp(X) == 1 || Y.Cmp(X) == 0 {
		return common.ServeError(0, "", "the account already has enough balance", http.StatusInternalServerError)
	}

	Z := X.Sub(X, Y)

	if Z.Cmp(faucetAccountBalance) == 1 {
		return common.ServeError(0, "", "faucet account doesn't have enough balance", http.StatusInternalServerError)
	}

	if Z.Cmp(M) == -1 {
		return common.ServeError(0, "", "transfer amount exceed the minimum amount", http.StatusInternalServerError)
	}

	// transfer Z
	if token == "0x0000000000000000000000000000000000000000" {
		a, _ := rpc.DialHTTPWithClient(nodeInfo.RPC+"/"+nodeInfo.RPCToken, new(http.Client))
		if len(nodeInfo.RPCSecret) > 0 {
			a.SetHeader("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(nodeInfo.RPCToken+":"+nodeInfo.RPCSecret)))
		}
		client := ethclient.NewClient(a)

		nonce, err := client.PendingNonceAt(context.Background(), faucetAddress)
		if err != nil {
			return common.ServeError(0, "failed to get nonce", err.Error(), http.StatusInternalServerError)
		}

		value := Z                // in wei (1 eth)
		gasLimit := uint64(21000) // in units
		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			return common.ServeError(0, "failed to get gas price", err.Error(), http.StatusInternalServerError)
		}

		toAddress := goEthCommon.HexToAddress(address)
		var data []byte
		tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

		chainID, err := client.NetworkID(context.Background())
		if err != nil {
			return common.ServeError(0, "failed to get chain id", err.Error(), http.StatusInternalServerError)
		}

		signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
		if err != nil {
			return common.ServeError(0, "failed to get signed Tx", err.Error(), http.StatusInternalServerError)
		}

		err = client.SendTransaction(context.Background(), signedTx)
		if err != nil {
			return common.ServeError(0, "failed to broadcast Tx", err.Error(), http.StatusInternalServerError)
		}

		// add new claim
		database.AddNewClaim(chain+address+token, time.Now().UTC())

		return signedTx.Hash().Hex(), nil, http.StatusOK
	} else {
		a, _ := rpc.DialHTTPWithClient(nodeInfo.RPC+"/"+nodeInfo.RPCToken, new(http.Client))
		if len(nodeInfo.RPCSecret) > 0 {
			a.SetHeader("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(nodeInfo.RPCToken+":"+nodeInfo.RPCSecret)))
		}
		client := ethclient.NewClient(a)

		nonce, err := client.PendingNonceAt(context.Background(), faucetAddress)
		if err != nil {
			return common.ServeError(0, "failed to get nonce", err.Error(), http.StatusInternalServerError)
		}

		value := big.NewInt(0) // in wei (0 eth)
		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			return common.ServeError(0, "failed to get gas price", err.Error(), http.StatusInternalServerError)
		}

		toAddress := goEthCommon.HexToAddress(address)
		tokenAddress := goEthCommon.HexToAddress(token)

		transferFnSignature := []byte("transfer(address,uint256)")
		methodID := crypto.Keccak256(transferFnSignature)[:4]

		paddedAddress := goEthCommon.LeftPadBytes(toAddress.Bytes(), 32)

		amount := Z
		paddedAmount := goEthCommon.LeftPadBytes(amount.Bytes(), 32)

		var data []byte
		data = append(data, methodID...)
		data = append(data, paddedAddress...)
		data = append(data, paddedAmount...)

		gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
			From: faucetAddress,
			To:   &toAddress,
			Data: data,
		})
		if err != nil {
			return common.ServeError(0, "failed to get gas limit", err.Error(), http.StatusInternalServerError)
		}

		tx := types.NewTransaction(nonce, tokenAddress, value, gasLimit*2, gasPrice, data)

		chainID, err := client.NetworkID(context.Background())
		if err != nil {
			return common.ServeError(0, "failed to get chain id", err.Error(), http.StatusInternalServerError)
		}

		signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
		if err != nil {
			return common.ServeError(0, "failed to get signed Tx", err.Error(), http.StatusInternalServerError)
		}

		err = client.SendTransaction(context.Background(), signedTx)
		if err != nil {
			return common.ServeError(0, "failed to broadcast Tx", err.Error(), http.StatusInternalServerError)
		}

		// add new claim
		database.AddNewClaim(chain+address+token, time.Now().UTC())

		return signedTx.Hash().Hex(), nil, http.StatusOK
	}
}

func queryEVMFaucetInfoFromNode(chainConfig *config.EVMConfig, nodeInfo config.EVMNodeConfig) (interface{}, interface{}, int) {
	client := jsonrpc2.NewRPCClient(nodeInfo.RPC + "/" + nodeInfo.RPCToken)
	if nodeInfo.RPCSecret != "" {
		client.SetBasicAuth(nodeInfo.RPCToken, nodeInfo.RPCSecret)
	}

	privateKey, _ := crypto.HexToECDSA(chainConfig.Faucet.PrivateKey)
	publicKey := privateKey.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
	faucetAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// FaucetAccountInfo is a struct to be used for Faucet Account Info
	type FaucetAccountInfo struct {
		Address  string            `json:"address"`
		Balances map[string]string `json:"balances"`
	}
	faucetInfo := FaucetAccountInfo{}
	faucetInfo.Address = faucetAddress.String()
	faucetInfo.Balances = make(map[string]string)

	for k := range chainConfig.Faucet.FaucetAmounts {
		if k == "0x0000000000000000000000000000000000000000" {
			// faucet account eth balance
			data, err := client.Call("eth_getBalance", faucetAddress.String(), "latest")
			if err != nil {
				return common.ServeError(0, "failed to get eth balance of faucet account", err.Error(), http.StatusInternalServerError)
			}
			ethBalanceString, err := data.GetString()
			if err != nil {
				return common.ServeError(0, "failed to get eth balance of faucet account", err.Error(), http.StatusInternalServerError)
			}

			ethBalanceStringBig := *new(big.Int)
			ethBalanceStringBig.SetString((ethBalanceString)[2:], 16)
			faucetInfo.Balances[k] = ethBalanceStringBig.String()
		} else {
			// faucet account token balance
			call := new(EVMCall)
			call.To = k
			call.Data = "0x70a08231000000000000000000000000" + faucetAddress.String()[2:] // balanceOf
			data, err := client.Call("eth_call", *call, "latest")
			if err != nil {
				return common.ServeError(0, "failed to get token balances", err.Error(), http.StatusInternalServerError)
			}
			tokenBalancesString, err := data.GetString()
			if err != nil {
				return common.ServeError(0, "failed to get token balances", err.Error(), http.StatusInternalServerError)
			}
			if tokenBalancesString != "0x" {
				tokenBalancesStringBig := *new(big.Int)
				tokenBalancesStringBig.SetString((tokenBalancesString)[2:], 16)
				faucetInfo.Balances[k] = tokenBalancesStringBig.String()
			} else {
				faucetInfo.Balances[k] = "0"
			}
		}
	}

	return faucetInfo, nil, http.StatusOK
}

func queryEVMFaucetRequestHandle(r *http.Request, chain string) (interface{}, interface{}, int) {
	isSupportedChain, chainConfig := GetChainConfig(chain)
	if !isSupportedChain {
		return common.ServeError(0, "", "unsupported chain", http.StatusBadRequest)
	}

	_ = r.ParseForm()

	address := r.FormValue("claim")
	token := r.FormValue("token")

	if len(address) == 0 {
		res, err, statusCode := queryEVMFaucetInfoFromNode(chainConfig, chainConfig.QuickNode)
		if err == nil {
			return res, err, statusCode
		}

		res, err, statusCode = queryEVMFaucetInfoFromNode(chainConfig, chainConfig.Pokt)
		if err == nil {
			return res, err, statusCode
		}

		return queryEVMFaucetInfoFromNode(chainConfig, chainConfig.Infura)
	}
	if len(token) == 0 {
		token = "0x0000000000000000000000000000000000000000"
	}

	res, err, statusCode := queryEVMFaucetFromNode(chain, chainConfig, chainConfig.QuickNode, address, token)
	if err == nil {
		return res, err, statusCode
	}

	res, err, statusCode = queryEVMFaucetFromNode(chain, chainConfig, chainConfig.Pokt, address, token)
	if err == nil {
		return res, err, statusCode
	}

	return queryEVMFaucetFromNode(chain, chainConfig, chainConfig.Infura, address, token)
}

// RegisterEVMFaucetRequest is a function to faucet evm tokens
func RegisterEVMFaucetRequest(rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		queries := mux.Vars(r)
		chain := queries["chain"]
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		log.CustomLogger().Info("[query-evm-faucet] Entering transactions execute: ", chain)

		if !common.RPCMethods["GET"][config.QueryEVMFaucet].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryEVMFaucet].CacheEnabled {

				log.CustomLogger().Info("Starting search cache for `RegisterEVMFaucetRequest` request...")

				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					log.CustomLogger().Info("[query-evm-faucet] Returning from the cache: ", chain)
					return
				}
			}

			response.Response, response.Error, statusCode = queryEVMFaucetRequestHandle(r, chain)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryEVMBalances].CacheEnabled)
	}
}
