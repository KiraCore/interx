package bitcoin

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/test"
	"github.com/btcsuite/btcd/btcjson"
	jrpc "github.com/gumeniukcom/golang-jsonrpc2"
	"github.com/stretchr/testify/suite"
)

type TransferQueryTestSuite struct {
	suite.Suite

	Chain     string
	Response1 SearchRawTransactionsResult
	Response2 DecodeScriptResponse
	Response3 RawTxResponse
	Response4 SendTxResult
}

func (suite *TransferQueryTestSuite) SetupTest() {
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

func (suite *TransferQueryTestSuite) TestRawTxDecodeQuery() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	q := r.URL.Query()
	q.Add("decode", "true")
	r.URL.RawQuery = q.Encode()

	form := url.Values{}
	form.Add("rawTx", "testrawtx")
	r.PostForm = form

	response, error, statusCode := queryBtcTransferRequestHandle(r, suite.Chain)

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

	suite.Require().EqualValues(result.Size, suite.Response1.Size)
	suite.Require().EqualValues(result.Weight, suite.Response1.Weight)
	suite.Require().EqualValues(result.Version, suite.Response1.Version)
	suite.Require().EqualValues(result.Hash, suite.Response1.Hash)
}

func (suite *TransferQueryTestSuite) TestScriptTxDecodeQuery() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	q := r.URL.Query()
	q.Add("decode", "true")
	r.URL.RawQuery = q.Encode()

	form := url.Values{}
	form.Add("scriptTx", "testscripttx")
	r.PostForm = form

	response, error, statusCode := queryBtcTransferRequestHandle(r, suite.Chain)

	byteData, err := json.Marshal(response)
	if err != nil {
		suite.Assert()
	}

	result := DecodeScriptResponse{}
	err = json.Unmarshal(byteData, &result)
	if err != nil {
		suite.Assert()
	}
	suite.Require().NoError(err)
	suite.Require().Nil(error)
	suite.Require().EqualValues(statusCode, http.StatusOK)

	suite.Require().EqualValues(result.Asm, suite.Response2.Asm)
	suite.Require().EqualValues(result.Hex, suite.Response2.Hex)
	suite.Require().EqualValues(result.Type, suite.Response2.Type)
	suite.Require().EqualValues(result.P2sh, suite.Response2.P2sh)
}

func (suite *TransferQueryTestSuite) TestCreateRawTx() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	q := r.URL.Query()
	q.Add("decode", "false")
	r.URL.RawQuery = q.Encode()

	form := url.Values{}
	form.Add("data", "{}")
	r.PostForm = form

	response, error, statusCode := queryBtcTransferRequestHandle(r, suite.Chain)

	byteData, err := json.Marshal(response)
	if err != nil {
		suite.Assert()
	}

	result := RawTxResponse{}
	err = json.Unmarshal(byteData, &result)
	if err != nil {
		suite.Assert()
	}
	suite.Require().NoError(err)
	suite.Require().Nil(error)
	suite.Require().EqualValues(statusCode, http.StatusOK)

	suite.Require().EqualValues(result.RawTx, suite.Response3.RawTx)
}

func (suite *TransferQueryTestSuite) TestSendRawTx() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	q := r.URL.Query()
	q.Add("decode", "false")
	r.URL.RawQuery = q.Encode()

	form := url.Values{}
	form.Add("rawTx", "testrawtx")
	r.PostForm = form

	response, error, statusCode := queryBtcTransferRequestHandle(r, suite.Chain)

	byteData, err := json.Marshal(response)
	if err != nil {
		suite.Assert()
	}

	result := SendTxResult{}
	err = json.Unmarshal(byteData, &result)
	if err != nil {
		suite.Assert()
	}
	suite.Require().NoError(err)
	suite.Require().Nil(error)
	suite.Require().EqualValues(statusCode, http.StatusOK)

	suite.Require().EqualValues(result.Txid, suite.Response4.Txid)
}

func TestTransferQueryTestSuite(t *testing.T) {
	testSuite := *new(TransferQueryTestSuite)
	testSuite.Chain = "testnet"
	testSuite.Response1 = *new(SearchRawTransactionsResult)
	testSuite.Response1.BlockConfirmations = "100+"
	testSuite.Response1.Size = 100
	testSuite.Response1.Weight = 200
	testSuite.Response1.Version = 1
	testSuite.Response1.Hash = "testhash"

	testSuite.Response2 = *new(DecodeScriptResponse)
	testSuite.Response2.Asm = "testasm"
	testSuite.Response2.Hex = "testhex"
	testSuite.Response2.Type = "testtype"
	testSuite.Response2.P2sh = "testsegwit"

	testSuite.Response3.RawTx = "resultrawtx"
	testSuite.Response4.Txid = "resulttxid"

	serv := jrpc.New()
	if err := serv.RegisterMethod("decoderawtransaction", decodeRawTransaction); err != nil {
		panic(err)
	}
	if err := serv.RegisterMethod("decodescript", decodeScriptForTx); err != nil {
		panic(err)
	}
	if err := serv.RegisterMethod("createrawtransaction", createRawTransaction); err != nil {
		panic(err)
	}
	if err := serv.RegisterMethod("sendrawtransaction", sendRawTransaction); err != nil {
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

func decodeRawTransaction(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	result := SearchRawTransactionsResult{
		Size:    100,
		Weight:  200,
		Version: 1,
		Hash:    "testhash",
	}

	mdata, err := json.Marshal(result)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

func decodeScriptForTx(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	result := DecodeScriptResult{}
	result.Segwit.Asm = "testasm"
	result.Segwit.Hex = "testhex"
	result.Segwit.Type = "testtype"
	result.Segwit.P2shSegwit = "testsegwit"

	mdata, err := json.Marshal(result)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

func createRawTransaction(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	result := "resultrawtx"
	mdata, err := json.Marshal(result)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

func sendRawTransaction(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	result := btcjson.TxRawResult{
		Txid: "resulttxid",
	}
	mdata, err := json.Marshal(result)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}
