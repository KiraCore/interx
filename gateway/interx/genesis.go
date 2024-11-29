package interx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/global"
	"github.com/KiraCore/interx/log"
	tmjson "github.com/cometbft/cometbft/libs/json"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// RegisterGenesisQueryRoutes registers genesis query routers.
func RegisterGenesisQueryRoutes(r *mux.Router, gwCosmosmux *runtime.ServeMux, rpcAddr string) {
	r.HandleFunc(config.QueryGenesis, QueryGenesis(rpcAddr)).Methods("GET")
	r.HandleFunc(config.QueryGenesisSum, QueryGenesisSum(rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryGenesis, "This is an API to query genesis.", true)
	common.AddRPCMethod("GET", config.QueryGenesisSum, "This is an API to get genesis checksum.", true)
}

func genesisPath() string {
	return config.GetReferenceCacheDir() + "/genesis.json"
}

func JSONRemarshal(bytes []byte) ([]byte, error) {
	var ifce interface{}
	err := json.Unmarshal(bytes, &ifce)
	if err != nil {
		return nil, err
	}
	return json.Marshal(ifce)
}

func getChunkedGenesisData(rpcAddr string, chunkedNum int) ([]byte, int, error) {

	log.CustomLogger().Info("Starting `getChunkedGenesisData` request...")

	data, _, _ := common.MakeTendermintRPCRequest(rpcAddr, "/genesis_chunked", fmt.Sprintf("chunk=%d", chunkedNum))

	type GenesisChunkedResponse struct {
		Chunk string `json:"chunk"`
		Total string `json:"total"`
		Data  []byte `json:"data"`
	}

	genesis := GenesisChunkedResponse{}
	byteData, err := json.Marshal(data)
	if err != nil {
		return nil, 0, err
	}

	err = json.Unmarshal(byteData, &genesis)
	if err != nil {
		log.CustomLogger().Error("`getChunkedGenesisData` Failed to unmarshal genesisi data.",
			"chunked_Num", chunkedNum,
		)
		return nil, 0, err
	}

	total, err := strconv.Atoi(genesis.Total)
	if err != nil {
		log.CustomLogger().Error("`getChunkedGenesisData` Failed to unmarshal genesisi data.",
			"chunked_Num", chunkedNum,
			"error", err,
		)
		return nil, 0, err
	}

	log.CustomLogger().Info("Completed `getChunkedGenesisData`.")

	return genesis.Data, total, nil
}

func saveGenesis(rpcAddr string) error {

	log.CustomLogger().Info("Starting `saveGenesis` request...")

	_, err := getGenesisCheckSum()
	if err == nil {
		log.CustomLogger().Error("[saveGenesis][getGenesisCheckSum] Failed to fetch genesis checksum.",
			"error", err,
		)
		return nil
	}

	cBytes, cTotal, err := getChunkedGenesisData(rpcAddr, 0)
	if err != nil {
		log.CustomLogger().Error("[saveGenesis][getChunkedGenesisData] Failed to fetch genesis chunked data.",
			"error", err,
		)
		return err
	}

	if cTotal > 1 {
		for i := 1; i < cTotal; i++ {
			nextBytes, _, _ := getChunkedGenesisData(rpcAddr, i)
			cBytes = append(cBytes, nextBytes...)
		}
	}

	genesis := tmtypes.GenesisDoc{}
	err = tmjson.Unmarshal(cBytes, &genesis)
	if err != nil {
		return err
	}

	err = genesis.ValidateAndComplete()
	if err != nil {
		log.CustomLogger().Error("[saveGenesis][ValidateAndComplete] Failed to validate genesis data.",
			"error", err,
		)
		return err
	}

	global.Mutex.Lock()
	err = os.WriteFile(genesisPath(), cBytes, 0644)
	global.Mutex.Unlock()

	log.CustomLogger().Info("Completed `saveGenesis`.")

	return err
}

func getGenesisCheckSum() (string, error) {
	global.Mutex.Lock()
	data, err := os.ReadFile(genesisPath())
	global.Mutex.Unlock()

	if err != nil {
		log.CustomLogger().Error("[getGenesisCheckSum][ReadFile] Failed to read from genesis file.",
			"genesis_Path", genesisPath,
			"error", err,
		)
		return "", err
	}

	return common.GetSha256SumFromBytes(data), nil
}

func GetGenesisResults(rpcAddr string) (*tmtypes.GenesisDoc, string, error) {
	err := saveGenesis(rpcAddr)
	if err != nil {
		return nil, "", err
	}

	global.Mutex.Lock()
	data, err := os.ReadFile(genesisPath())
	global.Mutex.Unlock()

	if err != nil {
		return nil, "", err
	}

	genesis := tmtypes.GenesisDoc{}
	err = tmjson.Unmarshal(data, &genesis)
	log.CustomLogger().Error("[GetGenesisResults][Unmarshal] Failed to unmarshal genesis doc.",
		"error", err,
	)

	return &genesis, common.GetSha256SumFromBytes(data), err
}

// QueryGenesis is a function to query genesis.
func QueryGenesis(rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int

		log.CustomLogger().Info("Starting 'QueryGenesis' request...")

		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		if saveGenesis(rpcAddr) != nil {

			log.CustomLogger().Info("Processed 'QueryGenesis' request.",
				"method", request.Method,
				"endpoint", request.Endpoint,
				"params", request.Params,
				"error", response.Error,
			)

			response.Response, response.Error, statusCode = common.ServeError(0, "", "interx error", http.StatusInternalServerError)
			common.WrapResponse(w, request, *response, statusCode, false)
		} else {
			http.ServeFile(w, r, genesisPath())
		}

		log.CustomLogger().Info("Finished 'QueryGenesis' request.")
	}
}

func queryGenesisSumHandler(rpcAddr string) (interface{}, interface{}, int) {
	err := saveGenesis(rpcAddr)
	if err != nil {
		return common.ServeError(0, "", "interx error", http.StatusInternalServerError)
	}

	checksum, err := getGenesisCheckSum()
	if err != nil {
		return common.ServeError(0, "", "interx error", http.StatusInternalServerError)
	}

	type GenesisChecksumResponse struct {
		Checksum string `json:"checksum,omitempty"`
	}
	result := GenesisChecksumResponse{
		Checksum: "0x" + checksum,
	}

	return result, nil, http.StatusOK
}

// QueryGenesisSum is a function to get genesis checksum.
func QueryGenesisSum(rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int

		log.CustomLogger().Info("Starting 'QueryGenesisSum' request...")

		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		response.Response, response.Error, statusCode = queryGenesisSumHandler(rpcAddr)

		log.CustomLogger().Info("Processed 'QueryGenesis' request.",
			"method", request.Method,
			"endpoint", request.Endpoint,
			"params", request.Params,
			"error", response.Error,
		)

		common.WrapResponse(w, request, *response, statusCode, false)

		log.CustomLogger().Info("Finished 'QueryGenesisSum' request.")
	}
}
