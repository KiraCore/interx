package types

type GenesisChunkedResponse struct {
	Chunk string `json:"chunk"`
	Total string `json:"total"`
	Data  []byte `json:"data"`
}
