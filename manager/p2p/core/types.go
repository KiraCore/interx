package core

import "time"

type NodeID string

type Request struct {
	ID        string     `json:"id"`
	StartTime time.Time  `json:"start_time"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Method    string     `json:"method"`
	Path      string     `json:"path"`
	FromPeer  bool       `json:"from_peer"`
}

type Message struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
}

type MessageType string

const (
	MessageTypeJoinRequest  MessageType = "join_request"
	MessageTypeJoinResponse MessageType = "join_response"
	MessageTypeMetrics      MessageType = "metrics"
	MessageTypePing         MessageType = "ping"
	MessageTypePong         MessageType = "pong"
)

type JoinRequest struct {
	NodeID       NodeID   `json:"node_id"`
	Address      string   `json:"address"`
	VisitedNodes []NodeID `json:"visited_nodes,omitempty"`
}

type JoinResponse struct {
	Success          bool       `json:"success"`
	Error            string     `json:"error,omitempty"`
	AlternativePeers []PeerInfo `json:"alternative_peers,omitempty"`
	NATPort          int        `json:"nat_port,omitempty"`
}

type PeerInfo struct {
	NodeID    NodeID `json:"node_id"`
	Address   string `json:"address"`
	Connected bool   `json:"connected"`
}
