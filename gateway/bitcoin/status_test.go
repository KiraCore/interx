package bitcoin

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/test"
	"github.com/btcsuite/btcd/btcjson"
	jrpc "github.com/gumeniukcom/golang-jsonrpc2"
	"github.com/stretchr/testify/suite"
)

type StatusQueryTestSuite struct {
	suite.Suite

	Address  string
	Chain    string
	Response BitcoinStatus
}

func (suite *StatusQueryTestSuite) SetupTest() {
	btcConfig := config.BitcoinConfig{}
	btcConfig.RPC = test.BTC_RPC
	btcConfig.RPC_CRED = "admin:1234"
	btcConfig.BTC_CONFIRMATIONS = 100
	btcConfig.BTC_MAX_RESCANS = 4
	btcConfig.BTC_WATCH_ADDRESSES = []string{}
	btcConfig.BTC_WALLETS = []string{}
	btcConfig.BTC_WATCH_REGEX = ""
	btcConfig.BTC_FAUCET = ""

	config.Config.Bitcoin = make(map[string]config.BitcoinConfig)
	config.Config.Bitcoin[suite.Chain] = btcConfig
}

func (suite *StatusQueryTestSuite) TestStatusQuery() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	response, error, statusCode := queryBitcoinStatusHandle(r, suite.Chain)

	byteData, err := json.Marshal(response)
	if err != nil {
		suite.Assert()
	}

	result := BitcoinStatus{}
	err = json.Unmarshal(byteData, &result)
	if err != nil {
		suite.Assert()
	}
	suite.Require().NoError(err)
	suite.Require().Nil(error)
	suite.Require().EqualValues(statusCode, http.StatusOK)

	suite.Require().EqualValues(result.NodeInfo.Version.Net, suite.Response.NodeInfo.Version.Net)
	suite.Require().EqualValues(result.NodeInfo.Version.Sub, suite.Response.NodeInfo.Version.Sub)
	suite.Require().EqualValues(result.NodeInfo.Version.Protocol, suite.Response.NodeInfo.Version.Protocol)
	suite.Require().EqualValues(result.GasPrice, suite.Response.GasPrice)
}

func TestStatusQueryTestSuite(t *testing.T) {
	testSuite := *new(StatusQueryTestSuite)
	testSuite.Address = "tb1qmf2r4ylqhq2zs8xt0mnzhdz503l2x2s4p7x3wc"
	testSuite.Chain = "testnet"
	testSuite.Response = *new(BitcoinStatus)
	testSuite.Response.NodeInfo.Version.Net = 1
	testSuite.Response.NodeInfo.Version.Sub = "1"
	testSuite.Response.NodeInfo.Version.Protocol = 11
	testSuite.Response.GasPrice = big.NewFloat(0.12 * 1e8).String()

	serv := jrpc.New()
	if err := serv.RegisterMethod("getblockchaininfo", getBlockchainInfo); err != nil {
		panic(err)
	}
	if err := serv.RegisterMethod("getnetworkinfo", getNetworkInfo); err != nil {
		panic(err)
	}
	if err := serv.RegisterMethod("getblockstats", getBlockStats); err != nil {
		panic(err)
	}
	if err := serv.RegisterMethod("estimatesmartfee", estimateSmartFee); err != nil {
		panic(err)
	}

	btcServer := http.Server{
		Addr: ":18332",
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
		_ = btcServer.ListenAndServe()
	}()

	time.Sleep(1 * time.Second)
	suite.Run(t, &testSuite)
	btcServer.Close()
}

func getNetworkInfo(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	result := btcjson.GetNetworkInfoResult{
		Version:         1,
		SubVersion:      "1",
		ProtocolVersion: 11,
		RelayFee:        0,
		IncrementalFee:  0,
	}

	mdata, err := json.Marshal(result)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

func estimateSmartFee(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	feeRate := 0.12
	result := btcjson.EstimateSmartFeeResult{
		FeeRate: &feeRate,
	}

	mdata, err := json.Marshal(result)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}
