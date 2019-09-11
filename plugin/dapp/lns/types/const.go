package types

// action类型id和name，这些常量可以自定义修改
const (
	TyUnknowAction = iota + 100
	TyOpenAction
	TyDepositAction
	TyWithdrawAction
	TyCloseAction
	TyUpdateProofAction
	TySettleAction

	NameOpenAction            = "Open"
	NameDepositChannelAction  = "DepositChannel"
	NameWithdrawChannelAction = "WithdrawChannel"
	NameCloseAction           = "Close"
	NameUpdateProofAction     = "UpdateProof"
	NameSettleAction          = "Settle"

	NameOpenLog        = "LogChannelOpen"
	NameDepositLog     = "LogChannelDeposit"
	NameWithdrawLog    = "LogChannelWithdraw"
	NameCloseLog       = "LogChannelClose"
	NameUpdateProofLog = "LogChannelUpdateProof"
	NameSettleLog      = "LogChannelSettle"
)

// log类型id值
const (
	TyUnknownLog = iota + 1000
	TyOpenLog
	TyDepositLog
	TyWithdrawLog
	TyCloseLog
	TyUpdateProofLog
	TySettleLog
)

// channel状态
const (
	StateNonExistent = iota
	StateOpen
	StateClosed
	StateSettled
	StateRemoved
)

// query function name
const (
	FuncQueryGetChannel = "GetChannel"
)
