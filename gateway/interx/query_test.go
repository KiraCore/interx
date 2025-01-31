package interx

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/database"
	"github.com/KiraCore/interx/test"
	"github.com/KiraCore/interx/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/p2p"
	tmRPCTypes "github.com/cometbft/cometbft/rpc/core/types"
	tmJsonRPCTypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
	tmTypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/go-bip39"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

type DashboardInfo struct {
	ConsensusHealth       string `json:"consensus_health"`
	CurrentBlockValidator struct {
		Moniker string `json:"moniker"`
		Address string `json:"address"`
	} `json:"current_block_validator"`
	Validators struct {
		ActiveValidators   int `json:"active_validators"`
		PausedValidators   int `json:"paused_validators"`
		InactiveValidators int `json:"inactive_validators"`
		JailedValidators   int `json:"jailed_validators"`
		TotalValidators    int `json:"total_validators"`
		WaitingValidators  int `json:"waiting_validators"`
	} `json:"validators"`
	Blocks struct {
		CurrentHeight       int     `json:"current_height"`
		SinceGenesis        int     `json:"since_genesis"`
		PendingTransactions int     `json:"pending_transactions"`
		CurrentTransactions int     `json:"current_transactions"`
		LatestTime          float64 `json:"latest_time"`
		AverageTime         float64 `json:"average_time"`
	} `json:"blocks"`
	Proposals struct {
		Total      int    `json:"total"`
		Active     int    `json:"active"`
		Enacting   int    `json:"enacting"`
		Finished   int    `json:"finished"`
		Successful int    `json:"successful"`
		Proposers  string `json:"proposers"`
		Voters     string `json:"voters"`
	} `json:"proposals"`
}

type StatusTestSuite struct {
	suite.Suite

	blockHeightQueryResponse tmJsonRPCTypes.RPCResponse
	blockQueryResponse       tmJsonRPCTypes.RPCResponse
	netInfoQueryResponse     tmJsonRPCTypes.RPCResponse
	consensusQueryResponse   types.RPCResponse
	kiraStatusResponse       types.RPCResponse
}

func (suite *StatusTestSuite) SetupTest() {
}

func (suite *StatusTestSuite) TestDashboardQuery() {
	config.Config.Cache.CacheDir = "./"
	config.Config.RPC = test.TENDERMINT_RPC
	_ = os.Mkdir("./db", 0777)

	database.LoadBlockDbDriver()
	database.LoadBlockNanoDbDriver()
	common.NodeStatus.Block = 100
	common.NodeStatus.Blocktime = time.Now().String()

	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	q := r.URL.Query()
	q.Add("account", "test_account")
	q.Add("limit", "100")
	r.URL.RawQuery = q.Encode()

	gwCosmosmux, err := GetGrpcServeMux(*addr)
	if err != nil {
		panic("failed to serve grpc")
	}
	res, _, statusCode := queryDashboardHandler(test.TENDERMINT_RPC, r, gwCosmosmux)

	interxRes := DashboardInfo{}
	bz, err := json.Marshal(res)
	if err != nil {
		panic("parse error")
	}

	err = json.Unmarshal(bz, &interxRes)
	if err != nil {
		panic(err)
	}

	suite.Require().EqualValues(statusCode, http.StatusOK)
	os.RemoveAll("./db")
}

func (suite *StatusTestSuite) TestAddrBookQuery() {
	err := os.Mkdir("./config", 0777)
	if err != nil {
		panic(err)
	}

	_, err = os.Create("./config/test_addr.json")
	if err != nil {
		panic(err)
	}

	config.Config.AddrBooks = []string{
		"./config/test_addr.json",
	}
	_, _, statusCode := queryAddrBookHandler("")
	suite.Require().EqualValues(statusCode, http.StatusOK)
	err = os.RemoveAll("./config")
	if err != nil {
		panic(err)
	}
}

func (suite *StatusTestSuite) TestStatusHandler() {
	config.Config.Cache.CacheDir = "./"
	err := os.Mkdir("./reference", 0777)
	if err != nil {
		panic(err)
	}

	f, err := os.Create("./reference/genesis.json")
	if err != nil {
		panic(err)
	}

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

	_, err = f.WriteString(string(resBytes))
	if err != nil {
		panic(err)
	}

	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	if err != nil {
		panic(err)
	}
	master, ch := hd.ComputeMastersFromSeed(seed)
	priv, err := hd.DerivePrivateKeyForPath(master, ch, "44'/118'/0'/0/0")
	config.Config.PrivKey = &secp256k1.PrivKey{Key: priv}
	config.Config.PubKey = config.Config.PrivKey.PubKey()

	if err != nil {
		panic(err)
	}

	res, _, statusCode := queryStatusHandle(test.TENDERMINT_RPC)

	suite.Require().EqualValues(res.(types.InterxStatus).InterxInfo.Moniker, "test_moniker")
	suite.Require().EqualValues(statusCode, http.StatusOK)
	os.RemoveAll("./reference")
}

func (suite *StatusTestSuite) TestNetInfoHandler() {
	res, _, statusCode := queryNetInfoHandler(test.TENDERMINT_RPC)
	bz, _ := json.Marshal(res)

	tmRes := tmRPCTypes.ResultNetInfo{}
	suiteRes := tmRPCTypes.ResultNetInfo{}

	err := tmjson.Unmarshal(suite.netInfoQueryResponse.Result, &suiteRes)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(bz, &tmRes)
	if err != nil {
		panic(err)
	}

	suite.Require().EqualValues(suiteRes.NPeers, tmRes.NPeers)
	suite.Require().EqualValues(statusCode, http.StatusOK)
}

type StatusTestSuite1 struct {
	suite.Suite
	kiraStatusResponse       tmRPCTypes.ResultStatus
	netInfoQueryResponse     tmRPCTypes.ResultNetInfo
	consensusQueryResponse   tmRPCTypes.ResultDumpConsensusState
	blockQueryResponse       tmRPCTypes.ResultBlockchainInfo
	blockHeightQueryResponse tmRPCTypes.ResultBlock
}

func TestStatusTestSuite(t *testing.T) {
	testSuite := new(StatusTestSuite1)

	// ✅ FIXED: Use correct struct for NodeInfo
	testSuite.kiraStatusResponse = tmRPCTypes.ResultStatus{
		NodeInfo: p2p.DefaultNodeInfo{Moniker: "KIRA TEST LOCAL VALIDATOR NODE"},
		SyncInfo: tmRPCTypes.SyncInfo{LatestBlockHeight: 100, CatchingUp: true},
	}

	// ✅ FIXED: Use proper struct assignment, NOT raw bytes
	testSuite.netInfoQueryResponse = tmRPCTypes.ResultNetInfo{
		Listening: true,
		NPeers:    100,
		Peers:     make([]tmRPCTypes.Peer, 0), // ✅ Use correct type
	}

	testSuite.consensusQueryResponse = tmRPCTypes.ResultDumpConsensusState{
		RoundState: json.RawMessage(`{
			"height": "100",
			"last_commit": {
				"votes_bit_array": "test_votesbitarray"
			}
		}`),
	}

	testSuite.blockQueryResponse = tmRPCTypes.ResultBlockchainInfo{
		LastHeight: 100,
		BlockMetas: []*tmTypes.BlockMeta{
			{
				NumTxs: 10,
				Header: tmTypes.Header{
					Time: time.Now(),
				},
			},
			{
				Header: tmTypes.Header{
					Time: time.Now(),
				},
			},
		},
	}

	testSuite.blockHeightQueryResponse = tmRPCTypes.ResultBlock{
		Block: &tmTypes.Block{
			Header: tmTypes.Header{
				Time: time.Now(),
			},
		},
	}

	// ✅ Mock GRPC Server
	flag.Parse()
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("[Error] Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	go func() {
		_ = s.Serve(lis)
	}()

	// ✅ Mock Tendermint Server
	tmServer := http.Server{
		Addr: ":26657",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var response []byte
			var err error

			switch r.URL.Path {
			case "/status":
				response, err = json.Marshal(testSuite.kiraStatusResponse)
			case "/net_info":
				response, err = json.Marshal(testSuite.netInfoQueryResponse)
			case "/dump_consensus_state":
				response, err = json.Marshal(testSuite.consensusQueryResponse)
			case "/blockchain":
				response, err = json.Marshal(testSuite.blockQueryResponse)
			case "/block":
				response, err = json.Marshal(testSuite.blockHeightQueryResponse)
			default:
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}

			if err != nil {
				http.Error(w, "JSON Error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(response)
		}),
	}

	go func() {
		_ = tmServer.ListenAndServe()
	}()

	suite.Run(t, testSuite)

	tmServer.Close()
	s.Stop()
}
