package network

import (
	"net"
	"time"

	"github.com/saiset-co/sai-interx-manager/p2p/core"
)

type PeerStatus struct {
	Connected   bool
	LastPing    time.Time
	FailedPings int
	NATPort     int
}

type Peer struct {
	NodeID       core.NodeID
	Address      string
	Conn         net.Conn
	OutboundConn net.Conn
	Status       PeerStatus
}
