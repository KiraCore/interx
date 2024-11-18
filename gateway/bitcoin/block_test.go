package bitcoin

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
	"github.com/btcsuite/btcd/btcjson"
	jrpc "github.com/gumeniukcom/golang-jsonrpc2"
	"github.com/stretchr/testify/suite"
)

type BlockQueryTestSuite struct {
	suite.Suite

	Hash     string
	Number   string
	Chain    string
	Response BlockResult
}

func (suite *BlockQueryTestSuite) SetupTest() {
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

func (suite *BlockQueryTestSuite) TestBlockQueryWithHash() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	response, error, statusCode := queryBitcoinBlockRequestHandle(r, suite.Chain, suite.Hash)

	byteData, err := json.Marshal(response)
	if err != nil {
		suite.Assert()
	}

	result := BlockResult{}
	err = json.Unmarshal(byteData, &result)
	if err != nil {
		suite.Assert()
	}
	suite.Require().NoError(err)
	suite.Require().Nil(error)
	suite.Require().EqualValues(statusCode, http.StatusOK)

	suite.Require().EqualValues(result.Height, suite.Response.Height)
	suite.Require().EqualValues(result.BlockConfirmations, suite.Response.BlockConfirmations)
	suite.Require().EqualValues(result.GetBlockVerboseResult.Hash, suite.Response.GetBlockVerboseResult.Hash)
}

func (suite *BlockQueryTestSuite) TestBlockQueryWithNumber() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	response, error, statusCode := queryBitcoinBlockRequestHandle(r, suite.Chain, suite.Number)

	byteData, err := json.Marshal(response)
	if err != nil {
		suite.Assert()
	}

	result := BlockResult{}
	err = json.Unmarshal(byteData, &result)
	if err != nil {
		suite.Assert()
	}
	suite.Require().NoError(err)
	suite.Require().Nil(error)
	suite.Require().EqualValues(statusCode, http.StatusOK)

	suite.Require().EqualValues(result.Height, suite.Response.Height)
	suite.Require().EqualValues(result.BlockConfirmations, suite.Response.BlockConfirmations)
	suite.Require().EqualValues(result.GetBlockVerboseResult.Hash, suite.Response.GetBlockVerboseResult.Hash)
}

func TestBlockQueryTestSuite(t *testing.T) {
	testSuite := *new(BlockQueryTestSuite)
	testSuite.Hash = "0xabc"
	testSuite.Number = "1"
	testSuite.Chain = "testnet"
	testSuite.Response = *new(BlockResult)
	testSuite.Response.BlockConfirmations = "100+"
	testSuite.Response.GetBlockVerboseResult.Hash = "0xabc"
	testSuite.Response.Height = 20
	testSuite.Response.Stats.AvgFee = 10
	testSuite.Response.Stats.TotalFee = 20

	serv := jrpc.New()
	if err := serv.RegisterMethod("getblock", getBlock); err != nil {
		panic(err)
	}
	if err := serv.RegisterMethod("getblockstats", getBlockStats); err != nil {
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

func getBlock(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	result := BlockResult{
		GetBlockVerboseResult: btcjson.GetBlockVerboseResult{
			Height: 20,
			Hash:   "0xabc",
		},
		Confirmations: 200,
	}

	mdata, err := json.Marshal(result)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

func getBlockStats(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	result := BlockStats{
		AvgFee:   10,
		TotalFee: 20,
	}

	mdata, err := json.Marshal(result)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}
