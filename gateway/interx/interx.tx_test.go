package interx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/database"
	"github.com/KiraCore/interx/test"
	"github.com/KiraCore/interx/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	tmRPCTypes "github.com/cometbft/cometbft/rpc/core/types"
	tmJsonRPCTypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
	tmTypes "github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/suite"
)

type InterxTxTestSuite struct {
	suite.Suite
	blockQueryResponse                   tmJsonRPCTypes.RPCResponse
	txQueryResponse                      tmJsonRPCTypes.RPCResponse
	blockTransactionsQueryResponse       tmJsonRPCTypes.RPCResponse
	unconfirmedTransactionsQueryResponse tmJsonRPCTypes.RPCResponse
}

type TransactionSearchResult struct {
	Transactions []types.TransactionResponse `json:"transactions"`
	TotalCount   int                         `json:"total_count"`
}

func (suite *InterxTxTestSuite) SetupTest() {
}

func (suite *InterxTxTestSuite) TestQueryUnconfirmedTransactionsHandler() {
	config.Config.Cache.CacheDir = "./"
	_ = os.Mkdir("./db", 0777)

	database.LoadBlockDbDriver()
	database.LoadBlockNanoDbDriver()
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	q := r.URL.Query()
	q.Add("account", "test_account")
	q.Add("limit", "100")
	r.URL.RawQuery = q.Encode()
	response, error, statusCode := queryUnconfirmedTransactionsHandler(test.TENDERMINT_RPC, r)

	byteData, err := json.Marshal(response)
	if err != nil {
		suite.Assert()
	}

	result := tmRPCTypes.ResultUnconfirmedTxs{}
	err = json.Unmarshal(byteData, &result)
	if err != nil {
		suite.Assert()
	}

	resultTxSearch := tmRPCTypes.ResultUnconfirmedTxs{}
	err = json.Unmarshal(suite.blockTransactionsQueryResponse.Result, &resultTxSearch)
	suite.Require().NoError(err)
	suite.Require().EqualValues(result.Total, resultTxSearch.Total)
	suite.Require().Nil(error)
	suite.Require().EqualValues(statusCode, http.StatusOK)
	err = os.RemoveAll("./db")
	if err != nil {
		suite.Assert()
	}
}

func (suite *InterxTxTestSuite) TestBlockTransactionsHandler() {
	config.Config.Cache.CacheDir = "./"
	_ = os.Mkdir("./db", 0777)

	database.LoadBlockDbDriver()
	database.LoadBlockNanoDbDriver()
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	q := r.URL.Query()
	q.Add("address", "test_account")
	q.Add("direction", "outbound")
	q.Add("page_size", "1")
	r.URL.RawQuery = q.Encode()
	response, error, statusCode := QueryBlockTransactionsHandler(test.TENDERMINT_RPC, r)

	byteData, err := json.Marshal(response)
	if err != nil {
		suite.Assert()
	}

	result := TransactionSearchResult{}
	err = json.Unmarshal(byteData, &result)
	if err != nil {
		suite.Assert()
	}

	resultTxSearch := TxsResponse{}
	err = json.Unmarshal(suite.blockTransactionsQueryResponse.Result, &resultTxSearch)
	suite.Require().NoError(err)
	suite.Require().EqualValues(result.TotalCount, resultTxSearch.TotalCount)
	suite.Require().EqualValues(len(result.Transactions[0].Txs), len(resultTxSearch.Transactions[0].Txs))
	suite.Require().Nil(error)
	suite.Require().EqualValues(statusCode, http.StatusOK)
	os.RemoveAll("./db")
}

func (suite *InterxTxTestSuite) TestBlockHeight() {
	height, err := getBlockHeight(test.TENDERMINT_RPC, "test_hash")
	suite.Require().EqualValues(height, 100)
	suite.Require().NoError(err)
}

type InterxTxSuite struct {
	suite.Suite
	txQueryResponse                      tmRPCTypes.ResultTx
	blockQueryResponse                   tmRPCTypes.ResultBlock
	blockTransactionsQueryResponse       TxsResponse
	unconfirmedTransactionsQueryResponse tmRPCTypes.ResultUnconfirmedTxs
}

func (suite *InterxTxSuite) SetupTest() {
	_, err := tmjson.Marshal(tmRPCTypes.ResultTx{
		Height: 100,
	})
	suite.Require().NoError(err)
	suite.txQueryResponse = tmRPCTypes.ResultTx{Height: 100}

	_, err = tmjson.Marshal(tmRPCTypes.ResultBlock{
		Block: &tmTypes.Block{
			Header: tmTypes.Header{
				Height: 100,
				Time:   time.Now(),
			},
		},
	})
	suite.Require().NoError(err)
	suite.blockQueryResponse = tmRPCTypes.ResultBlock{
		Block: &tmTypes.Block{
			Header: tmTypes.Header{
				Height: 100,
				Time:   time.Now(),
			},
		},
	}

	txMsg := map[string]interface{}{"type": "send"}

	suite.blockTransactionsQueryResponse = TxsResponse{
		TotalCount: 1,
		Transactions: []types.TransactionResponse{
			{
				Txs: []interface{}{txMsg},
			},
		},
	}

	suite.unconfirmedTransactionsQueryResponse = tmRPCTypes.ResultUnconfirmedTxs{
		Total: 0,
	}
}

func (suite *InterxTxSuite) TestHeightofBlock() {
	suite.EqualValues(int64(100), suite.blockQueryResponse.Block.Header.Height)
}

func (suite *InterxTxSuite) TestBlockTransactionsHandler() {
	suite.Equal(1, suite.blockTransactionsQueryResponse.TotalCount)
}

func TestInterxTxTestSuite(t *testing.T) {
	suite.Run(t, new(InterxTxSuite))
}

type TransactionResponse struct {
	Txs []interface{} `json:"txs"`
}
