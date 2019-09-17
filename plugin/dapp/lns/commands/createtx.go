package commands

import (
	"fmt"
	"os"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	lnsty "github.com/33cn/plugin/plugin/dapp/lns/types"
	"github.com/spf13/cobra"

	cmdtypes "github.com/33cn/chain33/system/dapp/commands/types"
)

func openChannelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open",
		Short: "open channel",
		Run:   openChannel,
	}
	cmd.Flags().StringP("assetExec", "e", "coins", "asset issue contract, default coins")
	cmd.Flags().StringP("assetSymbol", "s", "bty", "asset symbol, default bty")
	cmd.Flags().StringP("partner", "p", "", "partner addr")
	cmd.Flags().Int32P("settleTimeout", "t", 0, "settle timeout block num, default min settle block num")
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

	payLoad, err := types.PBToJSON(open)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ErrPbToJson:"+err.Error())
		return
	}
	create := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(lnsty.LnsX),
		ActionName: lnsty.NameOpenAction,
		Payload:    payLoad,
	}

	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", create, &res)
	ctx.RunWithoutMarshal()
}

func depositChannelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "depositChan",
		Short: "deposit to channel",
		Run:   depositChannel,
	}

	cmd.Flags().Int64P("channelID", "c", 0, "channel id")
	cmd.Flags().Float64P("totalDeposit", "a", 0, "total deposit amount")
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

	payLoad, err := types.PBToJSON(deposit)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ErrPbToJson:"+err.Error())
		return
	}
	create := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(lnsty.LnsX),
		ActionName: lnsty.NameDepositChannelAction,
		Payload:    payLoad,
	}

	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", create, &res)
	ctx.RunWithoutMarshal()
}

func withdrawChannelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdrawChan",
		Short: "withdraw from channel",
		Run:   withdrawChannel,
	}

	cmd.Flags().Int64P("channelID", "c", 0, "channel id")
	cmd.Flags().StringP("proof", "d", "", "withdraw proof hex data")
	cmd.Flags().StringP("partnerSign", "p", "", "partner signature hex data")

	cmd.MarkFlagRequired("channelID")
	cmd.MarkFlagRequired("proof")
	cmd.MarkFlagRequired("partnerSign")
	return cmd
}

func withdrawChannel(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	channelID, _ := cmd.Flags().GetInt64("channelID")
	proof, _ := cmd.Flags().GetString("proof")
	ps, _ := cmd.Flags().GetString("partnerSign")

	proofByte, _ := common.FromHex(proof)
	psByte, _ := common.FromHex(ps)
	withdrawProof := &lnsty.WithdrawConfirmProof{}
	sign2 := &types.Signature{}
	err2 := types.Decode(psByte, sign2)
	err3 := types.Decode(proofByte, withdrawProof)

	if err2 != nil || err3 != nil {
		fmt.Fprintln(os.Stderr, "ErrParams-", "withdrawProofErr:", err3, "-partnerSignErr:", err2)
		return
	}

	withdraw := &lnsty.WithdrawChannel{
		ChannelID:           channelID,
		Proof:               withdrawProof,
		PartnerSignature:    sign2,
	}

	payLoad, err := types.PBToJSON(withdraw)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ErrPbToJson:"+err.Error())
		return
	}
	create := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(lnsty.LnsX),
		ActionName: lnsty.NameWithdrawChannelAction,
		Payload:    payLoad,
	}

	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", create, &res)
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

	payLoad, err := types.PBToJSON(close)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ErrPbToJson:"+err.Error())
		return
	}
	create := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(lnsty.LnsX),
		ActionName: lnsty.NameCloseAction,
		Payload:    payLoad,
	}

	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", create, &res)
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

	payLoad, err := types.PBToJSON(update)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ErrPbToJson:"+err.Error())
		return
	}
	create := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(lnsty.LnsX),
		ActionName: lnsty.NameUpdateProofAction,
		Payload:    payLoad,
	}

	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", create, &res)
	ctx.RunWithoutMarshal()
}

func settleChannelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "settle",
		Short: "settle channel",
		Run:   settleChannel,
	}

	cmd.Flags().Int64P("channelID", "c", 0, "channel id")
	cmd.Flags().Float64P("selfTransAmount", "s", 0, "self transferred amount")
	cmd.Flags().Float64P("partnerTransAmount", "p", 0, "partner transferred amount")

	cmd.MarkFlagRequired("channelID")
	cmd.MarkFlagRequired("selfTransferredAmount")
	cmd.MarkFlagRequired("partnerTransferredAmount")
	return cmd
}

func settleChannel(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	channelID, _ := cmd.Flags().GetInt64("channelID")
	selfAmount, _ := cmd.Flags().GetFloat64("selfTransAmount")
	partnerAmount, _ := cmd.Flags().GetFloat64("partnerTransAmount")

	selfAmountInt64 := cmdtypes.FormatAmountDisplay2Value(selfAmount)
	partnerAmountInt64 := cmdtypes.FormatAmountDisplay2Value(partnerAmount)

	settle := &lnsty.Settle{
		ChannelID:                channelID,
		SelfTransferredAmount:    selfAmountInt64,
		PartnerTransferredAmount: partnerAmountInt64,
	}

	payLoad, err := types.PBToJSON(settle)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ErrPbToJson:"+err.Error())
		return
	}
	create := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(lnsty.LnsX),
		ActionName: lnsty.NameSettleAction,
		Payload:    payLoad,
	}

	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", create, &res)
	ctx.RunWithoutMarshal()
}
