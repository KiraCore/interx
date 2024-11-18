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
	"github.com/KiraCore/interx/global"
	"github.com/KiraCore/interx/test"
	"github.com/btcsuite/btcd/btcjson"
	jrpc "github.com/gumeniukcom/golang-jsonrpc2"
	"github.com/stretchr/testify/suite"
)

type BalancesQueryTestSuite struct {
	suite.Suite

	Address  string
	Chain    string
	Response BalancesResult
}

func (suite *BalancesQueryTestSuite) SetupTest() {
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

func (suite *BalancesQueryTestSuite) TestBalancesQuery() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	response, error, statusCode := queryBtcBalancesRequestHandle(r, suite.Chain, suite.Address)

	byteData, err := json.Marshal(response)
	if err != nil {
		suite.Assert()
	}

	result := BalancesResult{}
	err = json.Unmarshal(byteData, &result)
	if err != nil {
		suite.Assert()
	}
	suite.Require().NoError(err)
	suite.Require().Nil(error)
	suite.Require().EqualValues(statusCode, http.StatusOK)

	suite.Require().EqualValues(result.Tracking, suite.Response.Tracking)
	suite.Require().EqualValues(result.Blocks, suite.Response.Blocks)
	suite.Require().EqualValues(result.TxCount, suite.Response.TxCount)
	suite.Require().EqualValues(result.Wallet.Name, suite.Response.Wallet.Name)
	suite.Require().EqualValues(result.Wallet.Version, suite.Response.Wallet.Version)
	suite.Require().EqualValues(result.Wallet.Format, suite.Response.Wallet.Format)
	suite.Require().EqualValues(result.Wallet.Descriptors, suite.Response.Wallet.Descriptors)
	suite.Require().EqualValues(result.Wallet.Addresses, suite.Response.Wallet.Addresses)
	suite.Require().EqualValues(result.Balance.Confirmed, suite.Response.Balance.Confirmed)
	suite.Require().EqualValues(result.Balance.Denom, suite.Response.Balance.Denom)
	suite.Require().EqualValues(result.Balance.Decimals, suite.Response.Balance.Decimals)
	suite.Require().EqualValues(result.Scanning.Isscanning, suite.Response.Scanning.Isscanning)
	suite.Require().EqualValues(result.Scanning.Progress, suite.Response.Scanning.Progress)
	suite.Require().EqualValues(result.Scanning.Duration, suite.Response.Scanning.Duration)
}

func TestBalancesQueryTestSuite(t *testing.T) {
	testSuite := *new(BalancesQueryTestSuite)
	testSuite.Address = "tb1qmf2r4ylqhq2zs8xt0mnzhdz503l2x2s4p7x3wc"
	testSuite.Chain = "testnet"
	testSuite.Response = *new(BalancesResult)
	testSuite.Response.Tracking = false
	testSuite.Response.Blocks = 100
	testSuite.Response.TxCount = 100
	testSuite.Response.Wallet.AvoidReuse = true
	testSuite.Response.Wallet.Format = "format"
	testSuite.Response.Wallet.Version = 100
	testSuite.Response.Wallet.Addresses = []string{testSuite.Address}
	testSuite.Response.Wallet.Name = "name"
	testSuite.Response.Wallet.Descriptors = false
	testSuite.Response.Balance.Confirmed = 1.234
	testSuite.Response.Balance.Decimals = 8
	testSuite.Response.Balance.Denom = "satoshi"
	testSuite.Response.Scanning.Isscanning = true
	testSuite.Response.Scanning.Progress = 0.0011
	testSuite.Response.Scanning.Duration = 150
	global.AddressToWallet[testSuite.Address] = "tb1qmf2r4ylqhq2zs8xt0mnzhdz503l2x2s4p7x3wd"

	serv := jrpc.New()
	if err := serv.RegisterMethod("validateaddress", validateAddress); err != nil {
		panic(err)
	}
	if err := serv.RegisterMethod("getwalletinfo", getWalletInfo); err != nil {
		panic(err)
	}
	if err := serv.RegisterMethod("getbalances", getBalances); err != nil {
		panic(err)
	}
	if err := serv.RegisterMethod("listtransactions", listTransactions); err != nil {
		panic(err)
	}
	if err := serv.RegisterMethod("listunspent", listUnspent); err != nil {
		panic(err)
	}
	if err := serv.RegisterMethod("listaddressgroupings", listAddressGroupings); err != nil {
		panic(err)
	}
	if err := serv.RegisterMethod("getblockchaininfo", getBlockchainInfo); err != nil {
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

func getWalletInfo(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	result := GetWalletInfoResult{
		GetWalletInfoResult: btcjson.GetWalletInfoResult{
			TransactionCount: 100,
			AvoidReuse:       true,
			WalletVersion:    100,
			WalletName:       "name",
			Scanning: btcjson.ScanningOrFalse{
				Value: btcjson.ScanProgress{
					Progress: 0.0011,
					Duration: 150,
				},
			},
		},
		Format:      "format",
		Descriptors: false,
	}

	mdata, err := json.Marshal(result)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

func getBalances(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	result := btcjson.GetBalancesResult{}

	mdata, err := json.Marshal(result)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

func listTransactions(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	result := []btcjson.ListTransactionsResult{}

	mdata, err := json.Marshal(result)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

func listUnspent(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	result := []btcjson.ListUnspentResult{}

	mdata, err := json.Marshal(result)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

func listAddressGroupings(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	result := [][][]interface{}{
		{
			{
				"tb1qmf2r4ylqhq2zs8xt0mnzhdz503l2x2s4p7x3wc",
				1.234,
			},
		},
	}

	mdata, err := json.Marshal(result)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

func getBlockchainInfo(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	result := btcjson.GetBlockChainInfoResult{
		Blocks: 100,
	}

	mdata, err := json.Marshal(result)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}
