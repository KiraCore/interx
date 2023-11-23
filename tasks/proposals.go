package tasks

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/database"
	"github.com/KiraCore/interx/types/kira/gov"
	tmjson "github.com/cometbft/cometbft/libs/json"
	tmTypes "github.com/cometbft/cometbft/rpc/core/types"
	tmJsonRPCTypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	sekaitypes "github.com/KiraCore/sekai/types"
)

var (
	AllProposals gov.AllProposals
	ProposalsMap map[string]gov.Proposal = make(map[string]gov.Proposal)
)

const (
	// Undefined status
	VOTE_RESULT_PASSED string = "VOTE_RESULT_PASSED"
	// Active status
	VOTE_PENDING string = "VOTE_PENDING"
	// Inactive status
	VOTE_RESULT_ENACTMENT string = "VOTE_RESULT_ENACTMENT"
)

func QueryProposals(gwCosmosmux *runtime.ServeMux, gatewayAddr string, rpcAddr string) error {
	result := gov.ProposalsResponse{}

	limit := sekaitypes.PageIterationLimit - 1
	offset := 0
	for {
		proposalsQueryRequest, _ := http.NewRequest("GET", "http://"+gatewayAddr+"/kira/gov/proposals?pagination.offset="+strconv.Itoa(offset)+"&pagination.limit="+strconv.Itoa(limit), nil)

		proposalsQueryResponse, failure, _ := common.ServeGRPC(proposalsQueryRequest, gwCosmosmux)

		if proposalsQueryResponse == nil {
			return errors.New(ToString(failure))
		}

		byteData, err := json.Marshal(proposalsQueryResponse)
		if err != nil {
			return err
		}

		subResult := gov.ProposalsResponse{}
		err = json.Unmarshal(byteData, &subResult)
		if err != nil {
			return err
		}

		if len(subResult.Proposals) == 0 {
			break
		}

		result.Proposals = append(result.Proposals, subResult.Proposals...)

		offset += limit
	}

	for _, prop := range result.Proposals {
		ProposalsMap[prop.ProposalID] = prop
	}

	cachedProps, err := GetCachedProposals(gwCosmosmux, gatewayAddr, rpcAddr)
	if err != nil {
		return err
	}

	for _, cachedProp := range cachedProps {
		prop, found := ProposalsMap[cachedProp.ProposalID]
		if found {
			prop.VotersCount = cachedProp.VotersCount
			prop.VotesCount = cachedProp.VotesCount
			prop.Quorum = cachedProp.Quorum
			prop.Metadata = cachedProp.Metadata
			prop.Hash = cachedProp.Hash
			prop.Timestamp = cachedProp.Timestamp
			prop.BlockHeight = cachedProp.BlockHeight
			prop.Type = cachedProp.Type
			prop.Proposer = cachedProp.Proposer
			ProposalsMap[cachedProp.ProposalID] = prop
		}
	}

	allProposals := gov.AllProposals{}

	allProposals.Proposals = result.Proposals

	allProposals.Status.TotalProposals = len(result.Proposals)
	allProposals.Status.ActiveProposals = 0
	allProposals.Status.EnactingProposals = 0
	allProposals.Status.FinishedProposals = 0
	allProposals.Status.SuccessfulProposals = 0
	for _, proposal := range result.Proposals {
		if proposal.Result == VOTE_PENDING {
			allProposals.Status.ActiveProposals++
		}
		if proposal.Result == VOTE_RESULT_ENACTMENT {
			allProposals.Status.EnactingProposals++
		}
		if proposal.Result == VOTE_RESULT_PASSED {
			allProposals.Status.SuccessfulProposals++
		}
	}

	allProposals.Status.FinishedProposals = allProposals.Status.TotalProposals - allProposals.Status.ActiveProposals - allProposals.Status.EnactingProposals

	{
		request, _ := http.NewRequest("GET", "http://"+gatewayAddr+"/kira/gov/proposers_voters_count", nil)

		response, failure, _ := common.ServeGRPC(request, gwCosmosmux)

		if response == nil {
			return errors.New(ToString(failure))
		}

		byteData, err := json.Marshal(response)
		if err != nil {
			return err
		}
		result := gov.ProposalUserCount{}
		err = json.Unmarshal(byteData, &result)
		if err != nil {
			return err
		}

		allProposals.Users = result
	}

	AllProposals = allProposals

	return nil
}

// GetCachedProposals syncs with sekai for querying new proposals and return cached proposals
func GetCachedProposals(gwCosmosmux *runtime.ServeMux, gatewayAddr string, rpcAddr string) ([]gov.CachedProposal, error) {
	// fetch the block number for the latest cached proposal and sync with sekai for newer proposals
	lastBlock := database.GetLastBlockFetchedForProposals()
	cachedProps := []gov.CachedProposal{}
	propTxs := tmTypes.ResultTxSearch{
		Txs:        []*tmTypes.ResultTx{},
		TotalCount: 0,
	}
	page := 1
	for {
		var events = make([]string, 0, 5)
		events = append(events, "submit_proposal.proposal_id>0")
		events = append(events, fmt.Sprintf("tx.height>%d", lastBlock))

		endpoint := fmt.Sprintf("%s/tx_search?query=\"%s\"&page=%d&per_page=100&order_by=\"desc\"", rpcAddr, strings.Join(events, "%20AND%20"), page)
		resp, err := http.Get(endpoint)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		respBody, _ := ioutil.ReadAll(resp.Body)
		response := new(tmJsonRPCTypes.RPCResponse)

		if err := json.Unmarshal(respBody, response); err != nil {
			break
		}

		if response.Error != nil {
			break
		}

		result := new(tmTypes.ResultTxSearch)
		if err := tmjson.Unmarshal(response.Result, result); err != nil {
			break
		}

		if result.TotalCount == 0 {
			break
		}

		propTxs.Txs = append(propTxs.Txs, result.Txs...)

		if result.TotalCount < 100 {
			break
		}
		page++
	}
	propTxs.TotalCount = len(propTxs.Txs)

	// grab quorum through a gRPC call
	quorumStr := ""
	networkInfoQueryRequest, _ := http.NewRequest("GET", "http://"+gatewayAddr+"/kira/gov/network_properties", nil)
	success, _, _ := common.ServeGRPC(networkInfoQueryRequest, gwCosmosmux)
	if success != nil {
		networkInfo, err := common.QueryNetworkPropertiesFromGrpcResult(success)
		if err != nil {
			return nil, err
		}

		result := make(map[string]map[string]interface{})
		byteData, err := json.Marshal(networkInfo)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(byteData, &result)
		if err != nil {
			return nil, err
		}

		if result["properties"] != nil {
			quorum, err := strconv.Atoi(result["properties"]["voteQuorum"].(string))
			if err != nil {
				return nil, err
			}

			quorumStr = fmt.Sprintf("%.2f", float64(quorum)/100)
		}
	}

	// create new CachedProposal objects and store them into the cache
	for _, propTx := range propTxs.Txs {
		cachedProp := gov.CachedProposal{}
		cachedProp.Hash = fmt.Sprintf("0x%s", propTx.Hash)
		// grab proposer address, and proposal id from the proposal event
		for _, event := range propTx.TxResult.Events {
			if event.GetType() == "tx" {
				attr := event.GetAttributes()[0]
				key := string(attr.GetKey())
				if key == "acc_seq" {
					// acc_seq format is {addr}/{seq}
					accSeqs := strings.Split(string(attr.GetValue()), "/")
					cachedProp.Proposer = accSeqs[0]
				}
			}

			if event.GetType() == "submit_proposal" {
				attrs := event.GetAttributes()
				for _, attr := range attrs {
					key := string(attr.GetKey())
					if key == "proposal_id" {
						cachedProp.ProposalID = string(attr.GetValue())
					} else if key == "proposal_type" {
						cachedProp.Type = string(attr.GetValue())
					}
				}
			}
		}
		// grab height and block time from height
		cachedProp.BlockHeight = int(propTx.Height)

		txTime, err := common.GetBlockTime(rpcAddr, propTx.Height)
		if err != nil {
			common.GetLogger().Error("[query-transactions] Block not found: ", propTx.Height)
			continue
		}

		cachedProp.Timestamp = int(txTime)

		// grab metadata from the transaction
		tx, err := config.EncodingCg.TxConfig.TxDecoder()(propTx.Tx)
		if err != nil {
			common.GetLogger().Error("[query-transactions] Failed to decode transaction: ", err)
			continue
		}

		txSigning, ok := tx.(signing.Tx)
		if ok {
			cachedProp.Metadata = txSigning.GetMemo()
		}

		// grab voters through a gRPC call
		votersQueryRequest, _ := http.NewRequest("GET", "http://"+gatewayAddr+"/kira/gov/voters/"+cachedProp.ProposalID, nil)
		success, _, _ := common.ServeGRPC(votersQueryRequest, gwCosmosmux)

		if success != nil {
			voters, err := common.QueryVotersFromGrpcResult(success)
			if err != nil {
				return nil, err
			}

			if voters != nil {
				cachedProp.VotersCount = len(voters)
			}
		}

		// grab votes through a gRPC call
		votesQueryRequest, _ := http.NewRequest("GET", "http://"+gatewayAddr+"/kira/gov/votes/"+cachedProp.ProposalID, nil)
		success, _, _ = common.ServeGRPC(votesQueryRequest, gwCosmosmux)

		if success != nil {
			votes, err := common.QueryVotesFromGrpcResult(success)
			if err != nil {
				return nil, err
			}

			if votes != nil {
				cachedProp.VotesCount = len(votes)
			}
		}

		cachedProp.Quorum = quorumStr
		cachedProps = append(cachedProps, cachedProp)
	}

	err := database.SaveProposals(cachedProps)
	if err != nil {
		return nil, err
	}

	return database.GetProposals()
}

func SyncProposals(gwCosmosmux *runtime.ServeMux, gatewayAddr string, rpcAddr string, isLog bool) {
	lastBlock := int64(0)
	for {
		if common.NodeStatus.Block != lastBlock {
			err := QueryProposals(gwCosmosmux, gatewayAddr, rpcAddr)

			if err != nil && isLog {
				common.GetLogger().Error("[sync-proposals] Failed to query proposals: ", err)
			}

			lastBlock = common.NodeStatus.Block
		}

		time.Sleep(1 * time.Second)
	}
}
