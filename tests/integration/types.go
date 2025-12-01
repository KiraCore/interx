package integration

// Expected Response Types based on Miro Frontend DTOs
// Source: /home/ubuntu/Code/github.com/kiracore/miro/lib/infra/dto/
// These define what the frontend EXPECTS from the API

// ============================================================================
// Account Endpoint Types (/api/kira/accounts/{address})
// ============================================================================

// ExpectedPubKey represents the expected pub_key structure
type ExpectedPubKey struct {
	Type  string `json:"@type"`
	Value string `json:"value"`
}

// ExpectedAccountResponse - miro expects snake_case fields
// Source: lib/infra/dto/api_kira/query_account/response/
type ExpectedAccountResponse struct {
	Type          string          `json:"@type"`
	AccountNumber string          `json:"account_number"` // NOT accountNumber
	Address       string          `json:"address"`
	PubKey        *ExpectedPubKey `json:"pub_key"` // NOT pubKey
	Sequence      string          `json:"sequence"`
}

// ============================================================================
// Balance Endpoint Types (/api/kira/balances/{address})
// ============================================================================

// ExpectedCoin - shared type for amount/denom pairs
type ExpectedCoin struct {
	Amount string `json:"amount"` // String for precision
	Denom  string `json:"denom"`
}

// ExpectedPagination - miro expects total as string
type ExpectedPagination struct {
	Total string `json:"total"` // String, not int
}

// ExpectedBalanceResponse
// Source: lib/infra/dto/api_kira/query_balance/response/
type ExpectedBalanceResponse struct {
	Balances   []ExpectedCoin      `json:"balances"`
	Pagination *ExpectedPagination `json:"pagination"` // Required by miro
}

// ============================================================================
// Transaction Endpoint Types (/api/transactions)
// ============================================================================

// ExpectedTransaction - miro expects specific fields
// Source: lib/infra/dto/api/query_transactions/response/transaction.dart
type ExpectedTransaction struct {
	Time      int            `json:"time"`      // Unix timestamp as INT (not string, not ISO)
	Hash      string         `json:"hash"`
	Status    string         `json:"status"`    // "confirmed", "pending", "failed"
	Direction string         `json:"direction"` // "inbound" or "outbound"
	Memo      string         `json:"memo"`
	Fee       []ExpectedCoin `json:"fee"`
	Txs       []interface{}  `json:"txs"` // NOT "messages"
}

// ExpectedTransactionsResponse
// Source: lib/infra/dto/api/query_transactions/response/query_transactions_resp.dart
type ExpectedTransactionsResponse struct {
	Transactions []ExpectedTransaction `json:"transactions"`
	TotalCount   int                   `json:"total_count"` // INT, snake_case
}

// ============================================================================
// Validator Endpoint Types (/api/valopers)
// ============================================================================

// ExpectedValidatorStatus - counts as integers
// Source: lib/infra/dto/api/query_validators/response/status.dart
type ExpectedValidatorStatus struct {
	ActiveValidators   int `json:"active_validators"`
	PausedValidators   int `json:"paused_validators"`
	InactiveValidators int `json:"inactive_validators"`
	JailedValidators   int `json:"jailed_validators"`
	TotalValidators    int `json:"total_validators"`
	WaitingValidators  int `json:"waiting_validators"`
}

// ExpectedValidator - all fields as strings
// Source: lib/infra/dto/api/query_validators/response/validator.dart
type ExpectedValidator struct {
	Top                   string  `json:"top"`
	Address               string  `json:"address"`
	Valkey                string  `json:"valkey"`
	Pubkey                string  `json:"pubkey"`
	Proposer              string  `json:"proposer"`
	Moniker               string  `json:"moniker"`
	Status                string  `json:"status"`
	Rank                  string  `json:"rank"`
	Streak                string  `json:"streak"`
	Mischance             string  `json:"mischance"`
	MischanceConfidence   string  `json:"mischance_confidence"`
	StartHeight           string  `json:"start_height"`
	InactiveUntil         string  `json:"inactive_until"`
	LastPresentBlock      string  `json:"last_present_block"`
	MissedBlocksCounter   string  `json:"missed_blocks_counter"`
	ProducedBlocksCounter string  `json:"produced_blocks_counter"`
	StakingPoolId         *string `json:"staking_pool_id,omitempty"`
	StakingPoolStatus     *string `json:"staking_pool_status,omitempty"`
	Description           *string `json:"description,omitempty"`
	Website               *string `json:"website,omitempty"`
	Logo                  *string `json:"logo,omitempty"`
	Social                *string `json:"social,omitempty"`
	Contact               *string `json:"contact,omitempty"`
	ValidatorNodeId       *string `json:"validator_node_id,omitempty"`
	SentryNodeId          *string `json:"sentry_node_id,omitempty"`
}

// ExpectedValidatorsResponse
// Source: lib/infra/dto/api/query_validators/response/query_validators_resp.dart
type ExpectedValidatorsResponse struct {
	Waiting    []string                 `json:"waiting"`
	Validators []ExpectedValidator      `json:"validators"`
	Status     *ExpectedValidatorStatus `json:"status,omitempty"`
}

// ============================================================================
// Status Endpoint Types (/api/status)
// ============================================================================

// ExpectedNode - nested in InterxInfo
type ExpectedNode struct {
	NodeType       string `json:"node_type"`
	SeedNodeId     string `json:"seed_node_id"`
	SentryNodeId   string `json:"sentry_node_id"`
	SnapshotNodeId string `json:"snapshot_node_id"`
	ValidatorNodeId string `json:"validator_node_id"`
}

// ExpectedInterxInfo
// Source: lib/infra/dto/api/query_interx_status/interx_info.dart
type ExpectedInterxInfo struct {
	CatchingUp        bool            `json:"catching_up"` // BOOL
	ChainId           string          `json:"chain_id"`
	GenesisChecksum   string          `json:"genesis_checksum"`
	KiraAddr          string          `json:"kira_addr"`
	KiraPubKey        string          `json:"kira_pub_key"`
	LatestBlockHeight string          `json:"latest_block_height"`
	Moniker           string          `json:"moniker"`
	Version           string          `json:"version"`
	FaucetAddr        *string         `json:"faucet_addr,omitempty"`
	Node              *ExpectedNode   `json:"node,omitempty"`
	PubKey            *ExpectedPubKey `json:"pub_key,omitempty"`
}

// ExpectedProtocolVersion
type ExpectedProtocolVersion struct {
	App   string `json:"app"`
	Block string `json:"block"`
	P2p   string `json:"p2p"`
}

// ExpectedNodeInfo
// Source: lib/infra/dto/api/query_interx_status/node_info.dart
type ExpectedNodeInfo struct {
	Channels        string                   `json:"channels"`
	Id              string                   `json:"id"`
	ListenAddr      string                   `json:"listen_addr"`
	Moniker         string                   `json:"moniker"`
	Network         string                   `json:"network"`
	Version         string                   `json:"version"`
	ProtocolVersion *ExpectedProtocolVersion `json:"protocol_version,omitempty"`
	Other           map[string]interface{}   `json:"other,omitempty"`
}

// ExpectedSyncInfo
// Source: lib/infra/dto/api/query_interx_status/sync_info.dart
type ExpectedSyncInfo struct {
	EarliestAppHash     string `json:"earliest_app_hash"`
	EarliestBlockHash   string `json:"earliest_block_hash"`
	EarliestBlockHeight string `json:"earliest_block_height"`
	EarliestBlockTime   string `json:"earliest_block_time"`
	LatestAppHash       string `json:"latest_app_hash"`
	LatestBlockHash     string `json:"latest_block_hash"`
	LatestBlockHeight   string `json:"latest_block_height"`
	LatestBlockTime     string `json:"latest_block_time"`
	CatchingUp          bool   `json:"catching_up"`
}

// ExpectedStatusValidatorInfo
type ExpectedStatusValidatorInfo struct {
	Address     string          `json:"address"`
	PubKey      *ExpectedPubKey `json:"pub_key"`
	VotingPower string          `json:"voting_power"`
}

// ExpectedStatusResponse
// Source: lib/infra/dto/api/query_interx_status/query_interx_status_resp.dart
type ExpectedStatusResponse struct {
	Id            string                       `json:"id"`
	InterxInfo    *ExpectedInterxInfo          `json:"interx_info"`
	NodeInfo      *ExpectedNodeInfo            `json:"node_info"`
	SyncInfo      *ExpectedSyncInfo            `json:"sync_info"`
	ValidatorInfo *ExpectedStatusValidatorInfo `json:"validator_info"`
}

// ============================================================================
// Dashboard Endpoint Types (/api/dashboard)
// ============================================================================

// ExpectedCurrentBlockValidator
type ExpectedCurrentBlockValidator struct {
	Moniker string `json:"moniker"`
	Address string `json:"address"`
}

// ExpectedDashboardValidators - counts as integers
type ExpectedDashboardValidators struct {
	ActiveValidators   int `json:"active_validators"`
	PausedValidators   int `json:"paused_validators"`
	InactiveValidators int `json:"inactive_validators"`
	JailedValidators   int `json:"jailed_validators"`
	TotalValidators    int `json:"total_validators"`
	WaitingValidators  int `json:"waiting_validators"`
}

// ExpectedDashboardBlocks
// Source: lib/infra/dto/api/dashboard/blocks.dart
type ExpectedDashboardBlocks struct {
	CurrentHeight       int     `json:"current_height"`
	SinceGenesis        int     `json:"since_genesis"`
	PendingTransactions int     `json:"pending_transactions"`
	CurrentTransactions int     `json:"current_transactions"`
	LatestTime          float64 `json:"latest_time"`
	AverageTime         float64 `json:"average_time"`
}

// ExpectedDashboardProposals
// Source: lib/infra/dto/api/dashboard/proposals.dart
type ExpectedDashboardProposals struct {
	Total      int    `json:"total"`
	Active     int    `json:"active"`
	Enacting   int    `json:"enacting"`
	Finished   int    `json:"finished"`
	Successful int    `json:"successful"`
	Proposers  string `json:"proposers"`
	Voters     string `json:"voters"`
}

// ExpectedDashboardResponse
// Source: lib/infra/dto/api/dashboard/dashboard_resp.dart
type ExpectedDashboardResponse struct {
	ConsensusHealth       string                         `json:"consensus_health"`
	CurrentBlockValidator *ExpectedCurrentBlockValidator `json:"current_block_validator"`
	Validators            *ExpectedDashboardValidators   `json:"validators"`
	Blocks                *ExpectedDashboardBlocks       `json:"blocks"`
	Proposals             *ExpectedDashboardProposals    `json:"proposals"`
}

// ============================================================================
// Governance Endpoint Types
// ============================================================================

// ExpectedNetworkProperties - ~47 fields, all strings except 3 bools
// Source: lib/infra/dto/api_kira/query_network_properties/response/properties.dart
type ExpectedNetworkProperties struct {
	MinTxFee                    string `json:"min_tx_fee"`
	VoteQuorum                  string `json:"vote_quorum"`
	MinValidators               string `json:"min_validators"`
	UnstakingPeriod             string `json:"unstaking_period"`
	MaxDelegators               string `json:"max_delegators"`
	MinDelegationPushout        string `json:"min_delegation_pushout"`
	MaxMischance                string `json:"max_mischance"`
	MischanceConfidence         string `json:"mischance_confidence"`
	MaxJailedPercentage         string `json:"max_jailed_percentage"`
	MaxSlashingPercentage       string `json:"max_slashing_percentage"`
	UnjailMaxTime               string `json:"unjail_max_time"`
	EnableForeignFeePayments    bool   `json:"enable_foreign_fee_payments"`
	EnableTokenBlacklist        bool   `json:"enable_token_blacklist"`
	EnableTokenWhitelist        bool   `json:"enable_token_whitelist"`
	MinProposalEndBlocks        string `json:"min_proposal_end_blocks"`
	MinProposalEnactmentBlocks  string `json:"min_proposal_enactment_blocks"`
	MinimumProposalEndTime      string `json:"minimum_proposal_end_time"`
	ProposalEnactmentTime       string `json:"proposal_enactment_time"`
	InflationRate               string `json:"inflation_rate"`
	InflationPeriod             string `json:"inflation_period"`
	MaxAnnualInflation          string `json:"max_annual_inflation"`
	MaxProposalTitleSize        string `json:"max_proposal_title_size"`
	MaxProposalDescriptionSize  string `json:"max_proposal_description_size"`
	MaxProposalReferenceSize    string `json:"max_proposal_reference_size"`
	MaxProposalChecksumSize     string `json:"max_proposal_checksum_size"`
	MaxProposalPollOptionSize   string `json:"max_proposal_poll_option_size"`
	MaxProposalPollOptionCount  string `json:"max_proposal_poll_option_count"`
	ValidatorsFeeShare          string `json:"validators_fee_share"`
	ValidatorRecoveryBond       string `json:"validator_recovery_bond"`
	MaxAbstention               string `json:"max_abstention"`
	AbstentionRankDecreaseAmount string `json:"abstention_rank_decrease_amount"`
	MischanceRankDecreaseAmount string `json:"mischance_rank_decrease_amount"`
	InactiveRankDecreasePercent string `json:"inactive_rank_decrease_percent"`
	PoorNetworkMaxBankSend      string `json:"poor_network_max_bank_send"`
	MinIdentityApprovalTip      string `json:"min_identity_approval_tip"`
	UniqueIdentityKeys          string `json:"unique_identity_keys"`
	UbiHardcap                  string `json:"ubi_hardcap"`
}

// ExpectedNetworkPropertiesResponse
type ExpectedNetworkPropertiesResponse struct {
	Properties *ExpectedNetworkProperties `json:"properties"`
}

// ExpectedExecutionFee
// Source: lib/infra/dto/api_kira/query_execution_fee/response/fee.dart
type ExpectedExecutionFee struct {
	DefaultParameters string `json:"default_parameters"`
	ExecutionFee      string `json:"execution_fee"`
	FailureFee        string `json:"failure_fee"`
	Timeout           string `json:"timeout"`
	TransactionType   string `json:"transaction_type"`
}

// ExpectedExecutionFeeResponse
type ExpectedExecutionFeeResponse struct {
	Fee *ExpectedExecutionFee `json:"fee"`
}

// ============================================================================
// Token Endpoint Types
// ============================================================================

// ExpectedTokenAlias
// Source: lib/infra/dto/api_kira/query_kira_tokens_aliases/response/token_alias.dart
type ExpectedTokenAlias struct {
	Decimals int      `json:"decimals"` // INT
	Denoms   []string `json:"denoms"`
	Name     string   `json:"name"`
	Symbol   string   `json:"symbol"`
	Icon     string   `json:"icon"`
	Amount   string   `json:"amount"`
}

// ExpectedTokenAliasesResponse
// Source: lib/infra/dto/api_kira/query_kira_tokens_aliases/response/
type ExpectedTokenAliasesResponse struct {
	TokenAliasesData []ExpectedTokenAlias `json:"token_aliases_data"`
	DefaultDenom     string               `json:"default_denom"`
	Bech32Prefix     string               `json:"bech32_prefix"`
}

// ExpectedTokenRate
// Source: lib/infra/dto/api_kira/query_kira_tokens_rates/response/token_rate.dart
type ExpectedTokenRate struct {
	Denom       string `json:"denom"`
	FeePayments bool   `json:"fee_payments"` // BOOL
	FeeRate     string `json:"fee_rate"`
	StakeCap    string `json:"stake_cap"`
	StakeMin    string `json:"stake_min"`
	StakeToken  bool   `json:"stake_token"` // BOOL
}

// ExpectedTokenRatesResponse
type ExpectedTokenRatesResponse struct {
	Data []ExpectedTokenRate `json:"data"`
}

// ============================================================================
// Staking Endpoint Types
// ============================================================================

// ExpectedStakingPoolResponse
// Source: lib/infra/dto/api_kira/query_staking_pool/response/
type ExpectedStakingPoolResponse struct {
	Id              int            `json:"id"`               // INT
	TotalDelegators int            `json:"total_delegators"` // INT
	Commission      string         `json:"commission"`
	Slashed         string         `json:"slashed"`
	Tokens          []string       `json:"tokens"`
	VotingPower     []ExpectedCoin `json:"voting_power"`
}

// ExpectedValidatorInfo - for delegations
type ExpectedValidatorInfo struct {
	Moniker string  `json:"moniker"`
	Address string  `json:"address"`
	Valkey  string  `json:"valkey"`
	Website *string `json:"website,omitempty"`
	Logo    *string `json:"logo,omitempty"`
}

// ExpectedPoolInfo
type ExpectedPoolInfo struct {
	Id         int      `json:"id"` // INT
	Commission string   `json:"commission"`
	Status     string   `json:"status"`
	Tokens     []string `json:"tokens"`
}

// ExpectedDelegation
type ExpectedDelegation struct {
	ValidatorInfo *ExpectedValidatorInfo `json:"validator_info"`
	PoolInfo      *ExpectedPoolInfo      `json:"pool_info"`
}

// ExpectedDelegationsResponse
type ExpectedDelegationsResponse struct {
	Delegations []ExpectedDelegation `json:"delegations"`
	Pagination  *ExpectedPagination  `json:"pagination"`
}

// ============================================================================
// Identity Endpoint Types
// ============================================================================

// ExpectedIdentityRecord
// Source: lib/infra/dto/api_kira/query_identity_records/response/
type ExpectedIdentityRecord struct {
	Address   string   `json:"address"`
	Date      string   `json:"date"` // ISO 8601
	Id        string   `json:"id"`
	Key       string   `json:"key"`
	Value     string   `json:"value"`
	Verifiers []string `json:"verifiers"`
}

// ExpectedIdentityRecordsResponse
type ExpectedIdentityRecordsResponse struct {
	Records []ExpectedIdentityRecord `json:"records"`
}
