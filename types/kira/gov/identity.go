package gov

type VerifyRecord struct {
	Address            string   `json:"address,omitempty"`
	Id                 string   `json:"id,omitempty"`
	LastRecordEditDate string   `json:"lastRecordEditDate,omitempty"`
	RecordIds          []string `json:"recordIds,omitempty"`
	Tip                string   `json:"tip,omitempty"`
	Verifier           string   `json:"verifier,omitempty"`
}

// Used to parse response from sekai gRPC ("/kira/gov/identity_verify_requests_by_requester/{requester}")
type IdVerifyRequests struct {
	VerifyRecords []VerifyRecord `json:"verifyRecords,omitempty"`
	Pagination    interface{}    `json:"pagination,omitempty"`
}
