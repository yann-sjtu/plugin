package executor

import (
	"github.com/33cn/chain33/types"
	lnstypes "github.com/33cn/plugin/plugin/dapp/lns/types"
)

/*
 * 实现交易相关数据本地执行，数据不上链
 * 非关键数据，本地存储(localDB), 用于辅助查询，效率高
 */

func (l *lns) ExecLocal_Open(payload *lnstypes.OpenChannel, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	var dbSet *types.LocalDBSet
	//implement code
	return dbSet, nil
}

func (l *lns) ExecLocal_Deposit(payload *lnstypes.DepositChannel, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	var dbSet *types.LocalDBSet
	//implement code
	return dbSet, nil
}

func (l *lns) ExecLocal_Withdraw(payload *lnstypes.WithdrawChannel, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	var dbSet *types.LocalDBSet
	//implement code
	return dbSet, nil
}

func (l *lns) ExecLocal_Close(payload *lnstypes.CloseChannel, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	var dbSet *types.LocalDBSet
	//implement code
	return dbSet, nil
}

func (l *lns) ExecLocal_Update(payload *lnstypes.UpdateBalanceProof, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	var dbSet *types.LocalDBSet
	//implement code
	return dbSet, nil
}

func (l *lns) ExecLocal_Settle(payload *lnstypes.Settle, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	var dbSet *types.LocalDBSet
	//implement code
	return dbSet, nil
}

func (l *lns) addAutoRollBack(tx *types.Transaction, kv []*types.KeyValue) *types.LocalDBSet {

	dbSet := &types.LocalDBSet{}
	dbSet.KV = l.AddRollbackKV(tx, tx.Execer, kv)
	return dbSet
}
