package metamask

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/KiraCore/interx/config"
	tx "github.com/KiraCore/interx/proto-gen/cosmos/tx/v1beta1"
	custodytypes "github.com/KiraCore/sekai/x/custody/types"
	evidencetypes "github.com/KiraCore/sekai/x/evidence/types"
	customgovtypes "github.com/KiraCore/sekai/x/gov/types"
	multistakingtypes "github.com/KiraCore/sekai/x/multistaking/types"
	customslashingtypes "github.com/KiraCore/sekai/x/slashing/types"
	spendingtypes "github.com/KiraCore/sekai/x/spending/types"
	customstakingtypes "github.com/KiraCore/sekai/x/staking/types"
	tokenstypes "github.com/KiraCore/sekai/x/tokens/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	cosmostypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

const (
	MsgBankSend = iota
	// customgov
	MsgSubmitEvidence
	MsgSubmitProposal
	MsgVoteProposal
	MsgRegisterIdentityRecords
	MsgDeleteIdentityRecords
	MsgRequestIdentityRecordsVerify
	MsgHandleIdentityRecordsVerifyRequest
	MsgCancelIdentityRecordsVerifyRequest
	MsgSetNetworkProperties
	MsgSetExecutionFee
	MsgClaimCouncilor
	MsgWhitelistPermissions
	MsgBlacklistPermissions
	MsgCreateRole
	MsgAssignRole
	MsgUnassignRole
	MsgWhitelistRolePermission
	MsgBlacklistRolePermission
	MsgRemoveWhitelistRolePermission
	MsgRemoveBlacklistRolePermission
	// staking
	MsgClaimValidator
	// tokens
	MsgUpsertTokenAlias
	MsgUpsertTokenRate
	// slashing
	MsgActivate
	MsgPause
	MsgUnpause
	// spending
	MsgCreateSpendingPool
	MsgDepositSpendingPool
	MsgRegisterSpendingPoolBeneficiary
	MsgClaimSpendingPool
	// multistaking
	MsgUpsertStakingPool
	MsgDelegate
	MsgUndelegate
	MsgClaimRewards
	MsgClaimUndelegation
	MsgSetCompoundInfo
	MsgRegisterDelegator
	// custody module
	MsgCreateCustody
	MsgAddToCustodyWhiteList
	MsgAddToCustodyCustodians
	MsgRemoveFromCustodyCustodians
	MsgDropCustodyCustodians
	MsgRemoveFromCustodyWhiteList
	MsgDropCustodyWhiteList
	MsgApproveCustodyTx
	MsgDeclineCustodyTx
)

var grpcConn *grpc.ClientConn

type ListRoleParam struct {
	RoleIdentifier string `json:"role_identifier"`
	Permission     uint32 `json:"permission"`
}

type PermissionsParam struct {
	Permission     uint32 `json:"permission"`
	ControlledAddr string `json:"controlled_addr"`
}

type RoleParam struct {
	RoleId     uint32 `json:"role_id"`
	Controller string `json:"controller"`
}

type DelegateParam struct {
	Amounts string `json:"amounts"`
	To      string `json:"to"`
}

type CustodianParam struct {
	NewAddrs []string `json:"new_addrs"`
	OldKey   string   `json:"old_key"`
	NewKey   string   `json:"new_key"`
	Next     string   `json:"next"`
	Target   string   `json:"target"`
}

type SingleCustodianParam struct {
	NewAddr string `json:"new_addr"`
	OldKey  string `json:"old_key"`
	NewKey  string `json:"new_key"`
	Next    string `json:"next"`
	Target  string `json:"target"`
}

type CustodyParam struct {
	To   string `json:"to"`
	Hash string `json:"hash"`
}

type SpendingPoolParam struct {
	Name    string `json:"name"`
	Amounts string `json:"amounts"`
}

// decode 256bit param like bool, uint, hex-typed address etc
func Decode256Bit(data *[]byte, params *[][]byte) error {
	if len(*data) < 32 {
		return errors.New("decoding 256bit failed, not enough length")
	}
	*params = append(*params, (*data)[:32])
	*data = (*data)[32:]

	return nil
}

// decode string-typed param
// structure:
// * offset - offset of the string in the data 	: 32byte
// * length - length of the string 				: 32byte
// * content - content of the string			: (length/32+1)*32byte
func DecodeString(data *[]byte, params *[][]byte) error {
	// offset := data[:32] // string value offset
	*data = (*data)[32:]

	length, err := bytes2uint64((*data)[:32])
	if err != nil {
		return err
	}
	*data = (*data)[32:]

	*params = append(*params, (*data)[:length])
	*data = (*data)[(length/32+1)*32:]
	return nil
}

func DecodeParam(data []byte, txType int) ([][]byte, error) {
	if txType == MsgBankSend {
		return nil, nil
	}

	var params [][]byte

	// decode data field v, r, s, sender
	for i := 0; i < 4; i++ {
		err := Decode256Bit(&data, &params)
		if err != nil {
			return nil, err
		}
	}

	// decode param string
	err := DecodeString(&data, &params)

	return params, err
}

func sendTx(txRawData string, gwCosmosmux *runtime.ServeMux, r *http.Request) (string, error) {
	byteData, err := hex.DecodeString(txRawData[2:])
	if err != nil {
		return "", err
	}

	ethTxData, err := GetEthTxInfo(byteData)
	if err != nil {
		return "", err
	}

	txBytes, err := SignTx(ethTxData, byteData, gwCosmosmux, r)
	if err != nil {
		return "", err
	}

	txHash, err := sendCosmosTx(r.Context(), txBytes)
	if err != nil {
		return "", err
	}

	return txHash, nil
}

func getTxType(txData []byte) (int, error) {
	submitEvidencePrefix, _ := hex.DecodeString("85db2453")
	submitProposalPrefix, _ := hex.DecodeString("00000000")
	voteProposalPrefix, _ := hex.DecodeString("7f1f06dc")
	registerIdentityRecordsPrefix, _ := hex.DecodeString("bc05f106")
	deleteIdentityRecordsPrefix, _ := hex.DecodeString("2dbad8e8")
	requestIdentityRecordsVerifyPrefix, _ := hex.DecodeString("9765358e")
	handleIdentityRecordsVerifyRequestPrefix, _ := hex.DecodeString("4335ed0c")
	cancelIdentityRecordsVerifyRequestPrefix, _ := hex.DecodeString("eeaa0488")
	setNetworkPropertiesPrefix, _ := hex.DecodeString("f8060fb5")
	setExecutionFeePrefix, _ := hex.DecodeString("c7586de8")
	claimCouncilorPrefix, _ := hex.DecodeString("b7b8ff46")
	whitelistPermissionsPrefix, _ := hex.DecodeString("2f313ab8")
	blacklistPermissionsPrefix, _ := hex.DecodeString("3864f845")
	createRolePrefix, _ := hex.DecodeString("2d8abfdf")
	assignRolePrefix, _ := hex.DecodeString("fcc121b5")
	unassignRolePrefix, _ := hex.DecodeString("c79ca19d")
	whitelistRolePermissionPrefix, _ := hex.DecodeString("59472362")
	blacklistRolePermissionPrefix, _ := hex.DecodeString("99c557da")
	removeWhitelistRolePermissionPrefix, _ := hex.DecodeString("2a11d702")
	removeBlacklistRolePermissionPrefix, _ := hex.DecodeString("f5f865e4")
	claimValidatorPrefix, _ := hex.DecodeString("00000000")
	upsertTokenAliasPrefix, _ := hex.DecodeString("f69a4787")
	upsertTokenRatePrefix, _ := hex.DecodeString("3b30a97a")
	activatePrefix, _ := hex.DecodeString("a1374dc2")
	pausePrefix, _ := hex.DecodeString("1371cf19")
	unpausePrefix, _ := hex.DecodeString("b9179894")
	createSpendingPoolPrefix, _ := hex.DecodeString("4ed8a0a2")
	depositSpendingPoolPrefix, _ := hex.DecodeString("e10c925c")
	registerSpendingPoolBeneficiaryPrefix, _ := hex.DecodeString("7ab7eecf")
	claimSpendingPoolPrefix, _ := hex.DecodeString("efeed4a0")
	upsertStakingPoolPrefix, _ := hex.DecodeString("fb24f5cc")
	delegatePrefix, _ := hex.DecodeString("4b193c09")
	undelegatePrefix, _ := hex.DecodeString("94574f0c")
	claimRewardsPrefix, _ := hex.DecodeString("9838bc2f")
	claimUndelegationPrefix, _ := hex.DecodeString("2f608d76")
	setCompoundInfoPrefix, _ := hex.DecodeString("e2d6a093")
	registerDelegatorPrefix, _ := hex.DecodeString("99db185d")
	createCustodyPrefix, _ := hex.DecodeString("bebde6d1")
	addToCustodyWhiteListPrefix, _ := hex.DecodeString("25a1d834")
	addToCustodyCustodiansPrefix, _ := hex.DecodeString("8c7fdb91")
	removeFromCustodyCustodiansPrefix, _ := hex.DecodeString("90be51cf")
	dropCustodyCustodiansPrefix, _ := hex.DecodeString("0ca697b4")
	removeFromCustodyWhiteListPrefix, _ := hex.DecodeString("fa431c3e")
	dropCustodyWhiteListPrefix, _ := hex.DecodeString("bc65010a")
	approveCustodyTxPrefix, _ := hex.DecodeString("5da292d4")
	declineCustodyTxPrefix, _ := hex.DecodeString("dce4399a")

	var msgType int
	switch {
	case txData == nil:
		msgType = MsgBankSend
	case len(txData) == 0:
		msgType = MsgBankSend
	case bytes.Equal(txData, delegatePrefix):
		msgType = MsgDelegate
	case bytes.Equal(txData, undelegatePrefix):
		msgType = MsgUndelegate
	case bytes.Equal(txData, submitEvidencePrefix):
		msgType = MsgSubmitEvidence
	case bytes.Equal(txData, submitProposalPrefix):
		msgType = MsgSubmitProposal
	case bytes.Equal(txData, voteProposalPrefix):
		msgType = MsgVoteProposal
	case bytes.Equal(txData, registerIdentityRecordsPrefix):
		msgType = MsgRegisterIdentityRecords
	case bytes.Equal(txData, deleteIdentityRecordsPrefix):
		msgType = MsgDeleteIdentityRecords
	case bytes.Equal(txData, requestIdentityRecordsVerifyPrefix):
		msgType = MsgRequestIdentityRecordsVerify
	case bytes.Equal(txData, handleIdentityRecordsVerifyRequestPrefix):
		msgType = MsgHandleIdentityRecordsVerifyRequest
	case bytes.Equal(txData, cancelIdentityRecordsVerifyRequestPrefix):
		msgType = MsgCancelIdentityRecordsVerifyRequest
	case bytes.Equal(txData, setNetworkPropertiesPrefix):
		msgType = MsgSetNetworkProperties
	case bytes.Equal(txData, setExecutionFeePrefix):
		msgType = MsgSetExecutionFee
	case bytes.Equal(txData, claimCouncilorPrefix):
		msgType = MsgClaimCouncilor
	case bytes.Equal(txData, whitelistPermissionsPrefix):
		msgType = MsgWhitelistPermissions
	case bytes.Equal(txData, blacklistPermissionsPrefix):
		msgType = MsgBlacklistPermissions
	case bytes.Equal(txData, createRolePrefix):
		msgType = MsgCreateRole
	case bytes.Equal(txData, assignRolePrefix):
		msgType = MsgAssignRole
	case bytes.Equal(txData, unassignRolePrefix):
		msgType = MsgUnassignRole
	case bytes.Equal(txData, whitelistRolePermissionPrefix):
		msgType = MsgWhitelistRolePermission
	case bytes.Equal(txData, blacklistRolePermissionPrefix):
		msgType = MsgBlacklistRolePermission
	case bytes.Equal(txData, removeWhitelistRolePermissionPrefix):
		msgType = MsgRemoveWhitelistRolePermission
	case bytes.Equal(txData, removeBlacklistRolePermissionPrefix):
		msgType = MsgBlacklistRolePermission
	case bytes.Equal(txData, claimValidatorPrefix):
		msgType = MsgClaimValidator
	case bytes.Equal(txData, upsertTokenAliasPrefix):
		msgType = MsgUpsertTokenAlias
	case bytes.Equal(txData, upsertTokenRatePrefix):
		msgType = MsgUpsertTokenRate
	case bytes.Equal(txData, activatePrefix):
		msgType = MsgActivate
	case bytes.Equal(txData, pausePrefix):
		msgType = MsgPause
	case bytes.Equal(txData, unpausePrefix):
		msgType = MsgUnpause
	case bytes.Equal(txData, createSpendingPoolPrefix):
		msgType = MsgCreateSpendingPool
	case bytes.Equal(txData, depositSpendingPoolPrefix):
		msgType = MsgDepositSpendingPool
	case bytes.Equal(txData, registerSpendingPoolBeneficiaryPrefix):
		msgType = MsgRegisterSpendingPoolBeneficiary
	case bytes.Equal(txData, claimSpendingPoolPrefix):
		msgType = MsgClaimSpendingPool
	case bytes.Equal(txData, upsertStakingPoolPrefix):
		msgType = MsgUpsertStakingPool
	case bytes.Equal(txData, claimRewardsPrefix):
		msgType = MsgClaimRewards
	case bytes.Equal(txData, claimUndelegationPrefix):
		msgType = MsgClaimUndelegation
	case bytes.Equal(txData, setCompoundInfoPrefix):
		msgType = MsgSetCompoundInfo
	case bytes.Equal(txData, registerDelegatorPrefix):
		msgType = MsgRegisterDelegator
	case bytes.Equal(txData, createCustodyPrefix):
		msgType = MsgCreateCustody
	case bytes.Equal(txData, addToCustodyWhiteListPrefix):
		msgType = MsgAddToCustodyWhiteList
	case bytes.Equal(txData, addToCustodyCustodiansPrefix):
		msgType = MsgAddToCustodyCustodians
	case bytes.Equal(txData, removeFromCustodyCustodiansPrefix):
		msgType = MsgRemoveFromCustodyCustodians
	case bytes.Equal(txData, dropCustodyCustodiansPrefix):
		msgType = MsgDropCustodyCustodians
	case bytes.Equal(txData, removeFromCustodyWhiteListPrefix):
		msgType = MsgRemoveFromCustodyWhiteList
	case bytes.Equal(txData, dropCustodyWhiteListPrefix):
		msgType = MsgDropCustodyWhiteList
	case bytes.Equal(txData, approveCustodyTxPrefix):
		msgType = MsgApproveCustodyTx
	case bytes.Equal(txData, declineCustodyTxPrefix):
		msgType = MsgDeclineCustodyTx
	default:
		return 0, errors.New("no such functions")
	}
	return msgType, nil
}

func getInstanceOfTx(txType int) cosmostypes.Msg {
	switch txType {
	case MsgSubmitEvidence:
		return &evidencetypes.MsgSubmitEvidence{}
	case MsgSubmitProposal:
		return &customgovtypes.MsgSubmitProposal{}
	case MsgVoteProposal:
		return &customgovtypes.MsgVoteProposal{}
	case MsgRegisterIdentityRecords:
		return &customgovtypes.MsgRegisterIdentityRecords{}
	case MsgDeleteIdentityRecords:
		return &customgovtypes.MsgDeleteIdentityRecords{}
	case MsgRequestIdentityRecordsVerify:
		return &customgovtypes.MsgRequestIdentityRecordsVerify{}
	case MsgHandleIdentityRecordsVerifyRequest:
		return &customgovtypes.MsgHandleIdentityRecordsVerifyRequest{}
	case MsgCancelIdentityRecordsVerifyRequest:
		return &customgovtypes.MsgCancelIdentityRecordsVerifyRequest{}
	case MsgSetNetworkProperties:
		return &customgovtypes.MsgSetNetworkProperties{}
	case MsgSetExecutionFee:
		return &customgovtypes.MsgSetExecutionFee{}
	case MsgClaimCouncilor:
		return &customgovtypes.MsgClaimCouncilor{}
	case MsgWhitelistPermissions:
		return &customgovtypes.MsgWhitelistPermissions{}
	case MsgBlacklistPermissions:
		return &customgovtypes.MsgBlacklistPermissions{}
	case MsgCreateRole:
		return &customgovtypes.MsgCreateRole{}
	case MsgAssignRole:
		return &customgovtypes.MsgAssignRole{}
	case MsgUnassignRole:
		return &customgovtypes.MsgUnassignRole{}
	case MsgWhitelistRolePermission:
		return &customgovtypes.MsgWhitelistRolePermission{}
	case MsgBlacklistRolePermission:
		return &customgovtypes.MsgBlacklistRolePermission{}
	case MsgRemoveWhitelistRolePermission:
		return &customgovtypes.MsgRemoveWhitelistRolePermission{}
	case MsgRemoveBlacklistRolePermission:
		return &customgovtypes.MsgRemoveBlacklistRolePermission{}
	case MsgClaimValidator:
		return &customstakingtypes.MsgClaimValidator{}
	case MsgUpsertTokenAlias:
		return &tokenstypes.MsgUpsertTokenAlias{}
	case MsgUpsertTokenRate:
		return &tokenstypes.MsgUpsertTokenRate{}
	case MsgActivate:
		return &customslashingtypes.MsgActivate{}
	case MsgPause:
		return &customslashingtypes.MsgPause{}
	case MsgUnpause:
		return &customslashingtypes.MsgUnpause{}
	case MsgCreateSpendingPool:
		return &spendingtypes.MsgCreateSpendingPool{}
	case MsgDepositSpendingPool:
		return &spendingtypes.MsgDepositSpendingPool{}
	case MsgRegisterSpendingPoolBeneficiary:
		return &spendingtypes.MsgRegisterSpendingPoolBeneficiary{}
	case MsgClaimSpendingPool:
		return &spendingtypes.MsgClaimSpendingPool{}
	case MsgUpsertStakingPool:
		return &multistakingtypes.MsgUpsertStakingPool{}
	case MsgDelegate:
		return &multistakingtypes.MsgDelegate{}
	case MsgUndelegate:
		return &multistakingtypes.MsgUndelegate{}
	case MsgClaimRewards:
		return &multistakingtypes.MsgClaimRewards{}
	case MsgClaimUndelegation:
		return &multistakingtypes.MsgClaimUndelegation{}
	case MsgSetCompoundInfo:
		return &multistakingtypes.MsgSetCompoundInfo{}
	case MsgRegisterDelegator:
		return &multistakingtypes.MsgRegisterDelegator{}
	case MsgCreateCustody:
		return &custodytypes.MsgCreateCustodyRecord{}
	case MsgAddToCustodyWhiteList:
		return &custodytypes.MsgAddToCustodyWhiteList{}
	case MsgAddToCustodyCustodians:
		return &custodytypes.MsgAddToCustodyCustodians{}
	case MsgRemoveFromCustodyCustodians:
		return &custodytypes.MsgRemoveFromCustodyCustodians{}
	case MsgDropCustodyCustodians:
		return &custodytypes.MsgDropCustodyCustodians{}
	case MsgRemoveFromCustodyWhiteList:
		return &custodytypes.MsgRemoveFromCustodyWhiteList{}
	case MsgDropCustodyWhiteList:
		return &custodytypes.MsgDropCustodyWhiteList{}
	case MsgApproveCustodyTx:
		return &custodytypes.MsgApproveCustodyTransaction{}
	case MsgDeclineCustodyTx:
		return &custodytypes.MsgDeclineCustodyTransaction{}
	default:
		return nil
	}
}

func SignTx(ethTxData EthTxData, ethTxBytes []byte, gwCosmosmux *runtime.ServeMux, r *http.Request) ([]byte, error) {
	// Create a new TxBuilder.
	txBuilder := config.EncodingCg.TxConfig.NewTxBuilder()

	addr1, err := hex2bech32(ethTxData.From, TypeKiraAddr)
	if err != nil {
		return nil, err
	}

	var msg cosmostypes.Msg = &tokenstypes.MsgEthereumTx{
		TxType: "NativeSend",
		Sender: addr1,
		Hash:   ethTxData.Hash,
		Data:   ethTxBytes,
	}

	var signature []byte
	if len(ethTxData.Data) >= 4 {
		txType, err := getTxType(ethTxData.Data[:4])
		if err != nil {
			return nil, err
		}
		params, err := DecodeParam(ethTxData.Data[4:], txType)
		if err != nil {
			return nil, err
		}
		if len(params) < 5 {
			return nil, errors.New("insufficient number of params")
		}

		v, r, s := params[0][len(params[0])-1:], params[1], params[2]
		signature = getSignatureV2(r, s, v)

		msg = getInstanceOfTx(txType)
		if msg == nil {
			return nil, fmt.Errorf("unrecognized transaction type: %d", txType)
		}
		err = json.Unmarshal(params[4], &msg)
		if err != nil {
			return nil, err
		}
	}

	// fmt.Println(msg)
	err = txBuilder.SetMsgs(msg)
	if err != nil {
		return nil, err
	}
	txBuilder.SetGasLimit(ethTxData.GasLimit)
	// TODO: set fee amount - how can I get the fee amount from eth tx? or fix this?
	txBuilder.SetFeeAmount(cosmostypes.NewCoins(cosmostypes.NewInt64Coin(config.DefaultKiraDenom, 200)))
	// txBuilder.SetMemo()
	// txBuilder.SetTimeoutHeight()
	// txHash, err := sendTx(context.Background(), byteData)

	privs := []cryptotypes.PrivKey{config.Config.PrivKey}
	accSeqs := []uint64{ethTxData.Nonce} // The accounts' sequence numbers

	// First round: gather all the signer infos
	var sigsV2 []signing.SignatureV2
	for i, priv := range privs {
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  config.EncodingCg.TxConfig.SignModeHandler().DefaultMode(),
				Signature: signature,
			},
			Sequence: accSeqs[i],
		}

		sigsV2 = append(sigsV2, sigV2)
	}

	err = txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	txBytes, err := config.EncodingCg.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}

	return txBytes, err
}

func sendCosmosTx(ctx context.Context, txBytes []byte) (string, error) {
	// --snip--

	// Create a connection to the gRPC server.
	if grpcConn == nil {
		var err error

		grpcConn, err = grpc.Dial(
			"127.0.0.1:9090",    // Or your gRPC server address.
			grpc.WithInsecure(), // The Cosmos SDK doesn't support any transport security mechanism.
		)

		if err != nil {
			return "", err
		}

		// defer grpcConn.Close()
	}

	// Broadcast the tx via gRPC. We create a new client for the Protobuf Tx
	// service.
	txClient := tx.NewServiceClient(grpcConn)
	// We then call the BroadcastTx method on this client.
	grpcRes, err := txClient.BroadcastTx(
		ctx,
		&tx.BroadcastTxRequest{
			Mode:    tx.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes, // Proto-binary of the signed tx, see previous step.
		},
	)

	if err != nil {
		return "", err
	}

	// fmt.Println(grpcRes.TxResponse)

	if grpcRes.TxResponse.Code != 0 {
		return "", errors.New(fmt.Sprintln("send tx failed - result code: ", grpcRes.TxResponse.Code, grpcRes.TxResponse.RawLog))
	}

	return grpcRes.TxResponse.TxHash, nil
}
