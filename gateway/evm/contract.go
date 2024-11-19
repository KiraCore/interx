package evm

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	jsonrpc2 "github.com/KeisukeYamashita/go-jsonrpc"
	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/log"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/gorilla/mux"

	// "github.com/powerman/rpc-codec/jsonrpc2"
	goeth "github.com/ethereum/go-ethereum/common"
)

// RegisterEVMContractRoutes registers read/write of smart contract.
func RegisterEVMContractRoutes(r *mux.Router, rpcAddr string) {
	r.HandleFunc(config.QueryReadContract, QueryReadContractRequest(rpcAddr)).Methods("GET")
	r.HandleFunc(config.QueryWriteContract, QueryWriteContractRequest(rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryReadContract, "This is an API to read smart contract.", true)
	common.AddRPCMethod("GET", config.QueryWriteContract, "This is an API to write smart contract.", true)
}

func ReadContractCall(nodeInfo config.EVMNodeConfig, call *EVMCall, method abi.Method) (interface{}, error) {
	client := jsonrpc2.NewRPCClient(nodeInfo.RPC + "/" + nodeInfo.RPCToken)
	if nodeInfo.RPCSecret != "" {
		client.SetBasicAuth(nodeInfo.RPCToken, nodeInfo.RPCSecret)
	}

	data, err := client.Call("eth_call", *call, "latest")
	if err != nil {
		return nil, err
	}

	result, err := data.GetString()
	if err != nil {
		return nil, err
	}

	bytes, err := hex.DecodeString(result[2:])
	if err != nil {
		return nil, err
	}

	res, err := method.Outputs.Unpack(bytes)

	return res, err
}

func decodeAbiArgument(param string, arg *abi.Argument) (v interface{}, err error) {
	param = strings.TrimSpace(param)
	switch arg.Type.T {
	case abi.StringTy:
		str_val := new(string)
		v = str_val
		err = json.Unmarshal([]byte(param), v)
	case abi.UintTy, abi.IntTy:
		val := big.NewInt(0)
		_, success := val.SetString(param, 10)
		if !success {
			err = fmt.Errorf("invalid numeric (base 10) value: %v", param)
		}
		v = val
	case abi.AddressTy:
		if !((len(param) == (goeth.AddressLength*2 + 2)) || (len(param) == goeth.AddressLength*2)) {
			err = fmt.Errorf("invalid address length (%v), must be 40 (unprefixed) or 42 (prefixed) chars", len(param))
		} else {
			var addr goeth.Address
			if len(param) == (goeth.AddressLength*2 + 2) {
				addr = goeth.HexToAddress(param)
			} else {
				var data []byte
				data, err = hex.DecodeString(param)
				addr.SetBytes(data)
			}
			v = addr
		}
	case abi.HashTy:
		if !((len(param) == (goeth.HashLength*2 + 2)) || (len(param) == goeth.HashLength*2)) {
			err = fmt.Errorf("invalid hash length, must be 64 (unprefixed) or 66 (prefixed) chars")
		} else {
			var hash goeth.Hash
			if len(param) == (goeth.HashLength*2 + 2) {
				hash = goeth.HexToHash(param)
			} else {
				var data []byte
				data, err = hex.DecodeString(param)
				hash.SetBytes(data)
			}
			v = hash
		}
	case abi.BytesTy:
		if len(param) > 2 {
			if (param[0] == '0') && (param[1] == 'x') {
				param = param[2:] // cut 0x prefix
			}
		}
		decoded_bytes, tmperr := hex.DecodeString(param)
		v = decoded_bytes
		err = tmperr
	case abi.BoolTy:
		val := new(bool)
		v = val
		err = json.Unmarshal([]byte(param), v)
	default:
		err = fmt.Errorf("not supported parameter type: %v", arg.Type)
	}
	return v, err
}

func queryReadSmartContractHandle(r *http.Request, chain string, contract string) (interface{}, interface{}, int) {
	isSupportedChain, chainConfig := GetChainConfig(chain)
	if !isSupportedChain {
		return common.ServeError(0, "", "unsupported chain", http.StatusBadRequest)
	}

	_ = r.ParseForm()

	functionName := r.FormValue("function")
	if len(functionName) == 0 {
		return common.ServeError(0, "", "no function name", http.StatusBadRequest)
	}

	var abiDecoded abi.ABI
	abiJsonUrlEncoded := r.FormValue("abi")
	if len(abiJsonUrlEncoded) == 0 {
		result, err, statusCode := common.MakeGetRequest(chainConfig.Etherscan.API, "", "module=contract&action=getabi&address="+contract+"&apikey="+chainConfig.Etherscan.APIToken)
		if err != nil {
			return nil, err, statusCode
		}

		abiDecoded, err = abi.JSON(strings.NewReader(result.(map[string]interface{})["result"].(string)))
		if err != nil {
			return nil, err, http.StatusBadRequest
		}
	} else {
		abiJson, err := url.QueryUnescape(abiJsonUrlEncoded)
		if err != nil {
			return common.ServeError(0, "Invalid ABI JSON", err.Error(), http.StatusBadRequest)
		}

		abiDecoded, err = abi.JSON(strings.NewReader(abiJson))
		if err != nil {
			return common.ServeError(0, "Invalid ABI JSON", err.Error(), http.StatusBadRequest)
		}
	}

	method, exist := abiDecoded.Methods[functionName]
	if !exist {
		return common.ServeError(0, "", "function does not exist", http.StatusBadRequest)
	}

	argLen := len(method.Inputs)
	var args []interface{}
	for i := 0; i < argLen; i++ {
		arg := r.FormValue("key_" + strconv.Itoa(i+1))
		if len(arg) == 0 {
			return common.ServeError(0, "", "argument invalid", http.StatusBadRequest)
		}

		v, err := decodeAbiArgument(arg, &method.Inputs[i])
		if err != nil {
			return common.ServeError(0, "", "argument invalid", http.StatusBadRequest)
		}

		args = append(args, v)
	}

	bytes, err := abiDecoded.Pack(functionName, args...)
	if err != nil {
		return common.ServeError(0, "", err.Error(), http.StatusBadRequest)
	}

	call := new(EVMCall)
	call.To = contract
	call.Data = "0x" + hex.EncodeToString(bytes)

	res, err := ReadContractCall(chainConfig.QuickNode, call, method)
	if err != nil {
		res, err = ReadContractCall(chainConfig.Infura, call, method)
		if err != nil {
			res, err = ReadContractCall(chainConfig.Pokt, call, method)
		}
	}

	if err != nil {
		return common.ServeError(0, err.Error(), "contract call fail", http.StatusBadRequest)
	}

	return res, nil, http.StatusOK
}

// QueryReadContractRequest is a function to read smart contract.
func QueryReadContractRequest(rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		queries := mux.Vars(r)
		chain := queries["chain"]
		contract := queries["contract"]
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		log.CustomLogger().Info("`QueryReadContractRequest` Starting reading contract request...",
			"chain", chain,
		)

		if !common.RPCMethods["GET"][config.QueryReadContract].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryReadContract].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					log.CustomLogger().Info("`QueryReadContractRequest` Returning from the cache",
						"chain", chain,
					)
					return
				}
			}

			response.Response, response.Error, statusCode = queryReadSmartContractHandle(r, chain, contract)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryReadContract].CachingEnabled)
	}
}

func WriteContractCall(nodeInfo config.EVMNodeConfig, from string, contract string, value string, data string) (interface{}, interface{}, int) {
	client := jsonrpc2.NewRPCClient(nodeInfo.RPC + "/" + nodeInfo.RPCToken)
	if nodeInfo.RPCSecret != "" {
		client.SetBasicAuth(nodeInfo.RPCToken, nodeInfo.RPCSecret)
	}

	type UnsignedTransaction struct {
		From     string `json:"from"`
		To       string `json:"to"`
		ChainId  uint64 `json:"chainId"`
		Nonce    uint64 `json:"nonce"`
		GasLimit string `json:"gasLimit"`
		GasPrice string `json:"gasPrice"`
		Value    string `json:"value"`
		Data     string `json:"data"`
	}

	var response UnsignedTransaction
	response.From = from
	response.To = contract
	response.Value = value
	response.Data = data

	result, err := client.Call("eth_gasPrice")
	if err != nil {
		return common.ServeError(0, "failed to get gas price ", err.Error(), http.StatusInternalServerError)
	}
	gasPrice, err := result.GetString()
	if err != nil {
		return common.ServeError(0, "failed to get gas price ", err.Error(), http.StatusInternalServerError)
	}
	gasPriceBig := *new(big.Int)
	gasPriceBig.SetString((gasPrice)[2:], 16)
	response.GasPrice = gasPriceBig.String()

	result, err = client.Call("eth_chainId")
	if err != nil {
		return common.ServeError(0, "failed to get chain id", err.Error(), http.StatusInternalServerError)
	}
	chainId, err := result.GetString()
	if err != nil {
		return common.ServeError(0, "failed to get chain id", err.Error(), http.StatusInternalServerError)
	}
	response.ChainId, _ = strconv.ParseUint((chainId)[2:], 16, 64)

	result, err = client.Call("eth_getTransactionCount", from, "latest")
	if err != nil {
		return common.ServeError(0, "failed to get latest sequence", err.Error(), http.StatusInternalServerError)
	}
	sequence, err := result.GetString()
	if err != nil {
		return common.ServeError(0, "failed to get latest sequence", err.Error(), http.StatusInternalServerError)
	}
	response.Nonce, _ = strconv.ParseUint((sequence)[2:], 16, 64)

	type TransactionCall struct {
		From     string `json:"from"`
		To       string `json:"to"`
		GasPrice string `json:"gasPrice"`
		Value    string `json:"value"`
		Data     string `json:"data"`
	}
	var transactionCall TransactionCall
	transactionCall.From = from
	transactionCall.To = contract
	transactionCall.GasPrice = gasPrice
	transactionCall.Value = "0x" + hex.EncodeToString([]byte(value))
	transactionCall.Data = data
	log.CustomLogger().Info("transaction", transactionCall)

	result, err = client.Call("eth_estimateGas", transactionCall, "latest")
	if err != nil {
		return common.ServeError(0, "failed to estimate gas", err.Error(), http.StatusInternalServerError)
	}
	gasLimit, err := result.GetString()
	if err != nil {
		return common.ServeError(0, "failed to get gas limit ", err.Error(), http.StatusInternalServerError)
	}
	gasLimitBig := *new(big.Int)
	gasLimitBig.SetString((gasLimit)[2:], 16)
	response.GasLimit = gasLimitBig.String()

	return response, nil, http.StatusOK
}

func queryWriteSmartContractHandle(r *http.Request, chain string, contract string) (interface{}, interface{}, int) {
	isSupportedChain, chainConfig := GetChainConfig(chain)
	if !isSupportedChain {
		return common.ServeError(0, "", "unsupported chain", http.StatusBadRequest)
	}

	_ = r.ParseForm()

	functionName := r.FormValue("function")
	if len(functionName) == 0 {
		return common.ServeError(0, "", "no function name", http.StatusBadRequest)
	}

	from := r.FormValue("from")
	if len(from) == 0 {
		return common.ServeError(0, "", "no from address", http.StatusBadRequest)
	}

	value := r.FormValue("value")

	var abiDecoded abi.ABI
	abiJsonUrlEncoded := r.FormValue("abi")
	if len(abiJsonUrlEncoded) == 0 {
		result, err, statusCode := common.MakeGetRequest(chainConfig.Etherscan.API, "", "module=contract&action=getabi&address="+contract+"&apikey="+chainConfig.Etherscan.APIToken)
		if err != nil {
			return nil, err, statusCode
		}

		abiDecoded, err = abi.JSON(strings.NewReader(result.(map[string]interface{})["result"].(string)))
		if err != nil {
			return nil, err, http.StatusBadRequest
		}
	} else {
		abiJson, err := url.QueryUnescape(abiJsonUrlEncoded)
		if err != nil {
			return common.ServeError(0, "Invalid ABI JSON", err.Error(), http.StatusBadRequest)
		}

		abiDecoded, err = abi.JSON(strings.NewReader(abiJson))
		if err != nil {
			return common.ServeError(0, "Invalid ABI JSON", err.Error(), http.StatusBadRequest)
		}
	}

	method, exist := abiDecoded.Methods[functionName]
	if !exist {
		return common.ServeError(0, "", "function does not exist", http.StatusBadRequest)
	}

	argLen := len(method.Inputs)
	var args []interface{}
	for i := 0; i < argLen; i++ {
		arg := r.FormValue("key_" + strconv.Itoa(i+1))
		if len(arg) == 0 {
			return common.ServeError(0, "", "argument invalid", http.StatusBadRequest)
		}

		v, err := decodeAbiArgument(arg, &method.Inputs[i])
		if err != nil {
			return common.ServeError(0, "", "argument invalid", http.StatusBadRequest)
		}

		args = append(args, v)
	}

	bytes, err := abiDecoded.Pack(functionName, args...)
	if err != nil {
		return common.ServeError(0, "", err.Error(), http.StatusBadRequest)
	}

	res, fail, statusCode := WriteContractCall(chainConfig.QuickNode, from, contract, value, "0x"+hex.EncodeToString(bytes))
	if fail == nil {
		return res, fail, statusCode
	}

	res, fail, statusCode = WriteContractCall(chainConfig.Infura, from, contract, value, "0x"+hex.EncodeToString(bytes))
	if fail == nil {
		return res, fail, statusCode
	}

	res, fail, statusCode = WriteContractCall(chainConfig.Pokt, from, contract, value, "0x"+hex.EncodeToString(bytes))
	if fail == nil {
		return res, fail, statusCode
	}

	return res, fail, statusCode
}

// QueryWriteContractRequestf is a function to write smart contract.
func QueryWriteContractRequest(rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		queries := mux.Vars(r)
		chain := queries["chain"]
		contract := queries["contract"]
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		log.CustomLogger().Info("[query-evm-write-contract] Entering write smart contract: ", chain)

		if !common.RPCMethods["GET"][config.QueryWriteContract].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryWriteContract].CachingEnabled {
				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					log.CustomLogger().Info("[query-evm-write-contract] Returning from the cache: ", chain)
					return
				}
			}

			response.Response, response.Error, statusCode = queryWriteSmartContractHandle(r, chain, contract)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryWriteContract].CachingEnabled)
	}
}
