package interx

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/test"
	tmjson "github.com/cometbft/cometbft/libs/json"
	tmRPCTypes "github.com/cometbft/cometbft/rpc/core/types"
	tmJsonRPCTypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
	tmTypes "github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/suite"
)

type GenesisChecksumResponse struct {
	Checksum string `json:"checksum,omitempty"`
}

type GenesisQueryTestSuite struct {
	suite.Suite

	genesisQueryResponse tmJsonRPCTypes.RPCResponse
}

func (suite *GenesisQueryTestSuite) SetupTest() {
}

func (suite *GenesisQueryTestSuite) TestQueryGenesisSumHandler() {
	config.Config.Cache.CacheDir = "./"
	err := os.Mkdir("./reference", 0777)
	if err != nil {
		suite.Assert()
	}

	_, err = os.Create("./reference/genesis.json")
	if err != nil {
		suite.Assert()
	}

	response, _, statusCode := queryGenesisSumHandler(test.TENDERMINT_RPC)

	byteData, err := json.Marshal(response)
	if err != nil {
		suite.Assert()
	}

	result := GenesisChecksumResponse{}
	err = json.Unmarshal(byteData, &result)
	if err != nil {
		suite.Assert()
	}

	suite.Require().NoError(err)
	suite.Require().EqualValues(statusCode, http.StatusOK)
	os.RemoveAll("./reference")
}

func TestGenesisQueryTestSuite(t *testing.T) {
	testSuite := new(GenesisQueryTestSuite)
	resBytes, err := tmjson.Marshal(tmRPCTypes.ResultGenesis{
		Genesis: &tmTypes.GenesisDoc{
			GenesisTime:   time.Now(),
			ChainID:       "test",
			InitialHeight: 1,
		},
	})

	if err != nil {
		panic(err)
	}

	testSuite.genesisQueryResponse.Result = resBytes

	tendermintServer := http.Server{
		Addr: ":26657",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/genesis" {
				response, _ := tmjson.Marshal(testSuite.genesisQueryResponse)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write(response)
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
