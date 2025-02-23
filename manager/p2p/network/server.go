package network

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/saiset-co/sai-interx-manager/p2p/core"
	"github.com/saiset-co/sai-interx-manager/p2p/metrics"
)

type Server struct {
	nodeID   core.NodeID
	address  string
	maxPeers int
	peers    map[core.NodeID]*Peer
	metrics  *metrics.Collector
	listener net.Listener
	mutex    sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

type ServerConfig struct {
	NodeID   core.NodeID
	Address  string
	MaxPeers int
	Metrics  *metrics.Collector
}

func NewServer(sctx context.Context, config ServerConfig) *Server {
	ctx, cancel := context.WithCancel(sctx)
	return &Server{
		nodeID:   config.NodeID,
		address:  config.Address,
		maxPeers: config.MaxPeers,
		peers:    make(map[core.NodeID]*Peer),
		metrics:  config.Metrics,
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}
	s.listener = listener

	go s.acceptConnections()
	go s.startHealthCheck()

	return nil
}

func (s *Server) Stop() {
	s.cancel()
	if s.listener != nil {
		s.listener.Close()
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, peer := range s.peers {
		if peer.Conn != nil {
			peer.Conn.Close()
		}
		if peer.OutboundConn != nil {
			peer.OutboundConn.Close()
		}
	}
}

func (s *Server) acceptConnections() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.ctx.Err() != nil {
				return // Сервер остановлен
			}
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	var msg core.Message
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&msg); err != nil {
		return
	}

	switch msg.Type {
	case core.MessageTypeJoinRequest:
		s.handleJoinRequest(conn, msg)
	case core.MessageTypeMetrics:
		s.handleMetricsUpdate(conn, msg)
	case core.MessageTypePing:
		s.handlePing(conn, msg)
	}
}

func (s *Server) handleJoinRequest(conn net.Conn, msg core.Message) {
	var joinReq core.JoinRequest
	if err := json.Unmarshal(msg.Payload.([]byte), &joinReq); err != nil {
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	response := core.JoinResponse{Success: false}

	if len(s.peers) < s.maxPeers {
		natPort := GetNATPort(conn)
		peer := &Peer{
			NodeID:  joinReq.NodeID,
			Address: joinReq.Address,
			Conn:    conn,
			Status: PeerStatus{
				Connected: true,
				LastPing:  time.Now(),
				NATPort:   natPort,
			},
		}

		s.peers[joinReq.NodeID] = peer
		response.Success = true
		response.NATPort = natPort

		go func() {
			if err := EstablishNATConnection(peer); err != nil {
				s.removePeer(joinReq.NodeID)
			}
		}()
	} else {
		alternatives := make([]core.PeerInfo, 0)
		for nodeID, peer := range s.peers {
			if peer.Status.Connected {
				alternatives = append(alternatives, core.PeerInfo{
					NodeID:    nodeID,
					Address:   peer.Address,
					Connected: true,
				})
			}
		}
		response.AlternativePeers = alternatives
	}

	encoder := json.NewEncoder(conn)
	encoder.Encode(core.Message{
		Type:    core.MessageTypeJoinResponse,
		Payload: response,
	})
}

func (s *Server) handleMetricsUpdate(conn net.Conn, msg core.Message) {
	var nodeMetrics metrics.NodeMetrics
	if err := json.Unmarshal(msg.Payload.([]byte), &nodeMetrics); err != nil {
		return
	}

	latency := time.Since(nodeMetrics.Timestamp).Seconds() * 1000
	s.metrics.UpdateNodeMetrics(nodeMetrics, latency)
}

func (s *Server) handlePing(conn net.Conn, msg core.Message) {
	encoder := json.NewEncoder(conn)
	encoder.Encode(core.Message{
		Type:    core.MessageTypePong,
		Payload: time.Now(),
	})
}

func (s *Server) startHealthCheck() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.checkPeersHealth()
		}
	}
}

func (s *Server) checkPeersHealth() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	localMetrics := s.metrics.CollectLocalMetrics()

	for nodeID, peer := range s.peers {
		if !peer.Status.Connected {
			continue
		}

		encoder := json.NewEncoder(peer.Conn)
		err := encoder.Encode(core.Message{
			Type:    core.MessageTypeMetrics,
			Payload: localMetrics,
		})

		if err != nil {
			peer.Status.FailedPings++
			if peer.Status.FailedPings >= 3 {
				s.removePeer(nodeID)
			}
			continue
		}

		peer.Status.FailedPings = 0
		peer.Status.LastPing = time.Now()
	}
}

func (s *Server) removePeer(nodeID core.NodeID) {
	if peer, exists := s.peers[nodeID]; exists {
		if peer.Conn != nil {
			peer.Conn.Close()
		}
		if peer.OutboundConn != nil {
			peer.OutboundConn.Close()
		}
		delete(s.peers, nodeID)
	}
}
