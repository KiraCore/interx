package config

const (
	InterxVersion = "v0.4.16"
	SekaiVersion  = "v0.3.0.21"
	CosmosVersion = "v0.45.1"

	QueryDashboard = "/api/dashboard"

	QueryAccounts        = "/api/cosmos/auth/accounts/{address}"
	QueryTotalSupply     = "/api/cosmos/bank/supply"
	QueryBalances        = "/api/cosmos/bank/balances/{address}"
	PostTransaction      = "/api/cosmos/txs"
	QueryTransactionHash = "/api/cosmos/txs/{hash}"
	EncodeTransaction    = "/api/cosmos/txs/encode"

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

	QuerySpendingPools         = "/api/kira/spending-pools"
	QuerySpendingPoolProposals = "/api/kira/spending-pool-proposals"

	QueryUBIRecords = "/api/kira/ubi-records"

	QueryInterxFunctions = "/api/metadata"

	FaucetRequestURL         = "/api/faucet"
	QueryRPCMethods          = "/api/rpc_methods"
	QueryWithdraws           = "/api/withdraws"
	QueryDeposits            = "/api/deposits"
	QueryUnconfirmedTxs      = "/api/unconfirmed_txs"
	QueryBlocks              = "/api/blocks"
	QueryBlockByHeightOrHash = "/api/blocks/{height}"
	QueryBlockTransactions   = "/api/blocks/{height}/transactions"
	QueryTransactionResult   = "/api/transactions/{txHash}"
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
)

var SupportedEVMChains = [1]string{"ropsten"}
