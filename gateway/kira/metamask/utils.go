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

func intToBytes(num int64) []byte {
	byteArr := make([]byte, 8) // Assuming int64, which is 8 bytes
	binary.LittleEndian.PutUint64(byteArr, uint64(num))
	return byteArr
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

func string2cosmosAddr(param []byte) (cosmostypes.AccAddress, error) {
	addr, err := cosmostypes.AccAddressFromBech32(string(param))

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
