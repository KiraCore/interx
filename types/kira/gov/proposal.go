package gov

// VoteOption enumerates the valid vote options for a given governance proposal.
type VoteOption int32

const (
	// VOTE_OPTION_UNSPECIFIED defines a no-op vote option.
	VoteOption_VOTE_OPTION_UNSPECIFIED VoteOption = 0
	// VOTE_OPTION_YES defines a yes vote option.
	VoteOption_VOTE_OPTION_YES VoteOption = 1
	// VOTE_OPTION_ABSTAIN defines an abstain vote option.
	VoteOption_VOTE_OPTION_ABSTAIN VoteOption = 2
	// VOTE_OPTION_NO defines a no vote option.
	VoteOption_VOTE_OPTION_NO VoteOption = 3
	// VOTE_OPTION_NO_WITH_VETO defines a no with veto vote option.
	VoteOption_VOTE_OPTION_NO_WITH_VETO VoteOption = 4
)

// Enum value maps for VoteOption.
var (
	VoteOption_name = map[int32]string{
		0: "VOTE_OPTION_UNSPECIFIED",
		1: "VOTE_OPTION_YES",
		2: "VOTE_OPTION_ABSTAIN",
		3: "VOTE_OPTION_NO",
		4: "VOTE_OPTION_NO_WITH_VETO",
	}
	VoteOption_value = map[string]int32{
		"VOTE_OPTION_UNSPECIFIED":  0,
		"VOTE_OPTION_YES":          1,
		"VOTE_OPTION_ABSTAIN":      2,
		"VOTE_OPTION_NO":           3,
		"VOTE_OPTION_NO_WITH_VETO": 4,
	}
)

// Enum value maps for VoteResult.
var (
	VoteResult = map[string]string{
		"Unknown":            "VOTE_RESULT_UNKNOWN",
		"Passed":             "VOTE_RESULT_PASSED",
		"Rejected":           "VOTE_RESULT_REJECTED",
		"RejectedWithVeto":   "VOTE_RESULT_REJECTED_WITH_VETO",
		"Pending":            "VOTE_PENDING",
		"QuorumNotReached":   "VOTE_RESULT_QUORUM_NOT_REACHED",
		"Enactment":          "VOTE_RESULT_ENACTMENT",
		"PassedWithExecFail": "VOTE_RESULT_PASSED_WITH_EXEC_FAIL",
	}
)

// Used to parse response from sekai gRPC ("/kira/gov/votes/{proposal_id}")
type Vote struct {
	ProposalID uint64 `json:"proposal_id"`
	Voter      string `json:"voter"`
	Option     string `json:"option"`
}

// Used to sync proposals with sekaid
type Proposal struct {
	ProposalID                 string      `json:"proposalId"`
	Title                      string      `json:"title"`
	Description                string      `json:"description"`
	Content                    interface{} `json:"content"`
	SubmitTime                 string      `json:"submitTime"`
	VotingEndTime              string      `json:"votingEndTime"`
	EnactmentEndTime           string      `json:"enactmentEndTime"`
	MinVotingEndBlockHeight    string      `json:"minVotingEndBlockHeight"`
	MinEnactmentEndBlockHeight string      `json:"minEnactmentEndBlockHeight"`
	ExecResult                 string      `json:"execResult"`
	Result                     string      `json:"result"`
	// New fields
	VotersCount int    `json:"voters_count"`
	VotesCount  int    `json:"votes_count"`
	Quorum      string `json:"quorum"`
	Metadata    string `json:"meta_data"`

	// Extra fields for filtering
	Hash        string `json:"transaction_hash,omitempty"`
	Timestamp   int    `json:"timestamp,omitempty"`
	BlockHeight int    `json:"block_height,omitempty"`
	Type        string `json:"type,omitempty"`
	Proposer    string `json:"proposer,omitempty"`
}

type CachedProposal struct {
	ProposalID string `json:"proposal_id"`
	// New fields
	VotersCount int    `json:"voters_count"`
	VotesCount  int    `json:"votes_count"`
	Quorum      string `json:"quorum"`
	Metadata    string `json:"meta_data"`

	// Extra fields for filtering
	Hash        string `json:"transaction_hash"`
	Timestamp   int    `json:"timestamp"`
	BlockHeight int    `json:"block_height"`
	Type        string `json:"type"`
	Proposer    string `json:"proposer"`
}

// Used to return proposal response
type PropsResponse struct {
	TotalCount int        `json:"total_count"`
	Proposals  []Proposal `json:"proposals"`
}

// Used to parse proposal response from sekai
type ProposalsResponse = struct {
	Proposals  []Proposal  `json:"proposals,omitempty"`
	Pagination interface{} `json:"pagination,omitempty"`
}

// Used to sync proposals with sekaid
type ProposalUserCount struct {
	Proposers string `json:"proposers"`
	Voters    string `json:"voters"`
}

// Used to sync proposals with sekaid
type AllProposals struct {
	Status struct {
		TotalProposals      int `json:"total_proposals"`
		ActiveProposals     int `json:"active_proposals"`
		EnactingProposals   int `json:"enacting_proposals"`
		FinishedProposals   int `json:"finished_proposals"`
		SuccessfulProposals int `json:"successful_proposals"`
	} `json:"status"`
	Proposals []Proposal        `json:"proposals"`
	Users     ProposalUserCount `json:"users"`
}
