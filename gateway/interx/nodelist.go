package interx

import (
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/global"
	"github.com/KiraCore/interx/log"
	"github.com/KiraCore/interx/tasks"
	"github.com/KiraCore/interx/types"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// RegisterInterxQueryRoutes registers query routers.
func RegisterNodeListQueryRoutes(r *mux.Router, gwCosmosmux *runtime.ServeMux, rpcAddr string) {
	r.HandleFunc(config.QueryPubP2PList, QueryPubP2PNodeList(gwCosmosmux, rpcAddr)).Methods("GET")
	r.HandleFunc(config.QueryPrivP2PList, QueryPrivP2PNodeList(gwCosmosmux, rpcAddr)).Methods("GET")
	r.HandleFunc(config.QueryInterxList, QueryInterxList(gwCosmosmux, rpcAddr)).Methods("GET")
	r.HandleFunc(config.QuerySnapList, QuerySnapList(gwCosmosmux, rpcAddr)).Methods("GET")

	common.AddRPCMethod("GET", config.QueryPubP2PList, "This is an API to query pub node list.", true)
	common.AddRPCMethod("GET", config.QueryPrivP2PList, "This is an API to query priv node list.", true)
	common.AddRPCMethod("GET", config.QueryInterxList, "This is an API to query interx list.", true)
	common.AddRPCMethod("GET", config.QuerySnapList, "This is an API to query snap node list.", true)
}

func queryPubP2PNodeList(r *http.Request, rpcAddr string) (interface{}, interface{}, int) {
	global.Mutex.Lock()
	response := tasks.PubP2PNodeListResponse

	_ = r.ParseForm()
	ip_only := r.FormValue("ip_only") == "true"
	peers_only := r.FormValue("peers_only") == "true"
	is_random := r.FormValue("order") == "random"
	is_format_simple := r.FormValue("format") == "simple"

	if is_random {
		dest := make([]types.P2PNode, len(response.NodeList))
		perm := rand.Perm(len(response.NodeList))
		for i, v := range perm {
			dest[v] = response.NodeList[i]
		}
		response.NodeList = dest
	} else {
		sort.Sort(types.P2PNodes(response.NodeList))
	}

	if is_format_simple {
		indexOfPeer := make(map[string]string)
		for index, node := range response.NodeList {
			indexOfPeer[node.ID] = strconv.Itoa(index)
		}

		for nID := range response.NodeList {
			for pIndex := range response.NodeList[nID].Peers {
				if pid, isIn := indexOfPeer[response.NodeList[nID].Peers[pIndex]]; isIn {
					response.NodeList[nID].Peers[pIndex] = pid
				}
			}
		}
	}

	dest := make([]types.P2PNode, 0)
	for _, node := range response.NodeList {
		if (r.FormValue("synced") == "true" && node.Synced) ||
			(r.FormValue("synced") == "false" && !node.Synced) ||
			(r.FormValue("synced") != "true" && r.FormValue("synced") != "false") {

			behind, _ := strconv.Atoi(r.FormValue("behind"))
			if behind == 0 || (node.BlockDiff <= int64(behind) && node.BlockDiff >= -int64(behind)) {
				if r.FormValue("unsafe") == "true" || node.Safe {
					dest = append(dest, node)
				}
			}
		}
	}
	response.NodeList = dest

	global.Mutex.Unlock()

	if ip_only {
		ips := []string{}
		for _, node := range response.NodeList {
			if r.FormValue("connected") == "true" {
				if node.Connected {
					ips = append(ips, node.IP)
				}
			} else if r.FormValue("connected") == "false" {
				if !node.Connected {
					ips = append(ips, node.IP)
				}
			} else {
				ips = append(ips, node.IP)
			}
		}

		return strings.Join(ips, "\n"), nil, http.StatusOK
	}

	if peers_only {
		peers := []string{}
		for _, node := range response.NodeList {
			peer := node.ID + "@" + node.IP + ":" + strconv.Itoa(int(node.Port))
			if r.FormValue("connected") == "true" {
				if node.Connected {
					peers = append(peers, peer)
				}
			} else if r.FormValue("connected") == "false" {
				if !node.Connected {
					peers = append(peers, peer)
				}
			} else {
				peers = append(peers, peer)
			}
		}

		return strings.Join(peers, "\n"), nil, http.StatusOK
	}

	return response, nil, http.StatusOK
}

// QueryNodeList is a function to query node list.
func QueryPubP2PNodeList(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		log.CustomLogger().Info("[query-pub-node-list] Entering pub p2p node lists query")

		if !common.RPCMethods["GET"][config.QueryPubP2PList].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryPubP2PList].CacheEnabled {

				log.CustomLogger().Info("Starting search cache for `QueryPubP2PNodeList` request...")

				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					log.CustomLogger().Info("[query-pub-node-list] Returning from the cache")
					return
				}
			}

			response.Response, response.Error, statusCode = queryPubP2PNodeList(r, rpcAddr)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryStatus].CacheEnabled)
	}
}

func queryPrivP2PNodeList(r *http.Request, rpcAddr string) (interface{}, interface{}, int) {
	global.Mutex.Lock()
	response := tasks.PrivP2PNodeListResponse

	_ = r.ParseForm()
	ip_only := r.FormValue("ip_only") == "true"
	peers_only := r.FormValue("peers_only") == "true"
	is_random := r.FormValue("order") == "random"
	is_format_simple := r.FormValue("format") == "simple"

	if is_random {
		dest := make([]types.P2PNode, len(response.NodeList))
		perm := rand.Perm(len(response.NodeList))
		for i, v := range perm {
			dest[v] = response.NodeList[i]
		}
		response.NodeList = dest
	} else {
		sort.Sort(types.P2PNodes(response.NodeList))
	}

	if is_format_simple {
		indexOfPeer := make(map[string]string)
		for index, node := range response.NodeList {
			indexOfPeer[node.ID] = strconv.Itoa(index)
		}

		for nID := range response.NodeList {
			for pIndex := range response.NodeList[nID].Peers {
				if pid, isIn := indexOfPeer[response.NodeList[nID].Peers[pIndex]]; isIn {
					response.NodeList[nID].Peers[pIndex] = pid
				}
			}
		}
	}

	dest := make([]types.P2PNode, 0)
	for _, node := range response.NodeList {
		if (r.FormValue("synced") == "true" && node.Synced) ||
			(r.FormValue("synced") == "false" && !node.Synced) ||
			(r.FormValue("synced") != "true" && r.FormValue("synced") != "false") {

			behind, _ := strconv.Atoi(r.FormValue("behind"))
			if behind == 0 || (node.BlockDiff <= int64(behind) && node.BlockDiff >= -int64(behind)) {
				if r.FormValue("unsafe") == "true" || node.Safe {
					dest = append(dest, node)
				}
			}
		}
	}
	response.NodeList = dest

	global.Mutex.Unlock()

	if ip_only {
		ips := []string{}
		for _, node := range response.NodeList {
			if r.FormValue("connected") == "true" {
				if node.Connected {
					ips = append(ips, node.IP)
				}
			} else if r.FormValue("connected") == "false" {
				if !node.Connected {
					ips = append(ips, node.IP)
				}
			} else {
				ips = append(ips, node.IP)
			}
		}

		return strings.Join(ips, "\n"), nil, http.StatusOK
	}

	if peers_only {
		peers := []string{}
		for _, node := range response.NodeList {
			peer := node.ID + "@" + node.IP + ":" + strconv.Itoa(int(node.Port))
			if r.FormValue("connected") == "true" {
				if node.Connected {
					peers = append(peers, peer)
				}
			} else if r.FormValue("connected") == "false" {
				if !node.Connected {
					peers = append(peers, peer)
				}
			} else {
				peers = append(peers, peer)
			}
		}

		return strings.Join(peers, "\n"), nil, http.StatusOK
	}

	return response, nil, http.StatusOK
}

// QueryNodeList is a function to query node list.
func QueryPrivP2PNodeList(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		log.CustomLogger().Info("[query-priv-node-list] Entering priv p2p node lists query")

		if !common.RPCMethods["GET"][config.QueryPrivP2PList].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryPrivP2PList].CacheEnabled {

				log.CustomLogger().Info("Starting search cache for `QueryPrivP2PNodeList` request...")

				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					log.CustomLogger().Info("[query-priv-node-list] Returning from the cache")
					return
				}
			}

			response.Response, response.Error, statusCode = queryPrivP2PNodeList(r, rpcAddr)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryStatus].CacheEnabled)
	}
}

func queryInterxList(r *http.Request, rpcAddr string) (interface{}, interface{}, int) {
	global.Mutex.Lock()
	response := tasks.InterxP2PNodeListResponse

	_ = r.ParseForm()
	ip_only := r.FormValue("ip_only") == "true"
	is_random := r.FormValue("order") == "random"

	if is_random {
		dest := make([]types.InterxNode, len(response.NodeList))
		perm := rand.Perm(len(response.NodeList))
		for i, v := range perm {
			dest[v] = response.NodeList[i]
		}
		response.NodeList = dest
	} else {
		sort.Sort(types.InterxNodes(response.NodeList))
	}

	dest := make([]types.InterxNode, 0)
	for _, node := range response.NodeList {
		if (r.FormValue("synced") == "true" && node.Synced) ||
			(r.FormValue("synced") == "false" && !node.Synced) ||
			(r.FormValue("synced") != "true" && r.FormValue("synced") != "false") {

			behind, _ := strconv.Atoi(r.FormValue("behind"))
			if behind == 0 || (node.BlockDiff <= int64(behind) && node.BlockDiff >= -int64(behind)) {
				if r.FormValue("unsafe") == "true" || node.Safe {
					dest = append(dest, node)
				}
			}
		}
	}
	response.NodeList = dest
	global.Mutex.Unlock()

	if ip_only {
		ips := []string{}
		for _, node := range response.NodeList {
			ips = append(ips, node.IP)
		}

		return strings.Join(ips, "\n"), nil, http.StatusOK
	}

	return response, nil, http.StatusOK
}

// QueryNodeList is a function to query node list.
func QueryInterxList(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		log.CustomLogger().Info("[query-interx-list] Entering interx lists query")

		if !common.RPCMethods["GET"][config.QueryInterxList].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QueryInterxList].CacheEnabled {

				log.CustomLogger().Info("Starting search cache for `QueryInterxList` request...")

				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					log.CustomLogger().Info("[query-interx-list] Returning from the cache")
					return
				}
			}

			response.Response, response.Error, statusCode = queryInterxList(r, rpcAddr)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryStatus].CacheEnabled)
	}
}

func querySnapList(r *http.Request, rpcAddr string) (interface{}, interface{}, int) {
	global.Mutex.Lock()
	response := tasks.SnapNodeListResponse

	_ = r.ParseForm()
	ip_only := r.FormValue("ip_only") == "true"
	is_random := r.FormValue("order") == "random"

	if is_random {
		dest := make([]types.SnapNode, len(response.NodeList))
		perm := rand.Perm(len(response.NodeList))
		for i, v := range perm {
			dest[v] = response.NodeList[i]
		}
		response.NodeList = dest
	} else {
		sort.Sort(types.SnapNodes(response.NodeList))
	}

	dest := make([]types.SnapNode, 0)
	for _, node := range response.NodeList {
		if r.FormValue("synced") == "true" {
			if node.Synced {
				dest = append(dest, node)
			}
		} else if r.FormValue("synced") == "false" {
			if !node.Synced {
				dest = append(dest, node)
			}
		} else {
			dest = append(dest, node)
		}
	}
	response.NodeList = dest
	global.Mutex.Unlock()

	if ip_only {
		ips := []string{}
		for _, node := range response.NodeList {
			ips = append(ips, node.IP)
		}

		return strings.Join(ips, "\n"), nil, http.StatusOK
	}

	return response, nil, http.StatusOK
}

// QueryNodeList is a function to query node list.
func QuerySnapList(gwCosmosmux *runtime.ServeMux, rpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		request := common.GetInterxRequest(r)
		response := common.GetResponseFormat(request, rpcAddr)

		log.CustomLogger().Info("[query-snap-list] Entering snap lists query")

		if !common.RPCMethods["GET"][config.QuerySnapList].Enabled {
			response.Response, response.Error, statusCode = common.ServeError(0, "", "API disabled", http.StatusForbidden)
		} else {
			if common.RPCMethods["GET"][config.QuerySnapList].CacheEnabled {

				log.CustomLogger().Info("Starting search cache for `QuerySnapList` request...")

				found, cacheResponse, cacheError, cacheStatus := common.SearchCache(request, response)
				if found {
					response.Response, response.Error, statusCode = cacheResponse, cacheError, cacheStatus
					common.WrapResponse(w, request, *response, statusCode, false)

					log.CustomLogger().Info("[query-snap-list] Returning from the cache")
					return
				}
			}

			response.Response, response.Error, statusCode = querySnapList(r, rpcAddr)
		}

		common.WrapResponse(w, request, *response, statusCode, common.RPCMethods["GET"][config.QueryStatus].CacheEnabled)
	}
}
