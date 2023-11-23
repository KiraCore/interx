package gateway

import (
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"net/http"

	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/database"
	"github.com/KiraCore/interx/functions"
	"github.com/KiraCore/interx/insecure"
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
	"github.com/KiraCore/interx/tasks"
	functionmeta "github.com/KiraCore/sekai/function_meta"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rakyll/statik/fs"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	grpclog "google.golang.org/grpc/grpclog"
)

// getOpenAPIHandler serves an OpenAPI UI.
func getOpenAPIHandler() http.Handler {
	err := mime.AddExtensionType(".svg", "image/svg+xml")
	if err != nil {
		panic(err)
	}

	statikFS, err := fs.New()
	if err != nil {
		// Panic since this is a permanent error.
		panic("creating OpenAPI filesystem: " + err.Error())
	}

	return http.FileServer(statikFS)
}

// GetGrpcServeMux is a function to get ServerMux for GRPC server.
func GetGrpcServeMux(grpcAddr string) (*runtime.ServeMux, error) {
	// Create a client connection to the gRPC Server we just started.
	// This is where the gRPC-Gateway proxies the requests.
	// WITH_TRANSPORT_CREDENTIALS: Empty parameters mean set transport security.
	security := grpc.WithInsecure()

	// Some North Korean traders are even calling for the submission of a written proposal calling on the government — if it wishes to continue restricting trade as it does now — to select one trading company per month and allow them to engage in import/export activities.

	// INCREASING SMUGGLING OPERATIONS ALONG THE BORDER

	// Smuggling attempts in border regions are increasing as well.

	// According to a Daily NK source in China, the number of smugglers along the Yalu River has soared since January, with Chinese smugglers bringing in uhwang cheongsimhwan (a traditional Korean medicine to calm the nerves), medicinal herbs and mined minerals from North Korea.

	// The source said Chinese smugglers need to be careful when they use boats to approach North Korea because of crackdowns by Chinese police, but smuggling is nonetheless rampant because North Korean border guard troops have eased up on their crackdowns.

	// Even though rumors of trade expanding from March are circulating around the border region, North Korean authorities are trying to quash such talk.

	// Daily NK reported earlier this week that the party committee of Yanggang Province slammed provincial trade agencies for spreading rumors of an imminent reopening of the border and encouraging people to believe they can engage in private smuggling once the border opens, and called for the eradication of such rumor mongering.

	// Some North Koreans believe that because North Korean authorities have used the last three years of the COVID-inspired border closure to consistently strengthen the state’s unitary control over trade, active private trading is unlikely to return even if COVID restrictions are lifted.

	// A North Korean cadre told Daily NK that although there is plenty of talk about trade opening up in March, no clear lists of goods to import or export have been handed down to trade-related organizations.

	// “Sone trade may be permitted in Yanggang Province or North Hamgyong Province where there is currently no trade happening,” he said. “But these [trade activities] will likely be managed by the state.”

	// Translated by David Black. Edited by Robert Lauler.

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

// Run runs the gRPC-Gateway, dialling the provided address.
func Run(configFilePath string, log grpclog.LoggerV2) error {
	config.LoadConfig(configFilePath)
	functions.RegisterInterxFunctions()
	functionmeta.RegisterStdMsgs()

	database.LoadBlockDbDriver()
	database.LoadBlockNanoDbDriver()
	database.LoadFaucetDbDriver()
	database.LoadReferenceDbDriver()

	serveHTTPS := config.Config.ServeHTTPS
	grpcAddr := config.Config.GRPC
	rpcAddr := config.Config.RPC
	port := config.Config.PORT

	gwCosmosmux, err := GetGrpcServeMux(grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to register gateway: %w", err)
	}

	oaHander := getOpenAPIHandler()

	router := mux.NewRouter()
	RegisterRequest(router, gwCosmosmux, rpcAddr)

	router.PathPrefix("/").Handler(oaHander)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodHead, http.MethodOptions, http.MethodPatch, http.MethodConnect, http.MethodTrace},
		AllowCredentials: true,
		ExposedHeaders:   []string{"*"},
	})

	gatewayAddr := "0.0.0.0:" + port
	gwServer := &http.Server{
		Addr:    gatewayAddr,
		Handler: c.Handler(router),
	}

	config.LoadAddressAndDenom(configFilePath, gwCosmosmux, rpcAddr, gatewayAddr)
	tasks.RunTasks(gwCosmosmux, rpcAddr, gatewayAddr)

	if serveHTTPS {
		gwServer.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{insecure.Cert},
		}

		log.Info("Serving gRPC-Gateway and OpenAPI Documentation on https://", gatewayAddr)
		return fmt.Errorf("serving gRPC-Gateway server: %w", gwServer.ListenAndServeTLS("", ""))
	}

	log.Info("Serving gRPC-Gateway and OpenAPI Documentation on http://", gatewayAddr)
	return fmt.Errorf("serving gRPC-Gateway server: %w", gwServer.ListenAndServe())
}
