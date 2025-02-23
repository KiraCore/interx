package metrics

import (
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/saiset-co/sai-interx-manager/p2p/core"
	"github.com/saiset-co/sai-interx-manager/p2p/utils"
)

type Collector struct {
	nodeID         core.NodeID
	mutex          sync.RWMutex
	requests       map[string]*core.Request
	metrics        map[core.NodeID]NodeMetrics
	latencies      map[core.NodeID]float64
	requestHistory []requestStat
	weights        Weights
	startTime      time.Time
	windowSize     time.Duration
}

type requestStat struct {
	Path      string
	Duration  float64
	IsError   bool
	Timestamp time.Time
}

func NewCollector(nodeID core.NodeID, weights Weights, windowSize time.Duration) *Collector {
	return &Collector{
		nodeID:         nodeID,
		requests:       make(map[string]*core.Request),
		metrics:        make(map[core.NodeID]NodeMetrics),
		latencies:      make(map[core.NodeID]float64),
		requestHistory: make([]requestStat, 0),
		weights:        weights,
		startTime:      time.Now(),
		windowSize:     windowSize,
	}
}

func (c *Collector) CollectLocalMetrics() NodeMetrics {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	windowStart := time.Now().Add(-c.windowSize)
	var requests, errors int
	for _, stat := range c.requestHistory {
		if stat.Timestamp.After(windowStart) {
			requests++
			if stat.IsError {
				errors++
			}
		}
	}

	var totalLatency float64
	validStats := 0
	for _, stat := range c.requestHistory {
		if stat.Timestamp.After(windowStart) {
			totalLatency += stat.Duration
			validStats++
		}
	}

	rps := float64(requests) / c.windowSize.Seconds()
	errorRate := 0.0
	if requests > 0 {
		errorRate = float64(errors) / float64(requests) * 100
	}
	avgLatency := 0.0
	if validStats > 0 {
		avgLatency = totalLatency / float64(validStats)
	}

	return NodeMetrics{
		NodeID:         c.nodeID,
		CPUUsage:       utils.GetCPUUsage(),
		MemoryUsage:    utils.GetMemoryUsage(),
		RequestsPerSec: rps,
		AverageLatency: avgLatency,
		ActiveRequests: len(c.requests),
		ErrorRate:      errorRate,
		Timestamp:      time.Now(),
	}
}

func (c *Collector) UpdateNodeMetrics(metrics NodeMetrics, latency float64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.metrics[metrics.NodeID] = metrics
	c.latencies[metrics.NodeID] = latency
	c.cleanup()
}

func (c *Collector) CalculateScore(nodeID core.NodeID) Score {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	metrics, exists := c.metrics[nodeID]
	if !exists {
		return Score{Total: 1.0}
	}

	cpuScore := metrics.CPUUsage / 100.0
	memScore := metrics.MemoryUsage / 100.0
	rpsScore := math.Min(metrics.RequestsPerSec/1000.0, 1.0)
	latencyScore := math.Min(c.latencies[nodeID]/1000.0, 1.0)

	total := cpuScore*c.weights.CPU +
		memScore*c.weights.Memory +
		rpsScore*c.weights.RPS +
		latencyScore*c.weights.Latency

	return Score{
		CPUScore:     cpuScore,
		MemoryScore:  memScore,
		RPSScore:     rpsScore,
		LatencyScore: latencyScore,
		Total:        total,
	}
}

func (c *Collector) StartRequest(req *core.Request) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.requests[req.ID] = req
}

func (c *Collector) FinishRequest(reqID string, isError bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if req, exists := c.requests[reqID]; exists {
		endTime := time.Now()
		req.EndTime = &endTime
		duration := endTime.Sub(req.StartTime).Seconds()

		c.requestHistory = append(c.requestHistory, requestStat{
			Path:      req.Path,
			Duration:  duration,
			IsError:   isError,
			Timestamp: endTime,
		})

		delete(c.requests, reqID)
		c.cleanup()
	}
}

func (c *Collector) cleanup() {
	cutoff := time.Now().Add(-c.windowSize)

	newHistory := make([]requestStat, 0)
	for _, stat := range c.requestHistory {
		if stat.Timestamp.After(cutoff) {
			newHistory = append(newHistory, stat)
		}
	}
	c.requestHistory = newHistory

	for nodeID, metrics := range c.metrics {
		if metrics.Timestamp.Before(cutoff) {
			delete(c.metrics, nodeID)
			delete(c.latencies, nodeID)
		}
	}
}

func (c *Collector) GetAllNodes() map[core.NodeID]struct{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	nodes := make(map[core.NodeID]struct{})
	for nodeID := range c.metrics {
		nodes[nodeID] = struct{}{}
	}
	return nodes
}

func (c *Collector) GetNodeInfo(nodeID core.NodeID) (*core.PeerInfo, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	metrics, exists := c.metrics[nodeID]
	if !exists {
		return nil, false
	}

	return &core.PeerInfo{
		NodeID:    nodeID,
		Address:   metrics.Address,
		Connected: true,
	}, true
}

func (c *Collector) Address(nodeID core.NodeID) (string, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	metrics, exists := c.metrics[nodeID]
	if !exists {
		return "", false
	}
	return metrics.Address, true
}

func (c *Collector) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := &core.Request{
			ID:        utils.GenerateID(),
			StartTime: time.Now(),
			Method:    r.Method,
			Path:      r.URL.Path,
			FromPeer:  r.Header.Get("X-From-Peer") == "true",
		}

		c.StartRequest(req)

		rw := utils.NewResponseWriter(w)
		next.ServeHTTP(rw, r)

		c.FinishRequest(req.ID, rw.StatusCode >= 400)
	})
}
