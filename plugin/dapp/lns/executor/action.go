package executor

import (
	"math"

	"github.com/33cn/chain33/common/address"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"

	lnstypes "github.com/33cn/plugin/plugin/dapp/lns/types"
)

const (
	minSettleTimeout = 10
	maxSettleTimeout = 10000
)

type action struct {
	api          client.QueueProtocolAPI
	coinsAccount *account.DB
	db           dbm.KV
	txHash       []byte
	fromAddr     string
	blockTime    int64
	height       int64
	index        int
	execAddr     string
}

func newAction(l *lns, tx *types.Transaction, index int) *action {
	hash := tx.Hash()
	fromAddr := tx.From()
	return &action{l.GetAPI(), l.GetCoinsAccount(), l.GetStateDB(), hash, fromAddr,
		l.GetBlockTime(), l.GetHeight(), index, dapp.ExecAddress(string(tx.Execer))}
}

func (a *action) openChannel(open *lnstypes.OpenChannel) (*types.Receipt, error) {

	chanCount := &lnstypes.ChannelCount{}
	countKey := calcLnsChannelCountKey()

	err := getDBAndDecode(a.db, countKey, chanCount)
	if err != nil && err != types.ErrNotFound {
		elog.Error("ExecOpenChannel", "GetChannelCountErr", err)
		return nil, err
	}

	if chanCount.Count >= math.MaxInt64 {
		return nil, lnstypes.ErrChannelIDOverFlow
	}
	if open.TokenSymbol == "" {
		open.TokenSymbol = types.GetCoinSymbol()
	}

	elog.Debug("ExecOpenChannel", "ChannelCount", chanCount.Count)
	channel := &lnstypes.Channel{}
	chanCount.Count++
	channel.ChannelID = chanCount.Count
	if open.SettleTimeout < minSettleTimeout {
		open.SettleTimeout = minSettleTimeout
	} else if open.SettleTimeout > maxSettleTimeout {
		open.SettleTimeout = maxSettleTimeout
	}

	channel.IssueContract = open.IssueContract
	channel.TokenSymbol = open.TokenSymbol
	channel.SettleBlockHeight = int64(open.SettleTimeout)
	channel.State = lnstypes.StateOpen
	channel.Participant1 = &lnstypes.Participant{
		Addr: a.fromAddr,
	}
	channel.Participant2 = &lnstypes.Participant{
		Addr: open.Partner,
	}

	receipt := &types.Receipt{Ty: types.ExecOk}
	receipt.KV = append(receipt.KV, &types.KeyValue{
		Key:   countKey,
		Value: types.Encode(chanCount),
	})
	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{
		Ty: lnstypes.TyOpenLog,
		Log: types.Encode(&lnstypes.ReceiptOpen{
			Opener:         a.fromAddr,
			Partner:        open.Partner,
			ChannelID:      channel.ChannelID,
			InitialBalance: open.Amount,
			ChainName:      types.GetTitle(),
			IssueContract:  open.GetIssueContract(),
			TokenSymbol:    open.GetTokenSymbol(),
			SettleTimeOut:  open.SettleTimeout,
		}),
	})
	accDB, err := a.createAccountDB(open.IssueContract, open.TokenSymbol)
	if err != nil {
		elog.Error("ExecOpenChannel", "exec", open.IssueContract, "symbol", open.TokenSymbol, "createAccountDBErr", err)
		return nil, err
	}

	//channel的资产统一放在执行器账户的执行器地址上
	if open.Amount > 0 {
		channel.TotalAmount += open.Amount
		channel.Participant1.TotalDeposit += open.Amount
		transReceipt, err := accDB.ExecTransfer(a.fromAddr, a.execAddr, a.execAddr, open.Amount)
		if err != nil {
			elog.Error("ExecOpenChannel", "execTransferErr", err)
			return nil, err
		}

		receipt.KV = append(receipt.KV, transReceipt.KV...)
		receipt.Logs = append(receipt.Logs, transReceipt.Logs...)
	}
	receipt.KV = append(receipt.KV, &types.KeyValue{
		Key:   calcLnsChannelIDKey(channel.ChannelID),
		Value: types.Encode(channel),
	})

	return receipt, nil
}

func (a *action) depositChannel(deposit *lnstypes.DepositChannel) (*types.Receipt, error) {

	elog.Debug("ExecDepositChannel", "ChannelID", deposit.ChannelID)
	channel := &lnstypes.Channel{}
	chanKey := calcLnsChannelIDKey(deposit.ChannelID)

	err := getDBAndDecode(a.db, chanKey, channel)
	if err != nil {
		elog.Error("ExecDepositChannel", "GetChannelErr", err)
		return nil, err
	}
	participants := getParticipantMap(channel.Participant1, channel.Participant2)

	//check tx
	if channel.State != lnstypes.StateOpen {
		return nil, lnstypes.ErrChannelState
	}
	partner := getPartnerAddr(a.fromAddr, channel.Participant1.Addr, channel.Participant2.Addr)
	if !checkParticipantValidity(participants, a.fromAddr, partner) {
		return nil, lnstypes.ErrInvalidChannelParticipants
	}

	//本次存入额度
	depositAmount := deposit.TotalDeposit - participants[a.fromAddr].TotalDeposit
	channel.TotalAmount += depositAmount
	participants[a.fromAddr].TotalDeposit = deposit.TotalDeposit

	receipt := &types.Receipt{Ty: types.ExecOk}
	receipt.KV = append(receipt.KV, &types.KeyValue{
		Key:   chanKey,
		Value: types.Encode(channel),
	})

	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{
		Ty: lnstypes.TyDepositLog,
		Log: types.Encode(&lnstypes.ReceiptDeposit{
			Chain:         types.GetTitle(),
			IssueContract: channel.GetIssueContract(),
			TokenSymbol:   channel.GetTokenSymbol(),
			TotalDeposit:  deposit.TotalDeposit,
			ChannelID:     channel.ChannelID,
			Depositor:     a.fromAddr,
			Partner:       partner}),
	})

	accDB, err := a.createAccountDB(channel.IssueContract, channel.TokenSymbol)
	if err != nil {
		elog.Error("ExecDepositChannel", "exec", channel.IssueContract, "symbol", channel.TokenSymbol, "createAccountDBErr", err)
		return nil, err
	}

	//channel的资产统一放在执行器账户的执行器地址上
	transReceipt, err := accDB.ExecTransfer(a.fromAddr, a.execAddr, a.execAddr, depositAmount)
	if err != nil {
		elog.Error("ExecDepositChannel", "execTransferErr", err)
		return nil, err
	}

	receipt.KV = append(receipt.KV, transReceipt.KV...)
	receipt.Logs = append(receipt.Logs, transReceipt.Logs...)

	return receipt, nil
}

func (a *action) withdrawChannel(withdraw *lnstypes.WithdrawChannel) (*types.Receipt, error) {

	channelID := withdraw.GetChannelID()
	elog.Debug("ExecWithdrawChannel", "ChannelID", channelID)
	channel := &lnstypes.Channel{}
	chanKey := calcLnsChannelIDKey(channelID)

	err := getDBAndDecode(a.db, chanKey, channel)
	if err != nil {
		elog.Error("ExecWithdrawChannel", "GetChannelErr", err)
		return nil, err
	}

	participants := getParticipantMap(channel.Participant1, channel.Participant2)
	withdrawAddr := a.fromAddr
	partner := address.PubKeyToAddr(withdraw.GetPartnerSignature().GetPubkey())
	totalWithdraw := withdraw.GetProof().GetTotalWithdraw()
	//本次提取额度
	withdrawAmount := totalWithdraw - participants[withdrawAddr].TotalWithdraw

	//check tx
	if channel.State != lnstypes.StateOpen {
		return nil, lnstypes.ErrChannelState
	}

	if !checkParticipantValidity(participants, withdrawAddr, partner) {
		return nil, lnstypes.ErrInvalidChannelParticipants
	}

	if withdrawAmount <= 0 || withdrawAmount > channel.TotalAmount {
		return nil, lnstypes.ErrInvalidWithdrawAmount
	}

	//修改channel相关额度数据
	channel.TotalAmount -= withdrawAmount
	participants[withdrawAddr].TotalWithdraw = totalWithdraw

	receipt := &types.Receipt{Ty: types.ExecOk}
	receipt.KV = append(receipt.KV, &types.KeyValue{
		Key:   chanKey,
		Value: types.Encode(channel),
	})

	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{
		Ty: lnstypes.TyWithdrawLog,
		Log: types.Encode(&lnstypes.ReceiptWithdraw{
			TotalWithdraw: totalWithdraw,
			ChannelID:     channel.ChannelID,
			Withdrawer:    withdrawAddr,
			Partner:       partner,
		}),
	})

	accDB, err := a.createAccountDB(channel.IssueContract, channel.TokenSymbol)
	if err != nil {
		elog.Error("ExecWithdrawChannel", "exec", channel.IssueContract, "symbol", channel.TokenSymbol, "createAccountDBErr", err)
		return nil, err
	}
	//channel的资产统一放在执行器账户的执行器地址上
	transReceipt, err := accDB.ExecTransfer(a.execAddr, withdrawAddr, a.execAddr, withdrawAmount)
	if err != nil {
		elog.Error("ExecWithdrawChannel", "execTransferErr", err)
		return nil, err
	}

	receipt.KV = append(receipt.KV, transReceipt.KV...)
	receipt.Logs = append(receipt.Logs, transReceipt.Logs...)

	return receipt, nil
}

func (a *action) closeChannel(close *lnstypes.CloseChannel) (*types.Receipt, error) {

	elog.Debug("ExecCloseChannel", "ChannelID", close.ChannelID)
	channel := &lnstypes.Channel{}
	chanKey := calcLnsChannelIDKey(close.ChannelID)

	err := getDBAndDecode(a.db, chanKey, channel)
	if err != nil {
		elog.Error("ExecCloseChannel", "GetChannelErr", err)
		return nil, err
	}
	nonCloser := address.PubKeyToAddr(close.GetNonCloserSignature().GetPubkey())
	participants := getParticipantMap(channel.Participant1, channel.Participant2)

	//check tx
	if channel.State != lnstypes.StateOpen {
		return nil, lnstypes.ErrChannelState
	}

	if !checkParticipantValidity(participants, nonCloser, a.fromAddr) {
		return nil, lnstypes.ErrInvalidChannelParticipants
	}

	channel.State = lnstypes.StateClosed
	channel.Closer = a.fromAddr
	channel.SettleBlockHeight += a.height

	if close.NonCloserBalancePf.Nonce > participants[nonCloser].Nonce {
		//记录非关闭方的的总转移额度
		participants[nonCloser].Nonce = close.NonCloserBalancePf.Nonce
		participants[nonCloser].TransferredAmount = close.NonCloserBalancePf.TransferredAmount
	}

	receipt := &types.Receipt{Ty: types.ExecOk}
	receipt.KV = append(receipt.KV, &types.KeyValue{
		Key:   chanKey,
		Value: types.Encode(channel),
	})

	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{
		Ty: lnstypes.TyCloseLog,
		Log: types.Encode(&lnstypes.ReceiptClose{
			TokenCanonicalId: &lnstypes.TokenCanonicalId{
				Chain:         types.GetTitle(),
				IssueContract: channel.GetIssueContract(),
				TokenSymbol:   channel.GetTokenSymbol(),
			},
			ChannelID: channel.ChannelID,
			Closer:    a.fromAddr,
			Partner:   nonCloser,
		}),
	})

	return receipt, nil
}

func (a *action) updateBalanceProof(update *lnstypes.UpdateBalanceProof) (*types.Receipt, error) {

	elog.Debug("ExecUpdateChannel", "ChannelID", update.ChannelID)
	channel := &lnstypes.Channel{}
	chanKey := calcLnsChannelIDKey(update.ChannelID)

	err := getDBAndDecode(a.db, chanKey, channel)
	if err != nil {
		elog.Error("ExecUpdateChannel", "GetChannelErr", err)
		return nil, err
	}
	partner := address.PubKeyToAddr(update.GetPartnerSignature().GetPubkey())
	participants := getParticipantMap(channel.Participant1, channel.Participant2)

	//check tx 允许非close状态下更新对方的balanceProof
	if !checkParticipantValidity(participants, partner, a.fromAddr) {
		return nil, lnstypes.ErrInvalidChannelParticipants
	}

	if update.PartnerBalancePf.Nonce > participants[partner].Nonce {
		//更新对方的转移额度
		participants[partner].Nonce = update.PartnerBalancePf.Nonce
		participants[partner].TransferredAmount = update.PartnerBalancePf.TransferredAmount
	}

	receipt := &types.Receipt{Ty: types.ExecOk}
	receipt.KV = append(receipt.KV, &types.KeyValue{
		Key:   chanKey,
		Value: types.Encode(channel),
	})

	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{
		Ty: lnstypes.TyUpdateProofLog,
		Log: types.Encode(&lnstypes.ReceiptUpdate{
			TokenCanonicalId: &lnstypes.TokenCanonicalId{
				Chain:         types.GetTitle(),
				IssueContract: channel.GetIssueContract(),
				TokenSymbol:   channel.GetTokenSymbol(),
			},
			ChannelID: channel.ChannelID,
			Nonce:     update.PartnerBalancePf.Nonce,
			Partner:   partner,
			Updater:   a.fromAddr,
		}),
	})

	return receipt, nil
}

func (a *action) settleChannel(settle *lnstypes.Settle) (*types.Receipt, error) {

	elog.Debug("ExecSettleChannel", "ChannelID", settle.ChannelID)
	channel := &lnstypes.Channel{}
	chanKey := calcLnsChannelIDKey(settle.ChannelID)

	err := getDBAndDecode(a.db, chanKey, channel)
	if err != nil {
		elog.Error("ExecWithdrawChannel", "GetChannelErr", err)
		return nil, err
	}
	partner := channel.Participant1.Addr
	if partner == a.fromAddr {
		partner = channel.Participant2.Addr
	}
	participants := getParticipantMap(channel.Participant1, channel.Participant2)

	//check tx
	if channel.State != lnstypes.StateClosed {
		return nil, lnstypes.ErrChannelState
	}
	if channel.SettleBlockHeight > a.height {
		return nil, lnstypes.ErrChannelCloseChallengePeriod
	}

	if !checkParticipantValidity(participants, partner, a.fromAddr) {
		return nil, lnstypes.ErrInvalidChannelParticipants
	}

	myFinalTransAmount := settle.SelfTransferredAmount - settle.PartnerTransferredAmount
	mySettleAmount := participants[a.fromAddr].TotalDeposit -
		participants[a.fromAddr].TotalWithdraw - myFinalTransAmount
	partnerSettleAmount := participants[partner].TotalDeposit -
		participants[partner].TotalWithdraw + myFinalTransAmount

	if mySettleAmount < 0 || partnerSettleAmount < 0 {
		return nil, lnstypes.ErrInvalidTransferredAmount
	}

	channel.State = lnstypes.StateSettled
	receipt := &types.Receipt{Ty: types.ExecOk}
	receipt.KV = append(receipt.KV, &types.KeyValue{
		Key:   chanKey,
		Value: types.Encode(channel),
	})

	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{
		Ty: lnstypes.TySettleLog,
		Log: types.Encode(&lnstypes.ReceiptSettle{
			TokenCanonicalId: &lnstypes.TokenCanonicalId{
				Chain:         types.GetTitle(),
				IssueContract: channel.GetIssueContract(),
				TokenSymbol:   channel.GetTokenSymbol(),
			},
			ChannelID:          channel.ChannelID,
			Participant1:       partner,
			TransferredAmount1: settle.PartnerTransferredAmount,
			Participant2:       a.fromAddr,
			TransferredAmount2: settle.SelfTransferredAmount,
		}),
	})

	accDB, err := a.createAccountDB(channel.IssueContract, channel.TokenSymbol)
	if err != nil {
		elog.Error("ExecSettleChannel", "exec", channel.IssueContract, "symbol", channel.TokenSymbol, "createAccountDBErr", err)
		return nil, err
	}

	//结算自己的账户, 从通道地址结算
	if mySettleAmount > 0 {
		transReceipt, err := accDB.ExecTransfer(a.execAddr, a.fromAddr, a.execAddr, mySettleAmount)
		if err != nil {
			elog.Error("ExecSettleChannel", "execTransferErr", err)
			return nil, err
		}

		receipt.KV = append(receipt.KV, transReceipt.KV...)
		receipt.Logs = append(receipt.Logs, transReceipt.Logs...)
	}

	//结算对方账户
	if partnerSettleAmount > 0 {
		transReceipt, err := accDB.ExecTransfer(a.execAddr, partner, a.execAddr, partnerSettleAmount)
		if err != nil {
			elog.Error("ExecSettleChannel", "execTransferErr", err)
			return nil, err
		}

		receipt.KV = append(receipt.KV, transReceipt.KV...)
		receipt.Logs = append(receipt.Logs, transReceipt.Logs...)
	}

	return receipt, nil
}

func getDBAndDecode(db dbm.KV, key []byte, msg types.Message) error {
	val, err := db.Get(key)
	if err != nil {
		return err
	}

	err = types.Decode(val, msg)
	if err != nil {
		return types.ErrDecode
	}
	return nil
}

func getPartnerAddr(self, p1, p2 string) string {
	if self == p1 {
		return p2
	}
	return p1
}

func getParticipantMap(partner1 *lnstypes.Participant, partner2 *lnstypes.Participant) map[string]*lnstypes.Participant {

	partnerMap := make(map[string]*lnstypes.Participant, 2)
	partnerMap[partner1.Addr] = partner1
	partnerMap[partner2.Addr] = partner2
	return partnerMap
}

func checkParticipantValidity(partMap map[string]*lnstypes.Participant, p1 string, p2 string) bool {

	_, ok1 := partMap[p1]
	_, ok2 := partMap[p2]

	return ok1 && ok2 && (p1 != p2)
}

func (a *action) createAccountDB(exec, symbol string) (*account.DB, error) {

	if exec == "coins" {
		return a.coinsAccount, nil
	}

	return account.NewAccountDB(exec, symbol, a.db)
}
