package common

import (
	"encoding/json"

	"github.com/KiraCore/interx/log"
	govTypes "github.com/KiraCore/interx/types/kira/gov"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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

func QueryVotersFromGrpcResult(success interface{}) ([]govTypes.Voter, error) {
	result := struct {
		Voters []struct {
			Address     []byte                `json:"address,omitempty"`
			Roles       []string              `json:"roles,omitempty"`
			Status      string                `json:"status,omitempty"`
			Votes       []string              `json:"votes,omitempty"`
			Permissions *govTypes.Permissions `json:"permissions,omitempty"`
			Skin        uint64                `json:"skin,string,omitempty"`
		} `json:"voters,omitempty"`
	}{}

	byteData, err := json.Marshal(success)
	if err != nil {
		log.CustomLogger().Error("[QueryVotersFromGrpcResult] Failed to marshal response.",
			"error", err,
		)
		return nil, err
	}

	err = json.Unmarshal(byteData, &result)
	if err != nil {
		log.CustomLogger().Error("[QueryVotersFromGrpcResult] Failed to unmarshal response.",
			"error", err,
		)
		return nil, err
	}

	voters := make([]govTypes.Voter, 0)

	for _, voter := range result.Voters {
		newVoter := govTypes.Voter{}

		newVoter.Address = sdk.MustBech32ifyAddressBytes(sdk.GetConfig().GetBech32AccountAddrPrefix(), voter.Address)
		newVoter.Roles = voter.Roles
		newVoter.Status = voter.Status
		newVoter.Votes = voter.Votes

		newVoter.Permissions.Blacklist = make([]string, 0)
		for _, black := range voter.Permissions.Blacklist {
			newVoter.Permissions.Blacklist = append(newVoter.Permissions.Blacklist, govTypes.PermValue_name[int32(black)])
		}
		newVoter.Permissions.Whitelist = make([]string, 0)
		for _, white := range voter.Permissions.Whitelist {
			newVoter.Permissions.Whitelist = append(newVoter.Permissions.Whitelist, govTypes.PermValue_name[int32(white)])
		}

		newVoter.Skin = voter.Skin

		voters = append(voters, newVoter)
	}
	return voters, nil
}

func QueryVotesFromGrpcResult(success interface{}) ([]govTypes.Vote, error) {
	result := struct {
		Votes []struct {
			ProposalID uint64 `json:"proposalID,string,omitempty"`
			Voter      []byte `json:"voter,omitempty"`
			Option     string `json:"option,omitempty"`
		} `json:"votes,omitempty"`
	}{}

	byteData, err := json.Marshal(success)
	if err != nil {
		log.CustomLogger().Error("[QueryVotesFromGrpcResult] Failed to marshal response.",
			"error", err,
		)
		return nil, err
	}

	err = json.Unmarshal(byteData, &result)
	if err != nil {
		log.CustomLogger().Error("[QueryVotesFromGrpcResult] Failed to unmarshal response.",
			"error", err,
		)
		return nil, err
	}

	votes := make([]govTypes.Vote, 0)

	for _, vote := range result.Votes {
		newVote := govTypes.Vote{}

		newVote.ProposalID = vote.ProposalID
		newVote.Voter = sdk.MustBech32ifyAddressBytes(sdk.GetConfig().GetBech32AccountAddrPrefix(), vote.Voter)
		newVote.Option = vote.Option

		votes = append(votes, newVote)
	}

	return votes, nil
}

func QueryNetworkPropertiesFromGrpcResult(success interface{}) (NetworkPropertiesResponse, error) {
	result := NetworkPropertiesResponse{}
	byteData, err := json.Marshal(success)
	if err != nil {
		log.CustomLogger().Error("[QueryNetworkPropertiesFromGrpcResult] Failed to marshal response.",
			"error", err,
		)
		return NetworkPropertiesResponse{}, err
	}
	err = json.Unmarshal(byteData, &result)
	if err != nil {
		log.CustomLogger().Error("[QueryNetworkPropertiesFromGrpcResult] Failed to unmarshal response.",
			"error", err,
		)
		return NetworkPropertiesResponse{}, err
	}

	result.Properties.InactiveRankDecreasePercent = ConvertRate(result.Properties.InactiveRankDecreasePercent)
	result.Properties.ValidatorsFeeShare = ConvertRate(result.Properties.ValidatorsFeeShare)
	result.Properties.InflationRate = ConvertRate(result.Properties.InflationRate)
	result.Properties.MaxSlashingPercentage = ConvertRate(result.Properties.MaxSlashingPercentage)
	result.Properties.MaxAnnualInflation = ConvertRate(result.Properties.MaxAnnualInflation)
	result.Properties.DappVerifierBond = ConvertRate(result.Properties.DappVerifierBond)

	return result, nil
}
