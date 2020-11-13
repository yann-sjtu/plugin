package executor

import (
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/33cn/chain33/common"

	"github.com/33cn/chain33/rpc/grpcclient"

	"github.com/stretchr/testify/mock"

	"github.com/33cn/chain33/client/mocks"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	types2 "github.com/33cn/plugin/plugin/dapp/wasm/types"
	"github.com/stretchr/testify/assert"
)

var (
	PrivKeys = []string{
		"0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b", // 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4
		"0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4", // 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR
	}

	Addrs = []string{
		"1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4",
		"1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR",
	}
	wasmAddr string

	cfg *types.Chain33Config
)

func init() {
	cfg = types.NewChain33Config(strings.Replace(types.GetDefaultCfgstring(), "Title=\"local\"", "Title=\"chain33\"", 1))
	Init(types2.WasmX, cfg, nil)
}

func TestWasm_Exec(t *testing.T) {
	dir, ldb, kvdb := util.CreateTestDB()
	defer util.CloseTestDB(dir, ldb)
	acc := initAccount(t, ldb)

	testCreate(t, acc, kvdb)
	testCall(t, acc, kvdb)
}

func TestWasm_Callback(t *testing.T) {
	dir, ldb, kvdb := util.CreateTestDB()
	defer util.CloseTestDB(dir, ldb)
	wasmCB = newWasm().(*Wasm)
	acc := initAccount(t, ldb)
	wasmCB.SetCoinsAccount(acc)
	wasmCB.SetStateDB(kvdb)
	wasmCB.SetLocalDB(kvdb)
	wasmCB.execAddr = wasmAddr

	var err error
	testKey, testValue := []byte("test"), []byte("test")

	//test stateDB
	setStateDB(testKey, testValue)
	err = kvdb.Set(wasmCB.kvs[0].Key, wasmCB.kvs[0].Value)
	assert.Nil(t, err)
	stateValue, _ := getStateDB(testKey)
	assert.Equal(t, testValue, stateValue)

	//test localDB
	setLocalDB(testKey, testValue)
	set, err := execLocal(&types.ReceiptData{
		Ty:   types.ExecOk,
		Logs: wasmCB.receiptLogs,
	})
	assert.Nil(t, err)
	err = kvdb.Set(set.KV[0].Key, set.KV[0].Value)
	assert.Nil(t, err)
	localValue, _ := getLocalDB(testKey)
	assert.Equal(t, testValue, localValue)

	//test getBalance
	api := mocks.QueueProtocolAPI{}
	wasmCB.SetAPI(&api)
	api.On("GetLastHeader").Return(&types.Header{}, nil)
	api.On("StoreGet", mock.Anything).Return(&types.StoreReplyValue{
		Values: [][]byte{types.Encode(&types.Account{
			Addr:    Addrs[0],
			Balance: 1e8,
			Frozen:  1e10,
		})},
	}, nil)
	api.On("GetConfig").Return(cfg)
	balance, frozen, err := getBalance(Addrs[0], types2.WasmX)
	assert.Nil(t, err)
	assert.Equal(t, int64(1e8), balance)
	assert.Equal(t, int64(1e10), frozen)

	//test account operations
	//test transfer
	wasmCB.receiptLogs = nil
	err = transfer(Addrs[0], Addrs[1], 1e8)
	assert.Nil(t, err)
	accountTransfer := types.ReceiptAccountTransfer{}
	err = types.Decode(wasmCB.receiptLogs[0].Log, &accountTransfer)
	assert.Nil(t, err)
	assert.Equal(t, int64(1e10), accountTransfer.Prev.Balance)
	assert.Equal(t, int64(99e8), accountTransfer.Current.Balance)
	err = types.Decode(wasmCB.receiptLogs[1].Log, &accountTransfer)
	assert.Nil(t, err)
	assert.Equal(t, int64(1e10), accountTransfer.Prev.Balance)
	assert.Equal(t, int64(101e8), accountTransfer.Current.Balance)

	//test transfer to exec
	wasmCB.receiptLogs = nil
	err = transferToExec(Addrs[0], wasmAddr, 1e9)
	assert.Nil(t, err)
	err = types.Decode(wasmCB.receiptLogs[0].Log, &accountTransfer)
	assert.Nil(t, err)
	assert.Equal(t, int64(99e8), accountTransfer.Prev.Balance)
	assert.Equal(t, int64(89e8), accountTransfer.Current.Balance)
	err = types.Decode(wasmCB.receiptLogs[1].Log, &accountTransfer)
	assert.Nil(t, err)
	assert.Equal(t, int64(0), accountTransfer.Prev.Balance)
	assert.Equal(t, int64(1e9), accountTransfer.Current.Balance)

	//test transfer withdraw
	wasmCB.receiptLogs = nil
	err = transferWithdraw(Addrs[0], wasmAddr, 1e8)
	assert.Nil(t, err)
	execAccountTransfer := types.ReceiptExecAccountTransfer{}
	err = types.Decode(wasmCB.receiptLogs[0].Log, &execAccountTransfer)
	assert.Nil(t, err)
	assert.Equal(t, int64(1e9), execAccountTransfer.Prev.Balance)
	assert.Equal(t, int64(9e8), execAccountTransfer.Current.Balance)
	err = types.Decode(wasmCB.receiptLogs[1].Log, &accountTransfer)
	assert.Nil(t, err)
	assert.Equal(t, int64(1e9), accountTransfer.Prev.Balance)
	assert.Equal(t, int64(9e8), accountTransfer.Current.Balance)
	err = types.Decode(wasmCB.receiptLogs[2].Log, &accountTransfer)
	assert.Nil(t, err)
	assert.Equal(t, int64(89e8), accountTransfer.Prev.Balance)
	assert.Equal(t, int64(9e9), accountTransfer.Current.Balance)

	//test exec transfer
	wasmCB.receiptLogs = nil
	err = execTransfer(Addrs[0], Addrs[1], 1e8)
	assert.Nil(t, err)
	err = types.Decode(wasmCB.receiptLogs[0].Log, &execAccountTransfer)
	assert.Nil(t, err)
	assert.Equal(t, int64(9e8), execAccountTransfer.Prev.Balance)
	assert.Equal(t, int64(8e8), execAccountTransfer.Current.Balance)
	err = types.Decode(wasmCB.receiptLogs[1].Log, &execAccountTransfer)
	assert.Nil(t, err)
	assert.Equal(t, int64(0), execAccountTransfer.Prev.Balance)
	assert.Equal(t, int64(1e8), execAccountTransfer.Current.Balance)

	//test exec frozen
	wasmCB.receiptLogs = nil
	err = execFrozen(Addrs[0], 2e8)
	assert.Nil(t, err)
	err = types.Decode(wasmCB.receiptLogs[0].Log, &execAccountTransfer)
	assert.Nil(t, err)
	assert.Equal(t, int64(8e8), execAccountTransfer.Prev.Balance)
	assert.Equal(t, int64(0), execAccountTransfer.Prev.Frozen)
	assert.Equal(t, int64(6e8), execAccountTransfer.Current.Balance)
	assert.Equal(t, int64(2e8), execAccountTransfer.Current.Frozen)

	//test exec transfer frozen
	wasmCB.receiptLogs = nil
	err = execTransferFrozen(Addrs[0], Addrs[1], 1e8)
	assert.Nil(t, err)
	err = types.Decode(wasmCB.receiptLogs[0].Log, &execAccountTransfer)
	assert.Nil(t, err)
	assert.Equal(t, int64(6e8), execAccountTransfer.Prev.Balance)
	assert.Equal(t, int64(2e8), execAccountTransfer.Prev.Frozen)
	assert.Equal(t, int64(6e8), execAccountTransfer.Current.Balance)
	assert.Equal(t, int64(1e8), execAccountTransfer.Current.Frozen)
	err = types.Decode(wasmCB.receiptLogs[1].Log, &execAccountTransfer)
	assert.Nil(t, err)
	assert.Equal(t, int64(1e8), execAccountTransfer.Prev.Balance)
	assert.Equal(t, int64(0), execAccountTransfer.Prev.Frozen)
	assert.Equal(t, int64(2e8), execAccountTransfer.Current.Balance)
	assert.Equal(t, int64(0), execAccountTransfer.Current.Frozen)

	//test exec active
	wasmCB.receiptLogs = nil
	err = execActive(Addrs[0], 1e8)
	assert.Nil(t, err)
	err = types.Decode(wasmCB.receiptLogs[0].Log, &execAccountTransfer)
	assert.Nil(t, err)
	assert.Equal(t, int64(6e8), execAccountTransfer.Prev.Balance)
	assert.Equal(t, int64(1e8), execAccountTransfer.Prev.Frozen)
	assert.Equal(t, int64(7e8), execAccountTransfer.Current.Balance)
	assert.Equal(t, int64(0), execAccountTransfer.Current.Frozen)

	//test random
	gclient, err := grpcclient.NewMainChainClient(cfg, "")
	assert.Nil(t, err)
	seedGen := func() []byte {
		seed, _ := time.Now().GobEncode()
		return seed
	}
	api.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(&types.ReplyHash{
		Hash: common.Sha256(seedGen()),
	}, nil)
	wasmCB.SetExecutorAPI(&api, gclient)
	random := getRandom()
	t.Log(random)
}

func testCreate(t *testing.T, acc *account.DB, stateDB db.KV) {
	code, err := ioutil.ReadFile("../contracts/dice/dice.wasm")
	assert.Nil(t, err, "read wasm file error")
	payload := types2.WasmAction{
		Ty: types2.WasmActionCreate,
		Value: &types2.WasmAction_Create{
			Create: &types2.WasmCreate{
				Name: "dice",
				Code: code,
			},
		},
	}
	tx := &types.Transaction{
		Payload: types.Encode(&payload),
	}
	tx, err = types.FormatTx(cfg, types2.WasmX, tx)
	assert.Nil(t, err, "format tx error")

	wasm := newWasm()

	wasm.SetCoinsAccount(acc)
	wasm.SetStateDB(stateDB)
	receipt, err := wasm.Exec(tx, 0)
	assert.Nil(t, err, "tx exec error")
	assert.Equal(t, int32(types.ExecOk), receipt.Ty)
	assert.Equal(t, int32(types2.TyLogWasmCreate), receipt.Logs[0].Ty)
	assert.Equal(t, code, receipt.KV[0].Value)
	assert.Equal(t, contractKey("dice"), receipt.KV[0].Key)
	err = stateDB.Set(receipt.KV[0].Key, receipt.KV[0].Value)
	assert.Nil(t, err)
}

func testCall(t *testing.T, acc *account.DB, stateDB db.KV) {
	payload := types2.WasmAction{
		Ty: types2.WasmActionCall,
		Value: &types2.WasmAction_Call{
			Call: &types2.WasmCall{
				Contract:   "dice",
				Method:     "startgame",
				Parameters: []int64{1e9},
			},
		},
	}
	tx := &types.Transaction{
		Payload: types.Encode(&payload),
	}
	tx, err := types.FormatTx(cfg, types2.WasmX, tx)
	assert.Nil(t, err, "format tx error")
	err = signTx(tx, PrivKeys[0])
	assert.Nil(t, err)
	wasm := newWasm()

	wasm.SetCoinsAccount(acc)
	wasm.SetStateDB(stateDB)
	wasmCB = wasm.(*Wasm)
	err = transferToExec(Addrs[0], wasmAddr, 1e9)
	assert.Nil(t, err)
	receipt, err := wasm.Exec(tx, 0)
	assert.Nil(t, err, "tx exec error")
	assert.Equal(t, int32(types.ExecOk), receipt.Ty)
	assert.Equal(t, int32(types2.TyLogWasmCall), receipt.Logs[0].Ty)
}

func initAccount(t *testing.T, db db.KV) *account.DB {
	wasmAddr = address.ExecAddress(cfg.ExecName(types2.WasmX))
	acc, err := account.NewAccountDB(cfg, "coins", "bty", db)
	assert.Nil(t, err, "new account db error")
	acc.SaveAccount(&types.Account{
		Balance: 1e10,
		Addr:    Addrs[0],
	})
	acc.SaveAccount(&types.Account{
		Balance: 1e10,
		Addr:    Addrs[1],
	})
	return acc
}

func signTx(tx *types.Transaction, hexPrivKey string) error {
	signType := types.SECP256K1
	c, err := crypto.New(types.GetSignName("", signType))
	if err != nil {
		return err
	}

	bytes, err := common.FromHex(hexPrivKey[:])
	if err != nil {
		return err
	}

	privKey, err := c.PrivKeyFromBytes(bytes)
	if err != nil {
		return err
	}

	tx.Sign(int32(signType), privKey)
	return nil
}