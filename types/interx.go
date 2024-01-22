package types

type InterxStatus struct {
	ID         string `json:"id"`
	InterxInfo struct {
		PubKey struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"pub_key,omitempty"`
		Moniker           string `json:"moniker"`
		KiraAddr          string `json:"kira_addr"`
		AppAddr           string `json:"app_addr"`
		KiraPubKey        string `json:"kira_pub_key"`
		FaucetAddr        string `json:"faucet_addr"`
		GenesisChecksum   string `json:"genesis_checksum"`
		ChainID           string `json:"chain_id"`
		InterxVersion     string `json:"version,omitempty"`
		SekaiVersion      string `json:"sekai_version,omitempty"`
		LatestBlockHeight string `json:"latest_block_height"`
		CatchingUp        bool   `json:"catching_up"`
		NodeType          string `json:"node_type"`
	} `json:"interx_info,omitempty"`
	NodeInfo      NodeInfo      `json:"node_info,omitempty"`
	SyncInfo      SyncInfo      `json:"sync_info,omitempty"`
	ValidatorInfo ValidatorInfo `json:"validator_info,omitempty"`
	AppInfo       struct {
		Name string `json:"name"`
		Abr  uint64 `json:"abr"`
		Mode string `json:"mode"`
		Mock bool   `json:"mock"`
	} `json:"app_info"`
}

type SnapShotChecksumResponse struct {
	Size     int64  `json:"size,omitempty"`
	Checksum string `json:"checksum,omitempty"`
}
