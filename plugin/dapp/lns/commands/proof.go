package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/rpc/jsonclient"
	"github.com/33cn/chain33/types"
	lnsty "github.com/33cn/plugin/plugin/dapp/lns/types"
	"github.com/spf13/cobra"

	cmdtypes "github.com/33cn/chain33/system/dapp/commands/types"
)

func proofCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proof",
		Short: "create or sign proof",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		withdrawProofCmd(),
		balanceProofCmd(),
		decodeProofCmd(),
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
	cmd.Flags().Int64P("expiration", "b", 0, "withdraw expiration block height")
	cmd.Flags().StringP("title", "t", "bityuan", "chain title, default bityuan")
	cmd.Flags().StringP("assetExec", "e", "coins", "asset executor, default coins")
	cmd.Flags().StringP("assetSymbol", "s", "bty", "asset executor, default bty")

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
	title, _ := cmd.Flags().GetString("title")
	exec, _ := cmd.Flags().GetString("assetExec")
	symbol, _ := cmd.Flags().GetString("assetSymbol")

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
		TokenCanonicalId: &lnsty.TokenCanonicalId{
			Chain:         title,
			IssueContract: exec,
			TokenSymbol:   symbol,
		},
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
	cmd.Flags().StringP("additionHash", "x", "", "addition hash, hex string")
	cmd.Flags().StringP("title", "t", "bityuan", "chain title, default bityuan")
	cmd.Flags().StringP("assetExec", "e", "coins", "asset executor, default coins")
	cmd.Flags().StringP("assetSymbol", "s", "bty", "asset executor, default bty")

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

	additionHashHex, _ := cmd.Flags().GetString("additionHash")
	title, _ := cmd.Flags().GetString("title")
	exec, _ := cmd.Flags().GetString("assetExec")
	symbol, _ := cmd.Flags().GetString("assetSymbol")

	additionHash, err := common.FromHex(additionHashHex)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ErrDecodeHexAdditionHash:"+err.Error())
		return
	}

	if channelID <= 0 {
		fmt.Fprintln(os.Stderr, "ErrChannelID")
		return
	}

	proof := &lnsty.BalanceProof{
		Nonce:             nonce,
		TransferredAmount: amountInt64,
		ChannelID:         channelID,
		TokenCanonicalId: &lnsty.TokenCanonicalId{
			Chain:         title,
			IssueContract: exec,
			TokenSymbol:   symbol,
		},
		AdditionHash: additionHash[:],
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

func decodeProofCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decode",
		Short: "decode hex proof data",
		Run:   decodeProof,
	}
	cmd.Flags().StringP("proofData", "d", "", "raw proof hex data")
	cmd.Flags().StringP("type", "t", "", "proof type,[\"withdraw\" | \"balance\"]")
	cmd.MarkFlagRequired("proofData")
	cmd.MarkFlagRequired("type")
	return cmd
}

func decodeProof(cmd *cobra.Command, args []string) {

	data, _ := cmd.Flags().GetString("proofData")
	proofType, _ := cmd.Flags().GetString("type")

	var proof types.Message
	if proofType == "withdraw" {
		proof = &lnsty.WithdrawConfirmProof{}
	} else if proofType == "balance" {
		proof = &lnsty.BalanceProof{}
	} else {
		fmt.Fprintln(os.Stderr, "ErrParamsProofType")
		return
	}

	proofByte, _ := common.FromHex(data)
	err := types.Decode(proofByte, proof)

	if err != nil {
		fmt.Fprintln(os.Stderr, "ErrDecodeProofData:"+err.Error())
		return
	}

	jsonData, err := json.MarshalIndent(proof, "", "    ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "ErrMarshalProof:"+err.Error())
		return
	}
	fmt.Println(string(jsonData))
}
