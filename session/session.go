package session

import (
	"errors"

	"github.com/fox-one/mixin-sdk-go"
)

var (
	ErrKeystoreNotProvided = errors.New("keystore not provided, use --file or --stdin")
	ErrPinNotProvided      = errors.New("pin not provided, use --pin or include in keystore file")
)

type Session struct {
	Version string

	store *mixin.Keystore
	token string
	pin   string
}

func (s *Session) WithKeystore(store *mixin.Keystore) *Session {
	s.store = store
	return s
}

func (s *Session) WithAccessToken(token string) *Session {
	s.token = token
	return s
}

func (s *Session) WithPin(pin string) *Session {
	s.pin = pin
	return s
}

func (s *Session) GetKeystore() (*mixin.Keystore, error) {
	if s.store != nil {
		return s.store, nil
	}

	return nil, ErrKeystoreNotProvided
}

func (s *Session) GetPin() (string, error) {
	if s.pin != "" {
		return s.pin, nil
	}

	return "", ErrPinNotProvided
}

func (s *Session) GetClient() (*mixin.Client, error) {
	store, err := s.GetKeystore()
	if err != nil {
		return mixin.NewFromAccessToken(s.token), nil
	}

	return mixin.NewFromKeystore(store)
}
