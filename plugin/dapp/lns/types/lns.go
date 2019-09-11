package types

import (
	"reflect"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
)

/*
 * 交易相关类型定义
 * 交易action通常有对应的log结构，用于交易回执日志记录
 * 每一种action和log需要用id数值和name名称加以区分
 */

var (
	//LnsX 执行器名称定义
	LnsX = "lns"
	//定义actionMap
	actionMap = map[string]int32{
		NameOpenAction:            TyOpenAction,
		NameDepositChannelAction:  TyDepositAction,
		NameWithdrawChannelAction: TyWithdrawAction,
		NameCloseAction:           TyCloseAction,
		NameUpdateProofAction:     TyUpdateProofAction,
		NameSettleAction:          TySettleAction,
	}
	//定义log的id和具体log类型及名称，填入具体自定义log类型
	logMap = map[int64]*types.LogInfo{
		TyOpenLog:        {Ty: reflect.TypeOf(ReceiptOpen{}), Name: NameOpenLog},
		TyDepositLog:     {Ty: reflect.TypeOf(ReceiptDeposit{}), Name: NameDepositLog},
		TyWithdrawLog:    {Ty: reflect.TypeOf(ReceiptWithdraw{}), Name: NameWithdrawLog},
		TyCloseLog:       {Ty: reflect.TypeOf(ReceiptClose{}), Name: NameCloseLog},
		TyUpdateProofLog: {Ty: reflect.TypeOf(ReceiptUpdate{}), Name: NameUpdateProofLog},
		TySettleLog:      {Ty: reflect.TypeOf(ReceiptSettle{}), Name: NameSettleLog},
	}
	tlog = log.New("module", "lns.types")
)

func init() {
	types.AllowUserExec = append(types.AllowUserExec, []byte(LnsX))
	types.RegistorExecutor(LnsX, newType())
	//注册合约启用高度
	types.RegisterDappFork(LnsX, "Enable", 0)
}

type lnsType struct {
	types.ExecTypeBase
}

func newType() *lnsType {
	c := &lnsType{}
	c.SetChild(c)
	return c
}

// GetPayload 获取合约action结构
func (t *lnsType) GetPayload() types.Message {
	return &LnsAction{}
}

// GeTypeMap 获取合约action的id和name信息
func (t *lnsType) GetTypeMap() map[string]int32 {
	return actionMap
}

// GetLogMap 获取合约log相关信息
func (t *lnsType) GetLogMap() map[int64]*types.LogInfo {
	return logMap
}
