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
	"strconv"
	"time"

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
	Permission     uint32 `json:"permission,string"`
}

type PermissionsParam struct {
	Permission     uint32 `json:"permission,string"`
	ControlledAddr string `json:"controlled_addr"`
}

type RoleParam struct {
	RoleId     uint32 `json:"role_id,string"`
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

		type Equivocation struct {
			Height           int64     `json:"height,string"`
			Time             time.Time `json:"time"`
			Power            int64     `json:"power,string"`
			ConsensusAddress string    `json:"consensus_address"`
		}

		var evidenceJsonParam Equivocation
		err = json.Unmarshal(params[4], &evidenceJsonParam)
		if err != nil {
			return nil, err
		}

		evidenceParam := evidencetypes.Equivocation{
			Height:           evidenceJsonParam.Height,
			Time:             evidenceJsonParam.Time,
			Power:            evidenceJsonParam.Power,
			ConsensusAddress: evidenceJsonParam.ConsensusAddress,
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
			ProposalID uint64 `json:"proposal_id,string"`
			Option     uint64 `json:"option,string"`
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
			RecordIds []string `json:"record_ids"`
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

		var recordIds []uint64
		for _, idStr := range recordParam.RecordIds {
			id, err := strconv.ParseUint(idStr, 10, 64)
			if err != nil {
				return nil, err
			}
			recordIds = append(recordIds, id)
		}

		msg = customgovtypes.NewMsgRequestIdentityRecordsVerify(from, verifier, recordIds, balance)
	case MsgHandleIdentityRecordsVerifyRequest:
		// V, R, S, signer, requestId, isApprove
		verifier, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		type IdentityRecordParam struct {
			RequestID  uint64 `json:"request_id,string"`
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
			VerifyRequestId uint64 `json:"verify_request_id,string"`
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

		type NetworkProperties struct {
			MinTxFee                        uint64 `json:"min_tx_fee,string"`
			MaxTxFee                        uint64 `json:"max_tx_fee,string"`
			VoteQuorum                      uint64 `json:"vote_quorum,string"`
			MinimumProposalEndTime          uint64 `json:"minimum_proposal_end_time,string"`
			ProposalEnactmentTime           uint64 `json:"proposal_enactment_time,string"`
			MinProposalEndBlocks            uint64 `json:"min_proposal_end_blocks,string"`
			MinProposalEnactmentBlocks      uint64 `json:"min_proposal_enactment_blocks,string"`
			EnableForeignFeePayments        bool   `json:"enable_foreign_fee_payments,string"`
			MischanceRankDecreaseAmount     uint64 `json:"mischance_rank_decrease_amount,string"`
			MaxMischance                    uint64 `json:"max_mischance,string"`
			MischanceConfidence             uint64 `json:"mischance_confidence,string"`
			InactiveRankDecreasePercent     string `json:"inactive_rank_decrease_percent"`
			MinValidators                   uint64 `json:"min_validators,string"`
			PoorNetworkMaxBankSend          uint64 `json:"poor_network_max_bank_send,string"`
			UnjailMaxTime                   uint64 `json:"unjail_max_time,string"`
			EnableTokenWhitelist            bool   `json:"enable_token_whitelist,string"`
			EnableTokenBlacklist            bool   `json:"enable_token_blacklist,string"`
			MinIdentityApprovalTip          uint64 `json:"min_identity_approval_tip,string"`
			UniqueIdentityKeys              string `json:"unique_identity_keys,string"`
			UbiHardcap                      uint64 `json:"ubi_hardcap,string"`
			ValidatorsFeeShare              string `json:"validators_fee_share"`
			InflationRate                   string `json:"inflation_rate"`
			InflationPeriod                 uint64 `json:"inflation_period,string"`
			UnstakingPeriod                 uint64 `json:"unstaking_period,string"`
			MaxDelegators                   uint64 `json:"max_delegators,string"`
			MinDelegationPushout            uint64 `json:"min_delegation_pushout,string"`
			SlashingPeriod                  uint64 `json:"slashing_period,string"`
			MaxJailedPercentage             string `json:"max_jailed_percentage"`
			MaxSlashingPercentage           string `json:"max_slashing_percentage"`
			MinCustodyReward                uint64 `json:"min_custody_reward,string"`
			MaxCustodyBufferSize            uint64 `json:"max_custody_buffer_size,string"`
			MaxCustodyTxSize                uint64 `json:"max_custody_tx_size,string"`
			AbstentionRankDecreaseAmount    uint64 `json:"abstention_rank_decrease_amount,string"`
			MaxAbstention                   uint64 `json:"max_abstention,string"`
			MinCollectiveBond               uint64 `json:"min_collective_bond,string"`
			MinCollectiveBondingTime        uint64 `json:"min_collective_bonding_time,string"`
			MaxCollectiveOutputs            uint64 `json:"max_collective_outputs,string"`
			MinCollectiveClaimPeriod        uint64 `json:"min_collective_claim_period,string"`
			ValidatorRecoveryBond           uint64 `json:"validator_recovery_bond,string"`
			MaxAnnualInflation              string `json:"max_annual_inflation"`
			MaxProposalTitleSize            uint64 `json:"max_proposal_title_size,string"`
			MaxProposalDescriptionSize      uint64 `json:"max_proposal_description_size,string"`
			MaxProposalPollOptionSize       uint64 `json:"max_proposal_poll_option_size,string"`
			MaxProposalPollOptionCount      uint64 `json:"max_proposal_poll_option_count,string"`
			MaxProposalReferenceSize        uint64 `json:"max_proposal_reference_size,string"`
			MaxProposalChecksumSize         uint64 `json:"max_proposal_checksum_size,string"`
			MinDappBond                     uint64 `json:"min_dapp_bond,string"`
			MaxDappBond                     uint64 `json:"max_dapp_bond,string"`
			DappLiquidationThreshold        uint64 `json:"dapp_liquidation_threshold,string"`
			DappLiquidationPeriod           uint64 `json:"dapp_liquidation_period,string"`
			DappBondDuration                uint64 `json:"dapp_bond_duration,string"`
			DappVerifierBond                string `json:"dapp_verifier_bond"`
			DappAutoDenounceTime            uint64 `json:"dapp_auto_denounce_time,string"`
			DappMischanceRankDecreaseAmount uint64 `json:"dapp_mischance_rank_decrease_amount,string"`
			DappMaxMischance                uint64 `json:"dapp_max_mischance,string"`
			DappInactiveRankDecreasePercent uint64 `json:"dapp_inactive_rank_decrease_percent,string"`
			DappPoolSlippageDefault         string `json:"dapp_pool_slippage_default"`
			MintingFtFee                    uint64 `json:"minting_ft_fee,string"`
			MintingNftFee                   uint64 `json:"minting_nft_fee,string"`
			VetoThreshold                   string `json:"veto_threshold"`
		}

		type NetworkPropertiesParam struct {
			NetworkProperties NetworkProperties `json:"network_properties"`
		}

		var networkProperties NetworkPropertiesParam
		err = json.Unmarshal(params[4], &networkProperties)
		if err != nil {
			return nil, err
		}

		inActiveRankDecreasePercent, err := cosmostypes.NewDecFromStr(networkProperties.NetworkProperties.InactiveRankDecreasePercent)
		if err != nil {
			return nil, err
		}
		validatorsFeeShare, err := cosmostypes.NewDecFromStr(networkProperties.NetworkProperties.ValidatorsFeeShare)
		if err != nil {
			return nil, err
		}
		inflationRate, err := cosmostypes.NewDecFromStr(networkProperties.NetworkProperties.InflationRate)
		if err != nil {
			return nil, err
		}
		maxJailedPercentage, err := cosmostypes.NewDecFromStr(networkProperties.NetworkProperties.MaxJailedPercentage)
		if err != nil {
			return nil, err
		}
		maxSlashingPercentage, err := cosmostypes.NewDecFromStr(networkProperties.NetworkProperties.MaxSlashingPercentage)
		if err != nil {
			return nil, err
		}
		maxAnnualInflation, err := cosmostypes.NewDecFromStr(networkProperties.NetworkProperties.MaxAnnualInflation)
		if err != nil {
			return nil, err
		}
		dappVerifierBond, err := cosmostypes.NewDecFromStr(networkProperties.NetworkProperties.DappVerifierBond)
		if err != nil {
			return nil, err
		}
		dappPoolSlippageDefault, err := cosmostypes.NewDecFromStr(networkProperties.NetworkProperties.DappPoolSlippageDefault)
		if err != nil {
			return nil, err
		}
		vetoThreshold, err := cosmostypes.NewDecFromStr(networkProperties.NetworkProperties.VetoThreshold)
		if err != nil {
			return nil, err
		}

		networkPropertiesParam := customgovtypes.NetworkProperties{
			MinTxFee:                        networkProperties.NetworkProperties.MinTxFee,
			MaxTxFee:                        networkProperties.NetworkProperties.MaxTxFee,
			VoteQuorum:                      networkProperties.NetworkProperties.VoteQuorum,
			MinimumProposalEndTime:          networkProperties.NetworkProperties.MinimumProposalEndTime,
			ProposalEnactmentTime:           networkProperties.NetworkProperties.ProposalEnactmentTime,
			MinProposalEndBlocks:            networkProperties.NetworkProperties.MinProposalEndBlocks,
			MinProposalEnactmentBlocks:      networkProperties.NetworkProperties.MinProposalEnactmentBlocks,
			EnableForeignFeePayments:        networkProperties.NetworkProperties.EnableForeignFeePayments,
			MischanceRankDecreaseAmount:     networkProperties.NetworkProperties.MischanceRankDecreaseAmount,
			MaxMischance:                    networkProperties.NetworkProperties.MaxMischance,
			MischanceConfidence:             networkProperties.NetworkProperties.MischanceConfidence,
			InactiveRankDecreasePercent:     inActiveRankDecreasePercent,
			MinValidators:                   networkProperties.NetworkProperties.MinValidators,
			PoorNetworkMaxBankSend:          networkProperties.NetworkProperties.PoorNetworkMaxBankSend,
			UnjailMaxTime:                   networkProperties.NetworkProperties.UnjailMaxTime,
			EnableTokenWhitelist:            networkProperties.NetworkProperties.EnableTokenWhitelist,
			EnableTokenBlacklist:            networkProperties.NetworkProperties.EnableTokenBlacklist,
			MinIdentityApprovalTip:          networkProperties.NetworkProperties.MinIdentityApprovalTip,
			UniqueIdentityKeys:              networkProperties.NetworkProperties.UniqueIdentityKeys,
			UbiHardcap:                      networkProperties.NetworkProperties.UbiHardcap,
			ValidatorsFeeShare:              validatorsFeeShare,
			InflationRate:                   inflationRate,
			InflationPeriod:                 networkProperties.NetworkProperties.InflationPeriod,
			UnstakingPeriod:                 networkProperties.NetworkProperties.UnstakingPeriod,
			MaxDelegators:                   networkProperties.NetworkProperties.MaxDelegators,
			MinDelegationPushout:            networkProperties.NetworkProperties.MinDelegationPushout,
			SlashingPeriod:                  networkProperties.NetworkProperties.SlashingPeriod,
			MaxJailedPercentage:             maxJailedPercentage,
			MaxSlashingPercentage:           maxSlashingPercentage,
			MinCustodyReward:                networkProperties.NetworkProperties.MinCustodyReward,
			MaxCustodyBufferSize:            networkProperties.NetworkProperties.MaxCustodyBufferSize,
			MaxCustodyTxSize:                networkProperties.NetworkProperties.MaxCustodyTxSize,
			AbstentionRankDecreaseAmount:    networkProperties.NetworkProperties.AbstentionRankDecreaseAmount,
			MaxAbstention:                   networkProperties.NetworkProperties.MaxAbstention,
			MinCollectiveBond:               networkProperties.NetworkProperties.MinCollectiveBond,
			MinCollectiveBondingTime:        networkProperties.NetworkProperties.MinCollectiveBondingTime,
			MaxCollectiveOutputs:            networkProperties.NetworkProperties.MaxCollectiveOutputs,
			MinCollectiveClaimPeriod:        networkProperties.NetworkProperties.MinCollectiveClaimPeriod,
			ValidatorRecoveryBond:           networkProperties.NetworkProperties.ValidatorRecoveryBond,
			MaxAnnualInflation:              maxAnnualInflation,
			MaxProposalTitleSize:            networkProperties.NetworkProperties.MaxProposalTitleSize,
			MaxProposalDescriptionSize:      networkProperties.NetworkProperties.MaxProposalDescriptionSize,
			MaxProposalPollOptionSize:       networkProperties.NetworkProperties.MaxProposalPollOptionSize,
			MaxProposalPollOptionCount:      networkProperties.NetworkProperties.MaxProposalPollOptionCount,
			MaxProposalReferenceSize:        networkProperties.NetworkProperties.MaxProposalReferenceSize,
			MaxProposalChecksumSize:         networkProperties.NetworkProperties.MaxProposalChecksumSize,
			MinDappBond:                     networkProperties.NetworkProperties.MinDappBond,
			MaxDappBond:                     networkProperties.NetworkProperties.MaxDappBond,
			DappLiquidationThreshold:        networkProperties.NetworkProperties.DappLiquidationThreshold,
			DappLiquidationPeriod:           networkProperties.NetworkProperties.DappLiquidationPeriod,
			DappBondDuration:                networkProperties.NetworkProperties.DappBondDuration,
			DappVerifierBond:                dappVerifierBond,
			DappAutoDenounceTime:            networkProperties.NetworkProperties.DappAutoDenounceTime,
			DappMischanceRankDecreaseAmount: networkProperties.NetworkProperties.DappMischanceRankDecreaseAmount,
			DappMaxMischance:                networkProperties.NetworkProperties.DappMaxMischance,
			DappInactiveRankDecreasePercent: networkProperties.NetworkProperties.DappInactiveRankDecreasePercent,
			DappPoolSlippageDefault:         dappPoolSlippageDefault,
			MintingFtFee:                    networkProperties.NetworkProperties.MintingFtFee,
			MintingNftFee:                   networkProperties.NetworkProperties.MintingNftFee,
			VetoThreshold:                   vetoThreshold,
		}

		msg = customgovtypes.NewMsgSetNetworkProperties(proposer, &networkPropertiesParam)
	case MsgSetExecutionFee:
		// V, R, S, signer, executionFee, failureFee, timeout, defaultParams
		proposer, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		type ExecutionFeeParam struct {
			TransactionType string `json:"transaction_type"`
			ExecutionFee    uint64 `json:"execution_fee,string"`
			FailureFee      uint64 `json:"failure_fee,string"`
			Timeout         uint64 `json:"timeout,string"`
			DefaultParams   uint64 `json:"default_params,string"`
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

		// type ClaimParam struct {
		// 	Moniker string `json:"moniker"`
		// 	ValKey  string `json:"val_key"`
		// 	PubKey  string `json:"pub_key"`
		// }

		// var claimParam ClaimParam
		// err = json.Unmarshal(params[4], &claimParam)
		// if err != nil {
		// 	return nil, err
		// }

		// valKey, err := cosmostypes.ValAddressFromBech32(claimParam.ValKey)
		// if err != nil {
		// 	return nil, err
		// }

		// // how to get pub key from string?
		// cryptotypes

		// msg, err = stakingtypes.NewMsgClaimValidator(claimParam.Moniker, valKey, claimParam.PubKey)
		// if err != nil {
		// 	return nil, err
		// }
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
			Decimals    uint32   `json:"decimals,string"`
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
			Denom         string `json:"denom"`
			Rate          string `json:"rate"`
			IsFeePayments bool   `json:"is_fee_payments"`
			StakeCap      string `json:"stake_cap"`
			StakeMin      string `json:"stake_min"`
			IsStakeToken  bool   `json:"is_stake_token"`
			Invalidated   bool   `json:"invalidated"`
		}

		var upsertParam UpsertTokenRateParam
		err = json.Unmarshal(params[4], &upsertParam)
		if err != nil {
			return nil, err
		}

		rate, err := cosmostypes.NewDecFromStr(upsertParam.Rate)
		if err != nil {
			return nil, err
		}
		stakeCap, err := cosmostypes.NewDecFromStr(upsertParam.StakeCap)
		if err != nil {
			return nil, err
		}
		stakeMin, ok := cosmostypes.NewIntFromString(upsertParam.StakeMin)
		if !ok {
			return nil, errors.New(fmt.Sprintln("StakeMin - decoding from string to Int type is failed:", upsertParam.StakeMin))
		}

		msg = tokenstypes.NewMsgUpsertTokenRate(proposer, upsertParam.Denom, rate, upsertParam.IsFeePayments,
			stakeCap, stakeMin, upsertParam.IsStakeToken, upsertParam.Invalidated)
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

		type PermInfo struct {
			OwnerRoles    []string `json:"owner_roles"`
			OwnerAccounts []string `json:"owner_accounts"`
		}

		type WeightedRole struct {
			Role   uint64 `json:"role,string"`
			Weight string `json:"weight"`
		}

		type WeightedAccount struct {
			Account string `json:"account"`
			Weight  string `json:"weight"`
		}

		type WeightedPermInfo struct {
			Roles    []WeightedRole    `json:"roles"`
			Accounts []WeightedAccount `json:"accounts"`
		}

		type SpendingPoolParam struct {
			Name              string           `json:"name"`
			ClaimStart        uint64           `json:"claim_start,string"`
			ClaimEnd          uint64           `json:"claim_end,string"`
			Rates             []string         `json:"rates"`
			VoteQuorum        uint64           `json:"vote_quorum,string"`
			VotePeriod        uint64           `json:"vote_period,string"`
			VoteEnactment     uint64           `json:"vote_enactment,string"`
			Owners            PermInfo         `json:"owners"`
			Beneficiaries     WeightedPermInfo `json:"beneficiaries"`
			IsDynamicRate     bool             `json:"is_dynamic_rate"`
			DynamicRatePeriod uint64           `json:"dynamic_rate_period,string"`
		}

		var poolParam SpendingPoolParam
		err = json.Unmarshal(params[4], &poolParam)
		if err != nil {
			return nil, err
		}

		var rates cosmostypes.DecCoins
		for _, rate := range poolParam.Rates {
			coin, err := cosmostypes.ParseDecCoin(rate)
			if err != nil {
				return nil, err
			}
			rates.Add(coin)
		}

		var permInfo spendingtypes.PermInfo
		for _, roleStr := range poolParam.Owners.OwnerRoles {
			role, err := strconv.ParseUint(roleStr, 10, 64)
			if err != nil {
				return nil, err
			}
			permInfo.OwnerRoles = append(permInfo.OwnerRoles, role)
		}
		permInfo.OwnerAccounts = append(permInfo.OwnerAccounts, poolParam.Owners.OwnerAccounts...)

		var beneficiaries spendingtypes.WeightedPermInfo
		for _, account := range poolParam.Beneficiaries.Accounts {
			weight, err := cosmostypes.NewDecFromStr(account.Weight)
			if err != nil {
				return nil, err
			}
			weightedAccount := spendingtypes.WeightedAccount{
				Account: account.Account,
				Weight:  weight,
			}
			beneficiaries.Accounts = append(beneficiaries.Accounts, weightedAccount)
		}
		for _, role := range poolParam.Beneficiaries.Roles {
			weight, err := cosmostypes.NewDecFromStr(role.Weight)
			if err != nil {
				return nil, err
			}
			weightedRole := spendingtypes.WeightedRole{
				Role:   role.Role,
				Weight: weight,
			}
			beneficiaries.Roles = append(beneficiaries.Roles, weightedRole)
		}

		msg = spendingtypes.NewMsgCreateSpendingPool(poolParam.Name, poolParam.ClaimStart, poolParam.ClaimEnd, rates,
			poolParam.VoteQuorum, poolParam.VotePeriod, poolParam.VoteEnactment, permInfo, beneficiaries,
			sender, poolParam.IsDynamicRate, poolParam.DynamicRatePeriod)
	case MsgDepositSpendingPool:
		// V, R, S, signer, amount string, name string
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		type SpendingPoolParam struct {
			Name    string `json:"name"`
			Amounts string `json:"amounts"`
		}

		var poolParam SpendingPoolParam
		err = json.Unmarshal(params[4], &poolParam)
		if err != nil {
			return nil, err
		}

		amounts, err := cosmostypes.ParseCoinsNormalized(poolParam.Amounts)
		if err != nil {
			return nil, err
		}

		msg = spendingtypes.NewMsgDepositSpendingPool(poolParam.Name, amounts, sender)
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
			Validator  string `json:"validator"`
			Enabled    bool   `json:"enabled"`
			Commission string `json:"commission"`
		}

		var poolParam StakingPoolParam
		err = json.Unmarshal(params[4], &poolParam)
		if err != nil {
			return nil, err
		}

		commission, err := cosmostypes.NewDecFromStr(poolParam.Commission)
		if err != nil {
			return nil, err
		}

		msg = multistakingtypes.NewMsgUpsertStakingPool(from, poolParam.Validator, poolParam.Enabled, commission)
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

		amounts, err := cosmostypes.ParseCoinsNormalized(delegateParam.Amounts)
		if err != nil {
			return nil, err
		}

		msg = multistakingtypes.NewMsgDelegate(from, delegateParam.To, amounts)
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

		amounts, err := cosmostypes.ParseCoinsNormalized(delegateParam.Amounts)
		if err != nil {
			return nil, err
		}

		msg = multistakingtypes.NewMsgUndelegate(from, delegateParam.To, amounts)
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
			UndelegationId uint64 `json:"undelegation_id,string"`
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

		type CustodySettings struct {
			CustodyEnabled bool   `json:"custody_enabled"`
			CustodyMode    uint64 `json:"custody_mode,string"`
			UsePassword    bool   `json:"use_password"`
			UseWhiteList   bool   `json:"use_white_list"`
			UseLimits      bool   `json:"use_limits"`
			Key            string `json:"key"`
			NextController string `json:"next_controller"`
		}

		type CustodyParam struct {
			CustodySettings CustodySettings `json:"custody_settings"`
			OldKey          string          `json:"old_key"`
			NewKey          string          `json:"new_key"`
			Next            string          `json:"next"`
			Target          string          `json:"target"`
		}

		var custodyParam CustodyParam
		err = json.Unmarshal(params[4], &custodyParam)
		if err != nil {
			return nil, err
		}

		custodySettings := custodytypes.CustodySettings{
			CustodyEnabled: custodyParam.CustodySettings.CustodyEnabled,
			CustodyMode:    custodyParam.CustodySettings.CustodyMode,
			UsePassword:    custodyParam.CustodySettings.UsePassword,
			UseWhiteList:   custodyParam.CustodySettings.UseWhiteList,
			UseLimits:      custodyParam.CustodySettings.UseLimits,
			Key:            custodyParam.CustodySettings.Key,
			NextController: custodyParam.CustodySettings.NextController,
		}

		msg = custodytypes.NewMsgCreateCustody(from, custodySettings, custodyParam.OldKey, custodyParam.NewKey,
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

		var newAddrs []cosmostypes.AccAddress
		for _, addrStr := range custodyParam.NewAddrs {
			addr, err := cosmostypes.AccAddressFromBech32(addrStr)
			if err != nil {
				return nil, err
			}
			newAddrs = append(newAddrs, addr)
		}

		msg = custodytypes.NewMsgAddToCustodyWhiteList(from, newAddrs, custodyParam.OldKey, custodyParam.NewKey,
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

		var newAddrs []cosmostypes.AccAddress
		for _, addrStr := range custodyParam.NewAddrs {
			addr, err := cosmostypes.AccAddressFromBech32(addrStr)
			if err != nil {
				return nil, err
			}
			newAddrs = append(newAddrs, addr)
		}

		msg = custodytypes.NewMsgAddToCustodyCustodians(from, newAddrs, custodyParam.OldKey, custodyParam.NewKey,
			custodyParam.Next, custodyParam.Target)
	case MsgRemoveFromCustodyCustodians:
		// V, R, S, signer, newAddr cosmostypes.AccAddress,
		// oldKey string, newKey string, next string, target string
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		var custodyParam SingleCustodianParam
		err = json.Unmarshal(params[4], &custodyParam)
		if err != nil {
			return nil, err
		}

		newAddr, err := cosmostypes.AccAddressFromBech32(custodyParam.NewAddr)
		if err != nil {
			return nil, err
		}

		msg = custodytypes.NewMsgRemoveFromCustodyCustodians(from, newAddr, custodyParam.OldKey, custodyParam.NewKey,
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

		var custodyParam SingleCustodianParam
		err = json.Unmarshal(params[4], &custodyParam)
		if err != nil {
			return nil, err
		}

		newAddr, err := cosmostypes.AccAddressFromBech32(custodyParam.NewAddr)
		if err != nil {
			return nil, err
		}

		msg = custodytypes.NewMsgRemoveFromCustodyWhiteList(from, newAddr, custodyParam.OldKey, custodyParam.NewKey,
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

		var custodyParam CustodyParam
		err = json.Unmarshal(params[4], &custodyParam)
		if err != nil {
			return nil, err
		}

		to, err := cosmostypes.AccAddressFromBech32(custodyParam.To)
		if err != nil {
			return nil, err
		}

		msg = custodytypes.NewMsgApproveCustodyTransaction(from, to, custodyParam.Hash)
	case MsgDeclineCustodyTx:
		// V, R, S, signer, to string, hash string
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		var custodyParam CustodyParam
		err = json.Unmarshal(params[4], &custodyParam)
		if err != nil {
			return nil, err
		}

		to, err := cosmostypes.AccAddressFromBech32(custodyParam.To)
		if err != nil {
			return nil, err
		}

		msg = custodytypes.NewMsgDeclineCustodyTransaction(from, to, custodyParam.Hash)
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

	// fmt.Println(grpcRes.TxResponse)

	if grpcRes.TxResponse.Code != 0 {
		return "", errors.New(fmt.Sprintln("send tx failed - result code: ", grpcRes.TxResponse.Code))
	}

	return grpcRes.TxResponse.TxHash, nil
}
