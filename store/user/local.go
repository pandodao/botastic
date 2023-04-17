package user

import (
	"context"

	"github.com/pandodao/botastic/core"
	"github.com/shopspring/decimal"
)

type localModeStore struct {
	core.UserStore
	user *core.User
}

func NewLocalModeStore(user *core.User) *localModeStore {
	return &localModeStore{
		user: user,
	}
}

func (s *localModeStore) GetUser(ctx context.Context, id uint64) (*core.User, error) {
	return s.user, nil
}

func (s *localModeStore) UpdateCredits(ctx context.Context, id uint64, amount decimal.Decimal) error {
	return nil
}
