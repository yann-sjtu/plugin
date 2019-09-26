package executor

import (
	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/log"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/stretchr/testify/assert"

	"testing"

	lnsty "github.com/33cn/plugin/plugin/dapp/lns/types"
)

var (
	initBalance  = types.Coin * 10000
	initHeight   = 100
	partnerAddr1 = "12HKLEn6g4FH39yUbHh4EVJWcFo5CXg22d"
	partnerAddr2 = "1Ka7EPFRqs3v9yreXG6qA4RQbNmbPJCZPj"
	lnsExecAddr  = "1FvToJGahtiY8TYG7DhDTzyZGrZHrvGTV3"

	//partnerAddr1
	priv1 = "0x9d4f8ab11361be596468b265cb66946c87873d4a119713fd0c3d8302eae0a8e4"
	//partnerAddr2
	priv2 = "0xd165c84ed37c2a427fea487470ee671b7a0495d68d82607cafbc6348bf23bec5"
	priv3 = "0xc21d38be90493512a5c2417d565269a8b23ce8152010e404ff4f75efead8183a"
)

func init() {
	log.SetLogLevel("error")
}

func initEnv() (string, dapp.Driver, dbm.DB) {

	dir, ldb, kvdb := util.CreateTestDB()
	accountA := &types.Account{
		Balance: initBalance,
		Addr:    partnerAddr1,
	}

	accountB := &types.Account{
		Balance: initBalance,
		Addr:    partnerAddr2,
	}

	accountExec := &types.Account{
		Balance: initBalance,
		Addr:    lnsExecAddr,
	}

	accCoin := account.NewCoinsAccount()
	accCoin.SetDB(ldb)
	accCoin.SaveAccount(accountA)
	accCoin.SaveExecAccount(lnsExecAddr, accountA)
	accCoin.SaveExecAccount(lnsExecAddr, accountB)
	accCoin.SaveExecAccount(lnsExecAddr, accountExec)

	exec := newLns()
	exec.SetStateDB(ldb)
	exec.SetLocalDB(kvdb)
	exec.SetEnv(100, 1539918074, 1539918074)

	return dir, exec, ldb
}

type testcase struct {
	payload   types.Message
	expectErr error
	priv      string
	index     int
}

func createTx(payload types.Message, priv string) (*types.Transaction, error) {

	tx, err := types.CreateFormatTx(types.ExecName(lnsty.LnsX), types.Encode(payload))
	if err != nil {
		return nil, err
	}

	c, err := crypto.New(crypto.GetName(types.SECP256K1))
	if err != nil {
		return tx, err
	}

	bytes, err := common.FromHex(priv[:])
	if err != nil {
		return tx, err
	}

	privKey, err := c.PrivKeyFromBytes(bytes)
	if err != nil {
		return tx, err
	}

	tx.Sign(int32(types.SECP256K1), privKey)
	return tx, nil
}

func runTest(db dbm.DB, exec dapp.Driver, tcArr []*testcase, priv string, t *testing.T) {

	for i, tc := range tcArr {

		signPriv := priv
		if tc.priv != "" {
			signPriv = tc.priv
		}
		tx, err := createTx(tc.payload, signPriv)
		assert.NoError(t, err, "lnsTestExecCreateTxErr")
		recp, err := exec.Exec(tx, i)
		if err == nil && len(recp.GetKV()) > 0 {
			util.SaveKVList(db, recp.KV)
		}
		assert.Equalf(t, tc.expectErr, err, "testcase index %d", tc.index)

	}
}

func TestLns_Exec_Open(t *testing.T) {

	dir, exec, ldb := initEnv()
	defer util.CloseTestDB(dir, ldb)
	tcArr := []*testcase{
		{
			index: 0,
			payload: &lnsty.OpenChannel{
				IssueContract: "coins",
				Partner:       partnerAddr2,
				Amount:        types.Coin,
			},
			expectErr: nil,
		},
		{
			index: 1,
			payload: &lnsty.OpenChannel{
				IssueContract: "coins",
				Partner:       partnerAddr2,
				SettleTimeout: maxSettleTimeout + 1,
				Amount:        types.Coin,
			},
			expectErr: nil,
		},
		{
			index: 2,
			payload: &lnsty.OpenChannel{
				IssueContract: "coins",
				Partner:       partnerAddr2,
				SettleTimeout: minSettleTimeout + 1,
				Amount:        initBalance + types.Coin,
			},
			expectErr: types.ErrNoBalance,
		},
	}
	var err error
	for _, tc := range tcArr {
		tc.payload, err = formatLnsAction(tc.payload)
		assert.NoError(t, err)
	}

	runTest(ldb, exec, tcArr, priv1, t)
	chanCount := &lnsty.ChannelCount{}
	err = getDBAndDecode(ldb, calcLnsChannelCountKey(), chanCount)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), chanCount.Count)
	channel := &lnsty.Channel{}
	err = getDBAndDecode(ldb, calcLnsChannelIDKey(1), channel)
	assert.NoError(t, err)
	if err == nil {
		assert.Equal(t, partnerAddr1, channel.Participant1.Addr)
		assert.Equal(t, partnerAddr2, channel.Participant2.Addr)
		assert.Equal(t, types.Coin, channel.Participant1.TotalDeposit)
		assert.Equal(t, int64(0), channel.Participant2.TotalDeposit)
	}
}

func TestLns_Exec_DepositChannel(t *testing.T) {

	dir, exec, ldb := initEnv()
	defer util.CloseTestDB(dir, ldb)

	tcArr := []*testcase{
		{
			index: 1,
			payload: &lnsty.OpenChannel{
				IssueContract: "coins",
				Partner:       partnerAddr2,
				Amount:        types.Coin,
			},
			expectErr: nil,
		},
		{
			index:     2,
			payload:   &lnsty.DepositChannel{},
			expectErr: types.ErrNotFound,
		},
		{
			index: 3,
			payload: &lnsty.DepositChannel{
				ChannelID: 1,
			},
			expectErr: lnsty.ErrInvalidChannelParticipants,
			priv:      priv3,
		},
		{
			index: 4,
			payload: &lnsty.DepositChannel{
				ChannelID: 1,
			},
			expectErr: lnsty.ErrTotalDepositAmount,
		},
		{
			index: 5,
			payload: &lnsty.DepositChannel{
				ChannelID:    1,
				TotalDeposit: types.Coin,
			},
			expectErr: nil,
			priv:      priv2,
		},
	}
	var err error
	for _, tc := range tcArr {
		tc.payload, err = formatLnsAction(tc.payload)
		assert.NoError(t, err)
	}

	runTest(ldb, exec, tcArr, priv1, t)

	channel := &lnsty.Channel{}
	err = getDBAndDecode(ldb, calcLnsChannelIDKey(1), channel)
	assert.NoError(t, err)
	assert.Equal(t, types.Coin, channel.Participant2.TotalDeposit)
}

func createTestSign(privKey string) (*types.Signature, error) {

	keyByte, err := common.FromHex(privKey)
	if err != nil {
		return nil, err
	}
	cr, err := crypto.New(crypto.GetName(types.SECP256K1))
	if err != nil {
		return nil, err
	}
	priv, err := cr.PrivKeyFromBytes(keyByte)
	if err != nil {
		return nil, err
	}
	sign := &types.Signature{
		Ty:     0,
		Pubkey: priv.PubKey().Bytes(),
	}
	return sign, nil
}

func TestLns_Exec_WithdrawChannel(t *testing.T) {
	dir, exec, ldb := initEnv()
	defer util.CloseTestDB(dir, ldb)
	sign, err := createTestSign(priv2)
	assert.NoError(t, err)

	tcArr := []*testcase{
		{
			index: 1,
			payload: &lnsty.OpenChannel{
				IssueContract: "coins",
				Partner:       partnerAddr2,
				Amount:        types.Coin * 10,
			},
			expectErr: nil,
		},
		{
			index: 2,
			payload: &lnsty.WithdrawChannel{
				ChannelID: 0,
			},
			expectErr: types.ErrNotFound,
		},
		{
			index: 3,
			payload: &lnsty.WithdrawChannel{
				ChannelID:        1,
				Proof:            nil,
				PartnerSignature: sign,
			},
			expectErr: lnsty.ErrInvalidChannelParticipants,
			priv:      priv3,
		},
		{
			index: 4,
			payload: &lnsty.WithdrawChannel{
				ChannelID: 1,
				Proof: &lnsty.WithdrawConfirmProof{
					TotalWithdraw: types.Coin,
				},
				PartnerSignature: sign,
			},
			expectErr: nil,
		},
		{
			index: 5,
			payload: &lnsty.WithdrawChannel{
				ChannelID: 1,
				Proof: &lnsty.WithdrawConfirmProof{
					TotalWithdraw: types.Coin,
				},
				PartnerSignature: sign,
			},
			expectErr: lnsty.ErrInvalidWithdrawAmount,
		},
		{
			index: 6,
			payload: &lnsty.WithdrawChannel{
				ChannelID: 1,
				Proof: &lnsty.WithdrawConfirmProof{
					TotalWithdraw: types.Coin * 100,
				},
				PartnerSignature: sign,
			},
			expectErr: lnsty.ErrInvalidWithdrawAmount,
		},
	}
	for _, tc := range tcArr {
		tc.payload, err = formatLnsAction(tc.payload)
		assert.NoError(t, err)
	}
	runTest(ldb, exec, tcArr, priv1, t)
	channel := &lnsty.Channel{}
	err = getDBAndDecode(ldb, calcLnsChannelIDKey(1), channel)
	assert.NoError(t, err)
	assert.Equal(t, types.Coin, channel.Participant1.TotalWithdraw)
}

func TestLns_Exec_Close(t *testing.T) {

	dir, exec, ldb := initEnv()
	defer util.CloseTestDB(dir, ldb)
	sign, err := createTestSign(priv2)
	assert.NoError(t, err)

	tcArr := []*testcase{
		{
			index: 1,
			payload: &lnsty.OpenChannel{
				IssueContract: "coins",
				Partner:       partnerAddr2,
				Amount:        types.Coin * 10,
			},
			expectErr: nil,
		},
		{
			index:     2,
			payload:   &lnsty.CloseChannel{},
			expectErr: types.ErrNotFound,
		},
		{
			index: 3,
			payload: &lnsty.CloseChannel{
				ChannelID:          1,
				NonCloserBalancePf: nil,
				NonCloserSignature: sign,
			},
			expectErr: lnsty.ErrInvalidChannelParticipants,
			priv:      priv3,
		},
		{
			index: 4,
			payload: &lnsty.CloseChannel{
				ChannelID: 1,
				NonCloserBalancePf: &lnsty.BalanceProof{
					Nonce:             1,
					TransferredAmount: types.Coin,
				},
				NonCloserSignature: sign,
			},
			expectErr: nil,
		},
		{
			index: 5,
			payload: &lnsty.CloseChannel{
				ChannelID: 1,
				NonCloserBalancePf: &lnsty.BalanceProof{
					Nonce:             1,
					TransferredAmount: types.Coin,
				},
				NonCloserSignature: sign,
			},
			expectErr: lnsty.ErrChannelState,
		},
	}
	for _, tc := range tcArr {
		tc.payload, err = formatLnsAction(tc.payload)
		assert.NoError(t, err)
	}
	runTest(ldb, exec, tcArr, priv1, t)
	channel := &lnsty.Channel{}
	err = getDBAndDecode(ldb, calcLnsChannelIDKey(1), channel)
	assert.NoError(t, err)
	assert.Equal(t, int32(lnsty.StateClosed), channel.State)
	assert.Equal(t, int64(1), channel.Participant2.Nonce)
	assert.Equal(t, types.Coin, channel.Participant2.TransferredAmount)
}

func TestLns_Exec_UpdateProof(t *testing.T) {

	dir, exec, ldb := initEnv()
	defer util.CloseTestDB(dir, ldb)
	sign, err := createTestSign(priv2)
	assert.NoError(t, err)
	tcArr := []*testcase{
		{
			index: 1,
			payload: &lnsty.OpenChannel{
				IssueContract: "coins",
				Partner:       partnerAddr2,
				Amount:        types.Coin * 10,
			},
			expectErr: nil,
		},
		{
			index:     2,
			payload:   &lnsty.UpdateBalanceProof{},
			expectErr: types.ErrNotFound,
		},
		{
			index: 3,
			payload: &lnsty.UpdateBalanceProof{
				ChannelID:        1,
				PartnerBalancePf: nil,
				PartnerSignature: sign,
			},
			expectErr: lnsty.ErrInvalidChannelParticipants,
			priv:      priv3,
		},
		{
			index: 4,
			payload: &lnsty.CloseChannel{
				ChannelID: 1,
				NonCloserBalancePf: &lnsty.BalanceProof{
					Nonce:             1,
					TransferredAmount: types.Coin,
				},
				NonCloserSignature: sign,
			},
			expectErr: nil,
		},
		{
			index: 5,
			payload: &lnsty.UpdateBalanceProof{
				ChannelID: 1,
				PartnerBalancePf: &lnsty.BalanceProof{
					Nonce:             1,
					TransferredAmount: types.Coin,
				},
				PartnerSignature: sign,
			},
			expectErr: lnsty.ErrBalanceProofNonce,
		},
		{
			index: 6,
			payload: &lnsty.UpdateBalanceProof{
				ChannelID: 1,
				PartnerBalancePf: &lnsty.BalanceProof{
					Nonce:             2,
					TransferredAmount: 2 * types.Coin,
				},
				PartnerSignature: sign,
			},
			expectErr: nil,
		},
	}
	tcArr2 := []*testcase{
		{
			index: 7,
			payload: &lnsty.UpdateBalanceProof{
				ChannelID: 1,
				PartnerBalancePf: &lnsty.BalanceProof{
					Nonce:             2,
					TransferredAmount: 2 * types.Coin,
				},
				PartnerSignature: sign,
			},
			expectErr: lnsty.ErrChannelCloseChallengePeriod,
		},
	}

	tcArrs := [][]*testcase{tcArr, tcArr2}
	for _, arr := range tcArrs {
		for _, tc := range arr {
			tc.payload, err = formatLnsAction(tc.payload)
			assert.NoError(t, err)
		}
	}
	runTest(ldb, exec, tcArr, priv1, t)
	//关闭通道,并设置当前高度超过挑战期间
	exec.SetEnv(1000, 1539918074, 1539918074)
	runTest(ldb, exec, tcArr2, priv1, t)

	channel := &lnsty.Channel{}
	err = getDBAndDecode(ldb, calcLnsChannelIDKey(1), channel)
	assert.NoError(t, err)
	assert.Equal(t, int32(lnsty.StateClosed), channel.State)
	assert.Equal(t, int64(2), channel.Participant2.Nonce)
	assert.Equal(t, 2*types.Coin, channel.Participant2.TransferredAmount)
}

func TestLns_Exec_Settle(t *testing.T) {

	dir, exec, ldb := initEnv()
	defer util.CloseTestDB(dir, ldb)
	sign, err := createTestSign(priv2)
	assert.NoError(t, err)
	sign1, err := createTestSign(priv1)
	assert.NoError(t, err)
	tcArr := []*testcase{
		{
			index: 1,
			payload: &lnsty.OpenChannel{
				IssueContract: "coins",
				Partner:       partnerAddr2,
				Amount:        types.Coin,
			},
			expectErr: nil,
		},
		{
			index: 1,
			payload: &lnsty.OpenChannel{
				IssueContract: "coins",
				Partner:       partnerAddr2,
				Amount:        types.Coin,
			},
			expectErr: nil,
		},
		{
			index: 1,
			payload: &lnsty.OpenChannel{
				IssueContract: "coins",
				Partner:       partnerAddr1,
				Amount:        types.Coin,
			},
			expectErr: nil,
			priv:      priv2,
		},
		{
			index:     2,
			payload:   &lnsty.Settle{},
			expectErr: types.ErrNotFound,
		},
		{
			index: 3,
			payload: &lnsty.Settle{
				ChannelID: 1,
			},
			expectErr: lnsty.ErrInvalidChannelParticipants,
			priv:      priv3,
		},
		{
			index: 4,
			payload: &lnsty.Settle{
				ChannelID: 1,
			},
			expectErr: lnsty.ErrChannelState,
		},
		{
			index: 51,
			payload: &lnsty.CloseChannel{
				ChannelID: 1,
				NonCloserBalancePf: &lnsty.BalanceProof{
					Nonce:             1,
					TransferredAmount: types.Coin,
				},
				NonCloserSignature: sign,
			},
			expectErr: nil,
		},
		{
			index: 52,
			payload: &lnsty.CloseChannel{
				ChannelID: 2,
				NonCloserBalancePf: &lnsty.BalanceProof{
					Nonce:             1,
					TransferredAmount: 2 * types.Coin,
				},
				NonCloserSignature: sign1,
			},
			expectErr: nil,
			priv:      priv2,
		},
		{
			index: 53,
			payload: &lnsty.CloseChannel{
				ChannelID: 3,
				NonCloserBalancePf: &lnsty.BalanceProof{
					Nonce:             1,
					TransferredAmount: 2 * types.Coin,
				},
				NonCloserSignature: sign,
			},
			expectErr: nil,
		},
		{
			index: 6,
			payload: &lnsty.Settle{
				ChannelID: 1,
			},
			expectErr: lnsty.ErrChannelCloseChallengePeriod,
		},
		{
			index: 7,
			payload: &lnsty.UpdateBalanceProof{
				ChannelID: 3,
				PartnerBalancePf: &lnsty.BalanceProof{
					Nonce:             1,
					TransferredAmount: types.Coin,
				},
				PartnerSignature: sign1,
			},
			expectErr: nil,
			priv:      priv2,
		},
	}
	tcArr2 := []*testcase{
		{
			index: 81,
			payload: &lnsty.Settle{
				ChannelID: 1,
			},
			expectErr: nil,
		},
		{
			index: 82,
			payload: &lnsty.Settle{
				ChannelID: 2,
			},
			expectErr: nil,
		},
		{
			index: 83,
			payload: &lnsty.Settle{
				ChannelID: 3,
			},
			expectErr: nil,
		},
	}

	tcArrs := [][]*testcase{tcArr, tcArr2}
	for _, arr := range tcArrs {
		for _, tc := range arr {
			tc.payload, err = formatLnsAction(tc.payload)
			assert.NoError(t, err)
		}
	}
	runTest(ldb, exec, tcArr, priv1, t)
	//关闭通道,并设置可结算高度
	exec.SetEnv(1000, 1539918074, 1539918074)
	runTest(ldb, exec, tcArr2, priv1, t)

	channel := &lnsty.Channel{}
	err = getDBAndDecode(ldb, calcLnsChannelIDKey(1), channel)
	assert.NoError(t, err)
	assert.Equal(t, int32(lnsty.StateSettled), channel.State)

	acc := exec.GetCoinsAccount()
	acc1 := acc.LoadExecAccount(partnerAddr1, lnsExecAddr)
	acc2 := acc.LoadExecAccount(partnerAddr2, lnsExecAddr)
	assert.Equal(t, initBalance, acc1.Balance)
	assert.Equal(t, initBalance, acc2.Balance)
}

func formatLnsAction(param types.Message) (types.Message, error) {

	action := &lnsty.LnsAction{
		Value: nil,
		Ty:    0,
	}
	if open, ok := param.(*lnsty.OpenChannel); ok {
		action.Value = &lnsty.LnsAction_Open{Open: open}
		action.Ty = lnsty.TyOpenAction
	} else if deposit, ok := param.(*lnsty.DepositChannel); ok {
		action.Value = &lnsty.LnsAction_DepositChannel{DepositChannel: deposit}
		action.Ty = lnsty.TyDepositAction
	} else if withdraw, ok := param.(*lnsty.WithdrawChannel); ok {
		action.Value = &lnsty.LnsAction_WithdrawChannel{WithdrawChannel: withdraw}
		action.Ty = lnsty.TyWithdrawAction
	} else if close, ok := param.(*lnsty.CloseChannel); ok {
		action.Value = &lnsty.LnsAction_Close{Close: close}
		action.Ty = lnsty.TyCloseAction
	} else if update, ok := param.(*lnsty.UpdateBalanceProof); ok {
		action.Value = &lnsty.LnsAction_UpdateProof{UpdateProof: update}
		action.Ty = lnsty.TyUpdateProofAction

	} else if settle, ok := param.(*lnsty.Settle); ok {
		action.Value = &lnsty.LnsAction_Settle{Settle: settle}
		action.Ty = lnsty.TySettleAction
	} else {
		return nil, types.ErrActionNotSupport
	}

	return action, nil
}
