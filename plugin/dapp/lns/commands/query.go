package commands

import (
	"fmt"
	"os"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	lnsty "github.com/33cn/plugin/plugin/dapp/lns/types"
	"github.com/spf13/cobra"
)

func queryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "query lns",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(queryChannelCmd(), queryChannelCountCmd())
	return cmd
}

func queryChannelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chan",
		Short: "query channel",
		Run:   queryChannel,
	}

	cmd.Flags().Int64P("channelID", "c", 0, "channel id")
	cmd.MarkFlagRequired("channelID")
	return cmd
}

func queryChannel(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	channelID, _ := cmd.Flags().GetInt64("channelID")

	get := &lnsty.ReqGetChannel{
		ChannelID: channelID,
	}

	payLoad, err := types.PBToJSON(get)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ErrPbToJson:"+err.Error())
		return
	}

	query := rpctypes.Query4Jrpc{
		Execer:   lnsty.LnsX,
		FuncName: lnsty.FuncQueryGetChannel,
		Payload:  payLoad,
	}

	channel := &lnsty.Channel{}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", query, channel)
	ctx.Run()
}

func queryChannelCountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chanCount",
		Short: "query current channel count",
		Run:   queryChannelCount,
	}
	return cmd
}

func queryChannelCount(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	query := rpctypes.Query4Jrpc{
		Execer:   lnsty.LnsX,
		FuncName: lnsty.FuncQueryGetChannelCount,
	}

	channel := &lnsty.ChannelCount{}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", query, channel)
	ctx.Run()
}
