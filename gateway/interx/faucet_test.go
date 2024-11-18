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

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/database"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"os"

	cosmosAuth "github.com/KiraCore/interx/proto-gen/cosmos/auth/v1beta1"
	cosmosBank "github.com/KiraCore/interx/proto-gen/cosmos/bank/v1beta1"
	kiraGov "github.com/KiraCore/interx/proto-gen/kira/gov"
	kiraSlashing "github.com/KiraCore/interx/proto-gen/kira/slashing/v1beta1"
	kiraSpending "github.com/KiraCore/interx/proto-gen/kira/spending"
	kiraStaking "github.com/KiraCore/interx/proto-gen/kira/staking"
	kiraTokens "github.com/KiraCore/interx/proto-gen/kira/tokens"
	kiraUbi "github.com/KiraCore/interx/proto-gen/kira/ubi"
	kiraUpgrades "github.com/KiraCore/interx/proto-gen/kira/upgrade"
	"github.com/KiraCore/interx/test"
	"github.com/KiraCore/interx/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/go-bip39"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

var (
	port        = flag.Int("port", 50051, "The server port")
	addr        = flag.String("addr", "localhost:50051", "the address to connect to")
	faucet_addr = "cosmos1jae3cq9c8y2lnsmh9q0rf7gjlwsglenkttgu85"
	user_addr   = "cosmos18x8js8kfyrlmqtnzcqzfjs3qhackxep5ww4nx7"
	mnemonic    = "slush panic rifle trust delay exist reduce submit female figure alert ugly rally clever expose humor category regular engine casual blanket carry tape museum"
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

type bankServer struct {
	bankTypes.UnimplementedQueryServer
	bankTypes.UnimplementedMsgServer
}

func (s *bankServer) AllBalances(ctx context.Context, in *bankTypes.QueryAllBalancesRequest) (*bankTypes.QueryAllBalancesResponse, error) {
	if in.Address == faucet_addr {
		return &bankTypes.QueryAllBalancesResponse{Balances: sdk.Coins{{Denom: "ukex", Amount: sdk.NewInt(10000000)}}}, nil
	} else {
		return &bankTypes.QueryAllBalancesResponse{Balances: sdk.Coins{{Denom: "ukex", Amount: sdk.NewInt(0)}}}, nil
	}
}

func (*bankServer) Send(ctx context.Context, req *bankTypes.MsgSend) (*bankTypes.MsgSendResponse, error) {
	return nil, nil
}

type FaucetResponse struct {
	Hash string `json:"hash"`
}

type RPCTempResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		Height string `json:"height"`
		Hash   string `json:"hash"`
	} `json:"result,omitempty"`
	Error struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type FaucetTestSuite struct {
	suite.Suite

	faucetResponse RPCTempResponse
}

func (suite *FaucetTestSuite) SetupTest() {
}

func (suite *FaucetTestSuite) TestServerFaucet() {
	config.Config.Cache.CacheDir = "./"
	_ = os.Mkdir("./db", 0777)

	database.LoadFaucetDbDriver()

	config.Config.Faucet = config.FaucetConfig{
		Mnemonic:             mnemonic,
		FaucetAmounts:        map[string]string{"ukex": "1000"},
		FaucetMinimumAmounts: map[string]string{"ukex": "100"},
		FeeAmounts:           map[string]string{"ukex": "100ukex"},
		TimeLimit:            3600,
	}

	config.Config.Faucet.Address = faucet_addr
	config.Config.GRPC = *addr
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)

	gwCosmosmux, err := GetGrpcServeMux(*addr)
	if err != nil {
		panic("failed to serve grpc")
	}

	request := common.GetInterxRequest(r)

	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	if err != nil {
		panic(err)
	}
	master, ch := hd.ComputeMastersFromSeed(seed)
	priv, err := hd.DerivePrivateKeyForPath(master, ch, "44'/118'/0'/0/0")
	config.Config.Faucet.PrivKey = &secp256k1.PrivKey{Key: priv}
	config.Config.Faucet.PubKey = config.Config.Faucet.PrivKey.PubKey()

	if err != nil {
		panic(err)
	}

	resultInfo, _, _ := serveFaucet(r, gwCosmosmux, request, test.INTERX_RPC, user_addr, "ukex")
	resultHash := FaucetResponse{}
	bz, err := json.Marshal(resultInfo)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(bz, &resultHash)
	if err != nil {
		panic(err)
	}

	suite.Require().EqualValues(resultHash.Hash, suite.faucetResponse.Result.Hash)
	os.RemoveAll("./db")
}

func (suite *FaucetTestSuite) TestServerFaucetInfo() {

	config.Config.Faucet.Address = faucet_addr
	config.Config.GRPC = *addr
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)

	gwCosmosmux, err := GetGrpcServeMux(*addr)
	if err != nil {
		panic("failed to serve grpc")
	}
	faucetInfo, _, _ := serveFaucetInfo(r, gwCosmosmux)
	suite.Require().EqualValues(faucetInfo.(types.FaucetAccountInfo).Balances[0].Amount, "10000000")
}

func TestFaucetTestSuite(t *testing.T) {
	testSuite := new(FaucetTestSuite)

	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	bankTypes.RegisterQueryServer(s, &bankServer{})
	log.Printf("server listening at %v", lis.Addr())

	go func() {
		_ = s.Serve(lis)
	}()

	testSuite.faucetResponse.Result.Hash = "faucet_hash"
	interxServer := http.Server{
		Addr: ":11000",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/broadcast_tx_async" {
				response, _ := json.Marshal(testSuite.faucetResponse)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write(response)
				if err != nil {
					panic(err)
				}
			}
		}),
	}
	go func() {
		_ = interxServer.ListenAndServe()
	}()

	suite.Run(t, testSuite)
	interxServer.Close()
	s.Stop()
}
