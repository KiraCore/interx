package types

import (
	types2 "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type GenesisChunkedResponse struct {
	Chunk string `json:"chunk"`
	Total string `json:"total"`
	Data  []byte `json:"data"`
}

type KiraStatus struct {
	NodeInfo      NodeInfo      `json:"node_info,omitempty"`
	SyncInfo      SyncInfo      `json:"sync_info,omitempty"`
	ValidatorInfo ValidatorInfo `json:"validator_info,omitempty"`
}

type GenesisInfo struct {
	GenesisDoc  *types2.GenesisDoc
	GenesisData []byte
}

type Coin struct {
	Amount string `json:"amount"`
	Denom  string `json:"denom"`
}

type QueryBalancesResponse struct {
	Balances []Coin `json:"balances"`
}

type Delegation struct {
	ValidatorInfo struct {
		Moniker string `json:"moniker,omitempty"`
		Address string `json:"address,omitempty"`
		ValKey  string `json:"valkey,omitempty"`
		Website string `json:"website,omitempty"`
		Logo    string `json:"logo,omitempty"`
	} `json:"validator_info"`
	PoolInfo struct {
		ID         int64    `json:"id,omitempty"`
		Commission string   `json:"commission,omitempty"`
		Status     string   `json:"status,omitempty"`
		Tokens     []string `json:"tokens,omitempty"`
	} `json:"pool_info"`
}

type QueryDelegationsResult struct {
	Delegations []Delegation `json:"delegations"`
	Pagination  struct {
		Total int `json:"total,string,omitempty"`
	} `json:"pagination,omitempty"`
}

type VerifyRecord struct {
	Address            string   `json:"address,omitempty"`
	Id                 string   `json:"id,omitempty"`
	LastRecordEditDate string   `json:"lastRecordEditDate,omitempty"`
	RecordIds          []string `json:"recordIds,omitempty"`
	Tip                string   `json:"tip,omitempty"`
	Verifier           string   `json:"verifier,omitempty"`
}

type IdVerifyRequests struct {
	VerifyRecords []VerifyRecord `json:"verifyRecords,omitempty"`
	Pagination    interface{}    `json:"pagination,omitempty"`
}

type NetworkProperties struct {
	MinTxFee                     string `json:"minTxFee"`
	MaxTxFee                     string `json:"maxTxFee"`
	VoteQuorum                   string `json:"voteQuorum"`
	MinimumProposalEndTime       string `json:"minimumProposalEndTime"`
	ProposalEnactmentTime        string `json:"proposalEnactmentTime"`
	MinProposalEndBlocks         string `json:"minProposalEndBlocks"`
	MinProposalEnactmentBlocks   string `json:"minProposalEnactmentBlocks"`
	EnableForeignFeePayments     bool   `json:"enableForeignFeePayments"`
	MischanceRankDecreaseAmount  string `json:"mischanceRankDecreaseAmount"`
	MaxMischance                 string `json:"maxMischance"`
	MischanceConfidence          string `json:"mischanceConfidence"`
	InactiveRankDecreasePercent  string `json:"inactiveRankDecreasePercent"`
	MinValidators                string `json:"minValidators"`
	PoorNetworkMaxBankSend       string `json:"poorNetworkMaxBankSend"`
	UnjailMaxTime                string `json:"unjailMaxTime"`
	EnableTokenWhitelist         bool   `json:"enableTokenWhitelist"`
	EnableTokenBlacklist         bool   `json:"enableTokenBlacklist"`
	MinIdentityApprovalTip       string `json:"minIdentityApprovalTip"`
	UniqueIdentityKeys           string `json:"uniqueIdentityKeys"`
	UbiHardcap                   string `json:"ubiHardcap"`
	ValidatorsFeeShare           string `json:"validatorsFeeShare"`
	InflationRate                string `json:"inflationRate"`
	InflationPeriod              string `json:"inflationPeriod"`
	UnstakingPeriod              string `json:"unstakingPeriod"`
	MaxDelegators                string `json:"maxDelegators"`
	MinDelegationPushout         string `json:"minDelegationPushout"`
	SlashingPeriod               string `json:"slashingPeriod"`
	MaxJailedPercentage          string `json:"maxJailedPercentage"`
	MaxSlashingPercentage        string `json:"maxSlashingPercentage"`
	MinCustodyReward             string `json:"minCustodyReward"`
	MaxCustodyBufferSize         string `json:"maxCustodyBufferSize"`
	MaxCustodyTxSize             string `json:"maxCustodyTxSize"`
	AbstentionRankDecreaseAmount string `json:"abstentionRankDecreaseAmount"`
	MaxAbstention                string `json:"maxAbstention"`
	MinCollectiveBond            string `json:"minCollectiveBond"`
	MinCollectiveBondingTime     string `json:"minCollectiveBondingTime"`
	MaxCollectiveOutputs         string `json:"maxCollectiveOutputs"`
	MinCollectiveClaimPeriod     string `json:"minCollectiveClaimPeriod"`
	ValidatorRecoveryBond        string `json:"validatorRecoveryBond"`
	MaxAnnualInflation           string `json:"maxAnnualInflation"`
	MaxProposalTitleSize         string `json:"maxProposalTitleSize"`
	MaxProposalDescriptionSize   string `json:"maxProposalDescriptionSize"`
	MaxProposalPollOptionSize    string `json:"maxProposalPollOptionSize"`
	MaxProposalPollOptionCount   string `json:"maxProposalPollOptionCount"`
	MaxProposalReferenceSize     string `json:"maxProposalReferenceSize"`
	MaxProposalChecksumSize      string `json:"maxProposalChecksumSize"`
	MinDappBond                  string `json:"minDappBond"`
	MaxDappBond                  string `json:"maxDappBond"`
	DappBondDuration             string `json:"dappBondDuration"`
	DappVerifierBond             string `json:"dappVerifierBond"`
	DappAutoDenounceTime         string `json:"dappAutoDenounceTime"`
}

type NetworkPropertiesResponse struct {
	Properties *NetworkProperties `json:"properties"`
}

type QueryStakingPoolDelegatorsResponse struct {
	Pool       ValidatorPool `json:"pool"`
	Delegators []string      `json:"delegators,omitempty"`
}

type QueryValidatorPoolResult struct {
	ID              int64      `json:"id,omitempty"`
	Slashed         string     `json:"slashed"`
	Commission      string     `json:"commission"`
	TotalDelegators int64      `json:"total_delegators"`
	VotingPower     []sdk.Coin `json:"voting_power"`
	Tokens          []string   `json:"tokens"`
}

type Undelegation struct {
	ID         uint64   `json:"id,string"`
	Address    string   `json:"address"`
	ValAddress string   `json:"valaddress"`
	Expiry     string   `json:"expiry"`
	Amount     []string `json:"amount"`
}

type QueryUndelegationsResult struct {
	Undelegations []Undelegation `json:"undelegations"`
}
