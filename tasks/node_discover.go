package tasks

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/global"
	"github.com/KiraCore/interx/types"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	tmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/libs/protoio"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/p2p/conn"
	tmp2p "github.com/cometbft/cometbft/proto/tendermint/p2p"
	tmTypes "github.com/cometbft/cometbft/rpc/core/types"
	tmJsonRPCTypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
)

// check if block is within 20
var BLOCK_DIFF_LIMIT int64 = 20

var (
	PubP2PNodeListResponse    types.P2PNodeListResponse
	PrivP2PNodeListResponse   types.P2PNodeListResponse
	InterxP2PNodeListResponse types.InterxNodeListResponse
	SnapNodeListResponse      types.SnapNodeListResponse
)

func timeout() time.Duration {
	timeoutDuration, err := time.ParseDuration(config.Config.NodeDiscovery.ConnectionTimeout)

	if err != nil {
		return 3 * time.Second
	}

	return timeoutDuration
}

func getInterxAddress(ipAddr string) string {
	return "http://" + ipAddr + ":" + config.Config.NodeDiscovery.DefaultInterxPort
}
func QueryNetInfo(rpcAddr string) (*tmTypes.ResultNetInfo, error) {
	result := new(tmTypes.ResultNetInfo)

	u, err := url.Parse(rpcAddr)
	if err != nil {
		return result, err
	}

	_, err = net.DialTimeout("tcp", u.Host, timeout())
	if err != nil {
		return result, err
	}

	endpoint := fmt.Sprintf("%s/net_info", rpcAddr)

	client := http.Client{
		Timeout: timeout(),
	}

	resp, err := client.Get(endpoint)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	response := new(tmJsonRPCTypes.RPCResponse)

	if err := json.Unmarshal(respBody, response); err != nil {
		return result, err
	}

	if response.Error != nil {
		return result, errors.New(fmt.Sprint(response.Error))
	}

	if err := tmjson.Unmarshal(response.Result, result); err != nil {
		return result, err
	}

	return result, err
}

func QueryNetInfoFromInterx(interxAddr string) (*tmTypes.ResultNetInfo, error) {
	result := new(tmTypes.ResultNetInfo)

	u, err := url.Parse(interxAddr)
	if err != nil {
		return result, err
	}

	_, err = net.DialTimeout("tcp", u.Host, timeout())
	if err != nil {
		return result, err
	}

	endpoint := fmt.Sprintf("%s/api/net_info", interxAddr)

	client := http.Client{
		Timeout: timeout(),
	}

	resp, err := client.Get(endpoint)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	if err := json.Unmarshal(respBody, result); err != nil {
		return result, err
	}

	return result, err
}

func QueryPeers(ipAddr string) ([]tmTypes.Peer, error) {
	interxRPC := getInterxAddress(ipAddr)

	netInfo, err := QueryNetInfoFromInterx(interxRPC)
	if err == nil {
		return netInfo.Peers, err
	}

	netInfo, err = QueryNetInfo("http://" + ipAddr + ":16657")
	if err == nil {
		return netInfo.Peers, err
	}

	netInfo, err = QueryNetInfo("http://" + ipAddr + ":26657")
	if err == nil {
		return netInfo.Peers, err
	}

	netInfo, err = QueryNetInfo("http://" + ipAddr + ":36657")
	if err == nil {
		return netInfo.Peers, err
	}

	netInfo, err = QueryNetInfo("http://" + ipAddr + ":46657")
	if err == nil {
		return netInfo.Peers, err
	}

	netInfo, err = QueryNetInfo("http://" + ipAddr + ":56657")
	if err == nil {
		return netInfo.Peers, err
	}

	return []tmTypes.Peer{}, err
}

func QueryKiraStatus(rpcAddr string) (tmTypes.ResultStatus, error) {
	result := tmTypes.ResultStatus{}

	endpoint := fmt.Sprintf("%s/status", rpcAddr)

	resp, err := http.Get(endpoint)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	response := new(tmJsonRPCTypes.RPCResponse)

	if err := json.Unmarshal(respBody, response); err != nil {
		return result, err
	}

	if response.Error != nil {
		return result, errors.New(fmt.Sprint(response.Error))
	}

	if err := tmjson.Unmarshal(response.Result, &result); err != nil {
		return result, err
	}

	return result, nil
}

func QueryStatus(ipAddr string) (tmTypes.ResultStatus, error) {
	result, err := QueryKiraStatus("http://" + ipAddr + ":16657")
	if err == nil {
		return result, err
	}

	result, err = QueryKiraStatus("http://" + ipAddr + ":26657")
	if err == nil {
		return result, err
	}

	result, err = QueryKiraStatus("http://" + ipAddr + ":36657")
	if err == nil {
		return result, err
	}

	result, err = QueryKiraStatus("http://" + ipAddr + ":46657")
	if err == nil {
		return result, err
	}

	result, err = QueryKiraStatus("http://" + ipAddr + ":56657")
	if err == nil {
		return result, err
	}

	return result, err
}

func getBlock(rpcAddr string, height string) (*struct {
	Chainid string `json:"chain_id"`
	Hash    string `json:"hash"`
	Height  string `json:"height"`
	Time    string `json:"time"`
}, error) {
	url := fmt.Sprintf("%s/block?height=%s", rpcAddr, height)
	common.GetLogger().Info("getBlock", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := new(struct {
		Jsonrpc string `json:"jsonrpc"`
		ID      int    `json:"id"`
		Result  struct {
			BlockId struct {
				Hash string `json:"hash"`
			} `json:"block_id"`
			Block struct {
				Header struct {
					Chainid string `json:"chain_id"`
					Height  string `json:"height"`
					Time    string `json:"time"`
				} `json:"header"`
			} `json:"block"`
		} `json:"result"`
		Error interface{} `json:"error"`
	})
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		common.GetLogger().Error("[node-status] Unexpected response: ", url)
		return nil, err
	}

	response := new(struct {
		Chainid string `json:"chain_id"`
		Hash    string `json:"hash"`
		Height  string `json:"height"`
		Time    string `json:"time"`
	})

	response.Chainid = result.Result.Block.Header.Chainid
	response.Hash = result.Result.BlockId.Hash
	response.Height = result.Result.Block.Header.Height
	response.Time = result.Result.Block.Header.Time

	return response, nil
}

func QueryBlock(ipAddr string, height string) (*struct {
	Chainid string `json:"chain_id"`
	Hash    string `json:"hash"`
	Height  string `json:"height"`
	Time    string `json:"time"`
}, error) {
	result, err := getBlock("http://"+ipAddr+":16657", height)
	if err == nil {
		return result, err
	}

	result, err = getBlock("http://"+ipAddr+":26657", height)
	if err == nil {
		return result, err
	}

	result, err = getBlock("http://"+ipAddr+":36657", height)
	if err == nil {
		return result, err
	}

	result, err = getBlock("http://"+ipAddr+":46657", height)
	if err == nil {
		return result, err
	}

	result, err = getBlock("http://"+ipAddr+":56657", height)
	if err == nil {
		return result, err
	}

	return result, err
}

func getBlockFromInterx(rpcAddr string, height string) (*struct {
	Chainid string `json:"chain_id"`
	Hash    string `json:"hash"`
	Height  string `json:"height"`
	Time    string `json:"time"`
}, error) {
	url := fmt.Sprintf("%s/api/block/%s", rpcAddr, height)
	common.GetLogger().Info("getBlockFromInterx", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := new(struct {
		Jsonrpc string `json:"jsonrpc"`
		ID      int    `json:"id"`
		Result  struct {
			BlockId struct {
				Hash string `json:"hash"`
			} `json:"block_id"`
			Block struct {
				Header struct {
					Chainid string `json:"chain_id"`
					Height  string `json:"height"`
					Time    string `json:"time"`
				} `json:"header"`
			} `json:"block"`
		} `json:"result"`
		Error interface{} `json:"error"`
	})
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		common.GetLogger().Error("[node-status] Unexpected response: ", url)
		return nil, err
	}

	response := new(struct {
		Chainid string `json:"chain_id"`
		Hash    string `json:"hash"`
		Height  string `json:"height"`
		Time    string `json:"time"`
	})

	response.Chainid = result.Result.Block.Header.Chainid
	response.Hash = result.Result.BlockId.Hash
	response.Height = result.Result.Block.Header.Height
	response.Time = result.Result.Block.Header.Time

	return response, nil
}

func NodeDiscover(rpcAddr string, isLog bool) {
	initPrivateIps()

	idOfPubList := make(map[string]int)
	idOfPrivList := make(map[string]int)
	idOfInterxList := make(map[string]int)
	idOfSnapshotList := make(map[string]int)

	for {
		global.Mutex.Lock()
		PubP2PNodeListResponse.Scanning = true
		PrivP2PNodeListResponse.Scanning = true
		InterxP2PNodeListResponse.Scanning = true
		SnapNodeListResponse.Scanning = true

		global.Mutex.Unlock()

		// isIpInListPrep := make(map[string]bool) // check if ip is already queried
		isPrivNodeID := make(map[string]bool)
		isPubNodeId := make(map[string]bool)
		isInterxNodeId := make(map[string]bool)
		isSnapshotIP := make(map[string]bool)

		// uniqueIPAddressesPrep := config.LoadUniqueIPAddresses()
		// for i := 0; i < len(uniqueIPAddressesPrep); i++ {
		// 	isIpInListPrep[uniqueIPAddressesPrep[i]] = true
		// }

		isIpInList := make(map[string]bool) // check if ip is already queried
		var uniqueIPAddresses []string
		isLocalPeer := make(map[string]bool)

		host, _ := getHostname(rpcAddr)
		localPeers, _ := QueryPeers(host)

		for _, peer := range localPeers {
			isLocalPeer[string(peer.NodeInfo.ID())] = true
			ip, _ := getHostname(peer.NodeInfo.ListenAddr)
			if !isPrivateIP(ip) && isIp(ip) {
				if _, ok := isIpInList[ip]; !ok {
					uniqueIPAddresses = append(uniqueIPAddresses, ip)
					isIpInList[ip] = true
				}
			}
			ip = peer.RemoteIP
			if !isPrivateIP(ip) && isIp(ip) {
				if _, ok := isIpInList[ip]; !ok {
					uniqueIPAddresses = append(uniqueIPAddresses, ip)
					isIpInList[ip] = true
				}
			}
		}

		peersFromIP := make(map[string]([]tmTypes.Peer))

		index := 0
		for index < len(uniqueIPAddresses) {
			// sleep for 1 seconds
			// time.Sleep(1 * time.Second)

			ipAddr := uniqueIPAddresses[index]
			index++

			if isLog {
				common.GetLogger().Info("[node-discovery] ", ipAddr)
			}

			kiraStatus, err := QueryStatus(ipAddr)
			if err != nil {
				continue
			}

			nodeInfo := types.P2PNode{}
			nodeInfo.ID = string(kiraStatus.NodeInfo.ID())
			nodeInfo.IP = ipAddr
			nodeInfo.Port, _ = getPort(kiraStatus.NodeInfo.ListenAddr)
			nodeInfo.Peers = []string{}
			nodeInfo.Alive = true
			nodeInfo.Synced = false
			nodeInfo.BlockHeightAtSync = kiraStatus.SyncInfo.LatestBlockHeight
			nodeInfo.BlockDiff = common.NodeStatus.Block - kiraStatus.SyncInfo.LatestBlockHeight
			if nodeInfo.BlockDiff >= -BLOCK_DIFF_LIMIT && nodeInfo.BlockDiff <= BLOCK_DIFF_LIMIT {
				nodeInfo.Synced = true
			}

			nodeInfo.Safe = kiraStatus.NodeInfo.Network == common.NodeStatus.Chainid

			if nodeInfo.Safe {
				commonBlock := common.NodeStatus.Block
				if commonBlock > kiraStatus.SyncInfo.LatestBlockHeight {
					commonBlock = kiraStatus.SyncInfo.LatestBlockHeight
				}

				localBlockInfo, _ := QueryBlock(host, strconv.Itoa(int(commonBlock)))
				nodeBlockInfo, _ := QueryBlock(ipAddr, strconv.Itoa(int(commonBlock)))

				nodeInfo.Safe = localBlockInfo != nil && nodeBlockInfo != nil && localBlockInfo.Hash == nodeBlockInfo.Hash
			}

			// verify p2p node_id via p2p connect
			peerNodeInfo, ping := connect(p2p.NewNetAddressIPPort(parseIP(nodeInfo.IP), uint16(nodeInfo.Port)), timeout())
			if peerNodeInfo == nil || nodeInfo.ID != string(peerNodeInfo.ID()) {
				continue
			}

			nodeInfo.Ping = ping

			if _, ok := isLocalPeer[nodeInfo.ID]; ok {
				nodeInfo.Connected = true
			}

			if _, ok := peersFromIP[ipAddr]; !ok {
				peersFromIP[ipAddr], _ = QueryPeers(ipAddr)
			}

			peers := peersFromIP[ipAddr]
			for _, peer := range peers {
				nodeInfo.Peers = append(nodeInfo.Peers, string(peer.NodeInfo.ID()))

				ip, _ := getHostname(peer.NodeInfo.ListenAddr)
				if isPrivateIP(ip) {
					privNodeInfo := types.P2PNode{}
					privNodeInfo.ID = string(peer.NodeInfo.ID())
					privNodeInfo.IP = ip
					privNodeInfo.Port, _ = getPort(peer.NodeInfo.ListenAddr)
					privNodeInfo.Peers = []string{}
					privNodeInfo.Peers = append(privNodeInfo.Peers, nodeInfo.ID)
					privNodeInfo.Alive = true
					privNodeInfo.Synced = nodeInfo.Synced
					privNodeInfo.BlockHeightAtSync = nodeInfo.BlockHeightAtSync
					privNodeInfo.BlockDiff = nodeInfo.BlockDiff
					privNodeInfo.Safe = nodeInfo.Safe

					if _, ok := isLocalPeer[privNodeInfo.ID]; ok {
						privNodeInfo.Connected = true
					}

					if _, ok := isPrivNodeID[privNodeInfo.ID]; !ok {
						global.Mutex.Lock()
						if pid, isIn := idOfPrivList[privNodeInfo.ID]; isIn {
							PrivP2PNodeListResponse.NodeList[pid] = privNodeInfo
						} else {
							PrivP2PNodeListResponse.NodeList = append(PrivP2PNodeListResponse.NodeList, privNodeInfo)
							idOfPrivList[privNodeInfo.ID] = len(PrivP2PNodeListResponse.NodeList) - 1
						}
						global.Mutex.Unlock()
						isPrivNodeID[privNodeInfo.ID] = true
					}
				} else {
					if _, ok := isIpInList[ip]; !ok {
						uniqueIPAddresses = append(uniqueIPAddresses, ip)
						isIpInList[ip] = true
					}
				}
			}

			global.Mutex.Lock()
			if pid, isIn := idOfPubList[nodeInfo.ID]; isIn {
				PubP2PNodeListResponse.NodeList[pid] = nodeInfo
			} else {
				PubP2PNodeListResponse.NodeList = append(PubP2PNodeListResponse.NodeList, nodeInfo)
				idOfPubList[nodeInfo.ID] = len(PubP2PNodeListResponse.NodeList) - 1
			}
			global.Mutex.Unlock()
			isPubNodeId[nodeInfo.ID] = true

			interxStartTime := makeTimestamp()
			interxAddress := getInterxAddress(ipAddr)
			interxStatus := common.GetInterxStatus(interxAddress)

			if interxStatus != nil {
				interxEndTime := makeTimestamp()

				interxInfo := types.InterxNode{}
				interxInfo.ID = interxStatus.ID
				interxInfo.IP = ipAddr
				interxInfo.Ping = interxEndTime - interxStartTime
				interxInfo.Moniker = interxStatus.InterxInfo.Moniker
				interxInfo.Faucet = interxStatus.InterxInfo.FaucetAddr
				interxInfo.Type = interxStatus.InterxInfo.Node.NodeType
				interxInfo.InterxVersion = interxStatus.InterxInfo.InterxVersion
				interxInfo.SekaiVersion = interxStatus.InterxInfo.SekaiVersion
				interxInfo.Alive = true
				interxInfo.Synced = false
				interxInfo.BlockHeightAtSync, _ = strconv.ParseInt(interxStatus.SyncInfo.LatestBlockHeight, 10, 64)
				interxInfo.BlockDiff = common.NodeStatus.Block - interxInfo.BlockHeightAtSync
				if interxInfo.BlockDiff >= -BLOCK_DIFF_LIMIT && interxInfo.BlockDiff <= BLOCK_DIFF_LIMIT {
					interxInfo.Synced = true
				}

				interxInfo.Safe = interxStatus.InterxInfo.ChainID == common.NodeStatus.Chainid

				if nodeInfo.Safe {
					commonBlock := common.NodeStatus.Block
					if commonBlock > interxInfo.BlockHeightAtSync {
						commonBlock = interxInfo.BlockHeightAtSync
					}

					localBlockInfo, _ := QueryBlock(host, strconv.Itoa(int(commonBlock)))
					nodeBlockInfo, _ := getBlockFromInterx(getInterxAddress(ipAddr), strconv.Itoa(int(commonBlock)))

					nodeInfo.Safe = localBlockInfo != nil && nodeBlockInfo != nil && localBlockInfo.Hash == nodeBlockInfo.Hash
				}

				global.Mutex.Lock()
				if pid, isIn := idOfInterxList[interxInfo.ID]; isIn {
					InterxP2PNodeListResponse.NodeList[pid] = interxInfo
				} else {
					InterxP2PNodeListResponse.NodeList = append(InterxP2PNodeListResponse.NodeList, interxInfo)
					idOfInterxList[interxInfo.ID] = len(InterxP2PNodeListResponse.NodeList) - 1
				}
				global.Mutex.Unlock()
				isInterxNodeId[interxInfo.ID] = true

				// snapshotInfo := common.GetSnapshotInfo(getInterxAddress(ipAddr))
				snapshotInfo := common.GetSnapshotInfo(interxAddress)
				if snapshotInfo != nil {
					snapNode := types.SnapNode{}
					snapNode.IP = ipAddr
					snapNode.Port, _ = getPort(interxAddress)
					snapNode.Checksum = snapshotInfo.Checksum
					snapNode.Size = snapshotInfo.Size
					snapNode.Alive = true
					snapNode.Synced = interxInfo.Synced

					global.Mutex.Lock()
					if pid, isIn := idOfSnapshotList[snapNode.IP]; isIn {
						SnapNodeListResponse.NodeList[pid] = snapNode
					} else {
						SnapNodeListResponse.NodeList = append(SnapNodeListResponse.NodeList, snapNode)
						idOfSnapshotList[snapNode.IP] = len(SnapNodeListResponse.NodeList) - 1
					}
					global.Mutex.Unlock()
					isSnapshotIP[snapNode.IP] = true
				}
			}
		}

		lastUpdate := time.Now().Unix()

		// Remove disconnected nodes
		for index, value := range PrivP2PNodeListResponse.NodeList {
			if !isPrivNodeID[value.ID] {
				PrivP2PNodeListResponse.NodeList[index].Alive = false
			}
		}
		for index, value := range PubP2PNodeListResponse.NodeList {
			if !isPubNodeId[value.ID] {
				PubP2PNodeListResponse.NodeList[index].Alive = false
			}
		}
		for index, value := range InterxP2PNodeListResponse.NodeList {
			if !isInterxNodeId[value.ID] {
				InterxP2PNodeListResponse.NodeList[index].Alive = false
			}
		}
		for index, value := range SnapNodeListResponse.NodeList {
			if !isSnapshotIP[value.IP] {
				SnapNodeListResponse.NodeList[index].Alive = false
			}
		}

		global.Mutex.Lock()
		PubP2PNodeListResponse.LastUpdate = lastUpdate
		PrivP2PNodeListResponse.LastUpdate = lastUpdate
		InterxP2PNodeListResponse.LastUpdate = lastUpdate
		SnapNodeListResponse.LastUpdate = lastUpdate

		PubP2PNodeListResponse.Scanning = false
		PrivP2PNodeListResponse.Scanning = false
		InterxP2PNodeListResponse.Scanning = false
		SnapNodeListResponse.Scanning = false
		global.Mutex.Unlock()

		if isLog {
			common.GetLogger().Info("[node-discovery] finished!")
		}

		time.Sleep(10 * time.Second)
	}
}

var privateIPBlocks []*net.IPNet

func initPrivateIps() {
	for _, cidr := range []string{
		"127.0.0.0/8",    // IPv4 loopback
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"169.254.0.0/16", // RFC3927 link-local
		"::1/128",        // IPv6 loopback
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 unique local addr
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(fmt.Errorf("parse error on %q: %v", cidr, err))
		}
		privateIPBlocks = append(privateIPBlocks, block)
	}
}

func isPrivateIP(ipAddr string) bool {
	ip := net.ParseIP(ipAddr)

	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func getPort(listenAddr string) (uint16, error) {
	strs := strings.Split(listenAddr, ":")
	if len(strs) > 1 {
		portNumber, err := strconv.ParseUint(strs[len(strs)-1], 10, 16)
		if err == nil {
			return uint16(portNumber), nil
		}
	}
	return 0, nil
}

func getHostname(listenAddr string) (string, error) {
	u, err := url.Parse(listenAddr)
	if err == nil {
		return u.Hostname(), nil
	}
	common.GetLogger().Error("[node-discovery] unexpected listen addr: ", listenAddr)
	return "", err
}

func isIp(ipAddr string) bool {
	addr := net.ParseIP(ipAddr)
	return addr != nil
}

func connect(
	netAddress *p2p.NetAddress,
	timeoutDuration time.Duration,
) (*p2p.DefaultNodeInfo, int64) {
	// load node_key
	privKey := ed25519.GenPrivKey()
	nodeKey := &p2p.NodeKey{
		PrivKey: privKey,
	}

	// manual handshaking
	// dial to address

	startTime := makeTimestamp()
	connection, err := netAddress.DialTimeout(timeoutDuration)
	endTime := makeTimestamp()
	connectionTime := endTime - startTime

	if err != nil {
		return nil, 0
	}

	// create secret connection
	startTime = makeTimestamp()
	secretConn, err := upgradeSecretConn(connection, timeoutDuration, nodeKey.PrivKey)
	endTime = makeTimestamp()
	if endTime-startTime > connectionTime {
		connectionTime = endTime - startTime
	}

	if err != nil {
		return nil, 0
	}

	// handshake

	startTime = makeTimestamp()
	peerNodeInfo, err := handshake(secretConn, timeoutDuration, p2p.DefaultNodeInfo{})
	endTime = makeTimestamp()
	if endTime-startTime > connectionTime {
		connectionTime = endTime - startTime
	}

	if err != nil {
		return nil, 0
	}

	return peerNodeInfo, connectionTime
}

func makeTimestamp() int64 {
	return time.Now().UTC().UnixNano() / int64(time.Millisecond)
}

func parseIP(host string) net.IP {
	ip := net.ParseIP(host)
	if ip == nil {
		ips, err := net.LookupIP(host)
		if err != nil {
			return nil
		}
		ip = ips[0]
	}

	return ip
}

func upgradeSecretConn(
	c net.Conn,
	timeout time.Duration,
	privKey crypto.PrivKey,
) (*conn.SecretConnection, error) {
	if err := c.SetDeadline(time.Now().UTC().Add(timeout)); err != nil {
		return nil, err
	}

	sc, err := conn.MakeSecretConnection(c, privKey)
	if err != nil {
		return nil, err
	}

	return sc, sc.SetDeadline(time.Time{})
}

func handshake(
	c net.Conn,
	timeout time.Duration,
	nodeInfo p2p.NodeInfo,
) (*p2p.DefaultNodeInfo, error) {
	if err := c.SetDeadline(time.Now().UTC().Add(timeout)); err != nil {
		return nil, err
	}

	var (
		errc = make(chan error, 2)

		pbpeerNodeInfo tmp2p.DefaultNodeInfo
		peerNodeInfo   p2p.DefaultNodeInfo
		ourNodeInfo    = nodeInfo.(p2p.DefaultNodeInfo)
	)

	go func(errc chan<- error, c net.Conn) {
		_, err := protoio.NewDelimitedWriter(c).WriteMsg(ourNodeInfo.ToProto())
		errc <- err
	}(errc, c)
	go func(errc chan<- error, c net.Conn) {
		protoReader := protoio.NewDelimitedReader(c, p2p.MaxNodeInfoSize())
		_, err := protoReader.ReadMsg(&pbpeerNodeInfo)
		errc <- err
	}(errc, c)

	for i := 0; i < cap(errc); i++ {
		err := <-errc
		if err != nil {
			return nil, err
		}
	}

	peerNodeInfo, err := p2p.DefaultNodeInfoFromToProto(&pbpeerNodeInfo)
	if err != nil {
		return nil, err
	}

	return &peerNodeInfo, c.SetDeadline(time.Time{})
}
