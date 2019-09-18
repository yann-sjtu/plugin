package executor

import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	lnstypes "github.com/33cn/plugin/plugin/dapp/lns/types"
)

/*
 * 执行器相关定义
 * 重载基类相关接口
 */

var (
	//日志
	elog = log.New("module", "lns.executor")
)

var driverName = lnstypes.LnsX

func init() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&lns{}))
}

// Init register dapp
func Init(name string, sub []byte) {
	drivers.Register(GetName(), newLns, types.GetDappFork(driverName, "Enable"))
}

type lns struct {
	drivers.DriverBase
}

func newLns() drivers.Driver {
	t := &lns{}
	t.SetChild(t)
	t.SetExecutorType(types.LoadExecutorType(driverName))
	return t
}

// GetName get driver name
func GetName() string {
	return newLns().GetName()
}

func (*lns) GetDriverName() string {
	return driverName
}

// CheckTx 实现自定义检验交易接口，供框架调用
func (l *lns) CheckTx(tx *types.Transaction, index int) error {
	// implement code

	lnsAction := &lnstypes.LnsAction{}
	err := types.Decode(tx.Payload, lnsAction)
	if err != nil {
		return types.ErrActionNotSupport
	}
	fromAddr := tx.From()

	switch lnsAction.Ty {

	case lnstypes.TyOpenAction:
		open := lnsAction.GetOpen()
		//amount必须大于0, 进行链上实际转账,可以校验token是否非法, 否则如果是非法的token, 会导致开辟一个无效channel, 占用资源
		if open == nil || open.Partner == "" || open.IssueContract == "" ||
			(open.IssueContract != "coins" && open.TokenSymbol == "") && open.GetAmount() <= 0 {
			return types.ErrInvalidParam
		}

		if address.CheckAddress(open.Partner) != nil || fromAddr == open.Partner {
			return lnstypes.ErrInvalidChannelParticipants
		}

	case lnstypes.TyDepositAction:

		deposit := lnsAction.GetDepositChannel()
		if deposit == nil || deposit.ChannelID <= 0 || deposit.TotalDeposit <= 0 {
			return types.ErrInvalidParam
		}

	case lnstypes.TyWithdrawAction:

		withdraw := lnsAction.GetWithdrawChannel()
		proof := withdraw.GetProof()
		if withdraw == nil || withdraw.ChannelID <= 0 ||
			proof == nil || proof.TotalWithdraw <= 0 ||
			withdraw.PartnerSignature == nil {
			return types.ErrInvalidParam
		}

		//balance proof的channel信息必须匹配(包括id和链名称一致), 否则会导致利用其他通道的数据进行攻击
		if withdraw.GetChannelID() != proof.GetChannelID() ||
			proof.GetTokenCanonicalId().GetChain() != types.GetTitle() {
			return lnstypes.ErrChannelInfoNotMatch
		}

		if fromAddr != withdraw.GetProof().GetWithdrawer() {
			elog.Error("CheckWithdrawChannelTx", "ChannelID", proof.GetChannelID(), "err", lnstypes.ErrWithdrawSign)
			return lnstypes.ErrWithdrawSign
		}
		if proof.GetExpirationBlock() <= l.GetHeight() {
			elog.Error("CheckWithdrawChannelTx", "ChannelID", proof.GetChannelID(), "currentHeight", l.GetHeight(),
				"expireBlock", proof.GetExpirationBlock())
			return lnstypes.ErrWithdrawBlockExpiration
		}

		if err := checkSign(proof, withdraw.PartnerSignature); err != nil {
			elog.Error("CheckWithdrawChannelTx", "ChannelID", proof.GetChannelID(), "CheckPartnerSignErr", err)
			return lnstypes.ErrPartnerSign
		}

	case lnstypes.TyCloseAction:
		close := lnsAction.GetClose()
		if close == nil || close.ChannelID <= 0 ||
			close.NonCloserBalancePf.TransferredAmount < 0 {
			return types.ErrInvalidParam
		}
		//balance proof的channel信息必须匹配(包括id和链名称一致), 否则会导致利用其他通道的数据进行攻击
		if close.GetChannelID() != close.GetNonCloserBalancePf().GetChannelID() ||
			close.GetNonCloserBalancePf().GetTokenCanonicalId().GetChain() != types.GetTitle() {
			return lnstypes.ErrChannelInfoNotMatch
		}

		//verify signature
		if err := checkSign(close.GetNonCloserBalancePf(), close.GetNonCloserSignature()); err != nil {
			elog.Error("CheckCloseChannelTx", "ChannelID", close.GetChannelID(), "CheckNonCloserSignErr", err)
			return lnstypes.ErrPartnerSign
		}

	case lnstypes.TyUpdateProofAction:
		update := lnsAction.GetUpdateProof()
		if update == nil || update.ChannelID <= 0 ||
			update.PartnerBalancePf.TransferredAmount < 0 {
			return types.ErrInvalidParam
		}

		//balance proof的channel信息必须匹配(包括id和链名称一致), 否则会导致利用其他通道的数据进行攻击
		if update.GetChannelID() != update.GetPartnerBalancePf().GetChannelID() ||
			update.GetPartnerBalancePf().GetTokenCanonicalId().GetChain() != types.GetTitle() {
			return lnstypes.ErrChannelInfoNotMatch
		}

		//verify signature
		if err := checkSign(update.GetPartnerBalancePf(), update.GetPartnerSignature()); err != nil {
			elog.Error("CheckUpdateProofTx", "ChannelID", update.GetChannelID(), "CheckPartnerSignErr", err)
			return lnstypes.ErrPartnerSign
		}
	case lnstypes.TySettleAction:
		settle := lnsAction.GetSettle()
		if settle == nil || settle.ChannelID <= 0 {
			return types.ErrInvalidParam
		}
	default:
		return types.ErrActionNotSupport

	}

	return nil
}

func checkSign(msg types.Message, sign *types.Signature) error {

	signData := common.Sha256(types.Encode(msg))
	c, err := crypto.New(crypto.GetName(int(sign.GetTy())))
	if err != nil {
		return err
	}
	pub, err := c.PubKeyFromBytes(sign.Pubkey)
	if err != nil {
		return err
	}
	signBytes, err := c.SignatureFromBytes(sign.Signature)
	if err != nil {
		return err
	}

	if !pub.VerifyBytes(signData, signBytes) {
		return types.ErrSign
	}

	return nil
}
