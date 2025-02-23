package metrics

import (
	"time"

	"github.com/saiset-co/sai-interx-manager/p2p/core"
)

type NodeMetrics struct {
	NodeID         core.NodeID `json:"node_id"`
	Address        string      `json:"address"`
	CPUUsage       float64     `json:"cpu_usage"`
	MemoryUsage    float64     `json:"memory_usage"`
	RequestsPerSec float64     `json:"requests_per_sec"`
	AverageLatency float64     `json:"average_latency"`
	ActiveRequests int         `json:"active_requests"`
	ErrorRate      float64     `json:"error_rate"`
	Timestamp      time.Time   `json:"timestamp"`
}

type Score struct {
	CPUScore     float64
	MemoryScore  float64
	RPSScore     float64
	LatencyScore float64
	Total        float64
}

type Weights struct {
	CPU     float64
	Memory  float64
	RPS     float64
	Latency float64
}
