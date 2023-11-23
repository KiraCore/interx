package interx

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KiraCore/interx/tasks"
	"github.com/KiraCore/interx/test"
	"github.com/KiraCore/interx/types"
	multiStakingTypes "github.com/KiraCore/sekai/x/multistaking/types"
	slashingTypes "github.com/KiraCore/sekai/x/slashing/types"
	stakingTypes "github.com/KiraCore/sekai/x/staking/types"
	tmJsonRPCTypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

type ValidatorsTestSuite struct {
	suite.Suite

	dumpConsensusQuery tmJsonRPCTypes.RPCResponse
}

type slashingServer struct {
	slashingTypes.UnimplementedQueryServer
}

type multiStakingServer struct {
	multiStakingTypes.UnimplementedQueryServer
	multiStakingTypes.UnimplementedMsgServer
}

type ValidatorsStatus struct {
	Validators []types.QueryValidator `json:"validators,omitempty"`
	Pagination struct {
		Total int `json:"total,string,omitempty"`
	} `json:"pagination,omitempty"`
}

func (s *slashingServer) SigningInfos(ctx context.Context, req *slashingTypes.QuerySigningInfosRequest) (*slashingTypes.QuerySigningInfosResponse, error) {
	return &slashingTypes.QuerySigningInfosResponse{Validators: []stakingTypes.QueryValidator{
		{
			Address: "test_address",
		},
	}}, nil
}

func (s *multiStakingServer) StakingPools(ctx context.Context, req *multiStakingTypes.QueryStakingPoolsRequest) (*multiStakingTypes.QueryStakingPoolsResponse, error) {
	return &multiStakingTypes.QueryStakingPoolsResponse{Pools: []multiStakingTypes.StakingPool{
		{
			Id:      1,
			Enabled: true,
		},
	}}, nil
}

func (suite *ValidatorsTestSuite) SetupTest() {
	tasks.AllValidators = types.AllValidators{
		Status: struct {
			ActiveValidators   int `json:"active_validators"`
			PausedValidators   int `json:"paused_validators"`
			InactiveValidators int `json:"inactive_validators"`
			JailedValidators   int `json:"jailed_validators"`
			TotalValidators    int `json:"total_validators"`
			WaitingValidators  int `json:"waiting_validators"`
		}{
			ActiveValidators: 10,
		},
		Validators: []types.QueryValidator{
			{
				Address:           "test_addr",
				Valkey:            "test_valkey",
				Pubkey:            "test_pubkey",
				Proposer:          "test_proposer",
				Moniker:           "test_moniker",
				Status:            "test_status",
				StakingPoolId:     1,
				StakingPoolStatus: "ENABLED",
			},
		},
	}
}

func (suite *ValidatorsTestSuite) TestDumpConsensusStateHandler() {
	response, _, statusCode := queryDumpConsensusStateHandler(nil, nil, test.TENDERMINT_RPC)
	suite.Require().EqualValues(response, "test")
	suite.Require().EqualValues(statusCode, http.StatusOK)
}

func (suite *ValidatorsTestSuite) TestValidatorInfosQuery() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	response, error, statusCode := queryValidatorsHandle(r, nil, test.TENDERMINT_RPC)

	byteData, err := json.Marshal(response)
	if err != nil {
		suite.Assert()
	}

	result := struct {
		Validators []types.QueryValidator `json:"validators,omitempty"`
		Pagination struct {
			Total int `json:"total,string,omitempty"`
		} `json:"pagination,omitempty"`
	}{}

	err = json.Unmarshal(byteData, &result)
	suite.Require().NoError(err)
	suite.Require().EqualValues(len(result.Validators), 1)
	suite.Require().EqualValues(result.Validators[0].StakingPoolStatus, "ENABLED")
	suite.Require().EqualValues(result.Validators[0].StakingPoolId, 1)
	suite.Require().Nil(error)
	suite.Require().EqualValues(statusCode, http.StatusOK)
}

func (suite *ValidatorsTestSuite) TestSnapInfoQuery() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	q := r.URL.Query()
	q.Add("address", "test_addr")
	r.URL.RawQuery = q.Encode()
	response, _, statusCode := queryValidatorsHandle(r, nil, "")

	res := ValidatorsStatus{}
	bz, err := json.Marshal(response)
	if err != nil {
		panic("parse error")
	}

	err = json.Unmarshal(bz, &res)
	if err != nil {
		panic(err)
	}

	suite.Require().EqualValues(len(res.Validators), len(tasks.AllValidators.Validators))
	suite.Require().EqualValues(statusCode, http.StatusOK)
}

func TestValidatorsTestSuite(t *testing.T) {
	testSuite := new(ValidatorsTestSuite)

	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	slashingTypes.RegisterQueryServer(s, &slashingServer{})
	log.Printf("server listening at %v", lis.Addr())

	go func() {
		_ = s.Serve(lis)
	}()

	testSuite.dumpConsensusQuery.Result, _ = json.Marshal("test")
	tmServer := http.Server{
		Addr: ":26657",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/dump_consensus_state" {
				response, _ := json.Marshal(testSuite.dumpConsensusQuery)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write(response)
				if err != nil {
					panic(err)
				}
			}
		}),
	}
	go func() {
		_ = tmServer.ListenAndServe()
	}()

	suite.Run(t, testSuite)
	s.Stop()
	tmServer.Close()
}
