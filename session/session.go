package session

import (
	"context"
	"time"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pandodao/botastic/core"
)

type Session struct {
	Version string
	store   *mixin.Keystore

	token string
	pin   string

	JwtSecret []byte
}

type JwtClaims struct {
	UserID uint64 `json:"user_id"`
	jwt.RegisteredClaims
}

func (s *Session) WithJWTSecret(secret []byte) *Session {
	s.JwtSecret = secret
	return s
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

	return nil, core.ErrKeystoreNotProvided
}

func (s *Session) GetClient() (*mixin.Client, error) {
	store, err := s.GetKeystore()
	if err != nil {
		return mixin.NewFromAccessToken(s.token), nil
	}

	return mixin.NewFromKeystore(store)
}

// func (s *Session) LoginWithMixin(ctx context.Context, userz core.UserService, authUser *auth.User, lang string) (*core.User, string, error) {
// 	user, err := userz.LoginWithMixin(ctx, authUser, lang)
// 	if err != nil {
// 		return nil, "", err
// 	}

// 	expirationTime := time.Now().Add(24 * 365 * time.Hour)
// 	claims := &JwtClaims{
// 		UserID: user.ID,
// 		RegisteredClaims: jwt.RegisteredClaims{
// 			ExpiresAt: jwt.NewNumericDate(expirationTime),
// 		},
// 	}

// 	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

// 	tokenString, err := newToken.SignedString(s.JwtSecret)
// 	if err != nil {
// 		return nil, "", err
// 	}

// 	return user, tokenString, nil
// }

func (s *Session) GetAccessToken(ctx context.Context, userID uint64) (tokenString string, err error) {
	expirationTime := time.Now().Add(24 * 365 * time.Hour)
	claims := &JwtClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err = newToken.SignedString(s.JwtSecret)
	if err != nil {
		return
	}

	return
}
