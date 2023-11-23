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

func TestBlockQueryTestSuite(t *testing.T) {
	testSuite := new(BlockQueryTestSuite)
	resBytes, err := tmjson.Marshal(tmRPCTypes.ResultBlock{
		Block: &tmTypes.Block{
			Header: tmTypes.Header{
				Time: time.Now(),
			},
		},
	})

	if err != nil {
		panic(err)
	}

	testSuite.blockQueryResponse.Result = resBytes

	resBytes, err = tmjson.Marshal(tmRPCTypes.ResultTxSearch{
		TotalCount: 1,
	})

	if err != nil {
		panic(err)
	}
	testSuite.blockTransactionsQueryResponse.Result = resBytes

	tendermintServer := http.Server{
		Addr: ":26657",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/blockchain" {
				response, _ := tmjson.Marshal(testSuite.blockQueryResponse)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write(response)
				if err != nil {
					panic(err)
				}
			} else if r.URL.Path == "/block" {
				response, _ := tmjson.Marshal(testSuite.blockQueryResponse)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write(response)
				if err != nil {
					panic(err)
				}
			} else if r.URL.Path == "/tx_search" {
				response := tmJsonRPCTypes.RPCResponse{
					JSONRPC: "2.0",
					Result:  []byte(`{"txs":[{"hash":"DE0CAB9BF94391C2562A0AA2784BB8E9A75031B719ED9D144683D008D24BFD40","height":"127","index":0,"tx_result":{"code":0,"data":"Ch4KHC9jb3Ntb3MuYmFuay52MWJldGExLk1zZ1NlbmQ=","log":"[{\"events\":[{\"type\":\"coin_received\",\"attributes\":[{\"key\":\"receiver\",\"value\":\"kira1uttsny8adtugcvpwdewc9ykgsdez7xactughf0\"},{\"key\":\"amount\",\"value\":\"100ukex\"}]},{\"type\":\"coin_spent\",\"attributes\":[{\"key\":\"spender\",\"value\":\"kira1kvdklhm7kdmyhvga3ty7nwd2llz9q9hyq3lvfh\"},{\"key\":\"amount\",\"value\":\"100ukex\"}]},{\"type\":\"message\",\"attributes\":[{\"key\":\"action\",\"value\":\"/cosmos.bank.v1beta1.MsgSend\"},{\"key\":\"sender\",\"value\":\"kira1kvdklhm7kdmyhvga3ty7nwd2llz9q9hyq3lvfh\"},{\"key\":\"module\",\"value\":\"bank\"}]},{\"type\":\"transfer\",\"attributes\":[{\"key\":\"recipient\",\"value\":\"kira1uttsny8adtugcvpwdewc9ykgsdez7xactughf0\"},{\"key\":\"sender\",\"value\":\"kira1kvdklhm7kdmyhvga3ty7nwd2llz9q9hyq3lvfh\"},{\"key\":\"amount\",\"value\":\"100ukex\"}]}]}]","info":"","gas_wanted":"0","gas_used":"0","events":[{"type":"tx","attributes":[{"key":"YWNjX3NlcQ==","value":"a2lyYTFrdmRrbGhtN2tkbXlodmdhM3R5N253ZDJsbHo5cTloeXEzbHZmaC8w","index":true}]},{"type":"tx","attributes":[{"key":"c2lnbmF0dXJl","value":"VENzcit4NDRQU1FOVFh1U2JIZElMYlZOWUwzV2h2VTdwQVlUTUFDWHpQZ3dxMU9VdFNnTit3RFB1ZjhyQmlZNkRmTDlncVZUVk5ZUitVc1NMRkc0TVE9PQ==","index":true}]},{"type":"coin_spent","attributes":[{"key":"c3BlbmRlcg==","value":"a2lyYTFrdmRrbGhtN2tkbXlodmdhM3R5N253ZDJsbHo5cTloeXEzbHZmaA==","index":true},{"key":"YW1vdW50","value":"MTAwdWtleA==","index":true}]},{"type":"coin_received","attributes":[{"key":"cmVjZWl2ZXI=","value":"a2lyYTE3eHBmdmFrbTJhbWc5NjJ5bHM2Zjg0ejNrZWxsOGM1bHFrZncycw==","index":true},{"key":"YW1vdW50","value":"MTAwdWtleA==","index":true}]},{"type":"transfer","attributes":[{"key":"cmVjaXBpZW50","value":"a2lyYTE3eHBmdmFrbTJhbWc5NjJ5bHM2Zjg0ejNrZWxsOGM1bHFrZncycw==","index":true},{"key":"c2VuZGVy","value":"a2lyYTFrdmRrbGhtN2tkbXlodmdhM3R5N253ZDJsbHo5cTloeXEzbHZmaA==","index":true},{"key":"YW1vdW50","value":"MTAwdWtleA==","index":true}]},{"type":"message","attributes":[{"key":"c2VuZGVy","value":"a2lyYTFrdmRrbGhtN2tkbXlodmdhM3R5N253ZDJsbHo5cTloeXEzbHZmaA==","index":true}]},{"type":"tx","attributes":[{"key":"ZmVl","value":"MTAwdWtleA==","index":true}]},{"type":"message","attributes":[{"key":"YWN0aW9u","value":"L2Nvc21vcy5iYW5rLnYxYmV0YTEuTXNnU2VuZA==","index":true}]},{"type":"coin_spent","attributes":[{"key":"c3BlbmRlcg==","value":"a2lyYTFrdmRrbGhtN2tkbXlodmdhM3R5N253ZDJsbHo5cTloeXEzbHZmaA==","index":true},{"key":"YW1vdW50","value":"MTAwdWtleA==","index":true}]},{"type":"coin_received","attributes":[{"key":"cmVjZWl2ZXI=","value":"a2lyYTF1dHRzbnk4YWR0dWdjdnB3ZGV3Yzl5a2dzZGV6N3hhY3R1Z2hmMA==","index":true},{"key":"YW1vdW50","value":"MTAwdWtleA==","index":true}]},{"type":"transfer","attributes":[{"key":"cmVjaXBpZW50","value":"a2lyYTF1dHRzbnk4YWR0dWdjdnB3ZGV3Yzl5a2dzZGV6N3hhY3R1Z2hmMA==","index":true},{"key":"c2VuZGVy","value":"a2lyYTFrdmRrbGhtN2tkbXlodmdhM3R5N253ZDJsbHo5cTloeXEzbHZmaA==","index":true},{"key":"YW1vdW50","value":"MTAwdWtleA==","index":true}]},{"type":"message","attributes":[{"key":"c2VuZGVy","value":"a2lyYTFrdmRrbGhtN2tkbXlodmdhM3R5N253ZDJsbHo5cTloeXEzbHZmaA==","index":true}]},{"type":"message","attributes":[{"key":"bW9kdWxl","value":"YmFuaw==","index":true}]}],"codespace":""},"tx":"CooBCocBChwvY29zbW9zLmJhbmsudjFiZXRhMS5Nc2dTZW5kEmcKK2tpcmExa3Zka2xobTdrZG15aHZnYTN0eTdud2QybGx6OXE5aHlxM2x2ZmgSK2tpcmExdXR0c255OGFkdHVnY3Zwd2Rld2M5eWtnc2Rlejd4YWN0dWdoZjAaCwoEdWtleBIDMTAwEmMKTgpGCh8vY29zbW9zLmNyeXB0by5zZWNwMjU2azEuUHViS2V5EiMKIQMyhuutfXlSrOPslBJJa94LMTSe2koeuVQIh+f5UKsF/xIECgIIARIRCgsKBHVrZXgSAzEwMBDAmgwaQEwrK/seOD0kDU17kmx3SC21TWC91ob1O6QGEzAAl8z4MKtTlLUoDfsAz7n/KwYmOg3y/YKlU1TWEflLEixRuDE="}],"total_count":"1"}`),
				}
				response1, err := tmjson.Marshal(response)
				if err != nil {
					panic(err)
				}
				w.Header().Set("Content-Type", "application/json")
				_, err = w.Write(response1)
				if err != nil {
					panic(err)
				}
			}
		}),
	}
	go func() {
		_ = tendermintServer.ListenAndServe()
	}()

	suite.Run(t, testSuite)

	tendermintServer.Close()
}
