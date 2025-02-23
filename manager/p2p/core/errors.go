package core

import "errors"

var (
	ErrNetworkFull     = errors.New("network is full")
	ErrPeerNotFound    = errors.New("peer not found")
	ErrInvalidMessage  = errors.New("invalid message")
	ErrNATPunchFailed  = errors.New("NAT punch failed")
	ErrNetworkJoinLoop = errors.New("detected loop while trying to join network")
)
