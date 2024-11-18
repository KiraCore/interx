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

type AccountsQueryTestSuite struct {
	suite.Suite

	Address  string
	Chain    string
	Response AccountResult
}

func (suite *AccountsQueryTestSuite) SetupTest() {
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

func (suite *AccountsQueryTestSuite) TestAccountsQuery() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	response, error, statusCode := queryBtcAccountsRequestHandle(r, suite.Chain, suite.Address)

	byteData, err := json.Marshal(response)
	if err != nil {
		suite.Assert()
	}

	result := AccountResult{}
	err = json.Unmarshal(byteData, &result)
	if err != nil {
		suite.Assert()
	}
	suite.Require().NoError(err)
	suite.Require().Nil(error)
	suite.Require().EqualValues(statusCode, http.StatusOK)

	suite.Require().EqualValues(result.IsValid, suite.Response.IsValid)
	suite.Require().EqualValues(result.Address, suite.Response.Address)
	suite.Require().EqualValues(result.IsScript, suite.Response.IsScript)
	suite.Require().EqualValues(result.IsWitness, suite.Response.IsWitness)
	suite.Require().EqualValues(result.Witness.IsWitness, suite.Response.Witness.IsWitness)
	suite.Require().EqualValues(result.Witness.Program, suite.Response.Witness.Program)
	suite.Require().EqualValues(result.Witness.Version, suite.Response.Witness.Version)
}

func TestAccountsQueryTestSuite(t *testing.T) {
	testSuite := *new(AccountsQueryTestSuite)
	testSuite.Address = "tb1qmf2r4ylqhq2zs8xt0mnzhdz503l2x2s4p7x3wc"
	testSuite.Chain = "testnet"
	testSuite.Response = *new(AccountResult)
	testSuite.Response.IsValid = true
	testSuite.Response.Address = testSuite.Address
	testSuite.Response.Witness = AccountWitness{
		IsWitness: true,
		Program:   "testprogram",
	}

	serv := jrpc.New()
	if err := serv.RegisterMethod("validateaddress", validateAddress); err != nil {
		panic(err)
	}
	if err := serv.RegisterMethod("decodescript", decodeScript); err != nil {
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

func validateAddress(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	isWitness := true
	isScript := false
	program := "testprogram"
	version := int32(0)
	result := AccountResult{
		IsValid:        true,
		IsScript:       &isScript,
		Address:        "tb1qmf2r4ylqhq2zs8xt0mnzhdz503l2x2s4p7x3wc",
		IsWitness:      &isWitness,
		WitnessProgram: &program,
		WitnessVersion: &version,
	}

	mdata, err := json.Marshal(result)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

func decodeScript(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	result := AccountScript{
		IsScript: true,
		Address:  "tb1qmf2r4ylqhq2zs8xt0mnzhdz503l2x2s4p7x3wc",
		PubKey:   "pubkey",
		Asm:      "asm",
		Desc:     "desc",
	}

	mdata, err := json.Marshal(result)
	if err != nil {
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}
