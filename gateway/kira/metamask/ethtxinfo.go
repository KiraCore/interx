package metamask

import (
	"errors"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

type EthTxData struct {
	Nonce    uint64  `json:"nonce"`
	GasPrice big.Int `json:"gas_price"`
	GasLimit uint64  `json:"gas_limit"`
	From     string  `json:"from"`
	To       string  `json:"to"`
	Value    big.Int `json:"value"`
	Data     []byte  `json:"data"`
	Hash     string  `json:"hash"`
	V        big.Int `json:"signature_v"`
	R        big.Int `json:"signature_r"`
	S        big.Int `json:"signature_s"`
}

func GetSenderAddrFromRawTxBytes(rawTxBytes []byte) (common.Address, error) {
	var rawTx ethtypes.Transaction
	if err := rlp.DecodeBytes(rawTxBytes, &rawTx); err != nil {
		return common.Address{}, err
	}

	signer := ethtypes.NewEIP155Signer(rawTx.ChainId())
	sender, err := signer.Sender(&rawTx)
	if err != nil {
		return common.Address{}, err
	}
	return sender, nil
}

func validateTx(rawTxBytes []byte, sender common.Address) bool {
	senderFromTx, err := GetSenderAddrFromRawTxBytes(rawTxBytes)
	if err != nil {
		return false
	}

	if senderFromTx.Hex() == sender.Hex() {
		return true
	}
	return false
}

// SendRawTransaction send a raw Ethereum transaction.
func GetEthTxInfo(data hexutil.Bytes) (EthTxData, error) {
	tx := new(ethtypes.Transaction)
	rlp.DecodeBytes(data, &tx)

	msg, err := tx.AsMessage(ethtypes.NewEIP155Signer(tx.ChainId()), big.NewInt(1))
	if err != nil {
		log.Fatal(err)
		return EthTxData{}, errors.New("decoding transaction data is failed")
	}

	v, r, s := tx.RawSignatureValues()

	// check the local node config in case unprotected txs are disabled
	if !tx.Protected() {
		// Ensure only eip155 signed transactions are submitted if EIP155Required is set.
		return EthTxData{}, errors.New("only replay-protected (EIP-155) transactions allowed over RPC")
	}

	if !validateTx(data, msg.From()) {
		return EthTxData{}, errors.New("validation is failed")
	}

	return EthTxData{
		Nonce:    tx.Nonce(),
		GasPrice: *tx.GasPrice(),
		GasLimit: tx.Gas(),
		From:     msg.From().String(),
		To:       msg.To().String(),
		Value:    *tx.Value(),
		Data:     tx.Data(),
		Hash:     tx.Hash().Hex(),
		V:        *v,
		R:        *r,
		S:        *s,
	}, nil
}
