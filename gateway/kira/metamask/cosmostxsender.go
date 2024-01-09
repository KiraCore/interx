package metamask

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net/http"

	sdkmath "cosmossdk.io/math"
	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	tx "github.com/KiraCore/interx/proto-gen/cosmos/tx/v1beta1"
	custodytypes "github.com/KiraCore/sekai/x/custody/types"
	evidencetypes "github.com/KiraCore/sekai/x/evidence/types"
	customgovtypes "github.com/KiraCore/sekai/x/gov/types"
	multistakingtypes "github.com/KiraCore/sekai/x/multistaking/types"
	slashingtypes "github.com/KiraCore/sekai/x/slashing/types"
	spendingtypes "github.com/KiraCore/sekai/x/spending/types"
	stakingtypes "github.com/KiraCore/sekai/x/staking/types"
	tokenstypes "github.com/KiraCore/sekai/x/tokens/types"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	cosmostypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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
	MsgDeleteIdentityRecord
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
	Amounts cosmostypes.Coins `json:"amounts"`
	To      string            `json:"to"`
}

type CustodianParam struct {
	NewAddrs []cosmostypes.AccAddress `json:"new_addrs"`
	OldKey   string                   `json:"old_key"`
	NewKey   string                   `json:"new_key"`
	Next     string                   `json:"next"`
	Target   string                   `json:"target"`
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

	var txBytes []byte

	submitEvidencePrefix, _ := hex.DecodeString("85db2453")
	submitProposalPrefix, _ := hex.DecodeString("00000000")
	voteProposalPrefix, _ := hex.DecodeString("7f1f06dc")
	registerIdentityRecordsPrefix, _ := hex.DecodeString("bc05f106")
	deleteIdentityRecordPrefix, _ := hex.DecodeString("25581f17")
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
	createSpendingPoolPrefix, _ := hex.DecodeString("00000000")
	depositSpendingPoolPrefix, _ := hex.DecodeString("e10c925c")
	registerSpendingPoolBeneficiaryPrefix, _ := hex.DecodeString("7ab7eecf")
	claimSpendingPoolPrefix, _ := hex.DecodeString("efeed4a0")
	upsertStakingPoolPrefix, _ := hex.DecodeString("fb24f5cc")
	delegatePrefix, _ := hex.DecodeString("4b193c09")
	undelegatePrefix, _ := hex.DecodeString("94574f0c")
	claimRewardsPrefix, _ := hex.DecodeString("9e796524")
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
	case ethTxData.Data == nil:
		msgType = MsgBankSend
	case bytes.Equal(ethTxData.Data[:4], delegatePrefix):
		msgType = MsgDelegate
	case bytes.Equal(ethTxData.Data[:4], undelegatePrefix):
		msgType = MsgUndelegate
	case bytes.Equal(ethTxData.Data[:4], submitEvidencePrefix):
		msgType = MsgSubmitEvidence
	case bytes.Equal(ethTxData.Data[:4], submitProposalPrefix):
		msgType = MsgSubmitProposal
	case bytes.Equal(ethTxData.Data[:4], voteProposalPrefix):
		msgType = MsgVoteProposal
	case bytes.Equal(ethTxData.Data[:4], registerIdentityRecordsPrefix):
		msgType = MsgRegisterIdentityRecords
	case bytes.Equal(ethTxData.Data[:4], deleteIdentityRecordPrefix):
		msgType = MsgDeleteIdentityRecord
	case bytes.Equal(ethTxData.Data[:4], requestIdentityRecordsVerifyPrefix):
		msgType = MsgRequestIdentityRecordsVerify
	case bytes.Equal(ethTxData.Data[:4], handleIdentityRecordsVerifyRequestPrefix):
		msgType = MsgHandleIdentityRecordsVerifyRequest
	case bytes.Equal(ethTxData.Data[:4], cancelIdentityRecordsVerifyRequestPrefix):
		msgType = MsgCancelIdentityRecordsVerifyRequest
	case bytes.Equal(ethTxData.Data[:4], setNetworkPropertiesPrefix):
		msgType = MsgSetNetworkProperties
	case bytes.Equal(ethTxData.Data[:4], setExecutionFeePrefix):
		msgType = MsgSetExecutionFee
	case bytes.Equal(ethTxData.Data[:4], claimCouncilorPrefix):
		msgType = MsgClaimCouncilor
	case bytes.Equal(ethTxData.Data[:4], whitelistPermissionsPrefix):
		msgType = MsgWhitelistPermissions
	case bytes.Equal(ethTxData.Data[:4], blacklistPermissionsPrefix):
		msgType = MsgBlacklistPermissions
	case bytes.Equal(ethTxData.Data[:4], createRolePrefix):
		msgType = MsgCreateRole
	case bytes.Equal(ethTxData.Data[:4], assignRolePrefix):
		msgType = MsgAssignRole
	case bytes.Equal(ethTxData.Data[:4], unassignRolePrefix):
		msgType = MsgUnassignRole
	case bytes.Equal(ethTxData.Data[:4], whitelistRolePermissionPrefix):
		msgType = MsgWhitelistRolePermission
	case bytes.Equal(ethTxData.Data[:4], blacklistRolePermissionPrefix):
		msgType = MsgBlacklistRolePermission
	case bytes.Equal(ethTxData.Data[:4], removeWhitelistRolePermissionPrefix):
		msgType = MsgRemoveWhitelistRolePermission
	case bytes.Equal(ethTxData.Data[:4], removeBlacklistRolePermissionPrefix):
		msgType = MsgBlacklistRolePermission
	case bytes.Equal(ethTxData.Data[:4], claimValidatorPrefix):
		msgType = MsgClaimValidator
	case bytes.Equal(ethTxData.Data[:4], upsertTokenAliasPrefix):
		msgType = MsgUpsertTokenAlias
	case bytes.Equal(ethTxData.Data[:4], upsertTokenRatePrefix):
		msgType = MsgUpsertTokenRate
	case bytes.Equal(ethTxData.Data[:4], activatePrefix):
		msgType = MsgActivate
	case bytes.Equal(ethTxData.Data[:4], pausePrefix):
		msgType = MsgPause
	case bytes.Equal(ethTxData.Data[:4], unpausePrefix):
		msgType = MsgUnpause
	case bytes.Equal(ethTxData.Data[:4], createSpendingPoolPrefix):
		msgType = MsgCreateSpendingPool
	case bytes.Equal(ethTxData.Data[:4], depositSpendingPoolPrefix):
		msgType = MsgDepositSpendingPool
	case bytes.Equal(ethTxData.Data[:4], registerSpendingPoolBeneficiaryPrefix):
		msgType = MsgRegisterSpendingPoolBeneficiary
	case bytes.Equal(ethTxData.Data[:4], claimSpendingPoolPrefix):
		msgType = MsgClaimSpendingPool
	case bytes.Equal(ethTxData.Data[:4], upsertStakingPoolPrefix):
		msgType = MsgUpsertStakingPool
	case bytes.Equal(ethTxData.Data[:4], claimRewardsPrefix):
		msgType = MsgClaimRewards
	case bytes.Equal(ethTxData.Data[:4], claimUndelegationPrefix):
		msgType = MsgClaimUndelegation
	case bytes.Equal(ethTxData.Data[:4], setCompoundInfoPrefix):
		msgType = MsgSetCompoundInfo
	case bytes.Equal(ethTxData.Data[:4], registerDelegatorPrefix):
		msgType = MsgRegisterDelegator
	case bytes.Equal(ethTxData.Data[:4], createCustodyPrefix):
		msgType = MsgClaimSpendingPool
	case bytes.Equal(ethTxData.Data[:4], addToCustodyWhiteListPrefix):
		msgType = MsgAddToCustodyWhiteList
	case bytes.Equal(ethTxData.Data[:4], addToCustodyCustodiansPrefix):
		msgType = MsgAddToCustodyCustodians
	case bytes.Equal(ethTxData.Data[:4], removeFromCustodyCustodiansPrefix):
		msgType = MsgRemoveFromCustodyCustodians
	case bytes.Equal(ethTxData.Data[:4], dropCustodyCustodiansPrefix):
		msgType = MsgDropCustodyCustodians
	case bytes.Equal(ethTxData.Data[:4], removeFromCustodyWhiteListPrefix):
		msgType = MsgRemoveFromCustodyWhiteList
	case bytes.Equal(ethTxData.Data[:4], dropCustodyWhiteListPrefix):
		msgType = MsgDropCustodyWhiteList
	case bytes.Equal(ethTxData.Data[:4], approveCustodyTxPrefix):
		msgType = MsgApproveCustodyTx
	case bytes.Equal(ethTxData.Data[:4], declineCustodyTxPrefix):
		msgType = MsgDeclineCustodyTx
	default:
		return "", errors.New("no such functions")
	}

	txBytes, err = SignTx(ethTxData, gwCosmosmux, r, msgType)

	if err != nil {
		return "", err
	}

	// if ethTxData

	txHash, err := sendCosmosTx(r.Context(), txBytes)
	if err != nil {
		return "", err
	}

	return txHash, nil
}

func getStructHash(txType int, valParam string) ethcommon.Hash {
	abiPack := abi.ABI{}
	var structABIPack []byte

	var funcSig []byte

	switch txType {
	case MsgSubmitEvidence:
		funcSig = crypto.Keccak256([]byte("submitEvidence(string param)"))
	case MsgSubmitProposal:
		funcSig = crypto.Keccak256([]byte("submitProposal(string param)"))
	case MsgVoteProposal:
		funcSig = crypto.Keccak256([]byte("voteProposal(string param)"))
	case MsgRegisterIdentityRecords:
		funcSig = crypto.Keccak256([]byte("registerIdentityRecords(string param)"))
	case MsgDeleteIdentityRecord:
		funcSig = crypto.Keccak256([]byte("deleteIdentityRecord(string param)"))
	case MsgRequestIdentityRecordsVerify:
		funcSig = crypto.Keccak256([]byte("requestIdentityRecordsVerify(string param)"))
	case MsgHandleIdentityRecordsVerifyRequest:
		funcSig = crypto.Keccak256([]byte("handleIdentityRecordsVerifyRequest(string param)"))
	case MsgCancelIdentityRecordsVerifyRequest:
		funcSig = crypto.Keccak256([]byte("cancelIdentityRecordsVerifyRequest(string param)"))
	case MsgSetNetworkProperties:
		funcSig = crypto.Keccak256([]byte("setNetworkProperties(string param)"))
	case MsgSetExecutionFee:
		funcSig = crypto.Keccak256([]byte("setExecutionFee(string param)"))
	case MsgClaimCouncilor:
		funcSig = crypto.Keccak256([]byte("claimCouncilor(string param)"))
	case MsgWhitelistPermissions:
		funcSig = crypto.Keccak256([]byte("whitelistPermissions(string param)"))
	case MsgBlacklistPermissions:
		funcSig = crypto.Keccak256([]byte("blacklistPermissions(string param)"))
	case MsgCreateRole:
		funcSig = crypto.Keccak256([]byte("createRole(string param)"))
	case MsgAssignRole:
		funcSig = crypto.Keccak256([]byte("assignRole(string param)"))
	case MsgUnassignRole:
		funcSig = crypto.Keccak256([]byte("unassignRole(string param)"))
	case MsgWhitelistRolePermission:
		funcSig = crypto.Keccak256([]byte("whitelistRolePermission(string param)"))
	case MsgBlacklistRolePermission:
		funcSig = crypto.Keccak256([]byte("blacklistRolePermission(string param)"))
	case MsgRemoveWhitelistRolePermission:
		funcSig = crypto.Keccak256([]byte("removeWhitelistRolePermission(string param)"))
	case MsgRemoveBlacklistRolePermission:
		funcSig = crypto.Keccak256([]byte("removeBlacklistRolePermission(string param)"))
	case MsgClaimValidator:
		// funcSig = crypto.Keccak256([]byte("delegate(string param)"))
	case MsgUpsertTokenAlias:
		funcSig = crypto.Keccak256([]byte("upsertTokenAlias(string param)"))
	case MsgUpsertTokenRate:
		funcSig = crypto.Keccak256([]byte("upsertTokenRate(string param)"))
	case MsgActivate:
		funcSig = crypto.Keccak256([]byte("activate()"))
	case MsgPause:
		funcSig = crypto.Keccak256([]byte("pause()"))
	case MsgUnpause:
		funcSig = crypto.Keccak256([]byte("unpause()"))
	case MsgCreateSpendingPool:
		// funcSig = crypto.Keccak256([]byte("delegate(string param)"))
	case MsgDepositSpendingPool:
		funcSig = crypto.Keccak256([]byte("depositSpendingPool(string param)"))
	case MsgRegisterSpendingPoolBeneficiary:
		funcSig = crypto.Keccak256([]byte("registerSpendingPoolBeneficiary(string param)"))
	case MsgClaimSpendingPool:
		funcSig = crypto.Keccak256([]byte("claimSpendingPool(string param)"))
	case MsgUpsertStakingPool:
		funcSig = crypto.Keccak256([]byte("upsertStakingPool(string param)"))
	case MsgDelegate:
		funcSig = crypto.Keccak256([]byte("delegate(string param)"))
	case MsgUndelegate:
		funcSig = crypto.Keccak256([]byte("undelegate(string param)"))
	case MsgClaimRewards:
		funcSig = crypto.Keccak256([]byte("claimRewards()"))
	case MsgClaimUndelegation:
		funcSig = crypto.Keccak256([]byte("claimUndelegation(string param)"))
	case MsgSetCompoundInfo:
		funcSig = crypto.Keccak256([]byte("setCompoundInfo(string param)"))
	case MsgRegisterDelegator:
		funcSig = crypto.Keccak256([]byte("registerDelegator()"))
	case MsgCreateCustody:
		funcSig = crypto.Keccak256([]byte("createCustody(string param)"))
	case MsgAddToCustodyWhiteList:
		funcSig = crypto.Keccak256([]byte("addToCustodyWhiteList(string param)"))
	case MsgAddToCustodyCustodians:
		funcSig = crypto.Keccak256([]byte("addToCustodyCustodians(string param)"))
	case MsgRemoveFromCustodyCustodians:
		funcSig = crypto.Keccak256([]byte("removeFromCustodyCustodians(string param)"))
	case MsgDropCustodyCustodians:
		funcSig = crypto.Keccak256([]byte("dropCustodyCustodians(string param)"))
	case MsgRemoveFromCustodyWhiteList:
		funcSig = crypto.Keccak256([]byte("removeFromCustodyWhiteList(string param)"))
	case MsgDropCustodyWhiteList:
		funcSig = crypto.Keccak256([]byte("dropCustodyWhiteList(string param)"))
	case MsgApproveCustodyTx:
		funcSig = crypto.Keccak256([]byte("approveCustodyTransaction(string param)"))
	case MsgDeclineCustodyTx:
		funcSig = crypto.Keccak256([]byte("declineCustodyTransaction(string param)"))
	default:
	}

	structJsonData := []byte(`[{"Type":"function","Name":"encode","Inputs":[{"Type":"bytes32","Name":"funcsig"},{"Type":"bytes32","Name":"param"}],"Outputs":[]}]`)
	_ = abiPack.UnmarshalJSON(structJsonData)

	var funcSignature [32]byte
	copy(funcSignature[:], funcSig)

	structABIPack, _ = PackABIParams(abiPack,
		"encode",
		convertByteArr2Bytes32(funcSig),
		convertByteArr2Bytes32(crypto.Keccak256([]byte(valParam))),
	)

	structHash := crypto.Keccak256Hash(structABIPack)

	return structHash
}

func ValidateEIP712Sign(v, r, s []byte, sender ethcommon.Address, valParam string, txType int) bool {
	abiPack := abi.ABI{}

	// get EIP712DomainHash
	jsonData := []byte(`[{"Type":"function","Name":"encode","Inputs":[{"Type":"bytes32","Name":"funcsig"},{"Type":"bytes32","Name":"name"},{"Type":"bytes32","Name":"version"},{"Type":"bytes32","Name":"chainId"}],"Outputs":[]}]`)
	abiPack.UnmarshalJSON(jsonData)
	funcSig := crypto.Keccak256([]byte("EIP712Domain(string name,string version,uint256 chainId)"))

	eip712DomainSeparatorABIPack, _ := PackABIParams(abiPack,
		"encode",
		convertByteArr2Bytes32(funcSig),
		convertByteArr2Bytes32(crypto.Keccak256([]byte("Kira"))),
		convertByteArr2Bytes32(crypto.Keccak256([]byte("1"))),
		convertByteArr2Bytes32(uint32To32Bytes(8789)),
	)
	eip712DomainSeparator := crypto.Keccak256Hash(eip712DomainSeparatorABIPack)

	// get StructHash
	structHash := getStructHash(txType, valParam)

	// Define the final hash
	hash := crypto.Keccak256Hash(
		append(append([]byte("\x19\x01"), eip712DomainSeparator.Bytes()...), structHash.Bytes()...),
	)

	signature := getSignature(r, s, v)

	// Recover the public key from the signature
	pubKey, err := crypto.SigToPub(hash.Bytes(), signature)
	// pbBytes, err := crypto.Ecrecover(hash.Bytes(), signature)
	// fmt.Println(string(pbBytes), err)
	// pubKey, err := crypto.UnmarshalPubkey(pbBytes)

	if err != nil {
		fmt.Println("eip712 err", err)
		return false
	}

	// Derive the signer's address from the public key
	signerAddress := crypto.PubkeyToAddress(*pubKey)

	return signerAddress.Hex() == sender.Hex()
}

func SignTx(ethTxData EthTxData, gwCosmosmux *runtime.ServeMux, r *http.Request, txType int) ([]byte, error) {
	// Create a new TxBuilder.
	txBuilder := config.EncodingCg.TxConfig.NewTxBuilder()

	addr1, err := hex2bech32(ethTxData.From, TypeKiraAddr)
	if err != nil {
		return nil, err
	}
	cosmosAddr1, err := cosmostypes.AccAddressFromBech32(addr1)
	if err != nil {
		return nil, err
	}

	params, err := DecodeParam(ethTxData.Data[4:], txType)
	if err != nil {
		return nil, err
	}

	if txType != MsgBankSend {
		valParam := string(params[4])

		validation := ValidateEIP712Sign(params[0][len(params[0])-1:], params[1], params[2], ethcommon.BytesToAddress(params[3][12:]), valParam, txType)
		if !validation {
			return nil, errors.New("eip712 validation is failed")
		}
	}

	var msg cosmostypes.Msg

	switch txType {
	case MsgBankSend:
		addr2, err := hex2bech32(ethTxData.To, TypeKiraAddr)
		if err != nil {
			return nil, err
		}
		cosmosAddr2, err := cosmostypes.AccAddressFromBech32(addr2)
		if err != nil {
			return nil, err
		}

		balance := ethTxData.Value.Div(&ethTxData.Value, big.NewInt(int64(math.Pow10(12))))
		msg = banktypes.NewMsgSend(cosmosAddr1, cosmosAddr2,
			cosmostypes.NewCoins(cosmostypes.NewInt64Coin(config.DefaultKiraDenom, balance.Int64())))
	case MsgSubmitEvidence:
		// V, R, S, signer, height, power, time, consensusAddr
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		var evidenceParam evidencetypes.Equivocation
		err = json.Unmarshal(params[4], &evidenceParam)
		if err != nil {
			return nil, err
		}

		msg, err = evidencetypes.NewMsgSubmitEvidence(from, &evidenceParam)
		if err != nil {
			return nil, err
		}
	case MsgSubmitProposal:
		// from, err := bytes2cosmosAddr(params[3][12:])
		// if err != nil {
		// 	return nil, err
		// }

		// type SubmitProposalParam struct {
		// 	Title       string `json:"title"`
		// 	Description string `json:"description"`
		// 	Content     string `json:"content"`
		// }

		// var proposalParam SubmitProposalParam
		// err = json.Unmarshal(params[4], &proposalParam)
		// if err != nil {
		// 	return nil, err
		// }

		// msg, err := customgovtypes.NewMsgSubmitProposal(from)
		// if err != nil {
		// 	return nil, err
		// }
	case MsgVoteProposal:
		// V, R, S, signer, param
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		type VoteProposalParam struct {
			ProposalID uint64 `json:"proposal_id"`
			Option     uint64 `json:"option"`
			Slash      string `json:"slash"`
		}

		var proposalParam VoteProposalParam
		err = json.Unmarshal(params[4], &proposalParam)
		if err != nil {
			return nil, err
		}

		slash, err := sdkmath.LegacyNewDecFromStr(proposalParam.Slash)
		if err != nil {
			return nil, err
		}

		msg = customgovtypes.NewMsgVoteProposal(proposalParam.ProposalID, from, customgovtypes.VoteOption(proposalParam.Option), slash)
	case MsgRegisterIdentityRecords:
		// V, R, S, signer, identityInfo,
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		type IdentityRecordParam struct {
			Infos []customgovtypes.IdentityInfoEntry `json:"identity_infos"`
		}

		var recordParam IdentityRecordParam
		err = json.Unmarshal(params[4], &recordParam)
		if err != nil {
			return nil, err
		}

		msg = customgovtypes.NewMsgRegisterIdentityRecords(from, recordParam.Infos)
	case MsgDeleteIdentityRecord:
		// V, R, S, signer, len, keys,
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		type IdentityRecordParam struct {
			Keys []string `json:"keys"`
		}

		var recordParam IdentityRecordParam
		err = json.Unmarshal(params[4], &recordParam)
		if err != nil {
			return nil, err
		}

		msg = customgovtypes.NewMsgDeleteIdentityRecords(from, recordParam.Keys)
	case MsgRequestIdentityRecordsVerify:
		// V, R, S, signer, tip, verifier, len, recordIds
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		type IdentityRecordParam struct {
			Balance   string   `json:"balance"`
			Verifier  string   `json:"verifier"`
			RecordIds []uint64 `json:"record_ids"`
		}

		var recordParam IdentityRecordParam
		err = json.Unmarshal(params[4], &recordParam)
		if err != nil {
			return nil, err
		}

		balance, err := cosmostypes.ParseCoinNormalized(recordParam.Balance)
		if err != nil {
			return nil, err
		}

		verifier, err := string2cosmosAddr(recordParam.Verifier)
		if err != nil {
			return nil, err
		}

		msg = customgovtypes.NewMsgRequestIdentityRecordsVerify(from, verifier, recordParam.RecordIds, balance)
	case MsgHandleIdentityRecordsVerifyRequest:
		// V, R, S, signer, requestId, isApprove
		verifier, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		type IdentityRecordParam struct {
			RequestID  uint64 `json:"request_id"`
			IsApproved bool   `json:"is_approved"`
		}

		var recordParam IdentityRecordParam
		err = json.Unmarshal(params[4], &recordParam)
		if err != nil {
			return nil, err
		}

		msg = customgovtypes.NewMsgHandleIdentityRecordsVerifyRequest(verifier, recordParam.RequestID, recordParam.IsApproved)
	case MsgCancelIdentityRecordsVerifyRequest:
		// V, R, S, signer, verifyRequestId
		executor, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		type IdentityRecordParam struct {
			VerifyRequestId uint64 `json:"verify_request_id"`
		}

		var recordParam IdentityRecordParam
		err = json.Unmarshal(params[4], &recordParam)
		if err != nil {
			return nil, err
		}

		msg = customgovtypes.NewMsgCancelIdentityRecordsVerifyRequest(executor, recordParam.VerifyRequestId)
	case MsgSetNetworkProperties:
		// V, R, S, signer, networkProperties
		proposer, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		type NetworkPropertiesParam struct {
			NetworkProperties customgovtypes.NetworkProperties `json:"network_properties"`
		}

		var networkProperties NetworkPropertiesParam
		err = json.Unmarshal(params[4], &networkProperties)
		if err != nil {
			return nil, err
		}

		msg = customgovtypes.NewMsgSetNetworkProperties(proposer, &networkProperties.NetworkProperties)
	case MsgSetExecutionFee:
		// V, R, S, signer, executionFee, failureFee, timeout, defaultParams
		proposer, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		type ExecutionFeeParam struct {
			TransactionType string `json:"transaction_type"`
			ExecutionFee    uint64 `json:"execution_fee"`
			FailureFee      uint64 `json:"failure_fee"`
			Timeout         uint64 `json:"timeout"`
			DefaultParams   uint64 `json:"default_params"`
		}

		var feeParam ExecutionFeeParam
		err = json.Unmarshal(params[4], &feeParam)
		if err != nil {
			return nil, err
		}

		msg = customgovtypes.NewMsgSetExecutionFee(feeParam.TransactionType, feeParam.ExecutionFee, feeParam.FailureFee,
			feeParam.Timeout, feeParam.DefaultParams, proposer)
	case MsgClaimCouncilor:
		// V, R, S, signer, moniker string, username string, description string, social string, contact string, avatar string
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		type ClaimCouncilorParam struct {
			Moniker     string `json:"moniker"`
			Username    string `json:"username"`
			Description string `json:"description"`
			Social      string `json:"social"`
			Contact     string `json:"contact"`
			Avatar      string `json:"avatar"`
		}

		var councilorParam ClaimCouncilorParam
		err = json.Unmarshal(params[4], &councilorParam)
		if err != nil {
			return nil, err
		}

		msg = customgovtypes.NewMsgClaimCouncilor(sender, councilorParam.Moniker, councilorParam.Username, councilorParam.Description,
			councilorParam.Social, councilorParam.Contact, councilorParam.Avatar)
	case MsgWhitelistPermissions:
		// V, R, S, signer, permission uint256, controlledAddr string
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		var permissionParam PermissionsParam
		err = json.Unmarshal(params[4], &permissionParam)
		if err != nil {
			return nil, err
		}

		controlledAddr, err := string2cosmosAddr(permissionParam.ControlledAddr)
		if err != nil {
			return nil, err
		}

		msg = customgovtypes.NewMsgWhitelistPermissions(sender, controlledAddr, permissionParam.Permission)
	case MsgBlacklistPermissions:
		// V, R, S, signer, permission uint256, controlledAddr string
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		var permissionParam PermissionsParam
		err = json.Unmarshal(params[4], &permissionParam)
		if err != nil {
			return nil, err
		}

		controlledAddr, err := string2cosmosAddr(permissionParam.ControlledAddr)
		if err != nil {
			return nil, err
		}

		msg = customgovtypes.NewMsgBlacklistPermissions(sender, controlledAddr, permissionParam.Permission)
	case MsgCreateRole:
		// V, R, S, signer, sid string, description string
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		type RoleParam struct {
			Sid         string `json:"sid"`
			Description string `json:"description"`
		}

		var roleParam RoleParam
		err = json.Unmarshal(params[4], &roleParam)
		if err != nil {
			return nil, err
		}

		msg = customgovtypes.NewMsgCreateRole(sender, roleParam.Sid, roleParam.Description)
	case MsgAssignRole:
		// V, R, S, signer, roleid uint32, controller string
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		var roleParam RoleParam
		err = json.Unmarshal(params[4], &roleParam)
		if err != nil {
			return nil, err
		}

		controller, err := string2cosmosAddr(roleParam.Controller)
		if err != nil {
			return nil, err
		}

		msg = customgovtypes.NewMsgAssignRole(sender, controller, roleParam.RoleId)
	case MsgUnassignRole:
		// V, R, S, signer, roleid uint32, controller address
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		var roleParam RoleParam
		err = json.Unmarshal(params[4], &roleParam)
		if err != nil {
			return nil, err
		}

		controller, err := string2cosmosAddr(roleParam.Controller)
		if err != nil {
			return nil, err
		}

		msg = customgovtypes.NewMsgUnassignRole(sender, controller, roleParam.RoleId)
	case MsgWhitelistRolePermission:
		// V, R, S, signer, permission uint32, roleIdentifier string
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		var roleParam ListRoleParam
		err = json.Unmarshal(params[4], &roleParam)
		if err != nil {
			return nil, err
		}

		msg = customgovtypes.NewMsgWhitelistRolePermission(sender, roleParam.RoleIdentifier, roleParam.Permission)
	case MsgBlacklistRolePermission:
		// V, R, S, signer, permission uint32, roleIdentifier string
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		var roleParam ListRoleParam
		err = json.Unmarshal(params[4], &roleParam)
		if err != nil {
			return nil, err
		}

		msg = customgovtypes.NewMsgBlacklistRolePermission(sender, roleParam.RoleIdentifier, roleParam.Permission)
	case MsgRemoveWhitelistRolePermission:
		// V, R, S, signer, permission uint32, roleIdentifier string
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		var roleParam ListRoleParam
		err = json.Unmarshal(params[4], &roleParam)
		if err != nil {
			return nil, err
		}

		msg = customgovtypes.NewMsgRemoveWhitelistRolePermission(sender, roleParam.RoleIdentifier, roleParam.Permission)
	case MsgRemoveBlacklistRolePermission:
		// V, R, S, signer, permission uint32, roleIdentifier string
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		var roleParam ListRoleParam
		err = json.Unmarshal(params[4], &roleParam)
		if err != nil {
			return nil, err
		}

		msg = customgovtypes.NewMsgRemoveBlacklistRolePermission(sender, roleParam.RoleIdentifier, roleParam.Permission)
	case MsgClaimValidator:
		// V, R, S, signer, moniker string, valKey cosmostypes.ValAddress, pubKey cryptotypes.PubKey

		type ClaimParam struct {
			Moniker string             `json:"moniker"`
			ValKey  string             `json:"val_key"`
			PubKey  cryptotypes.PubKey `json:"pub_key"`
		}

		var claimParam ClaimParam
		err = json.Unmarshal(params[4], &claimParam)
		if err != nil {
			return nil, err
		}

		valKey, err := cosmostypes.ValAddressFromBech32(claimParam.ValKey)
		if err != nil {
			return nil, err
		}

		msg, err = stakingtypes.NewMsgClaimValidator(claimParam.Moniker, valKey, claimParam.PubKey)
		if err != nil {
			return nil, err
		}
	case MsgUpsertTokenAlias:
		// V, R, S, signer, param
		proposer, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		type UpsertTokenAliasParam struct {
			Symbol      string   `json:"symbol"`
			Name        string   `json:"name"`
			Icon        string   `json:"icon"`
			Decimals    uint32   `json:"decimals"`
			Denoms      []string `json:"denoms"`
			Invalidated bool     `json:"invalidated"`
		}

		var upsertTokenAliasParam UpsertTokenAliasParam
		err = json.Unmarshal(params[4], &upsertTokenAliasParam)
		if err != nil {
			return nil, err
		}

		msg = tokenstypes.NewMsgUpsertTokenAlias(proposer, upsertTokenAliasParam.Symbol, upsertTokenAliasParam.Name,
			upsertTokenAliasParam.Icon, upsertTokenAliasParam.Decimals, upsertTokenAliasParam.Denoms, upsertTokenAliasParam.Invalidated)
	case MsgUpsertTokenRate:
		// V, R, S, signer, feePayments bool, stakeToken bool, invalidated bool, denom string, rate cosmostypes.Dec,
		// stakeCap cosmostypes.Dec, stakeMin cosmostypes.Int
		proposer, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		type UpsertTokenRateParam struct {
			Denom         string          `json:"denom"`
			Rate          cosmostypes.Dec `json:"rate"`
			IsFeePayments bool            `json:"is_fee_payments"`
			StakeCap      cosmostypes.Dec `json:"stake_cap"`
			StakeMin      cosmostypes.Int `json:"stake_min"`
			IsStakeToken  bool            `json:"is_stake_token"`
			Invalidated   bool            `json:"invalidated"`
		}

		var upsertParam UpsertTokenRateParam
		err = json.Unmarshal(params[4], &upsertParam)
		if err != nil {
			return nil, err
		}

		msg = tokenstypes.NewMsgUpsertTokenRate(proposer, upsertParam.Denom, upsertParam.Rate, upsertParam.IsFeePayments,
			upsertParam.StakeCap, upsertParam.StakeMin, upsertParam.IsStakeToken, upsertParam.Invalidated)
	case MsgActivate:
		// V, R, S, signer
		validator, err := bytes2cosmosValAddr(params[3][12:])
		if err != nil {
			return nil, err
		}
		msg = slashingtypes.NewMsgActivate(validator)
	case MsgPause:
		// V, R, S, signer
		validator, err := bytes2cosmosValAddr(params[3][12:])
		if err != nil {
			return nil, err
		}
		msg = slashingtypes.NewMsgPause(validator)
	case MsgUnpause:
		// V, R, S, signer
		validator, err := bytes2cosmosValAddr(params[3][12:])
		if err != nil {
			return nil, err
		}
		msg = slashingtypes.NewMsgUnpause(validator)
	case MsgCreateSpendingPool:
		// V, R, S, signer, name string, claimStart uint64, claimEnd uint64, rates cosmostypes.DecCoins, voteQuorum uint64,
		// votePeriod uint64, voteEnactment uint64, owners spendingtypes.PermInfo, beneficiaries spendingtypes.WeightedPermInfo,
		// sender cosmostypes.AccAddress, dynamicRate bool, dynamicRatePeriod uint64
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		type SpendingPoolParam struct {
			Name              string                         `json:"name"`
			ClaimStart        uint64                         `json:"claim_start"`
			ClaimEnd          uint64                         `json:"claim_end"`
			Rates             cosmostypes.DecCoins           `json:"rates"`
			VoteQuorum        uint64                         `json:"vote_quorum"`
			VotePeriod        uint64                         `json:"vote_period"`
			VoteEnactment     uint64                         `json:"vote_enactment"`
			Owners            spendingtypes.PermInfo         `json:"owners"`
			Beneficiaries     spendingtypes.WeightedPermInfo `json:"beneficiaries"`
			IsDynamicRate     bool                           `json:"is_dynamic_rate"`
			DynamicRatePeriod uint64                         `json:"dynamic_rate_period"`
		}

		var poolParam SpendingPoolParam
		err = json.Unmarshal(params[4], &poolParam)
		if err != nil {
			return nil, err
		}

		msg = spendingtypes.NewMsgCreateSpendingPool(poolParam.Name, poolParam.ClaimStart, poolParam.ClaimEnd, poolParam.Rates,
			poolParam.VoteQuorum, poolParam.VotePeriod, poolParam.VoteEnactment, poolParam.Owners, poolParam.Beneficiaries,
			sender, poolParam.IsDynamicRate, poolParam.DynamicRatePeriod)
	case MsgDepositSpendingPool:
		// V, R, S, signer, amount string, name string
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		type SpendingPoolParam struct {
			Name   string            `json:"Name"`
			Amount cosmostypes.Coins `json:"Amount"`
		}

		var poolParam SpendingPoolParam
		err = json.Unmarshal(params[4], &poolParam)
		if err != nil {
			return nil, err
		}

		msg = spendingtypes.NewMsgDepositSpendingPool(poolParam.Name, poolParam.Amount, sender)
	case MsgRegisterSpendingPoolBeneficiary:
		// V, R, S, signer, name string
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		type SpendingPoolParam struct {
			Name string `json:"Name"`
		}

		var poolParam SpendingPoolParam
		err = json.Unmarshal(params[4], &poolParam)
		if err != nil {
			return nil, err
		}

		msg = spendingtypes.NewMsgRegisterSpendingPoolBeneficiary(poolParam.Name, sender)
	case MsgClaimSpendingPool:
		// V, R, S, signer, name string
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		type SpendingPoolParam struct {
			Name string `json:"Name"`
		}

		var poolParam SpendingPoolParam
		err = json.Unmarshal(params[4], &poolParam)
		if err != nil {
			return nil, err
		}

		msg = spendingtypes.NewMsgClaimSpendingPool(poolParam.Name, sender)
	case MsgUpsertStakingPool:
		// V, R, S, signer, enabled bool, validator string, commission cosmostypes.Dec
		from, err := hex2bech32(hex.EncodeToString(params[3][12:]), TypeKiraAddr)
		if err != nil {
			return nil, err
		}

		type StakingPoolParam struct {
			Validator  string          `json:"validator"`
			Enabled    bool            `json:"enabled"`
			Commission cosmostypes.Dec `json:"commission"`
		}

		var poolParam StakingPoolParam
		err = json.Unmarshal(params[4], &poolParam)
		if err != nil {
			return nil, err
		}

		msg = multistakingtypes.NewMsgUpsertStakingPool(from, poolParam.Validator, poolParam.Enabled, poolParam.Commission)
	case MsgDelegate:
		// V, R, S, signer, param
		from, err := hex2bech32(hex.EncodeToString(params[3][12:]), TypeKiraAddr)
		if err != nil {
			return nil, err
		}

		var delegateParam DelegateParam
		err = json.Unmarshal(params[4], &delegateParam)
		if err != nil {
			return nil, err
		}

		msg = multistakingtypes.NewMsgDelegate(from, delegateParam.To, delegateParam.Amounts)
	case MsgUndelegate:
		// V, R, S, signer, amount, validator
		from, err := hex2bech32(hex.EncodeToString(params[3][12:]), TypeKiraAddr)
		if err != nil {
			return nil, err
		}

		var delegateParam DelegateParam
		err = json.Unmarshal(params[4], &delegateParam)
		if err != nil {
			return nil, err
		}

		msg = multistakingtypes.NewMsgUndelegate(from, delegateParam.To, delegateParam.Amounts)
	case MsgClaimRewards:
		// V, R, S, signer
		from, err := hex2bech32(hex.EncodeToString(params[3][12:]), TypeKiraAddr)
		if err != nil {
			return nil, err
		}
		msg = multistakingtypes.NewMsgClaimRewards(from)
	case MsgClaimUndelegation:
		// V, R, S, signer, undelegationId uint64
		from, err := hex2bech32(hex.EncodeToString(params[3][12:]), TypeKiraAddr)
		if err != nil {
			return nil, err
		}

		type UndelegationParam struct {
			UndelegationId uint64 `json:"undelegation_id"`
		}

		var undelegationParam UndelegationParam
		err = json.Unmarshal(params[4], &undelegationParam)
		if err != nil {
			return nil, err
		}

		msg = multistakingtypes.NewMsgClaimUndelegation(from, undelegationParam.UndelegationId)
	case MsgSetCompoundInfo:
		// V, R, S, signer, allDenom bool, len, denoms []string
		from, err := hex2bech32(hex.EncodeToString(params[3][12:]), TypeKiraAddr)
		if err != nil {
			return nil, err
		}

		type UndelegationParam struct {
			IsAllDenom bool     `json:"is_all_denom"`
			Denoms     []string `json:"denoms"`
		}

		var undelegationParam UndelegationParam
		err = json.Unmarshal(params[4], &undelegationParam)
		if err != nil {
			return nil, err
		}

		msg = multistakingtypes.NewMsgSetCompoundInfo(from, undelegationParam.IsAllDenom, undelegationParam.Denoms)
	case MsgRegisterDelegator:
		// V, R, S, signer
		from, err := hex2bech32(hex.EncodeToString(params[3][12:]), TypeKiraAddr)
		if err != nil {
			return nil, err
		}

		msg = multistakingtypes.NewMsgRegisterDelegator(from)
	case MsgCreateCustody:
		// V, R, S, signer, custodyMode uint256, key string, nextController string, len, boolArr, len, strArr
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		type CustodyParam struct {
			CustodySettings custodytypes.CustodySettings `json:"custody_settings"`
			OldKey          string                       `json:"old_key"`
			NewKey          string                       `json:"new_key"`
			Next            string                       `json:"next"`
			Target          string                       `json:"target"`
		}

		var custodyParam CustodyParam
		err = json.Unmarshal(params[4], &custodyParam)
		if err != nil {
			return nil, err
		}

		msg = custodytypes.NewMsgCreateCustody(from, custodyParam.CustodySettings, custodyParam.OldKey, custodyParam.NewKey,
			custodyParam.Next, custodyParam.Target)
	case MsgAddToCustodyWhiteList:
		// V, R, S, signer, oldKey string, newKey string, next string, target string
		// len, newAddr []string,
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		var custodyParam CustodianParam
		err = json.Unmarshal(params[4], &custodyParam)
		if err != nil {
			return nil, err
		}

		msg = custodytypes.NewMsgAddToCustodyWhiteList(from, custodyParam.NewAddrs, custodyParam.OldKey, custodyParam.NewKey,
			custodyParam.Next, custodyParam.Target)
	case MsgAddToCustodyCustodians:
		// V, R, S, signer, oldKey string, newKey string, next string, target string
		// len, newAddr []string
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		var custodyParam CustodianParam
		err = json.Unmarshal(params[4], &custodyParam)
		if err != nil {
			return nil, err
		}

		msg = custodytypes.NewMsgAddToCustodyCustodians(from, custodyParam.NewAddrs, custodyParam.OldKey, custodyParam.NewKey,
			custodyParam.Next, custodyParam.Target)
	case MsgRemoveFromCustodyCustodians:
		// V, R, S, signer, newAddr cosmostypes.AccAddress,
		// oldKey string, newKey string, next string, target string
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		type CustodianParam struct {
			NewAddr cosmostypes.AccAddress `json:"new_addr"`
			OldKey  string                 `json:"old_key"`
			NewKey  string                 `json:"new_key"`
			Next    string                 `json:"next"`
			Target  string                 `json:"target"`
		}

		var custodyParam CustodianParam
		err = json.Unmarshal(params[4], &custodyParam)
		if err != nil {
			return nil, err
		}

		msg = custodytypes.NewMsgRemoveFromCustodyCustodians(from, custodyParam.NewAddr, custodyParam.OldKey, custodyParam.NewKey,
			custodyParam.Next, custodyParam.Target)
	case MsgDropCustodyCustodians:
		// V, R, S, signer
		// oldKey string, newKey string, next string, target string
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		var custodyParam CustodianParam
		err = json.Unmarshal(params[4], &custodyParam)
		if err != nil {
			return nil, err
		}

		msg = custodytypes.NewMsgDropCustodyCustodians(from, custodyParam.OldKey, custodyParam.NewKey,
			custodyParam.Next, custodyParam.Target)
	case MsgRemoveFromCustodyWhiteList:
		// V, R, S, signer, newAddr string,
		// oldKey string, newKey string, next string, target string
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		type CustodianParam struct {
			NewAddr cosmostypes.AccAddress `json:"new_addr"`
			OldKey  string                 `json:"old_key"`
			NewKey  string                 `json:"new_key"`
			Next    string                 `json:"next"`
			Target  string                 `json:"target"`
		}

		var custodyParam CustodianParam
		err = json.Unmarshal(params[4], &custodyParam)
		if err != nil {
			return nil, err
		}

		msg = custodytypes.NewMsgRemoveFromCustodyWhiteList(from, custodyParam.NewAddr, custodyParam.OldKey, custodyParam.NewKey,
			custodyParam.Next, custodyParam.Target)
	case MsgDropCustodyWhiteList:
		// V, R, S, signer
		// oldKey string, newKey string, next string, target string
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		var custodyParam CustodianParam
		err = json.Unmarshal(params[4], &custodyParam)
		if err != nil {
			return nil, err
		}

		msg = custodytypes.NewMsgDropCustodyWhiteList(from, custodyParam.OldKey, custodyParam.NewKey,
			custodyParam.Next, custodyParam.Target)
	case MsgApproveCustodyTx:
		// V, R, S, signer, to string, hash string
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		type CustodianParam struct {
			To   cosmostypes.AccAddress `json:"to"`
			Hash string                 `json:"hash"`
		}

		var custodyParam CustodianParam
		err = json.Unmarshal(params[4], &custodyParam)
		if err != nil {
			return nil, err
		}

		msg = custodytypes.NewMsgApproveCustodyTransaction(from, custodyParam.To, custodyParam.Hash)
	case MsgDeclineCustodyTx:
		// V, R, S, signer, to string, hash string
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		type CustodianParam struct {
			To   cosmostypes.AccAddress `json:"to"`
			Hash string                 `json:"hash"`
		}

		var custodyParam CustodianParam
		err = json.Unmarshal(params[4], &custodyParam)
		if err != nil {
			return nil, err
		}

		msg = custodytypes.NewMsgDeclineCustodyTransaction(from, custodyParam.To, custodyParam.Hash)
	}

	fmt.Println(msg)
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

	accNum, _ := common.GetAccountNumberSequence(gwCosmosmux, r, addr1)

	privs := []cryptotypes.PrivKey{config.Config.PrivKey}
	accNums := []uint64{accNum}          // The accounts' account numbers
	accSeqs := []uint64{ethTxData.Nonce} // The accounts' sequence numbers

	// First round: gather all the signer infos
	var sigsV2 []signing.SignatureV2
	for i, priv := range privs {
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  config.EncodingCg.TxConfig.SignModeHandler().DefaultMode(),
				Signature: nil,
			},
			Sequence: accSeqs[i],
		}

		sigsV2 = append(sigsV2, sigV2)
	}

	err = txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	interxStatus := common.GetInterxStatus("http://127.0.0.1:11000")

	// Second round: all signer infos are set, so each signer can sign.
	sigsV2 = []signing.SignatureV2{}
	for i, priv := range privs {
		signerData := xauthsigning.SignerData{
			ChainID:       interxStatus.InterxInfo.ChainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		sigV2, err := clienttx.SignWithPrivKey(
			config.EncodingCg.TxConfig.SignModeHandler().DefaultMode(), signerData,
			txBuilder, priv, config.EncodingCg.TxConfig, accSeqs[i])
		if err != nil {
			return nil, err
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

	fmt.Println(grpcRes.TxResponse)

	if grpcRes.TxResponse.Code != 0 {
		return "", errors.New(fmt.Sprintln("send tx failed - result code: ", grpcRes.TxResponse.Code))
	}

	return grpcRes.TxResponse.TxHash, nil
}
