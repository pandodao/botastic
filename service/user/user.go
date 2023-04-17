package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/pandodao/botastic/core"
	"github.com/pandodao/passport-go/auth"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/logger"
)

func New(
	cfg Config,
	client *mixin.Client,
	users core.UserStore,
) *UserService {
	return &UserService{
		cfg:    cfg,
		client: client,
		users:  users,
	}
}

type Config struct {
	InitUserCredits float64
}

type UserService struct {
	cfg    Config
	client *mixin.Client
	users  core.UserStore
}

func (s *UserService) ReplaceStore(users core.UserStore) core.UserService {
	return New(s.cfg, s.client, users)
}

func (s *UserService) LoginWithMixin(ctx context.Context, authUser *auth.User, lang string) (*core.User, error) {

	if len(lang) >= 2 {
		lang = strings.ToLower(lang[:2])
	} else {
		lang = "en"
	}

	var user = &core.User{
		Lang:                lang,
		MixinUserID:         authUser.UserID,
		MixinIdentityNumber: authUser.IdentityNumber,
		FullName:            authUser.FullName,
		AvatarURL:           authUser.AvatarURL,
		MvmPublicKey:        authUser.MvmAddress.Hex(),
	}

	existing, err := s.users.GetUserByMixinID(ctx, user.MixinUserID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		fmt.Printf("err users.GetUserByMixinID: %v\n", err)
		return nil, err
	}

	// create
	if errors.Is(err, gorm.ErrRecordNotFound) {
		newUserId, err := s.users.CreateUser(ctx, user.FullName, user.AvatarURL, user.MixinUserID, user.MixinIdentityNumber,
			user.Lang, user.MvmPublicKey, decimal.NewFromFloat(s.cfg.InitUserCredits))
		if err != nil {
			fmt.Printf("err users.Create: %v\n", err)
			return nil, err
		}

		user.ID = newUserId

		// create conversation for messenger user.
		if user.IsMessengerUser() {
			if _, err := s.client.CreateContactConversation(ctx, user.MixinUserID); err != nil {
				return nil, err
			}
		}
		return user, nil
	}

	// update
	if err := s.users.UpdateInfo(ctx, existing.ID, user.FullName, user.AvatarURL, lang); err != nil {
		fmt.Printf("err users.Updates: %v\n", err)
		return nil, err
	}

	return existing, nil
}

func (s *UserService) Topup(ctx context.Context, user *core.User, amount decimal.Decimal) error {
	newAmount := user.Credits.Add(amount)

	err := s.users.UpdateCredits(ctx, user.ID, newAmount)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) ConsumeCreditsByModel(ctx context.Context, userID uint64, model core.Model, promptTokenCount, completionTokenCount int64) error {
	log := logger.FromContext(ctx).WithField("service", "user.ConsumeCreditsByModel")
	cost := model.CalculateTokenCost(promptTokenCount, completionTokenCount)
	log.Printf("model: %s:%s, cost: $%s, token: %d->%d, credits: $%s\n", model.Provider, model.ProviderModel,
		cost.StringFixed(8), promptTokenCount, completionTokenCount, cost.StringFixed(8))
	return s.ConsumeCredits(ctx, userID, cost)
}

func (s *UserService) ConsumeCredits(ctx context.Context, userID uint64, amount decimal.Decimal) error {
	user, err := s.users.GetUser(ctx, userID)
	if err != nil {
		return err
	}
	newAmount := user.Credits.Sub(amount)
	if newAmount.LessThan(decimal.Zero) {
		newAmount = decimal.Zero
	}
	if err := s.users.UpdateCredits(ctx, user.ID, newAmount); err != nil {
		return err
	}
	return nil
}
