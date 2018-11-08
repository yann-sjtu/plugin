package executor

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"gitlab.33.cn/chain33/chain33/account"
	"gitlab.33.cn/chain33/chain33/common"
	"gitlab.33.cn/chain33/chain33/common/address"
	"gitlab.33.cn/chain33/chain33/common/crypto"
	dbm "gitlab.33.cn/chain33/chain33/common/db"
	pty "gitlab.33.cn/chain33/chain33/plugin/dapp/unfreeze/types"
	"gitlab.33.cn/chain33/chain33/types"
)

type execEnv struct {
	blockTime   int64
	blockHeight int64
	difficulty  uint64
}

type orderArgs struct {
	total    int64
	startTs  int64
	period   int64
	duration int64
	except   int64
}

var (
	Symbol         = "TEST"
	AssetExecToken = "token"
	AssetExecPara  = "paracross"

	PrivKeyA = "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b" // 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4
	PrivKeyB = "0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4" // 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR
	PrivKeyC = "0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115" // 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k
	PrivKeyD = "0xcacb1f5d51700aea07fca2246ab43b0917d70405c65edea9b5063d72eb5c6b71" // 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs
	Nodes    = [][]byte{
		[]byte("1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"),
		[]byte("1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"),
		[]byte("1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"),
		[]byte("1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"),
	}
)

func TestUnfreeze(t *testing.T) {
	total := int64(100000)
	accountA := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    string(Nodes[0]),
	}
	accountB := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    string(Nodes[1]),
	}

	execAddr := address.ExecAddress(pty.UnfreezeX)
	stateDB, _ := dbm.NewGoMemDB("1", "2", 100)
	accA, _ := account.NewAccountDB(AssetExecPara, Symbol, stateDB)
	accA.SaveExecAccount(execAddr, &accountA)

	accB, _ := account.NewAccountDB(AssetExecPara, Symbol, stateDB)
	accB.SaveExecAccount(execAddr, &accountB)

	env := execEnv{
		10,
		2,
		1539918074,
	}

	// 创建
	opt := &pty.FixAmount{Period: 10, Amount: 2}
	p1 := &pty.UnfreezeCreate{
		StartTime:   10,
		AssetExec:   AssetExecPara,
		AssetSymbol: Symbol,
		TotalCount:  10000,
		Beneficiary: string(Nodes[1]),
		Means:       "FixAmount",
		MeansOpt:    &pty.UnfreezeCreate_FixAmount{FixAmount: opt},
	}
	createTx, err := pty.CreateUnfreezeCreateTx(p1)
	if err != nil {
		t.Error("CreateUnfreezeCreateTx", "err", err)
	}
	createTx, err = signTx(createTx, PrivKeyA)
	if err != nil {
		t.Error("CreateUnfreezeCreateTx sign", "err", err)
	}
	exec := newUnfreeze()
	exec.SetStateDB(stateDB)
	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	receipt, err := exec.Exec(createTx, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	//t.Log(receipt)
	accTmp := accA.LoadExecAccount(accountA.Addr, execAddr)
	assert.Equal(t, total-p1.TotalCount, accTmp.Balance)
	assert.Equal(t, p1.TotalCount, accTmp.Frozen)

	// 提币
	p2 := &pty.UnfreezeWithdraw{
		UnfreezeID: string(unfreezeID(string(createTx.Hash()))),
	}
	withdrawTx, err := pty.CreateUnfreezeWithdrawTx(p2)
	if err != nil {
		t.Error("CreateUnfreezeWithdrawTx", "err", err)
	}
	withdrawTx, err = signTx(withdrawTx, PrivKeyB)
	if err != nil {
		t.Error("CreateUnfreezeWithdrawTx sign", "err", err)
	}
	blockTime := int64(10)
	exec.SetEnv(env.blockHeight+1, env.blockTime+blockTime, env.difficulty)
	receipt, err = exec.Exec(withdrawTx, 1)
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	//t.Log(receipt)
	accATmp := accA.LoadExecAccount(accountA.Addr, execAddr)
	accBTmp := accB.LoadExecAccount(accountB.Addr, execAddr)
	assert.Equal(t, total-p1.TotalCount, accATmp.Balance)

	u := pty.Unfreeze{}
	e := types.Decode(receipt.KV[2].Value, &u)
	assert.Nil(t, e)
	assert.Equal(t, u.Remaining, accATmp.Frozen)
	assert.Equal(t, accountB.Balance+p1.TotalCount-u.Remaining, accBTmp.Balance)

	// 不是受益人提币
	{
		p2 := &pty.UnfreezeWithdraw{
			UnfreezeID: string(unfreezeID(string(createTx.Hash()))),
		}
		withdrawTx, err := pty.CreateUnfreezeWithdrawTx(p2)
		if err != nil {
			t.Error("CreateUnfreezeWithdrawTx", "err", err)
		}
		withdrawTx, err = signTx(withdrawTx, PrivKeyC)
		if err != nil {
			t.Error("CreateUnfreezeWithdrawTx sign", "err", err)
		}
		blockTime := int64(10)
		exec.SetEnv(env.blockHeight+1, env.blockTime+blockTime, env.difficulty)
		receipt, err = exec.Exec(withdrawTx, 1)
		assert.Equal(t, pty.ErrNoPrivilege, err)
		assert.Nil(t, receipt)
	}

	// 不是创建者终止
	{
		p3 := &pty.UnfreezeTerminate{
			UnfreezeID: string(unfreezeID(string(createTx.Hash()))),
		}
		terminateTx, err := pty.CreateUnfreezeTerminateTx(p3)
		if err != nil {
			t.Error("CreateUnfreezeTerminateTx", "err", err)
		}
		terminateTx, err = signTx(terminateTx, PrivKeyC)
		if err != nil {
			t.Error("CreateUnfreezeTerminateTx sign", "err", err)
		}
		receipt, err = exec.Exec(terminateTx, 1)
		assert.Equal(t, pty.ErrNoPrivilege, err)
		assert.Nil(t, receipt)
	}

	// 终止
	p3 := &pty.UnfreezeTerminate{
		UnfreezeID: string(unfreezeID(string(createTx.Hash()))),
	}
	terminateTx, err := pty.CreateUnfreezeTerminateTx(p3)
	if err != nil {
		t.Error("CreateUnfreezeTerminateTx", "err", err)
	}
	terminateTx, err = signTx(terminateTx, PrivKeyA)
	if err != nil {
		t.Error("CreateUnfreezeTerminateTx sign", "err", err)
	}
	receipt, err = exec.Exec(terminateTx, 1)
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	//t.Log(receipt)
	accATmp = accA.LoadExecAccount(accountA.Addr, execAddr)
	assert.Equal(t, total+total, accATmp.Balance+accBTmp.Balance)
	assert.Equal(t, int64(0), accATmp.Frozen)

	// 终止后不能继续提币
	{
		p2 := &pty.UnfreezeWithdraw{
			UnfreezeID: string(unfreezeID(string(createTx.Hash()))),
		}
		withdrawTx, err := pty.CreateUnfreezeWithdrawTx(p2)
		if err != nil {
			t.Error("CreateUnfreezeWithdrawTx", "err", err)
		}
		withdrawTx, err = signTx(withdrawTx, PrivKeyB)
		if err != nil {
			t.Error("CreateUnfreezeWithdrawTx sign", "err", err)
		}
		blockTime := int64(10)
		exec.SetEnv(env.blockHeight+1, env.blockTime+blockTime+blockTime, env.difficulty)
		receipt, err = exec.Exec(withdrawTx, 1)
		assert.Equal(t, pty.ErrUnfreezeEmptied, err)
		assert.Nil(t, receipt)
	}
}

func signTx(tx *types.Transaction, hexPrivKey string) (*types.Transaction, error) {
	signType := types.SECP256K1
	c, err := crypto.New(types.GetSignName(pty.UnfreezeX, signType))
	if err != nil {
		return tx, err
	}

	bytes, err := common.FromHex(hexPrivKey[:])
	if err != nil {
		return tx, err
	}

	privKey, err := c.PrivKeyFromBytes(bytes)
	if err != nil {
		return tx, err
	}

	tx.Sign(int32(signType), privKey)
	return tx, nil
}
