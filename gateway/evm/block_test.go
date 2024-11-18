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

type BlockQueryResponse struct {
	Hash      string `json:"hash"`
	Number    string `json:"number"`
	Timestamp string `json:"timestamp"`
}

type BlockQueryTestSuite struct {
	suite.Suite

	chain  string
	height string
	hash   string
}

func (suite *BlockQueryTestSuite) SetupTest() {
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

func (suite *BlockQueryTestSuite) TestQueryEVMBlockByHeight() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	response, error, statusCode := queryEVMBlockRequestHandle(r, suite.chain, suite.height)

	byteData, err := json.Marshal(response)
	if err != nil {
		suite.Assert()
	}

	result := BlockQueryResponse{}
	err = json.Unmarshal(byteData, &result)
	if err != nil {
		suite.Assert()
	}
	suite.Require().NoError(err)
	suite.Require().Nil(error)
	suite.Require().EqualValues(statusCode, http.StatusOK)
	suite.Require().EqualValues(result.Hash, suite.hash)
	suite.Require().EqualValues(result.Number, suite.height)
}

func (suite *BlockQueryTestSuite) TestQueryEVMBlockByHash() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	response, error, statusCode := queryEVMBlockRequestHandle(r, suite.chain, suite.hash)

	byteData, err := json.Marshal(response)
	if err != nil {
		suite.Assert()
	}

	result := BlockQueryResponse{}
	err = json.Unmarshal(byteData, &result)
	if err != nil {
		suite.Assert()
	}
	suite.Require().NoError(err)
	suite.Require().Nil(error)
	suite.Require().EqualValues(statusCode, http.StatusOK)

	suite.Require().EqualValues(result.Hash, suite.hash)
	suite.Require().EqualValues(result.Number, suite.height)
}

func TestBlockQueryTestSuite(t *testing.T) {
	testSuite := new(BlockQueryTestSuite)
	testSuite.chain = "goerli"
	testSuite.height = "0x0a"
	testSuite.hash = "0x0000000000000000000000000000000000000000000000000000000000000000"

	serv := jrpc.New()
	if err := serv.RegisterMethod("eth_getBlockByNumber", ethGetBlockByNumber); err != nil {
		panic(err)
	}
	if err := serv.RegisterMethod("eth_getBlockByHash", ethGetBlockByHash); err != nil {
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

func ethGetBlockByNumber(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	blockQueryResponse := BlockQueryResponse{
		Number:    "0x0a",
		Hash:      "0x0000000000000000000000000000000000000000000000000000000000000000",
		Timestamp: "0x00",
	}
	mdata, err := json.Marshal(blockQueryResponse)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

func ethGetBlockByHash(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	blockQueryResponse := BlockQueryResponse{
		Number:    "0x0a",
		Hash:      "0x0000000000000000000000000000000000000000000000000000000000000000",
		Timestamp: "0x0",
	}
	mdata, err := json.Marshal(blockQueryResponse)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}
