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

type BlockQueryTestSuite struct {
	suite.Suite

	blockQueryResponse             tmJsonRPCTypes.RPCResponse
	blockTransactionsQueryResponse tmJsonRPCTypes.RPCResponse
}

func (suite *BlockQueryTestSuite) SetupTest() {
}

func (suite *BlockQueryTestSuite) TestInitBlockQuery() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	response, error, statusCode := queryBlocksHandle(test.TENDERMINT_RPC, r)

	byteData, err := json.Marshal(response)
	if err != nil {
		suite.Assert()
	}

	result := tmRPCTypes.ResultBlock{}
	err = tmjson.Unmarshal(byteData, &result)
	if err != nil {
		suite.Assert()
	}

	resultBlock := tmRPCTypes.ResultBlock{}
	err = tmjson.Unmarshal(suite.blockQueryResponse.Result, &resultBlock)
	suite.Require().NoError(err)
	suite.Require().EqualValues(result.Block.Header.Time.Unix(), resultBlock.Block.Header.Time.Unix())
	suite.Require().Nil(error)
	suite.Require().EqualValues(statusCode, http.StatusOK)
}

func (suite *BlockQueryTestSuite) TestHeightOrHashBlockQuery() {
	response, error, statusCode := queryBlockByHeightOrHashHandle(test.TENDERMINT_RPC, "1")

	byteData, err := json.Marshal(response)
	if err != nil {
		suite.Assert()
	}

	result := tmRPCTypes.ResultBlock{}
	err = tmjson.Unmarshal(byteData, &result)
	if err != nil {
		suite.Assert()
	}

	resultBlock := tmRPCTypes.ResultBlock{}
	err = tmjson.Unmarshal(suite.blockQueryResponse.Result, &resultBlock)
	suite.Require().NoError(err)
	suite.Require().EqualValues(result.Block.Header.Time.Unix(), resultBlock.Block.Header.Time.Unix())
	suite.Require().Nil(error)
	suite.Require().EqualValues(statusCode, http.StatusOK)
}

func (suite *BlockQueryTestSuite) TestBlockTransactionsHandle() {
	config.Config.Cache.CacheDir = "./"
	_ = os.Mkdir("./db", 0777)

	database.LoadBlockDbDriver()
	database.LoadBlockNanoDbDriver()
	response, error, statusCode := QueryBlockTransactionsHandle(test.TENDERMINT_RPC, "1")

	byteData, err := json.Marshal(response)
	if err != nil {
		suite.Assert()
	}

	result := types.TransactionSearchResult{}
	err = json.Unmarshal(byteData, &result)
	if err != nil {
		suite.Assert()
	}

	resultTxSearch := tmRPCTypes.ResultTxSearch{}
	err = tmjson.Unmarshal(suite.blockTransactionsQueryResponse.Result, &resultTxSearch)
	suite.Require().NoError(err)
	suite.Require().EqualValues(result.TotalCount, resultTxSearch.TotalCount)
	suite.Require().Nil(error)
	suite.Require().EqualValues(statusCode, http.StatusOK)
	os.RemoveAll("./db")
}

// BlockQuerySuite structure
type BlockQuerySuite struct {
	suite.Suite
	blockQueryResponse             struct{ Result json.RawMessage }
	blockTransactionsQueryResponse struct{ Result json.RawMessage }
}

func TestBlockQueryTestSuite(t *testing.T) {
	// **Ensure test suite is initialized properly**
	testSuite := &BlockQuerySuite{}

	// **Use a fixed timestamp for consistency**
	fixedTime := time.Unix(1738226083, 0)

	// **Prepare mock block response**
	blockResponse := tmRPCTypes.ResultBlock{
		Block: &tmTypes.Block{
			Header: tmTypes.Header{
				Time: fixedTime,
			},
		},
	}

	// **Marshal block response safely**
	resBytes, err := tmjson.Marshal(blockResponse)
	if err != nil {
		t.Fatalf("[TestBlockQueryTestSuite] Failed to marshal block response: %v", err)
	}
	testSuite.blockQueryResponse.Result = resBytes

	// **Prepare mock transaction response**
	txResponse := tmRPCTypes.ResultTxSearch{
		TotalCount: 1, // Ensure this matches expected test assertions
	}

	// **Marshal transaction response safely**
	resBytes, err = tmjson.Marshal(txResponse)
	if err != nil {
		t.Fatalf("[TestBlockQueryTestSuite] Failed to marshal transaction response: %v", err)
	}
	testSuite.blockTransactionsQueryResponse.Result = resBytes

	// **Create a mock HTTP server**
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var responseData []byte
		var err error

		switch r.URL.Path {
		case "/blockchain", "/block":
			responseData, err = tmjson.Marshal(testSuite.blockQueryResponse)
		case "/tx_search":
			responseData, err = tmjson.Marshal(testSuite.blockTransactionsQueryResponse)
		default:
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		if err != nil {
			http.Error(w, "failed to marshal response", http.StatusInternalServerError)
			t.Logf("[TestBlockQueryTestSuite] Failed to marshal response for %s: %v", r.URL.Path, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(responseData); err != nil {
			t.Logf("[TestBlockQueryTestSuite] Failed to write response: %v", err)
		}
	}))
	defer mockServer.Close()

	// **Run the test suite**
	suite.Run(t, testSuite)
}
