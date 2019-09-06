/*Package commands implement dapp client commands*/
package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	lnsty "github.com/33cn/plugin/plugin/dapp/lns/types"
	"github.com/spf13/cobra"

	cmdtypes "github.com/33cn/chain33/system/dapp/commands/types"
)

/*
 * 实现合约对应客户端
 */

// Cmd lns client command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lns",
		Short: "lns command",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
		openChannelCmd(),
		depositChannelCmd(),
		withdrawChannelCmd(),
		closeChannelCmd(),
		updateChannelProofCmd(),
		settleChannelCmd(),
		proofCmd(),
	)
	return cmd
}

func openChannelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open",
		Short: "open channel",
		Run:   openChannel,
	}
	cmd.Flags().StringP("assetExec", "e", "coins", "asset issue contract, default coins")
	cmd.Flags().StringP("assetSymbol", "s", "", "asset symbol, default coins symbol")
	cmd.Flags().StringP("partner", "p", "", "partner addr")
	cmd.Flags().Int32P("settleTimeout", "t", 100, "settle timeout block num, default 100 block num")
	cmd.Flags().Float64P("amount", "a", 0, "initial deposit amount")
	cmd.MarkFlagRequired("partner")
	return cmd
}

func openChannel(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	exec, _ := cmd.Flags().GetString("assetExec")
	symbol, _ := cmd.Flags().GetString("assetSymbol")
	partner, _ := cmd.Flags().GetString("partner")
	timeout, _ := cmd.Flags().GetInt32("settleTimeout")
	amount, _ := cmd.Flags().GetFloat64("amount")
	amountInt64 := cmdtypes.FormatAmountDisplay2Value(amount)

	if exec == "" || (exec != "coins" && symbol == "") {
		fmt.Fprintln(os.Stderr, "ErrAssetExecOrSymbolParams")
		return
	}
	if address.CheckAddress(partner) != nil {
		fmt.Fprintln(os.Stderr, "ErrPartnerAddress")
		return
	}

	open := &lnsty.OpenChannel{
		IssueContract: exec,
		TokenSymbol:   symbol,
		Partner:       partner,
		SettleTimeout: timeout,
		Amount:        amountInt64,
	}

	payLoad, err := json.Marshal(open)
	if err != nil {
		return
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(lnsty.LnsX),
		ActionName: lnsty.NameOpenAction,
		Payload:    payLoad,
	}

	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}

func depositChannelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "depositChan",
		Short: "deposit to channel",
		Run:   depositChannel,
	}

	cmd.Flags().Int64P("channelID", "c", 0, "channel id")
	cmd.Flags().Float64P("totalDeposit", "t", 0, "total deposit")
	cmd.MarkFlagRequired("channelID")
	cmd.MarkFlagRequired("totalDeposit")
	return cmd
}

func depositChannel(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	totalDeposit, _ := cmd.Flags().GetFloat64("totalDeposit")
	channelID, _ := cmd.Flags().GetInt64("channelID")
	amountInt64 := cmdtypes.FormatAmountDisplay2Value(totalDeposit)

	if amountInt64 <= 0 {
		fmt.Fprintln(os.Stderr, "ErrTotalDepositAmount")
		return
	}

	if channelID <= 0 {
		fmt.Fprintln(os.Stderr, "ErrChannelID")
		return
	}

	deposit := &lnsty.DepositChannel{
		TotalDeposit: amountInt64,
		ChannelID:    channelID,
	}

	payLoad, err := json.Marshal(deposit)
	if err != nil {
		return
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(lnsty.LnsX),
		ActionName: lnsty.NameDepositChannelAction,
		Payload:    payLoad,
	}

	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}

func proofCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proof",
		Short: "create or sign proof",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		withdrawProofCmd(),
		balanceProofCmd(),

		signProofCmd(),
	)

	return cmd
}

func withdrawProofCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw",
		Short: "create withdraw proof",
		Run:   withdrawProof,
	}

	cmd.Flags().Int64P("channelID", "c", 0, "channel id")
	cmd.Flags().StringP("withdrawer", "w", "", "withdrawer addr")
	cmd.Flags().Float64P("totalWithdraw", "t", 0, "total withdraw")
	cmd.Flags().Int64P("expiration", "e", 0, "withdraw expiration block")

	cmd.MarkFlagRequired("channelID")
	cmd.MarkFlagRequired("totalWithdraw")
	cmd.MarkFlagRequired("withrawer")
	cmd.MarkFlagRequired("expiration")

	return cmd
}

func withdrawProof(cmd *cobra.Command, args []string) {

	withdrawer, _ := cmd.Flags().GetString("withdrawer")
	totalWithdraw, _ := cmd.Flags().GetFloat64("totalWithdraw")
	channelID, _ := cmd.Flags().GetInt64("channelID")
	expiration, _ := cmd.Flags().GetInt64("expiration")
	amountInt64 := cmdtypes.FormatAmountDisplay2Value(totalWithdraw)

	if address.CheckAddress(withdrawer) != nil {
		fmt.Fprintln(os.Stderr, "ErrWithdrawerAddress")
		return
	}
	if amountInt64 <= 0 {
		fmt.Fprintln(os.Stderr, "ErrTotalWithdrawAmount")
		return
	}

	if channelID <= 0 {
		fmt.Fprintln(os.Stderr, "ErrChannelID")
		return
	}
	proof := &lnsty.WithdrawConfirmProof{
		ChannelID:       channelID,
		Withdrawer:      withdrawer,
		TotalWithdraw:   amountInt64,
		ExpirationBlock: expiration,
	}
	fmt.Println(common.ToHex(types.Encode(proof)))
}

func balanceProofCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance",
		Short: "create balance proof",
		Run:   balanceProof,
	}

	cmd.Flags().Int64P("channelID", "c", 0, "channel id")
	cmd.Flags().Int64P("nonce", "n", 0, "transfer nonce")
	cmd.Flags().Float64P("amount", "a", 0, "transferred amount")

	cmd.MarkFlagRequired("channelID")
	cmd.MarkFlagRequired("nonce")
	cmd.MarkFlagRequired("amount")

	return cmd
}

func balanceProof(cmd *cobra.Command, args []string) {
	transferredAmount, _ := cmd.Flags().GetFloat64("amount")
	channelID, _ := cmd.Flags().GetInt64("channelID")
	nonce, _ := cmd.Flags().GetInt64("nonce")
	amountInt64 := cmdtypes.FormatAmountDisplay2Value(transferredAmount)

	if channelID <= 0 {
		fmt.Fprintln(os.Stderr, "ErrChannelID")
		return
	}

	proof := &lnsty.BalanceProof{
		Nonce:             nonce,
		TransferredAmount: amountInt64,
		ChannelID:         channelID,
	}
	fmt.Println(common.ToHex(types.Encode(proof)))
}

func signProofCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign",
		Short: "sign proof with key",
		Run:   signProof,
	}
	cmd.Flags().StringP("proofData", "d", "", "raw proof hex data")
	cmd.Flags().StringP("key", "k", "", "sign key or address")
	cmd.MarkFlagRequired("proofData")
	cmd.MarkFlagRequired("key")
	return cmd
}

func signProof(cmd *cobra.Command, args []string) {

	data, _ := cmd.Flags().GetString("proofData")
	key, _ := cmd.Flags().GetString("key")
	privHex := key
	if address.CheckAddress(key) == nil {
		rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
		params := types.ReqString{
			Data: key,
		}
		var res types.ReplyString
		ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.DumpPrivkey", params, &res)
		result, err := ctx.RunResult()
		if err != nil {
			fmt.Fprintln(os.Stderr, "ErrGetPrivKey:"+err.Error())
			return
		}
		val, ok := result.(*types.ReplyString)
		if !ok {
			fmt.Fprintln(os.Stderr, "ErrGetPrivKey")
			return
		}
		privHex = val.Data
	}

	keyByte, err := common.FromHex(privHex)
	if err != nil || len(keyByte) == 0 {
		fmt.Fprintln(os.Stderr, "ErrPrivKeyFromHex")
		return
	}
	cr, err := crypto.New(crypto.GetName(types.SECP256K1))
	if err != nil {
		fmt.Fprintln(os.Stderr, "ErrNewSECP256K1Crypto:"+err.Error())
		return
	}
	priv, err := cr.PrivKeyFromBytes(keyByte)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ErrNewPrivFromBytes"+err.Error())
		return
	}

	pub := priv.PubKey()
	byteData, err := common.FromHex(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ErrProofData:"+err.Error())
		return
	}
	sign := priv.Sign(common.Sha256(byteData))
	fmt.Println(common.ToHex(types.Encode(&types.Signature{
		Ty:        types.SECP256K1, //需要与钱包生成私钥的椭圆曲线的方法一致
		Pubkey:    pub.Bytes(),
		Signature: sign.Bytes(),
	})))
}

func withdrawChannelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdrawChan",
		Short: "withdraw from channel",
		Run:   withdrawChannel,
	}

	cmd.Flags().Int64P("channelID", "c", 0, "channel id")
	cmd.Flags().StringP("proof", "d", "", "withdraw proof hex data")

	cmd.Flags().StringP("withdrawerSign", "w", "", "withdrawer signature hex data")
	cmd.Flags().StringP("partnerSign", "p", "", "partner signature hex data")

	cmd.MarkFlagRequired("channelID")
	cmd.MarkFlagRequired("proof")
	cmd.MarkFlagRequired("withdrawerSign")
	cmd.MarkFlagRequired("partnerSign")
	return cmd
}

func withdrawChannel(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	channelID, _ := cmd.Flags().GetInt64("channelID")
	proof, _ := cmd.Flags().GetString("proof")
	ws, _ := cmd.Flags().GetString("withdrawerSign")
	ps, _ := cmd.Flags().GetString("partnerSign")

	proofByte, _ := common.FromHex(proof)
	wsByte, _ := common.FromHex(ws)
	psByte, _ := common.FromHex(ps)
	withdrawProof := &lnsty.WithdrawConfirmProof{}
	sign1 := &types.Signature{}
	sign2 := &types.Signature{}
	err1 := types.Decode(wsByte, sign1)
	err2 := types.Decode(psByte, sign2)
	err3 := types.Decode(proofByte, withdrawProof)

	if err1 != nil || err2 != nil || err3 != nil {
		fmt.Fprintln(os.Stderr, "ErrParams-", "withdrawProofErr:", err3,
			"-withdrawSignErr:", err1, "-partnerSignErr:", err2)
		return
	}

	withdraw := &lnsty.WithdrawChannel{
		ChannelID:           channelID,
		Proof:               withdrawProof,
		WithdrawerSignature: sign1,
		PartnerSignature:    sign2,
	}

	payLoad, err := types.PBToJSON(withdraw)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ErrPbToJson:"+err.Error())
		return
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(lnsty.LnsX),
		ActionName: lnsty.NameWithdrawChannelAction,
		Payload:    payLoad,
	}

	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}

func closeChannelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close",
		Short: "close channel",
		Run:   closeChannel,
	}

	cmd.Flags().Int64P("channelID", "c", 0, "channel id")
	cmd.Flags().StringP("proof", "d", "", "balance proof hex data")
	cmd.Flags().StringP("sign", "s", "", "non closer signature hex data")

	cmd.MarkFlagRequired("channelID")
	cmd.MarkFlagRequired("proof")
	cmd.MarkFlagRequired("sign")
	return cmd
}

func closeChannel(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	proof, _ := cmd.Flags().GetString("proof")
	channelID, _ := cmd.Flags().GetInt64("channelID")
	sign, _ := cmd.Flags().GetString("sign")

	proofByte, _ := common.FromHex(proof)
	signByte, _ := common.FromHex(sign)
	balanceProof := &lnsty.BalanceProof{}
	nonCloserSign := &types.Signature{}
	err1 := types.Decode(proofByte, balanceProof)
	err2 := types.Decode(signByte, nonCloserSign)

	if err1 != nil || err2 != nil {
		fmt.Fprintln(os.Stderr, "ErrParams-", "BalanceProofErr:", err1,
			"-nonCloserSignErr:", err2)
		return
	}

	close := &lnsty.CloseChannel{
		ChannelID:          channelID,
		NonCloserBalancePf: balanceProof,
		NonCloserSignature: nonCloserSign,
	}

	payLoad, err := json.Marshal(close)
	if err != nil {
		return
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(lnsty.LnsX),
		ActionName: lnsty.NameCloseAction,
		Payload:    payLoad,
	}

	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}

func updateChannelProofCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "updateProof",
		Short: "update channel balance proof",
		Run:   updateChannelProof,
	}

	cmd.Flags().Int64P("channelID", "c", 0, "channel id")
	cmd.Flags().StringP("proof", "d", "", "balance proof hex data")
	cmd.Flags().StringP("sign", "s", "", "partner signature hex data")

	cmd.MarkFlagRequired("channelID")
	cmd.MarkFlagRequired("proof")
	cmd.MarkFlagRequired("sign")
	return cmd
}

func updateChannelProof(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	proof, _ := cmd.Flags().GetString("proof")
	channelID, _ := cmd.Flags().GetInt64("channelID")
	sign, _ := cmd.Flags().GetString("sign")

	proofByte, _ := common.FromHex(proof)
	signByte, _ := common.FromHex(sign)
	balanceProof := &lnsty.BalanceProof{}
	partnerSign := &types.Signature{}
	err1 := types.Decode(proofByte, balanceProof)
	err2 := types.Decode(signByte, partnerSign)

	if err1 != nil || err2 != nil {
		fmt.Fprintln(os.Stderr, "ErrParams-", "BalanceProofErr:", err1,
			"-partnerSignErr:", err2)
		return
	}

	update := &lnsty.UpdateBalanceProof{
		ChannelID:        channelID,
		PartnerBalancePf: balanceProof,
		PartnerSignature: partnerSign,
	}

	payLoad, err := json.Marshal(update)
	if err != nil {
		return
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(lnsty.LnsX),
		ActionName: lnsty.NameUpdateProofAction,
		Payload:    payLoad,
	}

	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}

func settleChannelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "settle",
		Short: "settle channel",
		Run:   settleChannel,
	}

	cmd.Flags().Int64P("channelID", "c", 0, "channel id")
	cmd.Flags().Float64P("selfTransferredAmount", "s", 0, "self transferred amount")
	cmd.Flags().Float64P("partnerTransferredAmount", "p", 0, "partner transferred amount")

	cmd.MarkFlagRequired("channelID")
	cmd.MarkFlagRequired("selfTransferredAmount")
	cmd.MarkFlagRequired("partnerTransferredAmount")
	return cmd
}

func settleChannel(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	channelID, _ := cmd.Flags().GetInt64("channelID")
	selfAmount, _ := cmd.Flags().GetFloat64("selfTransferredAmount")
	partnerAmount, _ := cmd.Flags().GetFloat64("partnerTransferredAmount")

	selfAmountInt64 := cmdtypes.FormatAmountDisplay2Value(selfAmount)
	partnerAmountInt64 := cmdtypes.FormatAmountDisplay2Value(partnerAmount)

	settle := &lnsty.Settle{
		ChannelID:                channelID,
		SelfTransferredAmount:    selfAmountInt64,
		PartnerTransferredAmount: partnerAmountInt64,
	}

	payLoad, err := json.Marshal(settle)
	if err != nil {
		return
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(lnsty.LnsX),
		ActionName: lnsty.NameSettleAction,
		Payload:    payLoad,
	}

	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}
