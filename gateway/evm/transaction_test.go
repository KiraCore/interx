package evm

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/test"
	jrpc "github.com/gumeniukcom/golang-jsonrpc2"
	"github.com/stretchr/testify/suite"
)

type TransactionQueryResponse struct {
	Hash string `json:"hash"`
}

type TransactionQueryReceipt struct {
	Hash    string `json:"hash"`
	Receipt bool   `json:"receipt"`
}

type TransactionQueryTestSuite struct {
	suite.Suite

	chain    string
	hash     string
	Response TransactionQueryResponse
}

func (suite *TransactionQueryTestSuite) SetupTest() {
	evmConfig := config.EVMConfig{}
	evmConfig.Name = ""
	evmConfig.Infura.RPC = test.INFURA_RPC
	evmConfig.Infura.RPCToken = ""
	evmConfig.Infura.RPCSecret = ""
	evmConfig.QuickNode.RPC = test.INFURA_RPC
	evmConfig.QuickNode.RPCToken = ""
	evmConfig.Pokt.RPC = test.INFURA_RPC
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

	config.Config.Evm = make(map[string]config.EVMConfig)
	config.Config.Evm[suite.chain] = evmConfig
}

func (suite *TransactionQueryTestSuite) TestQueryTransactionByHash() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	r.URL.RawQuery = "receipt=false"
	response, error, statusCode := queryEVMTransactionRequestHandle(r, suite.chain, suite.hash)

	byteData, err := json.Marshal(response)
	if err != nil {
		suite.Assert()
	}

	result := TransactionQueryResponse{}
	err = json.Unmarshal(byteData, &result)
	if err != nil {
		suite.Assert()
	}
	suite.Require().NoError(err)
	suite.Require().Nil(error)
	suite.Require().EqualValues(statusCode, http.StatusOK)
	suite.Require().EqualValues(result.Hash, suite.hash)
}

func (suite *TransactionQueryTestSuite) TestQueryTransactionReceipt() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	r.URL.RawQuery = "receipt=true"
	response, error, statusCode := queryEVMTransactionRequestHandle(r, suite.chain, suite.hash)

	byteData, err := json.Marshal(response)
	if err != nil {
		suite.Assert()
	}

	result := TransactionQueryReceipt{}
	err = json.Unmarshal(byteData, &result)
	if err != nil {
		suite.Assert()
	}
	suite.Require().NoError(err)
	suite.Require().Nil(error)
	suite.Require().EqualValues(statusCode, http.StatusOK)
	suite.Require().EqualValues(result.Hash, suite.hash)
	suite.Require().EqualValues(result.Receipt, true)
}

func TestTransactionQueryTestSuite(t *testing.T) {
	testSuite := new(TransactionQueryTestSuite)
	testSuite.chain = "goerli"
	testSuite.hash = "0x0000000000000000000000000000000000000000000000000000000000000000"

	serv := jrpc.New()
	if err := serv.RegisterMethod("eth_getTransactionByHash", ethGetTransactionByHash); err != nil {
		panic(err)
	}
	if err := serv.RegisterMethod("eth_getTransactionReceipt", ethGetTransactionReceipt); err != nil {
		panic(err)
	}

	evmServer := http.Server{
		Addr: ":21000",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.Background()
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				panic(err)
			}
			defer r.Body.Close()

			w.Header().Set("Content-Type", "applicaition/json")
			w.WriteHeader(http.StatusOK)
			if _, err = w.Write(serv.HandleRPCJsonRawMessage(ctx, body)); err != nil {
				panic(err)
			}
		}),
	}
	go func() {
		_ = evmServer.ListenAndServe()
	}()

	time.Sleep(1 * time.Second)
	suite.Run(t, testSuite)
	evmServer.Close()
}

func ethGetTransactionByHash(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	transactionQueryResponse := TransactionQueryResponse{
		Hash: "0x0000000000000000000000000000000000000000000000000000000000000000",
	}
	mdata, err := json.Marshal(transactionQueryResponse)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

func ethGetTransactionReceipt(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	transactionQueryResponse := TransactionQueryReceipt{
		Hash:    "0x0000000000000000000000000000000000000000000000000000000000000000",
		Receipt: true,
	}
	mdata, err := json.Marshal(transactionQueryResponse)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}
