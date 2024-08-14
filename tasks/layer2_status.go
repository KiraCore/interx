package tasks

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/database"
	layer2types "github.com/KiraCore/sekai/x/layer2/types"
	cosmostypes "github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	legacytx "github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"

	// banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	types "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

func storeL2Status(gwCosmosmux *runtime.ServeMux, isLog bool) {
	for l2AppName, conf := range config.Config.Layer2 {
		rpcAddr := conf.RPC
		url := fmt.Sprintf("%s/api/v1/status", rpcAddr)
		resp, err := http.Get(url)
		if err != nil {
			common.GetLogger().Error("[layer2-status] Unable to connect to ", url)
			return
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			common.GetLogger().Error("[layer2-status] Unable to read status export result", err)
			return
		}

		database.SetLayer2State("chess", string(bodyBytes))

		if isLog {
			common.GetLogger().Info("[layer2-status] (new state): ", time.Now().UTC())
		}

		common.Layer2Status[l2AppName] = string(bodyBytes)

		r, err := http.NewRequest("GET", getInterxRPC(), strings.NewReader(""))
		if err != nil {
			common.GetLogger().Error("[layer2]: ", err)
			return
		}

		// If leader of session
		registrar := common.GetExecutionRegistrar(gwCosmosmux, r, l2AppName)

		// Workflow
		// if current session is empty and next session leader is myself, send MsgExecuteDappTx (this is needed at start phase)
		// if current session leader is myself and status is scheduled, send MsgExecuteDappTx (this is needed when session's accepted and next session moved to current session)
		// if current session leader is myself and status is on-going, send MsgTransitionDappTx
		// if current session looks okay, send MsgApproveDappTransitionTx if not, send MsgRejectDappTransitionTx (check registrar.ExecutionRegistrar.CurrSession.Start)
		// check all the state transition flow's working fine
		// how often should it execute approve tx? Since it will let sessions to be moved

		if registrar.ExecutionRegistrar != nil && registrar.ExecutionRegistrar.NextSession != nil {
			currSession := registrar.ExecutionRegistrar.CurrSession
			nextSession := registrar.ExecutionRegistrar.NextSession
			for _, operator := range registrar.Operators {
				if operator.Interx != config.Config.Address {
					continue
				}
				if currSession == nil && operator.Operator == nextSession.Leader {
					msg := &layer2types.MsgExecuteDappTx{
						Sender:   config.Config.Address,
						DappName: l2AppName,
						Gateway:  "interx.com",
					}

					sendTx(msg, conf.Fee, gwCosmosmux, r)
					time.Sleep(time.Second * 5)
				}
			}
		}

		if registrar.ExecutionRegistrar != nil && registrar.ExecutionRegistrar.CurrSession != nil {
			currSession := registrar.ExecutionRegistrar.CurrSession

			for _, operator := range registrar.Operators {
				if operator.Interx != config.Config.Address {
					continue
				}

				if currSession.Leader == operator.Operator {
					// sekai RPC submit
					// msg := &banktypes.MsgSend{
					// 	FromAddress: config.Config.Address,
					// 	ToAddress:   config.Config.Faucet.Address,
					// 	Amount:      sdk.Coins{sdk.NewInt64Coin("ukex", 100)},
					// }
					msg := &layer2types.MsgTransitionDappTx{
						Sender:          config.Config.Address,
						DappName:        l2AppName,
						StatusHash:      common.GetSha256SumFromBytes([]byte(common.Layer2Status[l2AppName] + fmt.Sprint(time.Now().Unix()))),
						OnchainMessages: []*types.Any{},
						Version:         "0cc0", // TODO: to be queried from dapp docker - to be equal to dapp.Bin[0].Hash
					}
					sendTx(msg, conf.Fee, gwCosmosmux, r)
					time.Sleep(time.Second * 5)
				}

				now := time.Now().Unix()
				start, err := strconv.Atoi(currSession.Start)
				if err != nil || start+60 < int(now) { // execute transition once per min for now
					msg := &layer2types.MsgApproveDappTransitionTx{
						Sender:   config.Config.Address,
						DappName: l2AppName,
						Version:  "0cc0", // TODO: to be queried from dapp docker - to be equal to dapp.Bin[0].Hash
					}
					// msg := layer2types.MsgRejectDappTransitionTx{
					// 	Sender:   config.Config.Address,
					// 	DappName: l2AppName,
					// 	Version:  "0cc0", // TODO: to be queried from dapp docker - to be equal to dapp.Bin[0].Hash
					// }
					sendTx(msg, conf.Fee, gwCosmosmux, r)
					time.Sleep(time.Second * 5)
				}
			}
		}
	}
}

func getInterxRPC() string {
	return "http://127.0.0.1:" + config.Config.PORT
}

// StoreL2Status is a function storing layer2 app status
func StoreL2Status(gwCosmosmux *runtime.ServeMux, isLog bool) {
	for {
		storeL2Status(gwCosmosmux, isLog)

		if isLog {
			common.GetLogger().Info("[node-status] Syncing node status")
			common.GetLogger().Info("[node-status] Chain_id = ", common.NodeStatus.Chainid)
			common.GetLogger().Info("[node-status] Block = ", common.NodeStatus.Block)
			common.GetLogger().Info("[node-status] Blocktime = ", common.NodeStatus.Blocktime)
		}

		time.Sleep(time.Duration(config.Config.Block.StatusSync) * time.Second)
	}
}

func sendTx(msg cosmostypes.Msg, feeStr string, gwCosmosmux *runtime.ServeMux, r *http.Request) error {
	accountNumber, sequence := common.GetAccountNumberSequence(gwCosmosmux, r.Clone(r.Context()), config.Config.Address)

	feeAmount, err := sdk.ParseCoinNormalized(feeStr)
	if err != nil {
		return err
	}

	msgs := []sdk.Msg{msg}
	fee := legacytx.NewStdFee(200000, sdk.NewCoins(feeAmount)) //Fee handling
	memo := "layer2-transition"

	// TODO: get it from somewhere
	common.NodeStatus.Chainid = "testing"
	sigs := make([]legacytx.StdSignature, 1)
	signBytes := legacytx.StdSignBytes(common.NodeStatus.Chainid, accountNumber, sequence, 0, fee, msgs, memo, nil)
	sig, err := config.Config.PrivKey.Sign(signBytes)
	if err != nil {
		common.GetLogger().Error("[layer2] Failed to sign transaction: ", err)
		panic(err)
	}

	sigs[0] = legacytx.StdSignature{PubKey: config.Config.PubKey, Signature: sig}

	stdTx := legacytx.NewStdTx(msgs, fee, sigs, memo)

	txBuilder := config.EncodingCg.TxConfig.NewTxBuilder()
	err = txBuilder.SetMsgs(stdTx.GetMsgs()...)
	if err != nil {
		common.GetLogger().Error("[layer2] Failed to set tx msgs: ", err)
		return err
	}

	sigV2, err := stdTx.GetSignaturesV2()
	if err != nil {
		common.GetLogger().Error("[layer2] Failed to get SignatureV2: ", err)
		return err
	}

	sigV2[0].Sequence = sequence

	err = txBuilder.SetSignatures(sigV2...)
	if err != nil {
		common.GetLogger().Error("[layer2] Failed to set SignatureV2: ", err)
		return err
	}

	txBuilder.SetMemo(stdTx.GetMemo())
	txBuilder.SetFeeAmount(stdTx.GetFee())
	txBuilder.SetGasLimit(stdTx.GetGas())

	txBytes, err := config.EncodingCg.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		common.GetLogger().Error("[layer2] Failed to get tx bytes: ", err)
		return err
	}

	txHash, err := common.BroadcastTransactionSync(config.Config.RPC, txBytes)
	if err != nil {
		common.GetLogger().Error("[layer2] Failed to broadcast transaction: ", err)
		return err
	}

	common.GetLogger().Info("txHash: ", txHash)
	return nil
}
