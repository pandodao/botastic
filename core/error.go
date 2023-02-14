package core

import (
	"errors"
)

const (
	FlagRefund  = 1 << iota
	FlagConsume = 1 << iota
)

type wrappedErr struct {
	err  error
	flag int
}

func (e *wrappedErr) Error() string {
	return e.err.Error()
}

func (e *wrappedErr) Unwrap() error {
	return e.err
}

func WrapError(err error, flags ...int) error {
	e := &wrappedErr{
		err: err,
	}

	for _, flag := range flags {
		e.flag = e.flag | flag
	}

	return e
}

func MatchFlag(err error, flag int) bool {
	for err != nil {
		var werr *wrappedErr
		if errors.As(err, &werr) && werr.flag&flag > 0 {
			return true
		}

		err = errors.Unwrap(err)
	}

	return false
}

var (
	ErrInsufficientQuota = errors.New("insufficient product quota")

	ErrProductIDNotMatch = errors.New("product id not match")

	ErrCorruptedEventExtra = errors.New("corrupted event extra")

	ErrMinAmountNotSatisfied = errors.New("min amount not satisfied")

	ErrIncorrectExchangeID = errors.New("incorrect exchange id")

	ErrCorruptedSwapFollowID = errors.New("corrupted swap follow id")

	ErrExchangeAndEventNotMatch = errors.New("exchange and event not match")

	ErrJPYCBotNotAvailable = errors.New("the jpyc bot is not available")

	ErrCorruptedUUID = errors.New("corrupted uuid")

	ErrUnsupportedSocialChannel = errors.New("unsupported social channel")

	ErrSocialTransferExpired = errors.New("social transfer expired")

	ErrSelfInvitation = errors.New("cannot invite oneself")

	ErrFailedToAddVerification = errors.New("cannot add verification record")
)
