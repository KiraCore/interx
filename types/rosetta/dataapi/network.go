package dataapi

import (
	"github.com/KiraCore/interx/types/rosetta"
)

// Used for parsing params of interx request
type MetadataRequest struct {
	Metadata interface{} `json:"metadata,omitempty"`
}

type NetworkListRequest MetadataRequest

// Used for interx response
type NetworkListResponse struct {
	NetworkIdentifiers []rosetta.NetworkIdentifier `json:"network_identifiers"`
}

type NetworkRequest struct {
	NetworkIdentifier rosetta.NetworkIdentifier `json:"network_identifier,omitempty"` // make it omitable
	Metadata          interface{}               `json:"metadata,omitempty"`
}

// Used for interx response
type NetworkOptionsResponse struct {
	Version rosetta.Version `json:"version"`
	Allow   rosetta.Allow   `json:"allow"`
}

// Used for parsing params of interx request
type NetworkStatusRequest NetworkRequest

// Used for interx response
type NetworkStatusResponse struct {
	CurrentBlockIdentifier rosetta.BlockIdentifier `json:"current_block_identifier"`
	CurrentBlockTimestamp  int64                   `json:"current_block_timestamp"`
	GenesisBlockIdentifier rosetta.BlockIdentifier `json:"genesis_block_identifier"`
	OldestBlockIdentifier  rosetta.BlockIdentifier `json:"oldest_block_identifier,omitempty"`
	SyncStatus             rosetta.SyncStatus      `json:"sync_status,omitempty"`
	Peers                  []rosetta.Peer          `json:"peers"`
}
