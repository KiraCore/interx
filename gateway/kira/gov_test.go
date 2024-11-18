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
	"os"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/database"
	cosmosAuth "github.com/KiraCore/interx/proto-gen/cosmos/auth/v1beta1"
	cosmosBank "github.com/KiraCore/interx/proto-gen/cosmos/bank/v1beta1"
	kiraGov "github.com/KiraCore/interx/proto-gen/kira/gov"
	kiraMultiStaking "github.com/KiraCore/interx/proto-gen/kira/multistaking"
	kiraSlashing "github.com/KiraCore/interx/proto-gen/kira/slashing/v1beta1"
	kiraSpending "github.com/KiraCore/interx/proto-gen/kira/spending"
	kiraStaking "github.com/KiraCore/interx/proto-gen/kira/staking"
	kiraTokens "github.com/KiraCore/interx/proto-gen/kira/tokens"
	kiraUbi "github.com/KiraCore/interx/proto-gen/kira/ubi"
	kiraUpgrades "github.com/KiraCore/interx/proto-gen/kira/upgrade"
	"github.com/KiraCore/interx/test"
	govTypes "github.com/KiraCore/sekai/x/gov/types"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
	addr = flag.String("0xaddr", "localhost:50051", "the address to connect to")
)

// GetGrpcServeMux is a function to get ServerMux for GRPC server.
func GetGrpcServeMux(grpcAddr string) (*runtime.ServeMux, error) {
	// Create a client connection to the gRPC Server we just started.
	// This is where the gRPC-Gateway proxies the requests.
	// WITH_TRANSPORT_CREDENTIALS: Empty parameters mean set transport security.
	security := grpc.WithInsecure()

	// With transport credentials
	// if strings.ToLower(os.Getenv("WITH_TRANSPORT_CREDENTIALS")) == "true" {
	// 	security = grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(insecure.CertPool, ""))
	// }

	conn, err := grpc.DialContext(
		context.Background(),
		grpcAddr,
		security,
		grpc.WithBlock(),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to dial server: %w", err)
	}

	gwCosmosmux := runtime.NewServeMux()
	err = cosmosBank.RegisterQueryHandler(context.Background(), gwCosmosmux, conn)
	if err != nil {
		return nil, fmt.Errorf("failed to register gateway: %w", err)
	}

	err = cosmosAuth.RegisterQueryHandler(context.Background(), gwCosmosmux, conn)
	if err != nil {
		return nil, fmt.Errorf("failed to register gateway: %w", err)
	}

	err = kiraGov.RegisterQueryHandler(context.Background(), gwCosmosmux, conn)
	if err != nil {
		return nil, fmt.Errorf("failed to register gateway: %w", err)
	}

	err = kiraStaking.RegisterQueryHandler(context.Background(), gwCosmosmux, conn)
	if err != nil {
		return nil, fmt.Errorf("failed to register gateway: %w", err)
	}

	err = kiraMultiStaking.RegisterQueryHandler(context.Background(), gwCosmosmux, conn)
	if err != nil {
		return nil, fmt.Errorf("failed to register gateway: %w", err)
	}

	err = kiraSlashing.RegisterQueryHandler(context.Background(), gwCosmosmux, conn)
	if err != nil {
		return nil, fmt.Errorf("failed to register gateway: %w", err)
	}

	err = kiraTokens.RegisterQueryHandler(context.Background(), gwCosmosmux, conn)
	if err != nil {
		return nil, fmt.Errorf("failed to register gateway: %w", err)
	}

	err = kiraUpgrades.RegisterQueryHandler(context.Background(), gwCosmosmux, conn)
	if err != nil {
		return nil, fmt.Errorf("failed to register gateway: %w", err)
	}

	err = kiraSpending.RegisterQueryHandler(context.Background(), gwCosmosmux, conn)
	if err != nil {
		return nil, fmt.Errorf("failed to register gateway: %w", err)
	}

	err = kiraUbi.RegisterQueryHandler(context.Background(), gwCosmosmux, conn)
	if err != nil {
		return nil, fmt.Errorf("failed to register gateway: %w", err)
	}

	return gwCosmosmux, nil
}

type govServer struct {
	govTypes.UnimplementedQueryServer
}

func (s *govServer) AllDataReferenceKeys(ctx context.Context, in *govTypes.QueryDataReferenceKeysRequest) (*govTypes.QueryDataReferenceKeysResponse, error) {
	return &govTypes.QueryDataReferenceKeysResponse{Keys: []string{
		"test1", "test2",
	}}, nil
}

func (s *govServer) DataReferenceByKey(ctx context.Context, in *govTypes.QueryDataReferenceRequest) (*govTypes.QueryDataReferenceResponse, error) {
	return &govTypes.QueryDataReferenceResponse{
		Data: &govTypes.DataRegistryEntry{
			Hash:      "test_hash",
			Reference: "test_reference",
		},
	}, nil
}

func (s *govServer) NetworkProperties(ctx context.Context, in *govTypes.NetworkPropertiesRequest) (*govTypes.NetworkPropertiesResponse, error) {
	return &govTypes.NetworkPropertiesResponse{
		Properties: &govTypes.NetworkProperties{
			MinTxFee:   777,
			MaxTxFee:   888,
			VoteQuorum: 999,
		},
	}, nil
}

func (s *govServer) ExecutionFee(ctx context.Context, in *govTypes.ExecutionFeeRequest) (*govTypes.ExecutionFeeResponse, error) {
	return &govTypes.ExecutionFeeResponse{
		Fee: &govTypes.ExecutionFee{
			TransactionType: "test_type",
		},
	}, nil
}

type DataReferenceTestSuite struct {
	suite.Suite
}

func (suite *DataReferenceTestSuite) SetupTest() {
}

func (suite *DataReferenceTestSuite) TestDataReferenceKeysHandler() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	gwCosmosmux, err := GetGrpcServeMux(*addr)
	if err != nil {
		panic("failed to serve grpc")
	}
	r.URL.Path = "/kira/gov/data_keys"
	refInfo, _, _ := queryDataReferenceKeysHandle(r, gwCosmosmux)
	res := govTypes.QueryDataReferenceKeysResponse{}
	bz, err := json.Marshal(refInfo)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(bz, &res)
	if err != nil {
		panic(err)
	}

	suite.Require().EqualValues(len(res.Keys), 2)
}

func (suite *DataReferenceTestSuite) TestDataReferenceByKeyHandler() {
	config.Config.Cache.CacheDir = "./"
	_ = os.Mkdir("./db", 0777)

	database.LoadReferenceDbDriver()
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	gwCosmosmux, err := GetGrpcServeMux(*addr)
	if err != nil {
		panic("failed to serve grpc")
	}
	r.URL.Path = "/kira/gov/data/test"
	refInfo, _, _ := queryDataReferenceHandle(r, gwCosmosmux, "test")
	res := govTypes.DataRegistryEntry{}
	bz, err := json.Marshal(refInfo)
	if err != nil {
		panic(err)
	}

	_ = json.Unmarshal(bz, &res)
	suite.Require().EqualValues(res.Hash, "test_hash")
	err = os.RemoveAll("./db")
	if err != nil {
		panic(err)
	}
}

func (suite *DataReferenceTestSuite) TestNetworkPropertiesHandler() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	gwCosmosmux, err := GetGrpcServeMux(*addr)
	if err != nil {
		panic("failed to serve grpc")
	}
	r.URL.Path = "/kira/gov/network_properties"
	_, _, statusCode := QueryNetworkPropertiesHandle(r, gwCosmosmux)
	suite.Require().EqualValues(statusCode, http.StatusOK)
}

func (suite *DataReferenceTestSuite) TestExecutionFeeHandler() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	gwCosmosmux, err := GetGrpcServeMux(*addr)
	if err != nil {
		panic("failed to serve grpc")
	}
	r.URL.Path = "/kira/gov/execution_fee/test"
	_, _, statusCode := QueryExecutionFeeHandle(r, gwCosmosmux)
	suite.Require().EqualValues(statusCode, http.StatusOK)
}

func TestDataReferenceTestSuite(t *testing.T) {
	testSuite := new(DataReferenceTestSuite)

	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	govTypes.RegisterQueryServer(s, &govServer{})
	log.Printf("server listening at %v", lis.Addr())

	go func() {
		_ = s.Serve(lis)
	}()

	suite.Run(t, testSuite)
	s.Stop()
}
