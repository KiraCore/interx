package kira

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

	tasks "github.com/KiraCore/interx/tasks"
	"github.com/KiraCore/interx/test"
	interxTypes "github.com/KiraCore/interx/types"
	multiStakingTypes "github.com/KiraCore/sekai/x/multistaking/types"
	tokenTypes "github.com/KiraCore/sekai/x/tokens/types"
	"github.com/cosmos/cosmos-sdk/types"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

type tokenServer struct {
	tokenTypes.UnimplementedQueryServer
}
type multiStakingServer struct {
	multiStakingTypes.UnimplementedQueryServer
}

type bankServer struct {
	bankTypes.UnimplementedQueryServer
	bankTypes.UnimplementedMsgServer
}

func (s *tokenServer) GetAllTokenRates(ctx context.Context, in *tokenTypes.AllTokenRatesRequest) (*tokenTypes.AllTokenRatesResponse, error) {
	return &tokenTypes.AllTokenRatesResponse{Data: []*tokenTypes.TokenRate{
		{
			Denom: "ukex",
		},
	}}, nil
}

func (s *bankServer) AllBalances(ctx context.Context, in *bankTypes.QueryAllBalancesRequest) (*bankTypes.QueryAllBalancesResponse, error) {
	return &bankTypes.QueryAllBalancesResponse{Balances: types.Coins{{Denom: "v1/ukex", Amount: types.NewInt(100)}}}, nil
}

func (s *multiStakingServer) StakingPoolDelegators(ctx context.Context, in *multiStakingTypes.QueryStakingPoolDelegatorsRequest) (*multiStakingTypes.QueryStakingPoolDelegatorsResponse, error) {
	return &multiStakingTypes.QueryStakingPoolDelegatorsResponse{
		Pool: multiStakingTypes.StakingPool{
			Id:        1,
			Validator: "test_validator",
			Enabled:   true,
			Slashed:   types.NewDec(1),
			TotalStakingTokens: []types.Coin{
				types.NewCoin("ukex", types.NewInt(100)),
			},
			TotalShareTokens: []types.Coin{
				types.NewCoin("v1/ukex", types.NewInt(100)),
			},
			TotalRewards: []types.Coin{},
			Commission:   types.NewDec(1),
		},
		Delegators: []string{
			"test_delegator",
		},
	}, nil
}

type MultistakingTestSuite struct {
	suite.Suite
}

func (suite *MultistakingTestSuite) SetupTest() {
}

func (suite *MultistakingTestSuite) TestQueryStakingPoolHandler() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	query := r.URL.Query()

	query.Add("validatorAddress", "test_validator")
	r.URL.RawQuery = query.Encode()

	tasks.PoolTokens = []string{"ukex"}
	tasks.AddrToValidator["test_validator"] = "test_valoper"
	gwCosmosmux, err := GetGrpcServeMux(*addr)
	if err != nil {
		panic("failed to serve grpc")
	}
	r.URL.Path = "/api/kira/staking-pool"
	response, _, statusCode := queryStakingPoolHandler(r, gwCosmosmux)

	res := interxTypes.QueryValidatorPoolResult{}
	bz, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(bz, &res)
	if err != nil {
		panic(err)
	}

	suite.Require().EqualValues(res.TotalDelegators, 1)
	suite.Require().EqualValues(res.ID, 1)
	suite.Require().EqualValues(res.Tokens, []string{"ukex"})
	suite.Require().EqualValues(res.Slashed, "1.0")
	suite.Require().EqualValues(res.Commission, "1.0")
	suite.Require().EqualValues(statusCode, http.StatusOK)
}

func (suite *MultistakingTestSuite) TestQueryDelegationsHandler() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	query := r.URL.Query()

	query.Add("delegatorAddress", "test_delegator")
	r.URL.RawQuery = query.Encode()

	valPool := interxTypes.ValidatorPool{
		ID:         1,
		Validator:  "test_val",
		Enabled:    true,
		Slashed:    "0",
		Commission: "0",
	}
	tasks.PoolTokens = []string{"ukex"}
	tasks.AllPools[1] = valPool
	tasks.PoolToValidator[1] = interxTypes.QueryValidator{
		Address: "test_address",
		Valkey:  "test_valkey",
		Moniker: "test_moniker",
		Website: "test_website",
		Logo:    "test_logo",
	}

	gwCosmosmux, err := GetGrpcServeMux(*addr)
	if err != nil {
		panic("failed to serve grpc")
	}
	r.URL.Path = "/api/kira/delegations"
	response, _, statusCode := queryDelegationsHandler(r, gwCosmosmux)

	res := interxTypes.QueryDelegationsResult{}
	bz, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(bz, &res)
	if err != nil {
		panic(err)
	}

	suite.Require().EqualValues(len(res.Delegations), 1)
	suite.Require().EqualValues(res.Delegations[0].PoolInfo.ID, 1)
	suite.Require().EqualValues(res.Delegations[0].PoolInfo.ID, 1)
	suite.Require().EqualValues(res.Delegations[0].PoolInfo.Status, "ENABLED")
	suite.Require().EqualValues(res.Delegations[0].ValidatorInfo.Address, "test_address")
	suite.Require().EqualValues(res.Delegations[0].ValidatorInfo.ValKey, "test_valkey")
	suite.Require().EqualValues(statusCode, http.StatusOK)
}

func TestMultistakingTestSuite(t *testing.T) {
	testSuite := new(MultistakingTestSuite)

	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	tokenTypes.RegisterQueryServer(s, &tokenServer{})
	multiStakingTypes.RegisterQueryServer(s, &multiStakingServer{})
	bankTypes.RegisterQueryServer(s, &bankServer{})
	log.Printf("server listening at %v", lis.Addr())

	go func() {
		_ = s.Serve(lis)
	}()

	suite.Run(t, testSuite)
	s.Stop()
}
