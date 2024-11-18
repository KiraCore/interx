package dataapi

import "github.com/KiraCore/interx/types/rosetta"

// Used for parsing params of interx request
type AccountBalanceRequest struct {
	NetworkIdentifier rosetta.NetworkIdentifier      `json:"network_identifier"`
	AccountIdentifier rosetta.AccountIdentifier      `json:"account_identifier"`
	BlockIdentifier   rosetta.PartialBlockIdentifier `json:"block_identifier,omitempty"`
	Currencies        []rosetta.Currency             `json:"currencies,omitempty"`
}

// Used for interx response
type AccountBalanceResponse struct {
	BlockIdentifier rosetta.BlockIdentifier `json:"block_identifier"`
	Balances        []rosetta.Amount        `json:"balances"`
	Metadata        interface{}             `json:"metadata,omitempty"`
}
