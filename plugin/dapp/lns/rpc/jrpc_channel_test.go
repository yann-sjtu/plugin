package rpc_test

import (
	"testing"

	"github.com/33cn/chain33/common/log"
	"github.com/33cn/chain33/rpc/jsonclient"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util/testnode"
	"github.com/stretchr/testify/assert"

	rpctypes "github.com/33cn/chain33/rpc/types"
	lnsty "github.com/33cn/plugin/plugin/dapp/lns/types"

	_ "github.com/33cn/chain33/system"
	_ "github.com/33cn/plugin/plugin"
)

func init() {
	log.SetLogLevel("error")
}

func testOpenChanTx(t *testing.T, cli *jsonclient.JSONClient) error {

	open := &lnsty.OpenChannel{}

	payLoad, err := types.PBToJSON(open)
	if err != nil {
		return err
	}
	create := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(lnsty.LnsX),
		ActionName: lnsty.NameOpenAction,
		Payload:    payLoad,
	}

	var res string
	return cli.Call("Chain33.CreateTransaction", create, &res)
}

func testDepositChanTx(t *testing.T, cli *jsonclient.JSONClient) error {

	params := &lnsty.DepositChannel{}

	payLoad, err := types.PBToJSON(params)
	if err != nil {
		return err
	}
	create := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(lnsty.LnsX),
		ActionName: lnsty.NameDepositChannelAction,
		Payload:    payLoad,
	}

	var res string
	return cli.Call("Chain33.CreateTransaction", create, &res)
}

func testWithdrawChanTx(t *testing.T, cli *jsonclient.JSONClient) error {

	params := &lnsty.WithdrawChannel{}

	payLoad, err := types.PBToJSON(params)
	if err != nil {
		return err
	}
	create := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(lnsty.LnsX),
		ActionName: lnsty.NameWithdrawChannelAction,
		Payload:    payLoad,
	}

	var res string
	return cli.Call("Chain33.CreateTransaction", create, &res)
}

func testCloseChanTx(t *testing.T, cli *jsonclient.JSONClient) error {

	params := &lnsty.CloseChannel{}

	payLoad, err := types.PBToJSON(params)
	if err != nil {
		return err
	}
	create := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(lnsty.LnsX),
		ActionName: lnsty.NameCloseAction,
		Payload:    payLoad,
	}

	var res string
	return cli.Call("Chain33.CreateTransaction", create, &res)
}

func testUpdateProofTx(t *testing.T, cli *jsonclient.JSONClient) error {

	params := &lnsty.UpdateBalanceProof{}

	payLoad, err := types.PBToJSON(params)
	if err != nil {
		return err
	}
	create := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(lnsty.LnsX),
		ActionName: lnsty.NameUpdateProofAction,
		Payload:    payLoad,
	}

	var res string
	return cli.Call("Chain33.CreateTransaction", create, &res)
}

func testSettleChanTx(t *testing.T, cli *jsonclient.JSONClient) error {

	params := &lnsty.Settle{}

	payLoad, err := types.PBToJSON(params)
	if err != nil {
		return err
	}
	create := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(lnsty.LnsX),
		ActionName: lnsty.NameSettleAction,
		Payload:    payLoad,
	}

	var res string
	return cli.Call("Chain33.CreateTransaction", create, &res)
}

func testQueryChannel(t *testing.T, cli *jsonclient.JSONClient) error {

	params := &lnsty.ReqGetChannel{ChannelID: 100}

	payLoad, err := types.PBToJSON(params)
	if err != nil {
		return err
	}
	query := rpctypes.Query4Jrpc{
		Execer:   lnsty.LnsX,
		FuncName: lnsty.FuncQueryGetChannel,
		Payload:  payLoad,
	}

	channel := &lnsty.Channel{}
	return cli.Call("Chain33.Query", query, channel)
}

func testQueryChannelCount(t *testing.T, cli *jsonclient.JSONClient) error {

	params := &types.ReqNil{}

	payLoad, err := types.PBToJSON(params)
	if err != nil {
		return err
	}
	query := rpctypes.Query4Jrpc{
		Execer:   lnsty.LnsX,
		FuncName: lnsty.FuncQueryGetChannelCount,
		Payload:  payLoad,
	}

	channel := &lnsty.ChannelCount{}
	return cli.Call("Chain33.Query", query, channel)
}

func TestJRPCChannel(t *testing.T) {
	// 启动RPCmocker
	mocker := testnode.New("--notset--", nil)
	defer func() {
		mocker.Close()
	}()
	mocker.Listen()

	jrpcClient := mocker.GetJSONC()
	assert.NotNil(t, jrpcClient)

	testCases := []struct {
		fn  func(*testing.T, *jsonclient.JSONClient) error
		err error
	}{
		{fn: testOpenChanTx},
		{fn: testDepositChanTx},
		{fn: testWithdrawChanTx},
		{fn: testCloseChanTx},
		{fn: testUpdateProofTx},
		{fn: testSettleChanTx},
		{fn: testQueryChannel, err: types.ErrNotFound},
		{fn: testQueryChannelCount, err: types.ErrNotFound},
	}

	for index, testCase := range testCases {
		err := testCase.fn(t, jrpcClient)
		assert.Equalf(t, testCase.err, err, "test case index %d", index)
	}
}
