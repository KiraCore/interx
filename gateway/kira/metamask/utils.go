package metamask

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/KiraCore/interx/config"
	cosmostypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

const (
	TypeKiraAddr = iota
	TypeKiraValAddr
)

func hex2bech32(hexAddr string, addrType int) (string, error) {
	hexAddress := hexAddr
	if strings.Contains(hexAddr, "0x") {
		hexAddress = hexAddr[2:]
	}
	byteArray, err := hex.DecodeString(hexAddress)
	if err != nil {
		return "", err
	}

	var bech32Addr string
	switch addrType {
	case TypeKiraAddr:
		bech32Addr, err = bech32.ConvertAndEncode(config.DefaultKiraAddrPrefix, byteArray)
	case TypeKiraValAddr:
		bech32Addr, err = bech32.ConvertAndEncode(config.DefaultKiraValAddrPrefix, byteArray)
	default:
		err = errors.New(fmt.Sprintln("invalid addrType: ", addrType))
	}
	if err != nil {
		return "", err
	}
	return bech32Addr, nil
}

func bech322hex(bech32Addr string) (string, error) {
	_, data, err := bech32.DecodeAndConvert(bech32Addr)
	if err != nil {
		return "", nil
	}

	// Encode the byte slice as a hex string
	hexStr := hex.EncodeToString(data)
	return hexStr, nil
}

func hex2int64(hexStr string) (int64, error) {
	hexString := hexStr
	if strings.Contains(hexStr, "0x") {
		hexString = hexStr[2:]
	}
	integer, err := strconv.ParseInt(hexString, 16, 64)
	if err != nil {
		return 0, err
	}
	return integer, nil
}

func hex2int32(hexStr string) (int32, error) {
	hexString := hexStr
	if strings.Contains(hexStr, "0x") {
		hexString = hexStr[2:]
	}
	integer, err := strconv.ParseInt(hexString, 16, 32)
	if err != nil {
		return 0, err
	}
	return int32(integer), nil
}

func hex2uint64(hexStr string) (uint64, error) {
	hexString := hexStr
	if strings.Contains(hexStr, "0x") {
		hexString = hexStr[2:]
	}
	integer, err := strconv.ParseUint(hexString, 16, 64)
	if err != nil {
		return 0, err
	}
	return integer, nil
}

func hex2uint32(hexStr string) (uint32, error) {
	hexString := hexStr
	if strings.Contains(hexStr, "0x") {
		hexString = hexStr[2:]
	}
	integer, err := strconv.ParseUint(hexString, 16, 32)
	if err != nil {
		return 0, err
	}
	return uint32(integer), nil
}

func bytes2uint64(param []byte) (uint64, error) {
	integer, err := strconv.ParseUint(hex.EncodeToString(param), 16, 64)
	if err != nil {
		return 0, err
	}
	return integer, nil
}

func bytes2uint32(param []byte) (uint32, error) {
	integer, err := strconv.ParseUint(hex.EncodeToString(param), 16, 32)
	if err != nil {
		return 0, err
	}
	return uint32(integer), nil
}

func bytes2int64(param []byte) (int64, error) {
	integer, err := strconv.ParseInt(hex.EncodeToString(param), 16, 64)
	if err != nil {
		return 0, err
	}
	return integer, nil
}

func bytes2int32(param []byte) (int32, error) {
	integer, err := strconv.ParseUint(hex.EncodeToString(param), 16, 32)
	if err != nil {
		return 0, err
	}
	return int32(integer), nil
}

func uint32To32Bytes(val uint32) []byte {
	byteArr := make([]byte, 4)
	binary.BigEndian.PutUint32(byteArr, val)

	paddedByteArr := addPaddingTo32Bytes(byteArr)

	return paddedByteArr
}

func addPaddingTo32Bytes(byteArr []byte) []byte {
	// Pad the byte array with leading zeros to get the desired length
	paddedByteArr := make([]byte, 32)
	copy(paddedByteArr[32-len(byteArr):], byteArr)

	return paddedByteArr
}

func bytes2cosmosAddr(param []byte) (cosmostypes.AccAddress, error) {
	addrBech32, err := hex2bech32(hex.EncodeToString(param), TypeKiraAddr)
	if err != nil {
		return nil, err
	}

	addr, err := cosmostypes.AccAddressFromBech32(addrBech32)
	if err != nil {
		return nil, err
	}

	return addr, nil
}

func string2cosmosAddr(param string) (cosmostypes.AccAddress, error) {
	addr, err := cosmostypes.AccAddressFromBech32(param)

	return addr, err
}

func bytes2cosmosValAddr(param []byte) (cosmostypes.ValAddress, error) {
	addr, err := cosmostypes.ValAddressFromHex("0x" + hex.EncodeToString(param))
	if err != nil {
		return nil, err
	}

	return addr, nil
}

func bytes2bool(param []byte) bool {
	return (param[len(param)-1] == 0x01)
}

func PackABIParams(abi abi.ABI, name string, args ...interface{}) ([]byte, error) {
	method, exist := abi.Methods[name]
	if !exist {
		return nil, fmt.Errorf("method '%s' not found", name)
	}
	arguments, err := method.Inputs.Pack(args...)
	if err != nil {
		return nil, err
	}
	// Pack up the method ID too if not a constructor and return
	return arguments, nil
}

func convertByteArr2Bytes32(val []byte) [32]byte {
	var returnVal [32]byte
	copy(returnVal[:], val)
	return returnVal
}

func getSignature(r, s, v []byte) []byte {
	sig := make([]byte, 65)
	copy(sig[:32], r)
	copy(sig[32:64], s)
	copy(sig[64:], v)
	sig[64] = sig[64] - 27
	return sig
}

func getSignatureV2(r, s, v []byte) []byte {
	sig := make([]byte, 65)
	copy(sig[:32], r)
	copy(sig[32:64], s)
	copy(sig[64:], v)
	return sig
}
