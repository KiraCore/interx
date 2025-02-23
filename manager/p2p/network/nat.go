package network

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/saiset-co/sai-interx-manager/p2p/core"
)

func GetNATPort(conn net.Conn) int {
	addr := conn.RemoteAddr().String()
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return 0
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0
	}
	return port
}

func EstablishNATConnection(peer *Peer) error {
	if peer.Status.NATPort <= 0 {
		return core.ErrNATPunchFailed
	}

	host, _, err := net.SplitHostPort(peer.Address)
	if err != nil {
		return fmt.Errorf("invalid peer address: %w", err)
	}

	natAddr := fmt.Sprintf("%s:%d", host, peer.Status.NATPort)
	conn, err := net.DialTimeout("tcp", natAddr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to establish NAT connection: %w", err)
	}

	peer.OutboundConn = conn
	return nil
}
