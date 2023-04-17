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
	ErrInternalServer = errors.New("internal server error")

	ErrInsufficientCredit = errors.New("insufficient credit")

	ErrMinAmountNotSatisfied = errors.New("min amount not satisfied")

	ErrCorruptedUUID = errors.New("corrupted uuid")

	ErrBotNotFound = errors.New("bot not found")

	ErrConvNotFound = errors.New("conversation not found")

	ErrMiddlewareNotFound = errors.New("middleware not found")

	ErrAppNotFound = errors.New("app not found")

	ErrUnauthorized = errors.New("unauthorized")

	ErrAppLimitReached = errors.New("reach app limit")

	ErrBadMvmLoginMethod = errors.New("bad mvm login method")

	ErrBadMvmLoginSignature = errors.New("bad mvm login signature")

	ErrBadMvmLoginMessage = errors.New("bad mvm login message")

	ErrInvalidUserID = errors.New("invalid user id")

	ErrNoRecord = errors.New("no record")

	ErrInvalidAuthParams = errors.New("invalid auth params")

	ErrKeystoreNotProvided = errors.New("keystore not provided")

	ErrConvTurnNotProcessed = errors.New("conversation turn is not processed")

	ErrBotIncorrectField = errors.New("bot incorrect field")

	ErrBotUnsupportedModel = errors.New("bot unsupported model")

	ErrInvalidModel = errors.New("invalid model")

	ErrTokenExceedLimit = errors.New("token exceed limit")

	ErrIncorrectOrderStatus = errors.New("incorrect order status")

	ErrIncorrectOrderUser = errors.New("incorrect order user")
)
