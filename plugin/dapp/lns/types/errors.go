package types

import "errors"

var (
	ErrChannelIDOverFlow           = errors.New("ErrChannelIDOverFlow")
	ErrChannelState                = errors.New("ErrChannelState")
	ErrInvalidChannelParticipants  = errors.New("ErrInvalidChannelParticipants")
	ErrInvalidWithdrawAmount       = errors.New("ErrInvalidWithdrawAmount")
	ErrInvalidTransferredAmount    = errors.New("ErrInvalidTransferredAmount")
	ErrChannelCloseChallengePeriod = errors.New("ErrChannelCloseChallengePeriod")
	ErrWithdrawBlockExpiration     = errors.New("ErrWithdrawBlockExpiration")
	ErrWithdrawSign                = errors.New("ErrWithdrawSign")
	ErrPartnerSign                 = errors.New("ErrPartnerSign")
	ErrTotalDepositAmount          = errors.New("ErrTotalDepositAmount")
	ErrChannelInfoNotMatch         = errors.New("ErrChannelInfoNotMatch")
)
