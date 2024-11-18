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
	jrpc "github.com/gumeniukcom/golang-jsonrpc2"
	"github.com/stretchr/testify/suite"
)

type TransactionsQueryTestSuite struct {
	suite.Suite

	Hash     string
	Chain    string
	Response SearchRawTransactionsResult
}

func (suite *TransactionsQueryTestSuite) SetupTest() {
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

func (suite *TransactionsQueryTestSuite) TestTransactionsQuery() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	response, error, statusCode := queryBtcTransactionRequestHandle(r, suite.Chain, suite.Hash)

	byteData, err := json.Marshal(response)
	if err != nil {
		suite.Assert()
	}

	result := SearchRawTransactionsResult{}
	err = json.Unmarshal(byteData, &result)
	if err != nil {
		suite.Assert()
	}
	suite.Require().NoError(err)
	suite.Require().Nil(error)
	suite.Require().EqualValues(statusCode, http.StatusOK)

	suite.Require().EqualValues(result.Size, suite.Response.Size)
	suite.Require().EqualValues(result.Weight, suite.Response.Weight)
	suite.Require().EqualValues(result.Version, suite.Response.Version)
	suite.Require().EqualValues(result.BlockConfirmations, suite.Response.BlockConfirmations)
	suite.Require().EqualValues(result.Hash, suite.Response.Hash)
}

func TestTransactionsQueryTestSuite(t *testing.T) {
	testSuite := *new(TransactionsQueryTestSuite)
	testSuite.Hash = "testhash"
	testSuite.Chain = "testnet"
	testSuite.Response = *new(SearchRawTransactionsResult)
	testSuite.Response.BlockConfirmations = "100+"
	testSuite.Response.Size = 100
	testSuite.Response.Weight = 200
	testSuite.Response.Version = 1
	testSuite.Response.Hash = "testhash"

	serv := jrpc.New()
	if err := serv.RegisterMethod("getrawtransaction", getRawTransaction); err != nil {
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

func getRawTransaction(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	result := SearchRawTransactionsResult{
		Confirmations: 1000,
		Size:          100,
		Weight:        200,
		Version:       1,
		Hash:          "testhash",
	}

	mdata, err := json.Marshal(result)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}
