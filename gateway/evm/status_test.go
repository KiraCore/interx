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

type StatusQueryTestSuite struct {
	suite.Suite

	Chain string
}

func (suite *StatusQueryTestSuite) SetupTest() {
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

func (suite *StatusQueryTestSuite) TestAccountsQuery() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	response, error, statusCode := queryEVMStatusHandle(r, suite.Chain)

	byteData, err := json.Marshal(response)
	if err != nil {
		suite.Assert()
	}

	result := EVMStatus{}
	err = json.Unmarshal(byteData, &result)
	if err != nil {
		suite.Assert()
	}
	suite.Require().NoError(err)
	suite.Require().Nil(error)
	suite.Require().EqualValues(statusCode, http.StatusOK)

	suite.Require().EqualValues(result.NodeInfo.Network, 1)
	suite.Require().EqualValues(result.NodeInfo.Version.Web3, "client-version")
	suite.Require().EqualValues(result.NodeInfo.Version.Net, "net-version")
	suite.Require().EqualValues(result.NodeInfo.Version.Protocol, "protocol-version")
	suite.Require().EqualValues(result.SyncInfo.CatchingUp, true)
	suite.Require().EqualValues(result.GasPrice, "1000000000")
}

func TestStatusQueryTestSuite(t *testing.T) {
	testSuite := *new(StatusQueryTestSuite)
	testSuite.Chain = "goerli"

	serv := jrpc.New()
	if err := serv.RegisterMethod("eth_getBlockByNumber", ethGetBlockByNumber); err != nil {
		panic(err)
	}
	if err := serv.RegisterMethod("eth_chainId", ethChainId); err != nil {
		panic(err)
	}
	if err := serv.RegisterMethod("web3_clientVersion", web3ClientVersion); err != nil {
		panic(err)
	}
	if err := serv.RegisterMethod("net_version", netVersion); err != nil {
		panic(err)
	}
	if err := serv.RegisterMethod("eth_protocolVersion", ethProtocolVersion); err != nil {
		panic(err)
	}
	if err := serv.RegisterMethod("eth_syncing", ethSyncing); err != nil {
		panic(err)
	}
	if err := serv.RegisterMethod("eth_gasPrice", ethGasPrice); err != nil {
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

func ethChainId(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	result := "0x1"

	mdata, err := json.Marshal(result)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

func web3ClientVersion(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	result := "client-version"

	mdata, err := json.Marshal(result)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

func netVersion(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	result := "net-version"

	mdata, err := json.Marshal(result)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

func ethProtocolVersion(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	result := "protocol-version"

	mdata, err := json.Marshal(result)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

func ethSyncing(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	result := "0x1"

	mdata, err := json.Marshal(result)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

func ethGasPrice(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	result := "0x" + fmt.Sprintf("%x", 1000000000)

	mdata, err := json.Marshal(result)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}
