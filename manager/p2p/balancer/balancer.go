package balancer

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/saiset-co/sai-interx-manager/p2p/core"
	"github.com/saiset-co/sai-interx-manager/p2p/metrics"
)

type LoadBalancer struct {
	nodeID     core.NodeID
	metrics    *metrics.Collector
	threshold  float64
	proxyCache map[core.NodeID]*httputil.ReverseProxy
}

func New(nodeID core.NodeID, metrics *metrics.Collector, threshold float64) *LoadBalancer {
	return &LoadBalancer{
		nodeID:     nodeID,
		metrics:    metrics,
		threshold:  threshold,
		proxyCache: make(map[core.NodeID]*httputil.ReverseProxy),
	}
}

func (lb *LoadBalancer) Middleware(next http.Handler) http.Handler {
	return lb.metrics.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Header.Get("X-From-Peer") == "true" {
			next.ServeHTTP(w, r)
			return
		}

		shouldHandle, targetNodeID := lb.shouldHandleRequest()
		if !shouldHandle {
			if err := lb.proxyRequest(w, r, targetNodeID); err != nil {
				http.Error(w, "Failed to delegate request", http.StatusInternalServerError)
				return
			}
			return
		}

		next.ServeHTTP(w, r)
	}))
}

func (lb *LoadBalancer) shouldHandleRequest() (bool, core.NodeID) {
	localScore := lb.metrics.CalculateScore(lb.nodeID)
	bestScore := localScore
	bestNodeID := lb.nodeID

	for nodeID := range lb.metrics.GetAllNodes() {
		if nodeID == lb.nodeID {
			continue
		}

		score := lb.metrics.CalculateScore(nodeID)
		scoreDiff := bestScore.Total - score.Total

		if scoreDiff > lb.threshold {
			bestScore = score
			bestNodeID = nodeID
		}
	}

	return bestNodeID == lb.nodeID, bestNodeID
}

func (lb *LoadBalancer) proxyRequest(w http.ResponseWriter, r *http.Request, targetNodeID core.NodeID) error {
	proxy, err := lb.getProxy(targetNodeID)
	if err != nil {
		return err
	}

	r.Header.Set("X-From-Peer", "true")
	r.Header.Set("X-Original-Node", string(lb.nodeID))

	proxy.ServeHTTP(w, r)
	return nil
}

func (lb *LoadBalancer) getProxy(nodeID core.NodeID) (*httputil.ReverseProxy, error) {
	if proxy, exists := lb.proxyCache[nodeID]; exists {
		return proxy, nil
	}

	nodeInfo, exists := lb.metrics.GetNodeInfo(nodeID)
	if !exists {
		return nil, fmt.Errorf("node %s not found", nodeID)
	}

	targetURL, err := url.Parse(fmt.Sprintf("http://%s", nodeInfo.Address))
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	lb.proxyCache[nodeID] = proxy

	return proxy, nil
}
