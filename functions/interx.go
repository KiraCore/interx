package functions

import (
	"encoding/json"

	"github.com/KiraCore/interx/config"
	"github.com/iancoleman/strcase"
)

type InterxFunctionParameter struct {
	Type        string                    `json:"type,omitempty"`
	Optional    bool                      `json:"optional,omitempty"`
	Description string                    `json:"description"`
	Fields      *InterxFunctionParameters `json:"fields,omitempty"`
}

type InterxFunctionParameters = map[string]InterxFunctionParameter

type InterxFunctionMeta struct {
	Endpoint    string                   `json:"endpoint"`
	Description string                   `json:"description"`
	Parameters  InterxFunctionParameters `json:"parameters"`
	Response    InterxFunctionParameters `json:"response"`
}

type InterxMetadata struct {
	Functions      map[string]InterxFunctionMeta `json:"functions"`
	ResponseHeader InterxFunctionParameters      `json:"response_header"`
}

var (
	interxMetadata InterxMetadata = InterxMetadata{
		Functions:      make(map[string]InterxFunctionMeta),
		ResponseHeader: InterxFunctionParameters{},
	}
)

// AddInterxFunction is a function to add a function
func AddInterxFunction(functionType string, endpoint string, meta string) {
	metadata := InterxFunctionMeta{}
	if err := json.Unmarshal([]byte(meta), &metadata); err != nil {
		panic(err)
	}
	metadata.Endpoint = endpoint

	interxMetadata.Functions[strcase.ToCamel(functionType)] = metadata
}

// RegisterInterxFunctions is a function to register all interx functions
func RegisterInterxFunctions() {
	interxMetadata.ResponseHeader["Interx_chain_id"] = InterxFunctionParameter{
		Type:        "number",
		Description: "This represents the current chain id.",
	}
	interxMetadata.ResponseHeader["Interx_block"] = InterxFunctionParameter{
		Type:        "number",
		Description: "This represents the current block number.",
	}
	interxMetadata.ResponseHeader["Interx_blocktime"] = InterxFunctionParameter{
		Type:        "number",
		Description: "This represents the current block timestamp.",
	}
	interxMetadata.ResponseHeader["Interx_timestamp"] = InterxFunctionParameter{
		Type:        "string",
		Description: "This represents the current interx timestamp.",
	}
	interxMetadata.ResponseHeader["Interx_request_hash"] = InterxFunctionParameter{
		Type:        "string",
		Description: "This represents the hash of request parameters.",
	}
	interxMetadata.ResponseHeader["Interx_signature"] = InterxFunctionParameter{
		Type:        "string",
		Description: "This represents the interx response signature.",
	}
	interxMetadata.ResponseHeader["Interx_hash"] = InterxFunctionParameter{
		Type:        "string",
		Description: "This represents the interx response hash.",
	}
	interxMetadata.ResponseHeader["Interx_ref"] = InterxFunctionParameter{
		Type:        "string",
		Description: "This represents link to download the data reference.",
	}

	AddInterxFunction(
		"QueryKiraStatus",
		config.QueryKiraStatus,
		`{
			"description": "QueryKiraStatus is a function to query the node status",
			"response": {
				"node_info": {
					"description": "The connected node information"
				},
				"sync_info": {
					"description": "The sync status of connected node"
				},
				"validator_info": {
					"description": "The validator information of connect node"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryAccount",
		config.QueryAccounts,
		`{
			"description": "QueryAccount is a function to query the account info.",
			"parameters": {
				"address": {
					"type":        "string",
					"description": "This represents the account address."
				}
			},
			"response": {
				"account": {
					"description": "The account info with address, pubkey and sequence."
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryTotalSupply",
		config.QueryTotalSupply,
		`{
			"description": "QueryTotalSupply is a function to query total supply.",
			"response": {
				"supply": {
					"type": "Coin[]",
					"description": "The total supply of the network"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryBalance",
		config.QueryBalances,
		`{
			"description": "QueryBalance is a function to query the account balances.",
			"parameters": {
				"address": {
					"type":        "string",
					"description": "This represents the account address."
				},
				"limit": {
					"type":        "number",
					"description": "This represents the page size"
				},
				"offset": {
					"type":        "number",
					"description": "This represents the page number"
				},
				"count_total": {
					"type":        "number",
					"description": "This represents the option to return total count of data reference keys.",
					"optional": true
				}
			},
			"response": {
				"balances": {
					"type": "Coin[]",
					"description": "The account balances with pagination"
				},
				"pagination": {
					"description": "The pagination response information like total and next_key"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryTransactionHash",
		config.QueryTransactionHash,
		`{
			"description": "QueryTransactionHash is a function to query transaction details from transaction hash.",
			"parameters": {
				"hash": {
					"type":        "string",
					"description": "This represents the transaction hash. (e.g. 0x20.....)"
				}
			},
			"response": {
				"hash": {
					"description": "The transaction hash"
				},
				"height": {
					"description": "The block height of transation"
				},
				"tx": {
					"description": "The base-64 encoded transaction"
				},
				"tx_result": {
					"description": "The result of transaction with events, gas info, logs and error code"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryDataReferenceKeys",
		config.QueryDataReferenceKeys,
		`{
			"description": "QueryDataReferenceKeys is a function to query data reference keys with pagination.",
			"parameters": {
				"limit": {
					"type":        "number",
					"description": "This represents the page size"
				},
				"offset": {
					"type":        "number",
					"description": "This represents the page number"
				},
				"count_total": {
					"type":        "number",
					"description": "This represents the option to return total count of data reference keys.",
					"optional": true
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryDataReference",
		config.QueryDataReference,
		`{
			"description": "QueryDataReference is a function to query data reference by a key.",
			"parameters": {
				"key": {
					"type":        "string",
					"description": "This represents data reference key."
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryRoles",
		config.QueryRoles,
		`{
			"description": "QueryRoles is a function to query all roles."
		}`,
	)

	AddInterxFunction(
		"QueryRolesByAddress",
		config.QueryRolesByAddress,
		`{
			"description": "QueryRolesByAddress is a function to query all roles by an address.",
			"parameters": {
				"val_addr": {
					"type":        "string",
					"description": "This represents the kira account address.",
					"optional": true
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryPermissionsByAddress",
		config.QueryPermissionsByAddress,
		`{
			"description": "QueryPermissionsByAddress is a function to query all permissions by an address.",
			"parameters": {
				"val_addr": {
					"type":        "string",
					"description": "This represents the kira account address.",
					"optional": true
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryProposals",
		config.QueryProposals,
		`{
			"description": "QueryProposals is a function to query all proposals.",
			"parameters": {
				"voter": {
					"type":        "string",
					"description": "This represents the voter address.",
					"optional": true
				},
				"proposer": {
					"type":        "string",
					"description": "This represents the proposer address.",
					"optional": true
				},
				"offset": {
					"type":        "string",
					"description": "This is an option to validators pagination. offset is a numeric offset that can be used when key is unavailable. It is less efficient than using key. Only one of offset or key should be set.",
					"optional": true
				},
				"limit": {
					"type":        "string",
					"description": "This is an option to validators pagination. limit is the total number of results to be returned in the result page. If left empty it will default to a value to be set by each app.",
					"optional": true
				},
				"status": {
					"type":        "string",
					"description": "This represents the proposal statuses(Unknown, Passed, Rejected, RejectedWithVeto, Pending, QuorumNotReached, Enactment, PassedWithExecFailed) separated by comma.",
					"optional": true
				},
				"sort": {
					"type":        "string",
					"description": "This represents how the proposal should be sorted(dateASC, dateDESC).",
					"optional": true
				},
				"types": {
					"type":        "string",
					"description": "This represents the proposal types separated by comma.",
					"optional": true
				},
				"dateStart": {
					"type":        "string",
					"description": "This represents the starting point in timestamp or date(DD/MM/YY) format.",
					"optional": true
				},
				"dateEnd": {
					"type":        "string",
					"description": "This represents the ending point in timestamp or date(DD/MM/YY) format.",
					"optional": true
				}
			},
			"response": {
				"proposals": {
					"description": "The array of proposals information"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryProposal",
		config.QueryProposal,
		`{
			"description": "QueryProposal is a function to query a proposal by a given proposal_id.",
			"parameters": {
				"proposal_id": {
					"type":        "number",
					"description": "This is an option of a proposal id"
				}
			},
			"response": {
				"proposal": {
					"description": "The proposal information"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryVoters",
		config.QueryVoters,
		`{
			"description": "QueryVoters is a function to query voters by a given proposal id.",
			"parameters": {
				"proposal_id": {
					"type":        "number",
					"description": "This is an option of a proposal id"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryVotes",
		config.QueryVotes,
		`{
			"description": "QueryVotes is a function to query votes by a given proposal id.",
			"parameters": {
				"proposal_id": {
					"type":        "number",
					"description": "This is an option of a proposal id"
				}
			},
			"response": {
				"votes": {
					"description": "The array of votes information"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryNetworkProperties",
		config.QueryNetworkProperties,
		`{
			"description": "QueryNetworkProperties is a function to query network properties."
		}`,
	)

	AddInterxFunction(
		"QueryExecutionFee",
		config.QueryExecutionFee,
		`{
			"description": "QueryExecutionFee is a function to query execution fee by transaction type.",
			"parameters": {
				"message": {
					"type":        "string",
					"description": "This is an option of a transaction type"
				}
			},
			"response": {
				"fee": {
					"description": "The execution fee info"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryExecutionFees",
		config.QueryExecutionFees,
		`{
			"description": "QueryExecutionFees is a function to query all execution fees.",
			"response": {
				"fees": {
					"type":        "array",
					"description": "All execution fees"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryKiraTokensAliases",
		config.QueryKiraTokensAliases,
		`{
			"description": "QueryKiraTokensAliases is a function to query all registered tokens."
		}`,
	)

	AddInterxFunction(
		"QueryKiraTokensRates",
		config.QueryKiraTokensRates,
		`{
			"description": "QueryKiraTokensRates is a function to query all registered token rates."
		}`,
	)

	AddInterxFunction(
		"QueryKiraTokensAliases",
		config.QueryKiraTokensAliases,
		`{
			"description": "QueryKiraTokensAliases is a function to query all tokens aliases."
		}`,
	)

	AddInterxFunction(
		"QueryKiraTokensRates",
		config.QueryKiraTokensRates,
		`{
			"description": "QueryKiraTokensRates is a function to query all tokens rates."
		}`,
	)

	AddInterxFunction(
		"QueryTransactions",
		config.QueryTransactions,
		`{
			"description": "QueryTransactions is a function to query transactions of the account filtered by various options like msg type, date, and so on.",
			"parameters": {
				"account": {
					"type":        "string",
					"description": "This represents the kira account address."
				},
				"type": {
					"type":        "string",
					"description": "This represents the transaction types separated by comma.",
					"optional": true
				},
				"page": {
					"type":        "int",
					"description": "This represents the page number of results.",
					"optional": true
				},
				"page_size": {
					"type":        "int",
					"description": "This represents the pageSize number of results. (1 ~ 100)",
					"optional": true
				},
				"offset": {
					"type":        "int",
					"description": "This represents the offset of the first transaction.",
					"optional": true
				},
				"limit": {
					"type":        "int",
					"description": "This represents the limit of total results to be shown. (1 ~ 100)",
					"optional": true
				},
				"dateStart": {
					"type":        "string",
					"description": "This represents the starting point in timestamp or date(DD/MM/YY) format.",
					"optional": true
				},
				"dateEnd": {
					"type":        "string",
					"description": "This represents the ending point in timestamp or date(DD/MM/YY) format.",
					"optional": true
				},
				"direction": {
					"type":        "string",
					"description": "This represents directions of the transaction(outbound, inbound) separated by comma.",
					"optional": true
				},
				"status": {
					"type":        "string",
					"description": "This represents the transaction statuses(pending, confirmed, failed) separated by comma.",
					"optional": true
				},
				"sort": {
					"type":        "string",
					"description": "This represents how the transactions should be sorted(dateASC, dateDESC).",
					"optional": true
				}
			},
			"response": {
				"transactions": {
					"description": "The array of transactions"
				},
				"total_count": {
					"description": "The total transaction count"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryUnconfirmedTxs",
		config.QueryUnconfirmedTxs,
		`{
			"description": "QueryUnconfirmedTxs is a function to query unconfirmed transactions.",
			"parameters": {
				"limit": {
					"type":        "int",
					"description": "This represents the limit of the transaction. (1 ~ 1000)",
					"optional": true
				}
			}
		}`,
	)

	AddInterxFunction(
		"Broadcast",
		config.PostTransaction,
		`{
			"description": "Broadcast is a function to broadcast signed transaction.",
			"parameters": {
				"tx": {
					"type":        "byte[]",
					"description": "This represents the transaction bytes."
				},
				"mode": {
					"type":        "string",
					"description": "This represents the broadcast mode. (block, sync, async)",
					"optional": true
				}
			}
		}`,
	)

	AddInterxFunction(
		"FaucetInfo",
		config.FaucetRequestURL,
		`{
			"description": "FaucetInfo is a function to return the available faucet amount",
			"response": {
				"address": {
					"description": "The faucet kira address"
				},
				"balances": {
					"description": "The balances array (amount & denom)"
				}
			}
		}`,
	)

	AddInterxFunction(
		"Faucet",
		config.FaucetRequestURL+"?",
		`{
			"description": "Faucet is a function to claim tokens to the account for free.",
			"parameters": {
				"claim": {
					"type":        "string",
					"description": "This represents the kira account address.",
					"optional": true
				},
				"token": {
					"type":        "string",
					"description": "This represents the token name.",
					"optional": true
				}
			},
			"response": {
				"hash": {
					"description": "The faucet transaction hash"
				}
			}
		}`,
	)

	AddInterxFunction(
		"Download",
		config.Download,
		`{
			"description": "Download is a function to download a data reference or arbitrary data.",
			"parameters": {
				"module": {
					"type":        "string",
					"description": "This represents the module name. (e.g. DRR for data reference registry.)"
				},
				"key": {
					"type":        "string",
					"description": "This represents the reference key. (It saves reference data with hashed name. e.g. 2CEE6B1689EDDDD6F08EB1EAEC7D3C4E.)"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryValidators",
		config.QueryValidators,
		`{
			"description": "QueryValidators is a function to query validators.",
			"parameters": {
				"all": {
					"type":        "string",
					"description": "This is an option to query all validators.",
					"optional": true
				},
				"status_only": {
					"type":        "string",
					"description": "This is an option to query only status of all validators.",
					"optional": true
				},
				"address": {
					"type":        "string",
					"description": "This is an option to query validator by a given kira address",
					"optional": true
				},
				"valkey": {
					"type":        "string",
					"description": "This is an option to query validator by a given valoper address",
					"optional": true
				},
				"pubkey": {
					"type":        "string",
					"description": "This is an option to query validator by a given pubkey",
					"optional": true
				},
				"moniker": {
					"type":        "string",
					"description": "This is an option to query validator by a given moniker",
					"optional": true
				},
				"status": {
					"type":        "string",
					"description": "This is an option to query validators by a given status",
					"optional": true
				},
				"proposer": {
					"type":        "string",
					"description": "This is an option to query validators by a given proposer address",
					"optional": true
				},
				"key": {
					"type":        "string",
					"description": "This is an option to validators pagination. key is a value returned in PageResponse.next_key to begin querying the next page most efficiently. Only one of offset or key should be set.",
					"optional": true
				},
				"offset": {
					"type":        "string",
					"description": "This is an option to validators pagination. offset is a numeric offset that can be used when key is unavailable. It is less efficient than using key. Only one of offset or key should be set.",
					"optional": true
				},
				"limit": {
					"type":        "string",
					"description": "This is an option to validators pagination. limit is the total number of results to be returned in the result page. If left empty it will default to a value to be set by each app.",
					"optional": true
				},
				"countTotal": {
					"type":        "string",
					"description": "This is an option to validators pagination. count_total is set to true  to indicate that the result set should include a count of the total number of items available for pagination in UIs. count_total is only respected when offset is used. It is ignored when key is set.",
					"optional": true
				}
			},
			"response": {
				"validators": {
					"description": "The array of validators information"
				},
				"pagination": {
					"description": "The pagination response information like total and next_key"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryValidatorInfos",
		config.QueryValidatorInfos,
		`{
			"description": "QueryValidatorInfos is a function to query validator infos.",
			"parameters": {
				"key": {
					"type":        "string",
					"description": "This is an option to validators pagination. key is a value returned in PageResponse.next_key to begin querying the next page most efficiently. Only one of offset or key should be set.",
					"optional": true
				},
				"offset": {
					"type":        "string",
					"description": "This is an option to validators pagination. offset is a numeric offset that can be used when key is unavailable. It is less efficient than using key. Only one of offset or key should be set.",
					"optional": true
				},
				"limit": {
					"type":        "string",
					"description": "This is an option to validators pagination. limit is the total number of results to be returned in the result page. If left empty it will default to a value to be set by each app.",
					"optional": true
				},
				"countTotal": {
					"type":        "string",
					"description": "This is an option to validators pagination. count_total is set to true  to indicate that the result set should include a count of the total number of items available for pagination in UIs. count_total is only respected when offset is used. It is ignored when key is set.",
					"optional": true
				}
			},
			"response": {
				"info": {
					"description": "The array of validators information"
				},
				"pagination": {
					"description": "The pagination response information like total and next_key"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryConsensus",
		config.QueryConsensus,
		`{
			"description": "QueryConsensus is a function to query consensus info.",
			"response": {
				"average_block_time": {
					"description": "average block time in seconds"
				},
				"commit_time": {
					"description": "The latest commit timestamp"
				},
				"consensus_stopped": {
					"description": "If the consensus is stopped or not"
				},
				"height": {
					"description": "The latest block height"
				},
				"noncommits": {
					"description": "The validators array with no commits"
				},
				"precommits": {
					"description": "The validators array with pre commits"
				},
				"prevotes": {
					"description": "The validators array with prevotes"
				},
				"proposer": {
					"description": "The current proposer kira address"
				},
				"round": {
					"description": "The consensus round"
				},
				"start_time": {
					"description": "The consensus start timestamp"
				},
				"step": {
					"description": "RoundStepNewHeight"
				},
				"triggered_timeout_precommit": {
					"description": "true or false"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryBlocks",
		config.QueryBlocks,
		`{
			"description": "QueryBlocks is a function to query blocks with pagination.",
			"parameters": {
				"minHeight": {
					"type":        "string",
					"description": "This is the option of the minimum block height.",
					"optional": true
				},
				"maxHeight": {
					"type":        "string",
					"description": "This is the option of the maximum block height.",
					"optional": true
				}
			},
			"response": {
				"block_metas": {
					"description": "The array of block informations"
				},
				"last_height": {
					"description": "The last block height"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryBlockByHeightOrHash",
		config.QueryBlockByHeightOrHash,
		`{
			"description": "QueryBlockByHeightOrHash is a function to query block by height or hash.",
			"parameters": {
				"height": {
					"type":        "string",
					"description": "This is an option of block height or hash.",
					"optional": true
				}
			},
			"response": {
				"block": {
					"description": "The block information"
				},
				"block_id": {
					"description": "The block hash inforamtion"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryBlockTransactions",
		config.QueryBlockTransactions,
		`{
			"description": "QueryBlockTransactions is a function to query block transactions by height.",
			"parameters": {
				"height": {
					"type":        "string",
					"description": "This is an option of block height.",
					"optional": true
				}
			},
			"response": {
				"txs": {
					"description": "The transaction information array"
				},
				"total_count": {
					"description": "The total transaction count"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryTransactionResult",
		config.QueryTransactionResult,
		`{
			"description": "QueryTransactionResult is a function to query transaction result by hash.",
			"parameters": {
				"txHash": {
					"type":        "string",
					"description": "This is an option of a transaction hash.",
					"optional": true
				}
			},
			"response": {
				"hash": {
					"description": "The transaction hash"
				},
				"status": {
					"description": "The transaction status"
				},
				"block_height": {
					"description": "The block height"
				},
				"block_timestamp": {
					"description": "The block timestamp"
				},
				"confirmation": {
					"description": "The block confirmations of this transaction"
				},
				"msgs": {
					"description": "The transaction msgs"
				},
				"fees": {
					"description": "The transaction fee"
				},
				"gas_wanted": {
					"description": "The gas limit amount for transaction"
				},
				"gas_used": {
					"description": "The gas amount used in transaction"
				},
				"memo": {
					"description": "The transaction memo"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryPrivP2PList",
		config.QueryPrivP2PList,
		`{
			"description": "QueryPrivP2PList is a function to query all private nodes list.",
			"parameters": {
				"ip_only": {
					"type":        "bool",
					"description": "This is an option to query only ip addresses separated by comma.",
					"optional": true
				},
				"order": {
					"type":        "string",
					"description": "This is an option to query nodes in a specific order. usecase: order=random",
					"optional": true
				},
				"format": {
					"type":        "string",
					"description": "This is an option to query nodes in a specific format. usecase: order=simple",
					"optional": true
				},
				"peers_only": {
					"type":        "bool",
					"description": "This is an option to query only peers separated by comma. <node_id>@<ip>:<port>",
					"optional": true
				},
				"connected": {
					"type":        "bool",
					"description": "This is an option to query only connected ips.",
					"optional": true
				}
			},
			"response": {
				"last_update": {
					"type": "number",
					"description": "The last updated timestamp"
				},
				"scanning": {
					"type": "bool",
					"description": "If discovery is still running or not"
				},
				"node_list": {
					"description": "The node list array",
					"fields": {
						"id": {
							"type":        "string",
							"description": "The private node id"
						},
						"ip": {
							"type":        "string",
							"description": "The local ip address"
						},
						"port": {
							"type":        "number",
							"description": "The p2p port number"
						},
						"ping": {
							"type":        "number",
							"description": "The time duration in miliseconds to dial p2p and verify p2p node id"
						},
						"connected": {
							"type":        "bool",
							"description": "If the node is connected with this node"
						},
						"peers": {
							"type":        "array<string>",
							"description": "The list of node ids"
						}
					}
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryPubP2PList",
		config.QueryPubP2PList,
		`{
			"description": "QueryPubP2PList is a function to query all public nodes list.",
			"parameters": {
				"ip_only": {
					"type":        "bool",
					"description": "This is an option to query only ip addresses separated by comma.",
					"optional": true
				},
				"order": {
					"type":        "string",
					"description": "This is an option to query nodes in a specific order. usecase: order=random",
					"optional": true
				},
				"format": {
					"type":        "string",
					"description": "This is an option to query nodes in a specific format. usecase: order=simple",
					"optional": true
				},
				"peers_only": {
					"type":        "bool",
					"description": "This is an option to query only peers separated by comma. <node_id>@<ip>:<port>",
					"optional": true
				},
				"connected": {
					"type":        "bool",
					"description": "This is an option to query only connected ips.",
					"optional": true
				}
			},
			"response": {
				"last_update": {
					"type": "number",
					"description": "The last updated timestamp"
				},
				"scanning": {
					"type": "bool",
					"description": "If discovery is still running or not"
				},
				"node_list": {
					"description": "The node list array",
					"fields": {
						"id": {
							"type":        "string",
							"description": "The public node id"
						},
						"ip": {
							"type":        "string",
							"description": "The public ip address"
						},
						"port": {
							"type":        "number",
							"description": "The p2p port number"
						},
						"ping": {
							"type":        "number",
							"description": "The time duration in miliseconds to dial p2p and verify p2p node id"
						},
						"connected": {
							"type":        "bool",
							"description": "If the node is connected with this node"
						},
						"peers": {
							"type":        "array<string>",
							"description": "The list of node ids"
						}
					}
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryInterxList",
		config.QueryInterxList,
		`{
			"description": "QueryInterxList is a function to query all interx list.",
			"parameters": {
				"ip_only": {
					"type":        "bool",
					"description": "This is an option to query only ip addresses separated by comma.",
					"optional": true
				}
			},
			"response": {
				"last_update": {
					"type": "number",
					"description": "The last updated timestamp"
				},
				"scanning": {
					"type": "bool",
					"description": "If discovery is still running or not"
				},
				"node_list": {
					"description": "The node list array",
					"fields": {
						"id": {
							"type":        "string",
							"description": "The interx public key"
						},
						"ip": {
							"type":        "string",
							"description": "The public ip address"
						},
						"ping": {
							"type":        "number",
							"description": "The time duration in miliseconds to dial INTERX and verify pub_key"
						},
						"moniker": {
							"type":        "string",
							"description": "From interx configuration"
						},
						"faucet": {
							"type":        "string",
							"description": "The faucet kira address"
						},
						"type": {
							"type":        "string",
							"description": "The node type from interx configuration"
						},
						"version": {
							"type":        "string",
							"description": "The interx version from interx configuration"
						}
					}
				}
			}
		}`,
	)

	AddInterxFunction(
		"QuerySnapList",
		config.QuerySnapList,
		`{
			"description": "QuerySnapList is a function to query all snapshot node list.",
			"parameters": {
				"ip_only": {
					"type":        "bool",
					"description": "This is an option to query only ip addresses separated by comma.",
					"optional": true
				}
			},
			"response": {
				"last_update": {
					"type": "number",
					"description": "The last updated timestamp"
				},
				"scanning": {
					"type": "bool",
					"description": "If discovery is still running or not"
				},
				"node_list": {
					"description": "The node list array",
					"fields": {
						"ip": {
							"type":        "string",
							"description": "The public ip address"
						},
						"port": {
							"type":        "number",
							"description": "The interx port number"
						},
						"size": {
							"type":        "number",
							"description": "The snapshot size in bytes"
						},
						"checksum": {
							"type":        "string",
							"description": "The snapshot checksum (SHA256)"
						}
					}
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryIdentityRecord",
		config.QueryIdentityRecord,
		`{
			"description": "QueryIdentityRecord is a function to query identity record by id.",
			"parameters": {
				"id": {
					"type":        "number",
					"description": "This is the identity record id.",
					"optional": false
				}
			},
			"response": {
				"record": {
					"description": "The identity record info",
					"fields": {
						"id": {
							"type":        "number",
							"description": "The identity record id"
						},
						"address": {
							"type":        "string",
							"description": "The address of identity record"
						},
						"key": {
							"type":        "string",
							"description": "The identity record key"
						},
						"value": {
							"type":        "string",
							"description": "The identity record value"
						},
						"date": {
							"type":        "string",
							"description": "The identity record timestamp"
						},
						"verifiers": {
							"type":        "array<string>",
							"description": "The address list of verifiers"
						}
					}
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryIdentityRecordsByAddress",
		config.QueryIdentityRecordsByAddress,
		`{
			"description": "QueryIdentityRecordsByAddress is a function to query identity records by address.",
			"parameters": {
				"creator": {
					"type":        "string",
					"description": "This is the identity record creator address.",
					"optional": false
				}
			},
			"response": {
				"records": {
					"type": "array",
					"description": "The identity records info",
					"fields": {
						"id": {
							"type":        "number",
							"description": "The identity record id"
						},
						"address": {
							"type":        "string",
							"description": "The address of identity record"
						},
						"key": {
							"type":        "string",
							"description": "The identity record key"
						},
						"value": {
							"type":        "string",
							"description": "The identity record value"
						},
						"date": {
							"type":        "string",
							"description": "The identity record timestamp"
						},
						"verifiers": {
							"type":        "array<string>",
							"description": "The address list of verifiers"
						}
					}
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryAllIdentityRecords",
		config.QueryAllIdentityRecords,
		`{
			"description": "QueryAllIdentityRecords is a function to query all identity records.",
			"parameters": {
				"key": {
					"type":        "string",
					"description": "This is an option for pagination. key is a value returned in PageResponse.next_key to begin querying the next page most efficiently. Only one of offset or key should be set.",
					"optional": true
				},
				"offset": {
					"type":        "string",
					"description": "This is an option for pagination. offset is a numeric offset that can be used when key is unavailable. It is less efficient than using key. Only one of offset or key should be set.",
					"optional": true
				},
				"limit": {
					"type":        "string",
					"description": "This is an option for pagination. limit is the total number of results to be returned in the result page. If left empty it will default to a value to be set by each app.",
					"optional": true
				},
				"countTotal": {
					"type":        "string",
					"description": "This is an option for pagination. count_total is set to true  to indicate that the result set should include a count of the total number of items available for pagination in UIs. count_total is only respected when offset is used. It is ignored when key is set.",
					"optional": true
				}
			},
			"response": {
				"records": {
					"type": "array",
					"description": "The identity records info",
					"fields": {
						"id": {
							"type":        "number",
							"description": "The identity record id"
						},
						"address": {
							"type":        "string",
							"description": "The address of identity record"
						},
						"key": {
							"type":        "string",
							"description": "The identity record key"
						},
						"value": {
							"type":        "string",
							"description": "The identity record value"
						},
						"date": {
							"type":        "string",
							"description": "The identity record timestamp"
						},
						"verifiers": {
							"type":        "array<string>",
							"description": "The address list of verifiers"
						}
					}
				},
				"pagination": {
					"description": "The pagination response information like total and next_key"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryIdentityRecordVerifyRequest",
		config.QueryIdentityRecordVerifyRequest,
		`{
			"description": "QueryIdentityRecordVerifyRequest is a function to query identity record verify request.",
			"parameters": {
				"request_id": {
					"type":        "number",
					"description": "This is the identity record verify request id.",
					"optional": false
				}
			},
			"response": {
				"verify_record": {
					"description": "The identity record verify request info",
					"fields": {
						"id": {
							"type":        "number",
							"description": "The verify request id"
						},
						"address": {
							"type":        "string",
							"description": "The request address of identity record"
						},
						"verifier": {
							"type":        "string",
							"description": "The verifier address of identity record"
						},
						"recordIds": {
							"type":        "array<number>",
							"description": "The array of identity record id"
						},
						"tip": {
							"type":        "Coin",
							"description": "The tip amount for verification",
							"fields": {
								"denom": {
									"type": "string"
								},
								"amount": {
									"type": "string"
								}
							}
						},
						"lastRecordEditDate": {
							"type":        "string",
							"description": "The latest edit timestamp"
						}
					}
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryIdentityRecordVerifyRequestsByRequester",
		config.QueryIdentityRecordVerifyRequestsByRequester,
		`{
			"description": "QueryIdentityRecordVerifyRequestsByRequester is a function to query identity record verify request by requester.",
			"parameters": {
				"requester": {
					"type":        "string",
					"description": "This is the identity record verify requester address.",
					"optional": false
				},
				"key": {
					"type":        "string",
					"description": "This is an option for pagination. key is a value returned in PageResponse.next_key to begin querying the next page most efficiently. Only one of offset or key should be set.",
					"optional": true
				},
				"offset": {
					"type":        "string",
					"description": "This is an option for pagination. offset is a numeric offset that can be used when key is unavailable. It is less efficient than using key. Only one of offset or key should be set.",
					"optional": true
				},
				"limit": {
					"type":        "string",
					"description": "This is an option for pagination. limit is the total number of results to be returned in the result page. If left empty it will default to a value to be set by each app.",
					"optional": true
				},
				"countTotal": {
					"type":        "string",
					"description": "This is an option for pagination. count_total is set to true  to indicate that the result set should include a count of the total number of items available for pagination in UIs. count_total is only respected when offset is used. It is ignored when key is set.",
					"optional": true
				}
			},
			"response": {
				"verify_records": {
					"type": "array",
					"description": "The identity record verify request info",
					"fields": {
						"id": {
							"type":        "number",
							"description": "The verify request id"
						},
						"address": {
							"type":        "string",
							"description": "The request address of identity record"
						},
						"verifier": {
							"type":        "string",
							"description": "The verifier address of identity record"
						},
						"recordIds": {
							"type":        "array<number>",
							"description": "The array of identity record id"
						},
						"tip": {
							"type":        "Coin",
							"description": "The tip amount for verification",
							"fields": {
								"denom": {
									"type": "string"
								},
								"amount": {
									"type": "string"
								}
							}
						},
						"lastRecordEditDate": {
							"type":        "string",
							"description": "The latest edit timestamp"
						}
					}
				},
				"pagination": {
					"description": "The pagination response information like total and next_key"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryIdentityRecordVerifyRequestsByApprover",
		config.QueryIdentityRecordVerifyRequestsByApprover,
		`{
			"description": "QueryIdentityRecordVerifyRequestsByApprover is a function to query identity record verify request by approver.",
			"parameters": {
				"requester": {
					"type":        "string",
					"description": "This is the identity record verify request approver address.",
					"optional": false
				},
				"key": {
					"type":        "string",
					"description": "This is an option for pagination. key is a value returned in PageResponse.next_key to begin querying the next page most efficiently. Only one of offset or key should be set.",
					"optional": true
				},
				"offset": {
					"type":        "string",
					"description": "This is an option for pagination. offset is a numeric offset that can be used when key is unavailable. It is less efficient than using key. Only one of offset or key should be set.",
					"optional": true
				},
				"limit": {
					"type":        "string",
					"description": "This is an option for pagination. limit is the total number of results to be returned in the result page. If left empty it will default to a value to be set by each app.",
					"optional": true
				},
				"countTotal": {
					"type":        "string",
					"description": "This is an option for pagination. count_total is set to true  to indicate that the result set should include a count of the total number of items available for pagination in UIs. count_total is only respected when offset is used. It is ignored when key is set.",
					"optional": true
				}
			},
			"response": {
				"verify_records": {
					"type": "array",
					"description": "The identity record verify request info",
					"fields": {
						"id": {
							"type":        "number",
							"description": "The verify request id"
						},
						"address": {
							"type":        "string",
							"description": "The request address of identity record"
						},
						"verifier": {
							"type":        "string",
							"description": "The verifier address of identity record"
						},
						"recordIds": {
							"type":        "array<number>",
							"description": "The array of identity record id"
						},
						"tip": {
							"type":        "Coin",
							"description": "The tip amount for verification",
							"fields": {
								"denom": {
									"type": "string"
								},
								"amount": {
									"type": "string"
								}
							}
						},
						"lastRecordEditDate": {
							"type":        "string",
							"description": "The latest edit timestamp"
						}
					}
				},
				"pagination": {
					"description": "The pagination response information like total and next_key"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryAllIdentityRecordVerifyRequests",
		config.QueryAllIdentityRecordVerifyRequests,
		`{
			"description": "QueryAllIdentityRecordVerifyRequests is a function to query all identity record verify requests.",
			"parameters": {
				"key": {
					"type":        "string",
					"description": "This is an option for pagination. key is a value returned in PageResponse.next_key to begin querying the next page most efficiently. Only one of offset or key should be set.",
					"optional": true
				},
				"offset": {
					"type":        "string",
					"description": "This is an option for pagination. offset is a numeric offset that can be used when key is unavailable. It is less efficient than using key. Only one of offset or key should be set.",
					"optional": true
				},
				"limit": {
					"type":        "string",
					"description": "This is an option for pagination. limit is the total number of results to be returned in the result page. If left empty it will default to a value to be set by each app.",
					"optional": true
				},
				"countTotal": {
					"type":        "string",
					"description": "This is an option for pagination. count_total is set to true  to indicate that the result set should include a count of the total number of items available for pagination in UIs. count_total is only respected when offset is used. It is ignored when key is set.",
					"optional": true
				}
			},
			"response": {
				"verify_records": {
					"type": "array",
					"description": "The identity record verify request info",
					"fields": {
						"id": {
							"type":        "number",
							"description": "The verify request id"
						},
						"address": {
							"type":        "string",
							"description": "The request address of identity record"
						},
						"verifier": {
							"type":        "string",
							"description": "The verifier address of identity record"
						},
						"recordIds": {
							"type":        "array<number>",
							"description": "The array of identity record id"
						},
						"tip": {
							"type":        "Coin",
							"description": "The tip amount for verification",
							"fields": {
								"denom": {
									"type": "string"
								},
								"amount": {
									"type": "string"
								}
							}
						},
						"lastRecordEditDate": {
							"type":        "string",
							"description": "The latest edit timestamp"
						}
					}
				},
				"pagination": {
					"description": "The pagination response information like total and next_key"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryStakingPool",
		config.QueryStakingPool,
		`{
			"description": "QueryStakingPool is a function to query a validator's staking pool.",
			"parameters": {
				"validatorAddress": {
					"type":        "string",
					"description": "This is the validator's address.",
					"optional": true
				}
			},
			"response": {
				"id": {
					"type":        "number",
					"description": "The pool id"
				},
				"commission": {
					"type":        "number",
					"description": "The pool's commission rate"
				},
				"slashed": {
					"type":        "number",
					"description": "The pool's total slashed percentage"
				},
				"voting_power": {
					"type":        "array<string>",
					"description": "The pool's total delegation tokens"
				},
				"total_delegators": {
					"type":        "number",
					"description": "Total number of delegators staked into the pool"
				},
				"tokens": {
					"type": 	   "array<string>",
					"description": "list of token types available to stake in the pool"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryUndelegations",
		config.QueryUndelegations,
		`{
			"description": "QueryUnelegations is a function to query all undelegation records for a delegator.",
			"parameters": {
				"undelegatorAddress": {
					"type":        "string",
					"description": "This is the delegator's address.",
					"optional": true
				},
				"offset": {
					"type":        "string",
					"description": "This is an option for pagination. offset is a numeric offset that specifies the starting point of the result set. If left empty it will default to a value to be set by each app.",
					"optional": true
				},
				"limit": {
					"type":        "string",
					"description": "This is an option for pagination. limit is the total number of results to be returned in the result page. If left empty it will default to a value to be set by each app.",
					"optional": true
				},
				"countTotal": {
					"type":        "string",
					"description": "This is an option for pagination. count_total is set to true  to indicate that the result set should include a count of the total number of items available for pagination in UIs.",
					"optional": true
				}
			},
			"response": {
				"undelegations": {
					"type": "array",
					"description": "The undelegations records info",
					"fields": {
						"validator_info": {
							"type":        "struct",
							"description": "The validator information",
							"fields": {
								"moniker": {
									"type":        "string",
									"description": "The validator's moniker"
								},
								"address": {
									"type":        "string",
									"description": "The validator's address"
								},
								"logo_url": {
									"type":        "string",
									"description": "The validator's logo"
								},
								"valkey": {
									"type":        "string",
									"description": "Validator’s valkey, required for MsgClaimUndelegation"
								}
							}
						},
						"tokens": {
							"type": "array",
							"description": "The token information",
							"fields": {
								"denom": {
									"type": "string",
									"description": "The token name"
								},
								"amount": {
									"type": "number",
									"description": "The token amount"
								}
							}
						},
						"expiry": {
                            "type": "string",
                            "description": "Epoch time value. After it passes, the tokens can be claimed"
						},
						"id": {
                            "type": "number",
                            "description": "Unique record marking, required for MsgClaimUndelegation"
						}
					}
				},
				"pagination": {
					"description": "The pagination response information that includes total count"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryDelegations",
		config.QueryDelegations,
		`{
			"description": "QueryDelegations is a function to query all delegation records for a delegator.",
			"parameters": {
				"delegatorAddress": {
					"type":        "string",
					"description": "This is the delegator's address.",
					"optional": true
				},
				"offset": {
					"type":        "string",
					"description": "This is an option for pagination. offset is a numeric offset that specifies the starting point of the result set. If left empty it will default to a value to be set by each app.",
					"optional": true
				},
				"limit": {
					"type":        "string",
					"description": "This is an option for pagination. limit is the total number of results to be returned in the result page. If left empty it will default to a value to be set by each app.",
					"optional": true
				},
				"countTotal": {
					"type":        "string",
					"description": "This is an option for pagination. count_total is set to true  to indicate that the result set should include a count of the total number of items available for pagination in UIs.",
					"optional": true
				}
			},
			"response": {
				"delegations": {
					"type": "array",
					"description": "The delegations records info",
					"fields": {
						"validator_info": {
							"type":        "struct",
							"description": "The validator information",
							"fields": {
								"moniker": {
									"type":        "string",
									"description": "The validator's moniker"
								},
								"address": {
									"type":        "string",
									"description": "The validator's address"
								},
								"valkey": {
									"type":        "string",
									"description": "The validator's valoper address"
								},
								"website": {
									"type":        "string",
									"description": "The validator's official website"
								},
								"logo": {
									"type":        "string",
									"description": "The validator's logo"
								}
							}
						},
						"pool_info": {
							"type":        "struct",
							"description": "The pool information",
							"fields": {
								"id": {
									"type": "number",
									"description": "The pool id"
								},
								"commission": {
									"type": "number",
									"description": "The pool's commission rate"
								},
								"status": {
									"type": "string",
									"description": "The pool's status"
								},
								"tokens": {
									"type": "array<string>",
									"description": "list of token types available to stake in the pool"
								}
							}
						}
					}
				},
				"pagination": {
					"description": "The pagination response information that includes total count"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryGenesis",
		config.QueryGenesis,
		`{
			"description": "QueryGenesis is a function to query genesis."
		}`,
	)

	AddInterxFunction(
		"QueryGenesisSum",
		config.QueryGenesisSum,
		`{
			"description": "QueryGenesisSum is a function to query genesis checksum."
		}`,
	)

	AddInterxFunction(
		"QueryInterxStatus",
		config.QueryStatus,
		`{
			"description": "QueryInterxStatus is a function to query interx informations."
		}`,
	)

	AddInterxFunction(
		"QueryRPCMethods",
		config.QueryRPCMethods,
		`{
			"description": "QueryRPCMethods is a function to query all rpc methods available."
		}`,
	)

	AddInterxFunction(
		"QueryKiraFunctions",
		config.QueryKiraFunctions,
		`{
			"description": "QueryKiraFunctions is a function to query kira functions."
		}`,
	)

	AddInterxFunction(
		"QueryInterxFunctions",
		config.QueryInterxFunctions,
		`{
			"description": "QueryInterxFunctions is a function to query interx functions."
		}`,
	)

	AddInterxFunction(
		"QuerySpendingPools",
		config.QuerySpendingPools,
		`{
			"description": "QuerySpendingPools is a function to query list of all spending pool names.",
			"parameters": {
				"account": {
					"type":        "string",
					"description": "This represents the kira account address.",
					"optional": true
				},
				"name": {
					"type":        "string",
					"description": "This represents the pool name.",
					"optional": true
				}
			},
			"response": {
				"names": {
					"description": "The list of all spending pools"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QuerySpendingPoolProposals",
		config.QuerySpendingPoolProposals,
		`{
			"description": "QuerySpendingPoolProposals is a function to query list of all spending pool proposals.",
			"response": {
				"names": {
					"description": "The list of all spending pool proposals"
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryUBIRecords",
		config.QueryUBIRecords,
		`{
			"description": "QueryUBIRecords is a function to query ubi records.",
			"parameters": {
				"name": {
					"type":        "string",
					"description": "This represents the pool name.",
					"optional": true
				}
			},
			"response": {
				"records": {
					"description": "All ubi records",
					"optional": true
				},
				"record": {
					"description": "ubi record",
					"optional": true
				}
			}
		}`,
	)

	AddInterxFunction(
		"QueryDashboard",
		config.QueryDashboard,
		`{
			"description": "QueryDashboard is a function to query data for the dashboard.",
			"response": {
				"consensus_health": {
					"description": "Float value between 0 and 1, represents the health status of the consensus."
				},
				"current_block_validator": {
					"type": "struct",
					"description": "The current block validator info",
					"fields": {
						"moniker": {
							"type":        "string",
							"description": "The validator's moniker"
						},
						"address": {
							"type":        "string",
							"description": "The validator's address"
						}
					}
				},
				"validators": {
					"type": "struct",
					"description": "The validators count info",
					"fields": {
						"total": {
							"type":        "number",
							"description": "The count of total validators"
						},
						"active": {
							"type":        "number",
							"description": "The count of active validators"
						},
						"inactive": {
							"type":        "number",
							"description": "The count of in-active validators"
						},
						"jailed": {
							"type":        "number",
							"description": "The count of jailed validators"
						},
						"paused": {
							"type":        "number",
							"description": "The count of paused validators"
						},
						"waiting": {
							"type":        "number",
							"description": "The count of waiting validators"
						}
					}
				},
				"blocks": {
					"type": "struct",
					"description": "The blocks consensus status",
					"fields": {
						"current_height": {
							"type":        "number",
							"description": "The current block height"
						},
						"since_genesis": {
							"type":        "number",
							"description": "The block count after genesis block"
						},
						"pending_transactions": {
							"type":        "number",
							"description": "The count of pending transactions"
						},
						"current_transactions": {
							"type":        "number",
							"description": "The count of current transactions"
						},
						"latest_time": {
							"type":        "number",
							"description": "The latest block confirm time"
						},
						"average_time": {
							"type":        "number",
							"description": "The average block confirm time"
						}
					}
				},
				"proposals": {
					"type": "struct",
					"description": "The proposals count info",
					"fields": {
						"total": {
							"type":        "number",
							"description": "The count of total proposals"
						},
						"active": {
							"type":        "number",
							"description": "The count of active proposals"
						},
						"enacting": {
							"type":        "number",
							"description": "The count of enacting proposals"
						},
						"finished": {
							"type":        "number",
							"description": "The count of finished proposals"
						},
						"successful": {
							"type":        "number",
							"description": "The count of successful proposals"
						},
						"proposers": {
							"type":        "number",
							"description": "The count of proposers"
						},
						"voters": {
							"type":        "number",
							"description": "The count of voters"
						}
					}
				}
			}
		}`,
	)
}
