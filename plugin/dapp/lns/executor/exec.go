package executor

import (
	"github.com/33cn/chain33/types"
	lnstypes "github.com/33cn/plugin/plugin/dapp/lns/types"
)

/*
 * 实现交易的链上执行接口
 * 关键数据上链（statedb）并生成交易回执（log）
 */

func (l *lns) Exec_Open(payload *lnstypes.OpenChannel, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(l, tx, index)
	return action.openChannel(payload)
}

func (l *lns) Exec_DepositChannel(payload *lnstypes.DepositChannel, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(l, tx, index)
	return action.depositChannel(payload)
}

func (l *lns) Exec_WithdrawChannel(payload *lnstypes.WithdrawChannel, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(l, tx, index)
	return action.withdrawChannel(payload)
}

func (l *lns) Exec_Close(payload *lnstypes.CloseChannel, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(l, tx, index)
	return action.closeChannel(payload)
}

func (l *lns) Exec_UpdateProof(payload *lnstypes.UpdateBalanceProof, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(l, tx, index)
	return action.updateBalanceProof(payload)
}

func (l *lns) Exec_Settle(payload *lnstypes.Settle, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(l, tx, index)
	return action.settleChannel(payload)
}
