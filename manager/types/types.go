package types

type GRPCResponse struct {
	Code    float64 `json:"code"`
	Message string  `json:"message"`
	Details []byte  `json:"details"`
}

type EVMStatus struct {
	NodeInfo struct {
		Network    uint64 `json:"network"`
		RPCAddress string `json:"rpcAddress"`
		Version    struct {
			Net      string `json:"net"`
			Web3     string `json:"web3"`
			Protocol string `json:"protocol"`
		} `json:"version"`
	} `json:"nodeInfo"`
	SyncInfo struct {
		CatchingUp          bool   `json:"catchingUp"`
		EarliestBlockHash   string `json:"earliestBlockHash"`
		EarliestBlockHeight uint64 `json:"earliestBlockHeight"`
		EarliestBlockTime   uint64 `json:"earliestBlockTime"`
		LatestBlockHash     string `json:"latestBlockHash"`
		LatestBlockHeight   uint64 `json:"latestBlockHeight"`
		LatestBlockTime     uint64 `json:"latestBlockTime"`
	} `json:"syncInfo"`
	GasPrice string `json:"gasPrice"`
}

type IdentityRecord struct {
	ID        uint64      `json:"id,string"`
	Key       string      `json:"key"`
	Value     string      `json:"value"`
	Date      interface{} `json:"date"`
	Verifiers []string    `json:"verifiers"`
}

type QueryValidator struct {
	Top int `json:"top,string"`

	Address             string           `json:"address"`
	Valkey              string           `json:"valkey"`
	Pubkey              string           `json:"pubkey"`
	Proposer            string           `json:"proposer"`
	Moniker             string           `json:"moniker"`
	Status              string           `json:"status"`
	Rank                int64            `json:"rank,string"`
	Streak              int64            `json:"streak,string"`
	Mischance           int64            `json:"mischance,string"`
	MischanceConfidence int64            `json:"mischanceConfidence,string"`
	Identity            []IdentityRecord `json:"identity,omitempty"`

	StartHeight           int64  `json:"startHeight,string"`
	InactiveUntil         string `json:"inactiveUntil"`
	LastPresentBlock      int64  `json:"lastPresentBlock,string"`
	MissedBlocksCounter   int64  `json:"missedBlocks_counter,string"`
	ProducedBlocksCounter int64  `json:"producedBlocksCounter,string"`
	StakingPoolId         int64  `json:"stakingPoolId,string,omitempty"`
	StakingPoolStatus     string `json:"stakingPool_status,omitempty"`

	Description       string `json:"description,omitempty"`
	Website           string `json:"website,omitempty"`
	Logo              string `json:"logo,omitempty"`
	Social            string `json:"social,omitempty"`
	Contact           string `json:"contact,omitempty"`
	Validator_node_id string `json:"validatorNodeId,omitempty"`
	Sentry_node_id    string `json:"sentryNodeId,omitempty"`
}

type QueryValidators []QueryValidator

func (s QueryValidators) Len() int {
	return len(s)
}
func (s QueryValidators) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s QueryValidators) Less(i, j int) bool {
	if s[i].Status != s[j].Status {
		if s[j].Status == "ACTIVE" {
			return false
		}
		if s[i].Status == "ACTIVE" {
			return true
		}
		return s[i].Status > s[j].Status
	}
	if s[i].Rank != s[j].Rank {
		return s[i].Rank > s[j].Rank
	}
	if s[i].Streak != s[j].Streak {
		return s[i].Streak > s[j].Streak
	}
	if s[i].MissedBlocksCounter != s[j].MissedBlocksCounter {
		return s[i].MissedBlocksCounter < s[j].MissedBlocksCounter
	}

	return false
}

type TokenRatesResponse struct {
	Data []TokenRate `json:"data"`
}

type ValidatorsResponse = struct {
	Validators []QueryValidator `json:"validators,omitempty"`
	Actors     []string         `json:"actors,omitempty"`
	Pagination struct {
		Total int `json:"total,string,omitempty"`
	} `json:"pagination,omitempty"`
}

type ValidatorInfoResponse = struct {
	ValValidatorInfos []ValidatorSigningInfo `json:"info,omitempty"`
	Validators        []QueryValidator       `json:"validators,omitempty"`
}

type AllValidators struct {
	Status struct {
		ActiveValidators   int `json:"activeValidators"`
		PausedValidators   int `json:"pausedValidators"`
		InactiveValidators int `json:"inactiveValidators"`
		JailedValidators   int `json:"jailedValidators"`
		TotalValidators    int `json:"totalValidators"`
		WaitingValidators  int `json:"waitingValidators"`
	} `json:"status"`
	PoolTokens      []string                 `json:"-"`
	AddrToValidator map[string]string        `json:"-"`
	PoolToValidator map[int64]QueryValidator `json:"-"`
	Waiting         []string                 `json:"waiting"`
	Validators      []QueryValidator         `json:"validators"`
}

type TokenRate struct {
	Denom       string `json:"denom"`
	FeePayments bool   `json:"feePayments"`
	FeeRate     string `json:"feeRate"`
	StakeCap    string `json:"stakeCap"`
	StakeMin    string `json:"stakeMin"`
	StakeToken  bool   `json:"stakeToken"`
}

type ValidatorSigningInfo struct {
	Address               string `json:"address"`
	StartHeight           int64  `json:"startHeight,string"`
	InactiveUntil         string `json:"inactiveUntil"`
	MischanceConfidence   int64  `json:"mischanceConfidence,string"`
	Mischance             int64  `json:"mischance,string"`
	LastPresentBlock      int64  `json:"lastPresentBlock,string"`
	MissedBlocksCounter   int64  `json:"missedBlocksCounter,string"`
	ProducedBlocksCounter int64  `json:"producedBlocksCounter,string"`
}

type ValidatorPool struct {
	ID                 int64    `json:"id,string"`
	Validator          string   `json:"validator,omitempty"`
	Enabled            bool     `json:"enabled,omitempty"`
	Slashed            string   `json:"slashed"`
	TotalStakingTokens []string `json:"totalStakingTokens"`
	TotalShareTokens   []string `json:"totalShareTokens"`
	TotalRewards       []string `json:"totalRewards"`
	Commission         string   `json:"commission"`
}

type AllPools struct {
	ValToPool map[string]ValidatorPool
	IdToPool  map[int64]ValidatorPool
}

type NodeConfig struct {
	NodeType        string `json:"nodeType"`
	SentryNodeID    string `json:"sentryNodeId"`
	SnapshotNodeID  string `json:"snapshotNodeId"`
	ValidatorNodeID string `json:"validatorNodeId"`
	SeedNodeID      string `json:"seedNodeId"`
}

type ProtocolVersion struct {
	P2P   string `json:"p2p,omitempty"`
	Block string `json:"block,omitempty"`
	App   string `json:"app,omitempty"`
}

type NodeOtherInfo struct {
	TxIndex    string `json:"txIndex,omitempty"`
	RpcAddress string `json:"rpcAddress,omitempty"`
}

type NodeInfo struct {
	ProtocolVersion ProtocolVersion `json:"protocol_version,omitempty"`
	Id              string          `json:"id,omitempty"`
	ListenAddr      string          `json:"listenAddr,omitempty"`
	Network         string          `json:"network,omitempty"`
	Version         string          `json:"version,omitempty"`
	Channels        string          `json:"channels,omitempty"`
	Moniker         string          `json:"moniker,omitempty"`
	Other           NodeOtherInfo   `json:"other,omitempty"`
}

type SyncInfo struct {
	LatestBlockHash     string `json:"latestBlockHash,omitempty"`
	LatestAppHash       string `json:"latestAppHash,omitempty"`
	LatestBlockHeight   string `json:"latestBlockHeight,omitempty"`
	LatestBlockTime     string `json:"latestBlockTime,omitempty"`
	EarliestBlockHash   string `json:"earliestBlockHash,omitempty"`
	EarliestAppHash     string `json:"earliestAppHash,omitempty"`
	EarliestBlockHeight string `json:"earliestBlockHeight,omitempty"`
	EarliestBlockTime   string `json:"earliestBlockTime,omitempty"`
	CatchingUp          bool   `json:"catchingUp,omitempty"`
}

type PubKey struct {
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

type ValidatorInfo struct {
	Address     string  `json:"address,omitempty"`
	PubKey      *PubKey `json:"pubKey,omitempty"`
	VotingPower string  `json:"votingPower,omitempty"`
}

type InterxStatus struct {
	ID         string `json:"id"`
	InterxInfo struct {
		PubKey struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"pubKey,omitempty"`
		Moniker           string     `json:"moniker"`
		KiraAddr          string     `json:"kiraAddr"`
		KiraPubKey        string     `json:"kiraPpubKey"`
		FaucetAddr        string     `json:"faucetAddr"`
		GenesisChecksum   string     `json:"genesisChecksum"`
		ChainID           string     `json:"chainId"`
		InterxVersion     string     `json:"version,omitempty"`
		SekaiVersion      string     `json:"sekaiVersion,omitempty"`
		LatestBlockHeight string     `json:"latestBlockHeight"`
		CatchingUp        bool       `json:"catchingUp"`
		Node              NodeConfig `json:"node"`
	} `json:"interxInfo,omitempty"`
	NodeInfo      NodeInfo      `json:"nodeInfo,omitempty"`
	SyncInfo      SyncInfo      `json:"syncInfo,omitempty"`
	ValidatorInfo ValidatorInfo `json:"validatorInfo,omitempty"`
}
