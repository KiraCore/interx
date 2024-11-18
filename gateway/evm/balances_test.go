package evm

import (
	"context"
	"encoding/hex"
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

type BalancesQueryTestSuite struct {
	suite.Suite

	Chain string
}

func (suite *BalancesQueryTestSuite) SetupTest() {
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

func (suite *BalancesQueryTestSuite) TestAccountsQuery() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	r.URL.RawQuery = "tokens=token"
	response, error, statusCode := queryEVMBalancesRequestHandle(r, suite.Chain, "0xaddr")

	byteData, err := json.Marshal(response)
	if err != nil {
		suite.Assert()
	}

	result := EVMBalancesResponse{}
	err = json.Unmarshal(byteData, &result)
	if err != nil {
		suite.Assert()
	}
	suite.Require().NoError(err)
	suite.Require().Nil(error)
	suite.Require().EqualValues(statusCode, http.StatusOK)

	suite.Require().EqualValues(len(result.Balances), 2)
	suite.Require().EqualValues(result.Balances[0].Contract, "")
	suite.Require().EqualValues(result.Balances[0].Symbol, "ETH")
	suite.Require().EqualValues(result.Balances[0].Decimals, 18)
	suite.Require().EqualValues(result.Balances[0].Amount, "1000000000000000000")
	suite.Require().EqualValues(result.Balances[1].Contract, "token")
	suite.Require().EqualValues(result.Balances[1].Symbol, "token")
	suite.Require().EqualValues(result.Balances[1].Decimals, 6)
	suite.Require().EqualValues(result.Balances[1].Amount, "1000000")
}

func TestBalancesQueryTestSuite(t *testing.T) {
	testSuite := *new(BalancesQueryTestSuite)
	testSuite.Chain = "goerli"

	serv := jrpc.New()
	if err := serv.RegisterMethod("eth_getCode", ethGetCode); err != nil {
		panic(err)
	}
	if err := serv.RegisterMethod("eth_getTransactionCount", ethGetTransactionCount); err != nil {
		panic(err)
	}
	if err := serv.RegisterMethod("eth_call", ethCall); err != nil {
		panic(err)
	}
	if err := serv.RegisterMethod("eth_getBalance", ethGetBalance); err != nil {
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

func ethCall(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	result := ""

	if string(data) == `[{"to":"token","data":"0x95d89b410000000000000000000000000000000000000000000000000000000000000000"},"latest"]` {
		// token symbol
		result = "0x00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000005" + hex.EncodeToString([]byte("token")) + "000000000000000000000000000000000000000000000000000000"
	} else if string(data) == `[{"to":"token","data":"0x313ce5670000000000000000000000000000000000000000000000000000000000000000"},"latest"]` {
		// token decimals
		result = "0x" + fmt.Sprintf("%x", 6)
	} else if string(data) == `[{"to":"token","data":"0x70a08231000000000000000000000000addr"},"latest"]` {
		// token balances
		result = "0x" + fmt.Sprintf("%x", 1000000)
	}

	mdata, err := json.Marshal(result)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

func ethGetBalance(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	result := "0x" + fmt.Sprintf("%x", 1000000000000000000)

	mdata, err := json.Marshal(result)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}
