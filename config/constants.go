package config

const (
	InterxVersion = "v0.4.43"
	SekaiVersion  = "v0.3.40"
	CosmosVersion = "v0.47.6"

	QueryDashboard = "/api/dashboard"

	QueryAccounts        = "/api/kira/accounts/{address}"
	QueryTotalSupply     = "/api/kira/supply"
	QueryBalances        = "/api/kira/balances/{address}"
	PostTransaction      = "/api/kira/txs"
	QueryTransactionHash = "/api/kira/txs/{hash}"
	EncodeTransaction    = "/api/kira/txs/encode"

	QueryRoles                = "/api/kira/gov/all_roles"
	QueryRolesByAddress       = "/api/kira/gov/roles_by_address/{val_addr}"
	QueryPermissionsByAddress = "/api/kira/gov/permissions_by_address/{val_addr}"
	QueryProposals            = "/api/kira/gov/proposals"
	QueryProposal             = "/api/kira/gov/proposals/{proposal_id}"
	QueryVoters               = "/api/kira/gov/voters/{proposal_id}"
	QueryVotes                = "/api/kira/gov/votes/{proposal_id}"
	QueryDataReferenceKeys    = "/api/kira/gov/data_keys"
	QueryDataReference        = "/api/kira/gov/data/{key}"
	QueryNetworkProperties    = "/api/kira/gov/network_properties"
	QueryExecutionFee         = "/api/kira/gov/execution_fee"
	QueryExecutionFees        = "/api/kira/gov/execution_fees"
	QueryKiraTokensAliases    = "/api/kira/tokens/aliases"
	QueryKiraTokensRates      = "/api/kira/tokens/rates"
	QueryKiraFunctions        = "/api/kira/metadata"
	QueryKiraStatus           = "/api/kira/status"

	QueryCurrentPlan = "/api/kira/upgrade/current_plan"
	QueryNextPlan    = "/api/kira/upgrade/next_plan"

	QueryIdentityRecord                          = "/api/kira/gov/identity_record/{id}"
	QueryIdentityRecordsByAddress                = "/api/kira/gov/identity_records/{creator}"
	QueryAllIdentityRecords                      = "/api/kira/gov/all_identity_records"
	QueryIdentityRecordVerifyRequest             = "/api/kira/gov/identity_verify_record/{request_id}"
	QueryIdentityRecordVerifyRequestsByRequester = "/api/kira/gov/identity_verify_requests_by_requester/{requester}"
	QueryIdentityRecordVerifyRequestsByApprover  = "/api/kira/gov/identity_verify_requests_by_approver/{approver}"
	QueryAllIdentityRecordVerifyRequests         = "/api/kira/gov/all_identity_verify_requests"

	QueryStakingPool           = "/api/kira/staking-pool"
	QueryDelegations           = "/api/kira/delegations"
	QueryUndelegations         = "/api/kira/undelegations"
	QuerySpendingPools         = "/api/kira/spending-pools"
	QuerySpendingPoolProposals = "/api/kira/spending-pool-proposals"

	QueryUBIRecords = "/api/kira/ubi-records"

	QueryInterxFunctions = "/api/metadata"

	FaucetRequestURL         = "/api/kira/faucet"
	QueryRPCMethods          = "/api/rpc_methods"
	QueryUnconfirmedTxs      = "/api/unconfirmed_txs"
	QueryBlocks              = "/api/blocks"
	QueryBlockByHeightOrHash = "/api/blocks/{height}"
	QueryBlockTransactions   = "/api/blocks/{height}/transactions"
	QueryTransactionResult   = "/api/transactions/{txHash}"
	QueryTransactions        = "/api/transactions"
	QueryStatus              = "/api/status"
	QueryConsensus           = "/api/consensus"
	QueryDumpConsensusState  = "/api/dump_consensus_state"
	QueryValidators          = "/api/valopers"
	QueryValidatorInfos      = "/api/valoperinfos"
	QueryGenesis             = "/api/genesis"
	QueryGenesisSum          = "/api/gensum"
	QuerySnapShot            = "/api/snapshot"
	QuerySnapShotInfo        = "/api/snapshot_info"
	QueryPubP2PList          = "/api/pub_p2p_list"
	QueryPrivP2PList         = "/api/priv_p2p_list"
	QueryInterxList          = "/api/interx_list"
	QuerySnapList            = "/api/snap_list"
	QueryAddrBook            = "/api/addrbook"
	QueryNetInfo             = "/api/net_info"

	Download              = "/download"
	DataReferenceRegistry = "DRR"
	DefaultInterxPort     = "11000"

	QueryRosettaNetworkList    = "/rosetta/network/list"
	QueryRosettaNetworkOptions = "/rosetta/network/options"
	QueryRosettaNetworkStatus  = "/rosetta/network/status"
	QueryRosettaAccountBalance = "/rosetta/account/balance"

	QueryEVMStatus      = "/api/{chain}/status"
	QueryEVMBlock       = "/api/{chain}/blocks/{identifier}"
	QueryEVMTransaction = "/api/{chain}/transactions/{hash}"
	QueryEVMTransfer    = "/api/{chain}/txs"
	QueryEVMAccounts    = "/api/{chain}/accounts/{address}"
	QueryEVMBalances    = "/api/{chain}/balances/{address}"
	QueryABI            = "/api/{chain}/abi/{contract}"
	QueryReadContract   = "/api/{chain}/read/{contract}"
	QueryWriteContract  = "/api/{chain}/write/{contract}"
	QueryEVMFaucet      = "/api/{chain}/faucet"

	QueryBitcoinStatus      = "/api/bitcoin/status"
	QueryBitcoinBlock       = "/api/bitcoin/blocks/{identifier}"
	QueryBitcoinTransaction = "/api/bitcoin/transactions/{hash}"
	QueryBitcoinTransfer    = "/api/bitcoin/txs"
	QueryBitcoinAccounts    = "/api/bitcoin/accounts/{address}"
	QueryBitcoinBalances    = "/api/bitcoin/balances/{address}"
	QueryBitcoinFaucet      = "/api/bitcoin/faucet"

	QueryKiraEndpoints = "/api/kira/"
)

// map msg type param from api to backend msg type
var MsgTypes = map[string]string{
	// Evidence
	"submit_evidence": "MsgSubmitEvidence",
	// Proposals
	"submit-proposal": "MsgSubmitProposal",
	"vote-proposal":   "MsgVoteProposal",
	// Permissions
	"whitelist-permissions":            "MsgWhitelistPermissions",
	"blacklist-permissions":            "MsgBlacklistPermissions",
	"whitelist-role-permission":        "MsgWhitelistRolePermission",
	"blacklist-role-permission":        "MsgBlacklistRolePermission",
	"remove-whitelist-role-permission": "MsgRemoveWhitelistRolePermission",
	"remove-blacklist-role-permission": "MsgRemoveBlacklistRolePermission",
	// Governance
	"claim-councilor":        "MsgClaimCouncilor",
	"set-network-properties": "MsgSetNetworkProperties",
	"set-execution-fee":      "MsgSetExecutionFee",
	// Roles
	"create-role": "MsgCreateRole",
	"assign-role": "MsgAssignRole",
	"remove-role": "MsgRemoveRole",
	// Identity records
	"register-identity-records":              "MsgRegisterIdentityRecords",
	"edit-identity-record":                   "MsgEditIdentityRecord",
	"request-identity-records-verify":        "MsgRequestIdentityRecordsVerify",
	"handle-identity-records-verify-request": "MsgHandleIdentityRecordsVerifyRequest",
	"cancel-identity-records-verify-request": "MsgCancelIdentityRecordsVerifyRequest",
	// Spending module
	"create-spending-pool":               "MsgCreateSpendingPool",
	"deposit-spending-pool":              "MsgDepositSpendingPool",
	"register-spending-pool-beneficiary": "MsgRegisterSpendingPoolBeneficiary",
	"claim-spending-pool":                "MsgClaimSpendingPool",
	// Staking module
	"claim-validator":     "MsgClaimValidator",
	"upsert_staking_pool": "MsgUpsertStakingPool",
	"delegate":            "MsgDelegate",
	"undelegate":          "MsgUndelegate",
	"claim_rewards":       "MsgClaimRewards",
	"claim_undelegation":  "MsgClaimUndelegation",
	"set_compound_info":   "MsgSetCompoundInfo",
	"register_delegator":  "MsgRegisterDelegator",
	// Tokens module
	"upsert-token-alias": "MsgUpsertTokenAlias",
	"upsert-token-rate":  "MsgUpsertTokenRate",
	// Cosmos SDK
	"send":      "/cosmos.bank.v1beta1.MsgSend",
	"multisend": "/cosmos.bank.v1beta1.MultiSend",
	// Slashing module
	"activate": "MsgActivate",
	"pause":    "MsgPause",
	"unpause":  "MsgUnpause",
	// Custody module
	"create-custody":                 "MsgCreteCustodyRecord",
	"add-to-custody-whitelist":       "MsgAddToCustodyWhiteList",
	"add-to-custody-custodians":      "MsgAddToCustodyCustodians",
	"remove-from-custody-custodians": "MsgRemoveFromCustodyCustodians",
	"drop-custody-custodians":        "MsgDropCustodyCustodians",
	"remove-from-custody-whitelist":  "MsgRemoveFromCustodyWhiteList",
	"drop-custody-whitelist":         "MsgDropCustodyWhiteList",
	"approve-custody-transaction":    "MsgApproveCustodyTransaction",
	"decline-custody-transaction":    "MsgDeclineCustodyTransaction",
	"password-confirm-transaction":   "MsgPasswordConfirmTransaction",
	"custody-send":                   "MsgSend",
	// Basket module
	"disable-basket-deposits":  "MsgDisableBasketDeposits",
	"disable-basket-withdraws": "MsgDisableBasketWithdraws",
	"disable-basket-swaps":     "MsgDisableBasketSwaps",
	"basket-token-mint":        "MsgBasketTokenMint",
	"basket-token-burn":        "MsgBasketTokenBurn",
	"basket-token-swap":        "MsgBasketTokenSwap",
	"basket-claim-rewards":     "MsgBasketClaimRewards",
}
var SupportedEVMChains = [1]string{"goerli"}
var SupportedBitcoinChains = [1]string{"testnet"}
