package interx

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/KiraCore/interx/config"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/go-bip39"
	"google.golang.org/grpc"

	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	FAUCET_ADDRESS   = "kira16gjh36qaxr2u2ltsraqnvvg2wywtk0gdug2r2u"
	RECEIVER_ADDRESS = "kira15xt4fm40hl50ufdf9sxmdlkn9cn5w9gk053y3l"
	TOKEN            = "ukex"
	CLAIM_AMOUNT     = 10000
	PREFIX           = "kira"
	MENEMONIC        = "marine code consider stuff paddle junk pond reduce undo they hamster rubber cereal purpose practice own early blast pipe match agent baby nice fetch"
)

func TestSendingTransaction1(t *testing.T) {

	sdk.GetConfig().SetBech32PrefixForAccount(PREFIX, sdk.PrefixPublic)

	txBuilder := config.EncodingCg.TxConfig.NewTxBuilder()
	msg := banktypes.NewMsgSend(sdk.MustAccAddressFromBech32(FAUCET_ADDRESS), sdk.MustAccAddressFromBech32(RECEIVER_ADDRESS), sdk.NewCoins(sdk.NewInt64Coin(TOKEN, CLAIM_AMOUNT)))
	err := txBuilder.SetMsgs(msg)
	if err != nil {
		fmt.Println("tx-builder error", err)
	}

	feeAmount := sdk.NewCoins(sdk.NewInt64Coin("ukex", 100))
	memo := "test api"
	gasLimit := uint64(200000)
	chainID := "localnet-1"

	txBuilder.SetGasLimit(gasLimit)
	txBuilder.SetFeeAmount(feeAmount)
	txBuilder.SetMemo(memo)
	txBuilder.SetTimeoutHeight(0)
	//
	mnemonic := MENEMONIC
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	if err != nil {
		panic(err)
	}
	master, ch := hd.ComputeMastersFromSeed(seed)
	priv, _ := hd.DerivePrivateKeyForPath(master, ch, "44'/118'/0'/0/0")
	config.Config.Faucet.PrivKey = &secp256k1.PrivKey{Key: priv}

	privs := []cryptotypes.PrivKey{config.Config.Faucet.PrivKey}
	// First round: we gather all the signer infos. We use the "set empty signature" hack to do that.
	var sigsV2 []signing.SignatureV2
	for _, priv := range privs {
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  config.EncodingCg.TxConfig.SignModeHandler().DefaultMode(),
				Signature: nil,
			},
			Sequence: 8,
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	txBuildErr := txBuilder.SetSignatures(sigsV2...)
	if txBuildErr != nil {
		fmt.Println("Test failed: first round signing failed")
		return
	}

	// Second round: all signer infos are set, so each signer can sign.
	sigsV2 = []signing.SignatureV2{}
	for _, priv := range privs {
		signerData := authsigning.SignerData{
			ChainID:       chainID,
			AccountNumber: 5,
			Sequence:      8,
		}
		sigV2, err := tx.SignWithPrivKey(
			config.EncodingCg.TxConfig.SignModeHandler().DefaultMode(), signerData,
			txBuilder, priv, config.EncodingCg.TxConfig, 5)
		if err != nil {
			fmt.Println("Test failed: first round signing failed")
			return
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err = txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		fmt.Println("Test failed: first round signing failed")
		return
	}

	// Generated Protobuf-encoded bytes.
	txBytes, err := config.EncodingCg.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		fmt.Println("Test failed: first round signing failed", txBytes)
		return
	}

	grpcConn, _ := grpc.Dial(
		"0.0.0.0:9090",      // Or your gRPC server address.
		grpc.WithInsecure(), // The Cosmos SDK doesn't support any transport security mechanism.
	)
	defer grpcConn.Close()

	// Simulation
	txClient1 := txtypes.NewServiceClient(grpcConn)
	grpcRes1, txClientErr := txClient1.Simulate(
		context.Background(),
		&txtypes.SimulateRequest{
			TxBytes: txBytes,
		},
	)
	if txClientErr != nil {
		fmt.Println("Test failed: first round signing failed", txClientErr)
		return
	}
	fmt.Println("gas info ---->", grpcRes1.GasInfo)

}

func AccountNumber(inputAddress string) (accountNum uint64, sequence uint64) {

	sdk.GetConfig().SetBech32PrefixForAccount("kira", sdk.PrefixPublic)

	grpcConn, err := grpc.Dial("0.0.0.0:9090", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to gRPC server: %v", err)
	}
	defer grpcConn.Close()

	// Set up a query client for account info
	queryClient := authtypes.NewQueryClient(grpcConn)

	// Replace with your address

	accountAddr, err := sdk.AccAddressFromBech32(inputAddress)
	if err != nil {
		log.Fatalf("failed to parse address: %v", err)
	}

	// Query the account
	accountRes, err := queryClient.Account(context.Background(), &authtypes.QueryAccountRequest{Address: accountAddr.String()})
	if err != nil {
		log.Fatalf("failed to query account: %v", err)
	}

	// Extract account information
	var account authtypes.BaseAccount
	err = account.Unmarshal(accountRes.Account.Value)
	if err != nil {
		log.Fatalf("failed to unmarshal account: %v", err)
	}

	fmt.Printf("Account Number: %d\n", account.AccountNumber)
	fmt.Printf("Sequence: %d\n", account.Sequence)
	return account.AccountNumber, account.Sequence
}
