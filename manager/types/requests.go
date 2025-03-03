// Package types provides common types for API and request/response handling
package types

// SaiRequest represents a request to the Sai service
type SaiRequest struct {
	Method   string      `json:"method"`
	Data     interface{} `json:"data"`
	Metadata interface{} `json:"metadata"`
}

// SaiResponse represents a response from the Sai service
type SaiResponse struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
