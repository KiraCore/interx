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
	ethcommon "github.com/ethereum/go-ethereum/common"
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
	MsgApproveCustodyTransaction
	MsgDeclineCustodyTransaction
)

var grpcConn *grpc.ClientConn

// decode 256bit param like bool, uint, hex-typed address etc
func Decode256Bit(data *[]byte, params *[][]byte) {
	*params = append(*params, (*data)[:32])
	*data = (*data)[32:]
}

// decode string-typed param
// structure:
// * offset - offset of the string in the data 	: 32byte
// * length - length of the string 				: 32byte
// * content - content of the string			: (length/32+1)*32byte
func DecodeString(data *[]byte, params *[][]byte) {
	// offset := data[:32] // string value offset
	*data = (*data)[32:]

	length, _ := bytes2uint64((*data)[:32])
	*data = (*data)[32:]

	*params = append(*params, (*data)[:length])
	*data = (*data)[(length/32+1)*32:]
}

// decode string-typed params - if string-typed params are consecutive, offsets are placed on top
// and length, contents are placed in order
// structure:
// * offsets - offsets of the strings in the data 	: length*32byte
// * array of below data
// ** length - length of the string 				: 32byte
// ** content - content of the string				: (length/32+1)*32byte
func DecodeStrings(data *[]byte, params *[][]byte, paramLen int) {
	// remove offset of each string
	*data = (*data)[32*paramLen:]
	for i := 0; i < paramLen; i++ {
		length, _ := bytes2uint64((*data)[:32])
		*data = (*data)[32:]

		*params = append(*params, (*data)[:length])
		*data = (*data)[(length/32+1)*32:]
	}
}

// decode string-array-typed param
// structure:
// * offset - offset of the strings in the data 	: 32byte
// * offsets - offsets of the string in the data 	: length*32byte
// * array of below data
// ** length - length of the string 				: 32byte
// ** content - content of the string				: (length/32+1)*32byte
func DecodeStringArray(data *[]byte, params *[][]byte) {
	// offset := data[:32] // string value offset
	*data = (*data)[32:]

	length, _ := bytes2uint64((*data)[:32])
	*params = append(*params, (*data)[:32])
	*data = (*data)[32:]

	// remove offset of each string
	*data = (*data)[32*length:]
	for i := uint64(0); i < length; i++ {
		length, _ := bytes2uint64((*data)[:32])
		*data = (*data)[32:]

		*params = append(*params, (*data)[:length])
		*data = (*data)[(length/32+1)*32:]
	}
}

// decode hex-array-typed param
// structure:
// * offset - offset of the hex-array in the data 	: 32byte
// * length - length of the hex 					: 32byte
// * array of below data
// ** content - content of the hex					: 32byte
func DecodeHexArray(data *[]byte, params *[][]byte) {
	// offset := data[:32] // string value offset
	*data = (*data)[32:]

	length, _ := bytes2uint64((*data)[:32])
	*data = (*data)[32:]

	*params = append(*params, (*data)[:length])
}

func DecodeParam(data []byte, txType int) [][]byte {
	if txType == MsgBankSend {
		return nil
	}

	var params [][]byte

	// decode data field v, r, s, sender
	for i := 0; i < 4; i++ {
		Decode256Bit(&data, &params)
	}

	switch txType {
	case MsgSubmitEvidence: // TODO: test
		// height, power, time, consensusAddr
		Decode256Bit(&data, &params)
		Decode256Bit(&data, &params)
		Decode256Bit(&data, &params)
		DecodeString(&data, &params)
	case MsgSubmitProposal:
	case MsgVoteProposal: // TODO: test
		// proposalID, option, slash
		Decode256Bit(&data, &params)
		Decode256Bit(&data, &params)
		DecodeString(&data, &params)
	case MsgRegisterIdentityRecords:
		// identityInfo,
		DecodeString(&data, &params)
	case MsgDeleteIdentityRecord:
		// len, keys
		DecodeStringArray(&data, &params)
	case MsgRequestIdentityRecordsVerify:
		// tip, verifier, recordIds
		DecodeStrings(&data, &params, 2)

		// recordIds
		DecodeHexArray(&data, &params)
	case MsgHandleIdentityRecordsVerifyRequest:
		// requestId, isApprove
		Decode256Bit(&data, &params)
		Decode256Bit(&data, &params)
	case MsgCancelIdentityRecordsVerifyRequest:
		// verifyRequestId
		Decode256Bit(&data, &params)
	case MsgSetNetworkProperties:
		// len, boolParamsList, len, uintParamsList, len, strParamsList, len, legacyDecParamsList

		// len, boolParamsList
		DecodeHexArray(&data, &params)

		// len, uintParamsList
		DecodeHexArray(&data, &params)

		// len, strParamsList
		DecodeStringArray(&data, &params)

		// len, legacyDecParamsList
		DecodeStringArray(&data, &params)
	case MsgSetExecutionFee:
		// executionFee, failureFee, timeout, defaultParams, transactionType
		for i := 0; i < 4; i++ {
			Decode256Bit(&data, &params)
		}

		// transactionType
		DecodeString(&data, &params)
	case MsgClaimCouncilor:
		// moniker string, username string, description string, social string, contact string, avatar string
		DecodeStrings(&data, &params, 6)
	case MsgWhitelistPermissions:
		// permission uint, controlledAddr address
		Decode256Bit(&data, &params)
		DecodeString(&data, &params)
	case MsgBlacklistPermissions:
		// permission uint, addr address
		Decode256Bit(&data, &params)
		DecodeString(&data, &params)
	case MsgCreateRole:
		// sid string, description string
		DecodeString(&data, &params)
		DecodeString(&data, &params)
	case MsgAssignRole:
		// roleid uint32, controller address
		Decode256Bit(&data, &params)
		DecodeString(&data, &params)
	case MsgUnassignRole:
		// roleid uint32, controller address
		Decode256Bit(&data, &params)
		DecodeString(&data, &params)
	case MsgWhitelistRolePermission:
		// permission uint32, roleIdentifier string
		Decode256Bit(&data, &params)
		DecodeString(&data, &params)
	case MsgBlacklistRolePermission:
		// permission uint32, roleIdentifier string
		Decode256Bit(&data, &params)
		DecodeString(&data, &params)
	case MsgRemoveWhitelistRolePermission:
		// permission uint32, roleIdentifier string
		Decode256Bit(&data, &params)
		DecodeString(&data, &params)
	case MsgRemoveBlacklistRolePermission:
		// permission uint32, roleIdentifier string
		Decode256Bit(&data, &params)
		DecodeString(&data, &params)
	case MsgClaimValidator:
	case MsgUpsertTokenAlias:
		// invalidated bool, decimals uint32, symbol string, name string, icon string, len, denoms []string
		Decode256Bit(&data, &params)
		Decode256Bit(&data, &params)
		DecodeStrings(&data, &params, 3)
		// denoms
		DecodeStringArray(&data, &params)
	case MsgUpsertTokenRate:
		// feePayments bool, stakeToken bool, invalidated bool, denom string, rate cosmostypes.Dec,
		// stakeCap cosmostypes.Dec, stakeMin cosmostypes.Int
		for i := 0; i < 3; i++ {
			Decode256Bit(&data, &params)
		}

		DecodeStrings(&data, &params, 4)
	case MsgActivate:
	case MsgPause:
	case MsgUnpause:
	case MsgCreateSpendingPool:
	case MsgDepositSpendingPool:
		// amount cosmostypes.Coins, name string
		DecodeStrings(&data, &params, 2)
	case MsgRegisterSpendingPoolBeneficiary:
		// name string
		DecodeString(&data, &params)
	case MsgClaimSpendingPool:
		// name string
		DecodeString(&data, &params)
	case MsgUpsertStakingPool:
		// enabled bool, validator string, commission cosmostypes.Dec
		Decode256Bit(&data, &params)
		DecodeStrings(&data, &params, 2)
	case MsgDelegate:
		// amount, validator
		DecodeStrings(&data, &params, 2)
	case MsgUndelegate:
		// amount, validator
		DecodeStrings(&data, &params, 2)
	case MsgClaimRewards:
	case MsgClaimUndelegation:
		// undelegationId uint64
		Decode256Bit(&data, &params)
	case MsgSetCompoundInfo:
		// allDenom bool, len, denoms []string
		Decode256Bit(&data, &params)
		DecodeStringArray(&data, &params)
	case MsgRegisterDelegator:
	case MsgCreateCustody:
		// V, R, S, signer, custodyMode uint256, key string, nextController string, len, boolArr, len, strArr
		for i := 0; i < 4; i++ {
			Decode256Bit(&data, &params)
		}

		DecodeStrings(&data, &params, 2)

		DecodeHexArray(&data, &params)
		DecodeStringArray(&data, &params)
	case MsgAddToCustodyWhiteList:
		// oldKey string, newKey string, next string, target string, len, newAddr []string
		DecodeStrings(&data, &params, 4)
		DecodeStringArray(&data, &params)
	case MsgAddToCustodyCustodians:
		// oldKey string, newKey string, next string, target string, len, newAddr []string
		DecodeStrings(&data, &params, 4)
		DecodeStringArray(&data, &params)
	case MsgRemoveFromCustodyCustodians:
		// newAddr string, oldKey string, newKey string, next string, target string
		DecodeStrings(&data, &params, 5)
	case MsgDropCustodyCustodians:
		// oldKey string, newKey string, next string, target string
		DecodeStrings(&data, &params, 4)
	case MsgRemoveFromCustodyWhiteList:
		// newAddr string, oldKey string, newKey string, next string, target string
		DecodeStrings(&data, &params, 5)
	case MsgDropCustodyWhiteList:
		// oldKey string, newKey string, next string, target string
		DecodeStrings(&data, &params, 4)
	case MsgApproveCustodyTransaction:
		// to string, hash string
		DecodeStrings(&data, &params, 2)
	case MsgDeclineCustodyTransaction:
		// to string, hash string
		DecodeStrings(&data, &params, 2)
	default:

	}

	return params
}

func SendTx(txRawData string, gwCosmosmux *runtime.ServeMux, r *http.Request) (string, error) {
	byteData, err := hex.DecodeString(txRawData[2:])
	if err != nil {
		return "", err
	}

	ethTxData, err := GetEthTxInfo(byteData)

	if err != nil {
		return "", err
	}

	var txBytes []byte

	submitEvidencePrefix, _ := hex.DecodeString("33104ff7")
	submitProposalPrefix, _ := hex.DecodeString("00000000")
	voteProposalPrefix, _ := hex.DecodeString("24d03a8b")
	registerIdentityRecordsPrefix, _ := hex.DecodeString("bc05f106")
	deleteIdentityRecordPrefix, _ := hex.DecodeString("2ac0ca03")
	requestIdentityRecordsVerifyPrefix, _ := hex.DecodeString("f7c51612")
	handleIdentityRecordsVerifyRequestPrefix, _ := hex.DecodeString("2facbfe8")
	cancelIdentityRecordsVerifyRequestPrefix, _ := hex.DecodeString("9b66f097")
	setNetworkPropertiesPrefix, _ := hex.DecodeString("3a91d2ba")
	setExecutionFeePrefix, _ := hex.DecodeString("fa02785e")
	claimCouncilorPrefix, _ := hex.DecodeString("9dd85dbf")
	whitelistPermissionsPrefix, _ := hex.DecodeString("4db4bc0b")
	blacklistPermissionsPrefix, _ := hex.DecodeString("675ed828")
	createRolePrefix, _ := hex.DecodeString("7b0fadab")
	assignRolePrefix, _ := hex.DecodeString("83b3d393")
	unassignRolePrefix, _ := hex.DecodeString("2fa24c7a")
	whitelistRolePermissionPrefix, _ := hex.DecodeString("557f0e63")
	blacklistRolePermissionPrefix, _ := hex.DecodeString("7e7595fe")
	removeWhitelistRolePermissionPrefix, _ := hex.DecodeString("e9830bd8")
	removeBlacklistRolePermissionPrefix, _ := hex.DecodeString("45e2144c")
	claimValidatorPrefix, _ := hex.DecodeString("00000000")
	upsertTokenAliasPrefix, _ := hex.DecodeString("59122e45")
	upsertTokenRatePrefix, _ := hex.DecodeString("fb7e4c17")
	activatePrefix, _ := hex.DecodeString("a1374dc2")
	pausePrefix, _ := hex.DecodeString("1371cf19")
	unpausePrefix, _ := hex.DecodeString("b9179894")
	createSpendingPoolPrefix, _ := hex.DecodeString("00000000")
	depositSpendingPoolPrefix, _ := hex.DecodeString("9fff5508")
	registerSpendingPoolBeneficiaryPrefix, _ := hex.DecodeString("7ab7eecf")
	claimSpendingPoolPrefix, _ := hex.DecodeString("efeed4a0")
	upsertStakingPoolPrefix, _ := hex.DecodeString("1f0da7b4")
	delegatePrefix, _ := hex.DecodeString("4563230c")
	undelegatePrefix, _ := hex.DecodeString("d8e05398")
	claimRewardsPrefix, _ := hex.DecodeString("9e796524")
	claimUndelegationPrefix, _ := hex.DecodeString("505b0330")
	setCompoundInfoPrefix, _ := hex.DecodeString("fb1257e6")
	registerDelegatorPrefix, _ := hex.DecodeString("99db185d")
	createCustodyPrefix, _ := hex.DecodeString("d486e9de")
	addToCustodyWhiteListPrefix, _ := hex.DecodeString("25d1857d")
	addToCustodyCustodiansPrefix, _ := hex.DecodeString("9c4f31a9")
	removeFromCustodyCustodiansPrefix, _ := hex.DecodeString("8ec62788")
	dropCustodyCustodiansPrefix, _ := hex.DecodeString("07d33244")
	removeFromCustodyWhiteListPrefix, _ := hex.DecodeString("b99abc7d")
	dropCustodyWhiteListPrefix, _ := hex.DecodeString("4db07a4e")
	approveCustodyTransactionPrefix, _ := hex.DecodeString("e6d41e78")
	declineCustodyTransactionPrefix, _ := hex.DecodeString("21ebd09e")

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
	case bytes.Equal(ethTxData.Data[:4], approveCustodyTransactionPrefix):
		msgType = MsgApproveCustodyTransaction
	case bytes.Equal(ethTxData.Data[:4], declineCustodyTransactionPrefix):
		msgType = MsgDeclineCustodyTransaction
	default:
		return "", errors.New("no such functions")
	}

	txBytes, err = SignTx(ethTxData, gwCosmosmux, r, msgType)

	if err != nil {
		return "", err
	}

	// if ethTxData

	txHash, err := sendTx(r.Context(), txBytes)
	if err != nil {
		return "", err
	}

	return txHash, nil
}

func ValidateEIP712Sign(v, r, s []byte, sender ethcommon.Address, amount int, transactionType int) bool {
	// eip712DomainSeparator := crypto.Keccak256Hash(
	// 	[]byte("EIP712Domain(string name,string version,uint256 chainId)"),
	// 	crypto.Keccak256([]byte("SetTest")),
	// 	crypto.Keccak256([]byte("1")),
	// 	intToBytes(config.DefaultChainID),
	// )

	// // Define the hash of the struct
	// hashStruct := crypto.Keccak256Hash(
	// 	[]byte("delegate(uint amount, address sender)"),
	// 	intToBytes(int64(amount)),
	// 	sender.Bytes(),
	// )

	// // Define the final hash
	// hash := crypto.Keccak256Hash(
	// 	[]byte("\x19\x01"),
	// 	eip712DomainSeparator.Bytes(),
	// 	hashStruct.Bytes(),
	// )

	// signature := append(append(r, s...), v...)

	// // Recover the public key from the signature
	// pubKey, err := crypto.Ecrecover(hash.Bytes(), signature)
	// if err != nil {
	// 	return false
	// }

	// // Derive the signer's address from the public key
	// signerAddress := crypto.PubkeyToAddress(*pubKey)

	// if signerAddress != senderAddr {
	// 	return false
	// }
	return true
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

	params := DecodeParam(ethTxData.Data[4:], txType)

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

		height, _ := bytes2int64(params[4])
		power, _ := bytes2int64(params[5])
		unixTime, _ := bytes2int64(params[6])

		consensusAddr := string(params[7])

		evidence := &evidencetypes.Equivocation{
			Height:           height,
			Power:            power,
			Time:             time.UnixMicro(unixTime),
			ConsensusAddress: consensusAddr,
		}

		msg, err = evidencetypes.NewMsgSubmitEvidence(from, evidence)
		if err != nil {
			return nil, err
		}
	case MsgSubmitProposal:
		// msg := customgovtypes.NewMsgSubmitProposal(from)
	case MsgVoteProposal:
		// V, R, S, signer, proposalID, option, slash
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		proposalID, _ := bytes2uint64(params[4])
		option, _ := bytes2uint64(params[5])
		slash, _ := sdkmath.LegacyNewDecFromStr(string(params[6]))

		msg = customgovtypes.NewMsgVoteProposal(proposalID, from, customgovtypes.VoteOption(option), slash)
	case MsgRegisterIdentityRecords:
		// V, R, S, signer, identityInfo,
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		var infosStr map[string]string
		err = json.Unmarshal(params[4], &infosStr)
		if err != nil {
			return nil, err
		}

		var infos []customgovtypes.IdentityInfoEntry
		for key := range infosStr {
			infos = append(infos, customgovtypes.IdentityInfoEntry{
				Key:  key,
				Info: infosStr[key],
			})
		}

		msg = customgovtypes.NewMsgRegisterIdentityRecords(from, infos)
	case MsgDeleteIdentityRecord:
		// V, R, S, signer, len, keys,
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		var keys []string
		len, _ := bytes2uint64(params[4])
		for i := uint64(0); i < len; i++ {
			keys = append(keys, string(params[5+i]))
		}

		msg = customgovtypes.NewMsgDeleteIdentityRecords(from, keys)
	case MsgRequestIdentityRecordsVerify:
		// V, R, S, signer, tip, verifier, len, recordIds
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		balance, err := cosmostypes.ParseCoinNormalized(string(params[4]))
		if err != nil {
			return nil, err
		}

		verifier, err := string2cosmosAddr(params[5])
		if err != nil {
			return nil, err
		}

		var recordIds []uint64
		len, _ := bytes2int64(params[6])
		for i := 0; i < int(len); i++ {
			recordId, _ := bytes2uint64(params[7+i])
			recordIds = append(recordIds, recordId)
		}

		msg = customgovtypes.NewMsgRequestIdentityRecordsVerify(from, verifier, recordIds, balance)
	case MsgHandleIdentityRecordsVerifyRequest:
		// V, R, S, signer, requestId, isApprove
		verifier, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		isApprove := bytes2bool(params[5])
		requestId, _ := bytes2uint64(params[4])

		msg = customgovtypes.NewMsgHandleIdentityRecordsVerifyRequest(verifier, requestId, isApprove)
	case MsgCancelIdentityRecordsVerifyRequest:
		// V, R, S, signer, verifyRequestId
		executor, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		verifyRequestId, _ := bytes2uint64(params[4])

		msg = customgovtypes.NewMsgCancelIdentityRecordsVerifyRequest(executor, verifyRequestId)
	case MsgSetNetworkProperties:
		// V, R, S, signer, len, boolParamsList, len, uintParamsList, len, strParamsList, len, legacyDecParamsList
		proposer, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		var boolParams []bool
		var uintParams []uint64
		var strParams []string
		var legacyDecParams []sdkmath.LegacyDec

		boolLen, _ := bytes2uint64(params[4])
		for i := uint64(0); i < boolLen; i++ {
			boolParams = append(boolParams, bytes2bool(params[5+i]))
		}

		uintLen, _ := bytes2uint64(params[5+boolLen])
		for i := uint64(0); i < uintLen; i++ {
			uintParam, _ := bytes2uint64(params[6+boolLen+i])
			uintParams = append(uintParams, uintParam)
		}

		strLen, _ := bytes2uint64(params[6+boolLen+uintLen])
		for i := uint64(0); i < strLen; i++ {
			strParams = append(strParams, string(params[7+boolLen+uintLen+i]))
		}

		legacyDecLen, _ := bytes2uint64(params[7+boolLen+uintLen+strLen])
		for i := uint64(0); i < legacyDecLen; i++ {
			legacyDecParam, _ := sdkmath.LegacyNewDecFromStr(string(params[8+boolLen+uintLen+strLen+i]))
			legacyDecParams = append(legacyDecParams, legacyDecParam)
		}

		networkProperties := customgovtypes.NetworkProperties{
			EnableForeignFeePayments:        boolParams[0],
			EnableTokenWhitelist:            boolParams[1],
			EnableTokenBlacklist:            boolParams[2],
			MinTxFee:                        uintParams[0],
			MaxTxFee:                        uintParams[1],
			VoteQuorum:                      uintParams[2],
			MinimumProposalEndTime:          uintParams[3],
			ProposalEnactmentTime:           uintParams[4],
			MinProposalEndBlocks:            uintParams[5],
			MinProposalEnactmentBlocks:      uintParams[6],
			MaxMischance:                    uintParams[7],
			MischanceConfidence:             uintParams[8],
			MinValidators:                   uintParams[9],
			PoorNetworkMaxBankSend:          uintParams[10],
			UnjailMaxTime:                   uintParams[11],
			MinIdentityApprovalTip:          uintParams[12],
			UbiHardcap:                      uintParams[13],
			InflationPeriod:                 uintParams[14],
			UnstakingPeriod:                 uintParams[15],
			MaxDelegators:                   uintParams[16],
			MinDelegationPushout:            uintParams[17],
			SlashingPeriod:                  uintParams[18],
			MinCustodyReward:                uintParams[19],
			MaxCustodyBufferSize:            uintParams[20],
			MaxCustodyTxSize:                uintParams[21],
			AbstentionRankDecreaseAmount:    uintParams[22],
			MaxAbstention:                   uintParams[23],
			MinCollectiveBond:               uintParams[24],
			MinCollectiveBondingTime:        uintParams[25],
			MaxCollectiveOutputs:            uintParams[26],
			MinCollectiveClaimPeriod:        uintParams[27],
			ValidatorRecoveryBond:           uintParams[28],
			MaxProposalTitleSize:            uintParams[29],
			MaxProposalDescriptionSize:      uintParams[30],
			MaxProposalPollOptionSize:       uintParams[31],
			MaxProposalPollOptionCount:      uintParams[32],
			MaxProposalReferenceSize:        uintParams[33],
			MaxProposalChecksumSize:         uintParams[34],
			MinDappBond:                     uintParams[35],
			MaxDappBond:                     uintParams[36],
			DappLiquidationThreshold:        uintParams[37],
			DappLiquidationPeriod:           uintParams[38],
			DappBondDuration:                uintParams[39],
			DappAutoDenounceTime:            uintParams[40],
			DappMaxMischance:                uintParams[41],
			DappInactiveRankDecreasePercent: uintParams[42],
			MintingFtFee:                    uintParams[43],
			MintingNftFee:                   uintParams[44],
			UniqueIdentityKeys:              strParams[0],
			InactiveRankDecreasePercent:     legacyDecParams[0],
			ValidatorsFeeShare:              legacyDecParams[1],
			InflationRate:                   legacyDecParams[2],
			MaxJailedPercentage:             legacyDecParams[3],
			MaxSlashingPercentage:           legacyDecParams[4],
			MaxAnnualInflation:              legacyDecParams[5],
			DappVerifierBond:                legacyDecParams[6],
			DappPoolSlippageDefault:         legacyDecParams[7],
			VetoThreshold:                   legacyDecParams[8],
		}

		msg = customgovtypes.NewMsgSetNetworkProperties(proposer, &networkProperties)
	case MsgSetExecutionFee:
		// V, R, S, signer, executionFee, failureFee, timeout, defaultParams, transactionType
		proposer, err := bytes2cosmosAddr(params[3][12:])
		executionFee, _ := bytes2uint64(params[4])
		failureFee, _ := bytes2uint64(params[5])
		timeout, _ := bytes2uint64(params[6])
		defaultParams, _ := bytes2uint64(params[7])
		if err != nil {
			return nil, err
		}

		msg = customgovtypes.NewMsgSetExecutionFee(string(params[8]), executionFee, failureFee, timeout, defaultParams, proposer)
	case MsgClaimCouncilor:
		// V, R, S, signer, moniker string, username string, description string, social string, contact string, avatar string
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		msg = customgovtypes.NewMsgClaimCouncilor(sender, string(params[4]), string(params[5]), string(params[6]),
			string(params[7]), string(params[8]), string(params[9]))
	case MsgWhitelistPermissions:
		// V, R, S, signer, permission uint256, controlledAddr string
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		permission, _ := bytes2uint32(params[4])
		controlledAddr, err := string2cosmosAddr(params[5])
		if err != nil {
			return nil, err
		}

		msg = customgovtypes.NewMsgWhitelistPermissions(sender, controlledAddr, permission)
	case MsgBlacklistPermissions:
		// V, R, S, signer, permission uint256, controlledAddr string
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		permission, _ := bytes2uint32(params[4])
		controlledAddr, err := string2cosmosAddr(params[5])
		if err != nil {
			return nil, err
		}

		msg = customgovtypes.NewMsgBlacklistPermissions(sender, controlledAddr, permission)
	case MsgCreateRole:
		// V, R, S, signer, sid string, description string
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		msg = customgovtypes.NewMsgCreateRole(sender, string(params[4]), string(params[5]))
	case MsgAssignRole:
		// V, R, S, signer, roleid uint32, controller string
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		roleId, _ := bytes2uint32(params[4])
		controller, err := string2cosmosAddr(params[5])
		if err != nil {
			return nil, err
		}

		msg = customgovtypes.NewMsgAssignRole(sender, controller, roleId)
	case MsgUnassignRole:
		// V, R, S, signer, roleid uint32, controller address
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		roleId, _ := bytes2uint32(params[4])
		controller, err := string2cosmosAddr(params[5])
		if err != nil {
			return nil, err
		}

		msg = customgovtypes.NewMsgUnassignRole(sender, controller, roleId)
	case MsgWhitelistRolePermission:
		// V, R, S, signer, permission uint32, roleIdentifier string
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		permission, _ := bytes2uint32(params[4])

		msg = customgovtypes.NewMsgWhitelistRolePermission(sender, string(params[5]), permission)
	case MsgBlacklistRolePermission:
		// V, R, S, signer, permission uint32, roleIdentifier string
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		permission, _ := bytes2uint32(params[4])

		msg = customgovtypes.NewMsgBlacklistRolePermission(sender, string(params[5]), permission)
	case MsgRemoveWhitelistRolePermission:
		// V, R, S, signer, permission uint32, roleIdentifier string
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		permission, _ := bytes2uint32(params[4])

		msg = customgovtypes.NewMsgRemoveWhitelistRolePermission(sender, string(params[5]), permission)
	case MsgRemoveBlacklistRolePermission:
		// V, R, S, signer, permission uint32, roleIdentifier string
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		permission, _ := bytes2uint32(params[4])

		msg = customgovtypes.NewMsgRemoveBlacklistRolePermission(sender, string(params[5]), permission)
	case MsgClaimValidator:
		// V, R, S, signer, moniker string, valKey cosmostypes.ValAddress, pubKey cryptotypes.PubKey

		// valKey, err := cosmostypes.ValAddressFromBech32(string(params[5]))
		// if err != nil {
		// 	return nil, err
		// }

		// TODO: get public key from str
		// publicKey, err := secp256k1.PubKeySecp256k1([]byte(params[6]))
		// msg, err := stakingtypes.NewMsgClaimValidator(string(params[4]), valKey, publicKey)
		// if err != nil {
		// 	return nil, err
		// }
	case MsgUpsertTokenAlias:
		// V, R, S, signer, invalidated bool, decimals uint32, symbol string, name string, icon string, len, denoms []string
		proposer, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		isInvalidated := bytes2bool(params[4])
		decimals, _ := bytes2uint32(params[5])

		var denoms []string
		len, _ := bytes2int32(params[9])
		for i := int32(0); i < len; i++ {
			denoms = append(denoms, string(params[i+10]))
		}

		msg = tokenstypes.NewMsgUpsertTokenAlias(proposer, string(params[6]), string(params[7]), string(params[8]),
			uint32(decimals), denoms, isInvalidated)
	case MsgUpsertTokenRate:
		// V, R, S, signer, feePayments bool, stakeToken bool, invalidated bool, denom string, rate cosmostypes.Dec,
		// stakeCap cosmostypes.Dec, stakeMin cosmostypes.Int
		proposer, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}

		isFeePayments := bytes2bool(params[4])
		isStakeToken := bytes2bool(params[5])
		isInvalidated := bytes2bool(params[6])
		denom := string(params[7])
		rate, err := cosmostypes.NewDecFromStr(hex.EncodeToString(params[8]))
		if err != nil {
			return nil, err
		}
		stakeCap, err := cosmostypes.NewDecFromStr(hex.EncodeToString(params[9]))
		if err != nil {
			return nil, err
		}
		stakeMin, ok := cosmostypes.NewIntFromString(hex.EncodeToString(params[10]))
		if !ok {
			return nil, err
		}

		msg = tokenstypes.NewMsgUpsertTokenRate(proposer, denom, rate, isFeePayments, stakeCap, stakeMin,
			isStakeToken, isInvalidated)
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

		// msg = spendingtypes.NewMsgCreateSpendingPool()
	case MsgDepositSpendingPool:
		// V, R, S, signer, amount string, name string
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}
		amount, err := cosmostypes.ParseCoinsNormalized(string(params[4]))
		if err != nil {
			return nil, err
		}

		msg = spendingtypes.NewMsgDepositSpendingPool(string(params[5]), amount, sender)
	case MsgRegisterSpendingPoolBeneficiary:
		// V, R, S, signer, name string
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}
		msg = spendingtypes.NewMsgRegisterSpendingPoolBeneficiary(string(params[4]), sender)
	case MsgClaimSpendingPool:
		// V, R, S, signer, name string
		sender, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}
		msg = spendingtypes.NewMsgClaimSpendingPool(string(params[4]), sender)
	case MsgUpsertStakingPool:
		// V, R, S, signer, enabled bool, validator string, commission cosmostypes.Dec
		from, err := hex2bech32(hex.EncodeToString(params[3][12:]), TypeKiraAddr)
		if err != nil {
			return nil, err
		}

		enabled := bytes2bool(params[4])
		commission, err := cosmostypes.NewDecFromStr(string(params[5]))
		if err != nil {
			return nil, err
		}

		msg = multistakingtypes.NewMsgUpsertStakingPool(from, string(params[5]), enabled, commission)
	case MsgDelegate:
		// TODO: check eip712 validation
		// V, R, S, signer, amount, validator
		// validateTx()

		balance, err := cosmostypes.ParseCoinsNormalized(string(params[4]))
		if err != nil {
			return nil, err
		}

		from, err := hex2bech32(hex.EncodeToString(params[3][12:]), TypeKiraAddr)
		if err != nil {
			return nil, err
		}

		to := string(params[5])

		msg = multistakingtypes.NewMsgDelegate(from, to, balance)
	case MsgUndelegate:
		// V, R, S, signer, amount, validator
		from, err := hex2bech32(hex.EncodeToString(params[3][12:]), TypeKiraAddr)
		if err != nil {
			return nil, err
		}
		balance, err := cosmostypes.ParseCoinsNormalized(string(params[4]))
		if err != nil {
			return nil, err
		}
		to := string(params[5])

		msg = multistakingtypes.NewMsgUndelegate(from, to, balance)
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
		undelegationId, _ := bytes2uint64(params[4])

		msg = multistakingtypes.NewMsgClaimUndelegation(from, undelegationId)
	case MsgSetCompoundInfo:
		// V, R, S, signer, allDenom bool, len, denoms []string
		from, err := hex2bech32(hex.EncodeToString(params[3][12:]), TypeKiraAddr)
		if err != nil {
			return nil, err
		}
		allDenom := bytes2bool(params[4])
		var denoms []string
		len, _ := bytes2uint64(params[5])
		for i := uint64(0); i < len; i++ {
			denoms = append(denoms, string(params[i+6]))
		}
		msg = multistakingtypes.NewMsgSetCompoundInfo(from, allDenom, denoms)
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

		custodyMode, _ := bytes2uint64(params[4])

		var boolArr []bool
		boolArrLen, _ := bytes2uint64(params[7])
		for i := uint64(0); i < boolArrLen; i++ {
			boolParam := bytes2bool(params[i+8])
			boolArr = append(boolArr, boolParam)
		}

		var strArr []string
		strArrLen, _ := bytes2uint64(params[8+boolArrLen])
		for i := uint64(0); i < strArrLen; i++ {
			param := string(params[i+9+boolArrLen])
			strArr = append(strArr, param)
		}

		custodySettings := custodytypes.CustodySettings{
			CustodyEnabled: boolArr[0],
			UsePassword:    boolArr[1],
			UseWhiteList:   boolArr[2],
			UseLimits:      boolArr[3],
			CustodyMode:    custodyMode,
			Key:            string(params[5]),
			NextController: string(params[6]),
		}
		msg = custodytypes.NewMsgCreateCustody(from, custodySettings, strArr[0], strArr[1],
			strArr[2], strArr[3])
	case MsgAddToCustodyWhiteList:
		// V, R, S, signer, oldKey string, newKey string, next string, target string
		// len, newAddr []string,
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}
		var newAddrs []cosmostypes.AccAddress
		len, _ := bytes2uint64(params[8])
		for i := uint64(0); i < len; i++ {
			newddr, err := string2cosmosAddr(params[i+9])
			if err != nil {
				return nil, err
			}
			newAddrs = append(newAddrs, newddr)
		}

		msg = custodytypes.NewMsgAddToCustodyWhiteList(from, newAddrs, string(params[4]), string(params[5]),
			string(params[6]), string(params[7]))
	case MsgAddToCustodyCustodians:
		// V, R, S, signer, oldKey string, newKey string, next string, target string
		// len, newAddr []string
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}
		var newAddrs []cosmostypes.AccAddress
		len, _ := bytes2uint64(params[8])
		for i := uint64(0); i < len; i++ {
			newddr, err := string2cosmosAddr(params[i+9])
			if err != nil {
				return nil, err
			}
			newAddrs = append(newAddrs, newddr)
		}

		msg = custodytypes.NewMsgAddToCustodyCustodians(from, newAddrs, string(params[4]), string(params[5]),
			string(params[6]), string(params[7]))
	case MsgRemoveFromCustodyCustodians:
		// V, R, S, signer, newAddr cosmostypes.AccAddress,
		// oldKey string, newKey string, next string, target string
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}
		newAddr, err := string2cosmosAddr(params[4])
		if err != nil {
			return nil, err
		}
		msg = custodytypes.NewMsgRemoveFromCustodyCustodians(from, newAddr, string(params[5]), string(params[6]),
			string(params[7]), string(params[8]))
	case MsgDropCustodyCustodians:
		// V, R, S, signer
		// oldKey string, newKey string, next string, target string
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}
		msg = custodytypes.NewMsgDropCustodyCustodians(from, string(params[4]), string(params[5]),
			string(params[6]), string(params[7]))
	case MsgRemoveFromCustodyWhiteList:
		// V, R, S, signer, newAddr string,
		// oldKey string, newKey string, next string, target string
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}
		newAddr, err := string2cosmosAddr(params[4])
		if err != nil {
			return nil, err
		}
		msg = custodytypes.NewMsgRemoveFromCustodyWhiteList(from, newAddr, string(params[5]), string(params[6]),
			string(params[7]), string(params[8]))
	case MsgDropCustodyWhiteList:
		// V, R, S, signer
		// oldKey string, newKey string, next string, target string
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}
		msg = custodytypes.NewMsgDropCustodyWhiteList(from, string(params[4]), string(params[5]),
			string(params[6]), string(params[7]))
	case MsgApproveCustodyTransaction:
		// V, R, S, signer, to string, hash string
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}
		to, err := string2cosmosAddr(params[4])
		if err != nil {
			return nil, err
		}
		msg = custodytypes.NewMsgApproveCustodyTransaction(from, to, string(params[5]))
	case MsgDeclineCustodyTransaction:
		// V, R, S, signer, to string, hash string
		from, err := bytes2cosmosAddr(params[3][12:])
		if err != nil {
			return nil, err
		}
		to, err := string2cosmosAddr(params[4])
		if err != nil {
			return nil, err
		}
		msg = custodytypes.NewMsgDeclineCustodyTransaction(from, to, string(params[5]))
	}
	err = txBuilder.SetMsgs(msg)
	if err != nil {
		return nil, err
	}
	txBuilder.SetGasLimit(ethTxData.GasLimit)
	// TODO: set fee amount - how can I get the fee amount from eth transaction? or fix this?
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

func sendTx(ctx context.Context, txBytes []byte) (string, error) {
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
			TxBytes: txBytes, // Proto-binary of the signed transaction, see previous step.
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
