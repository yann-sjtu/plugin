package executor

import (
	"github.com/33cn/chain33/types"
	lnstypes "github.com/33cn/plugin/plugin/dapp/lns/types"
)

/*
 * 实现交易相关数据本地执行，数据不上链
 * 非关键数据，本地存储(localDB), 用于辅助查询，效率高
 */

/*
 * TODO, 本地执行暂时没有存储数据, 后期根据查询需求实现内建table, 包括扩展为更广义的闪电网络等需求
 */

func (l *lns) ExecLocal_Open(payload *lnstypes.OpenChannel, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	return dbSet, nil
}

func (l *lns) ExecLocal_DepositChannel(payload *lnstypes.DepositChannel, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	return dbSet, nil
}

func (l *lns) ExecLocal_WithdrawChannel(payload *lnstypes.WithdrawChannel, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	return dbSet, nil
}

func (l *lns) ExecLocal_Close(payload *lnstypes.CloseChannel, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	return dbSet, nil
}

func (l *lns) ExecLocal_UpdateProof(payload *lnstypes.UpdateBalanceProof, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	return dbSet, nil
}

func (l *lns) ExecLocal_Settle(payload *lnstypes.Settle, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	return dbSet, nil
}

func (l *lns) addAutoRollBack(tx *types.Transaction, kv []*types.KeyValue) *types.LocalDBSet {

	dbSet := &types.LocalDBSet{}
	dbSet.KV = l.AddRollbackKV(tx, tx.Execer, kv)
	return dbSet
}
