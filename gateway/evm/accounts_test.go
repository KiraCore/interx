package evm

import (
	"context"
	"encoding/json"
	"fmt"
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

type AccountsQueryTestSuite struct {
	suite.Suite

	Address  string
	Chain    string
	Response EVMAccountResponse
}

func (suite *AccountsQueryTestSuite) SetupTest() {
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
	config.Config.Evm[suite.Chain] = evmConfig
}

func (suite *AccountsQueryTestSuite) TestAccountsQuery() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	response, error, statusCode := queryEVMAccountsRequestHandle(r, suite.Chain, suite.Address)

	byteData, err := json.Marshal(response)
	if err != nil {
		suite.Assert()
	}

	result := EVMAccountResponse{}
	err = json.Unmarshal(byteData, &result)
	if err != nil {
		suite.Assert()
	}
	suite.Require().NoError(err)
	suite.Require().Nil(error)
	suite.Require().EqualValues(statusCode, http.StatusOK)

	suite.Require().EqualValues(result.Account.Type, suite.Response.Account.Type)
	suite.Require().EqualValues(result.Account.Address, suite.Response.Account.Address)
	suite.Require().EqualValues(result.Account.Pending, suite.Response.Account.Pending)
	suite.Require().EqualValues(result.Account.Sequence, suite.Response.Account.Sequence)
	suite.Require().EqualValues(result.ContractCode, suite.Response.ContractCode)
}

func TestAccountsQueryTestSuite(t *testing.T) {
	testSuite := *new(AccountsQueryTestSuite)
	testSuite.Address = "0xaddr"
	testSuite.Chain = "goerli"
	testSuite.Response = *new(EVMAccountResponse)
	testSuite.Response.Account.Type = "contract"
	testSuite.Response.Account.Address = "0xaddr"
	testSuite.Response.Account.Pending = 10
	testSuite.Response.Account.Sequence = 9
	testSuite.Response.ContractCode = "code"

	serv := jrpc.New()
	if err := serv.RegisterMethod("eth_getCode", ethGetCode); err != nil {
		panic(err)
	}
	if err := serv.RegisterMethod("eth_getTransactionCount", ethGetTransactionCount); err != nil {
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
	suite.Run(t, &testSuite)
	evmServer.Close()
}

func ethGetCode(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	result := "code"
	mdata, err := json.Marshal(result)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

func ethGetTransactionCount(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	result := "0x" + fmt.Sprintf("%x", 10)

	if string(data) == `["0xaddr","latest"]` {
		result = "0x" + fmt.Sprintf("%x", 9)
	}

	mdata, err := json.Marshal(result)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}
