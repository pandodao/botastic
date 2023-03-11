package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/pandodao/botastic/core"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/ethereum/go-ethereum/common"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/passport-go/mvm"
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
	MixinClientSecret string
}

type UserService struct {
	cfg    Config
	client *mixin.Client
	users  core.UserStore
}

func (s *UserService) LoginWithMixin(ctx context.Context, token, pubkey, lang string) (*core.User, error) {
	var cli *mixin.Client
	if lang == "" {
		lang = "en"
	}

	if len(lang) >= 2 {
		lang = strings.ToLower(lang[:2])
	} else {
		lang = "en"
	}

	var user = &core.User{
		Lang: lang,
	}

	if token != "" {
		cli = mixin.NewFromAccessToken(token)
		profile, err := cli.UserMe(ctx)
		if err != nil {
			fmt.Printf("err cli.UserMe: %v\n", err)
			return nil, err
		}

		contractAddr, err := mvm.GetUserContract(ctx, profile.UserID)
		if err != nil {
			fmt.Printf("err mvm.GetUserContract: %v\n", err)
			return nil, err
		}

		// if contractAddr is not 0x000..00, it means the user has already registered a mvm account
		// we should not allow the user to login with mixin token
		emptyAddr := common.Address{}
		if contractAddr != emptyAddr {
			return nil, core.ErrBadMvmLoginMethod
		}

		user.MixinUserID = profile.UserID
		user.MixinIdentityNumber = profile.IdentityNumber
		user.FullName = profile.FullName
		user.AvatarURL = profile.AvatarURL

	} else if pubkey != "" {
		addr := common.HexToAddress(pubkey)
		mvmUser, err := mvm.GetBridgeUser(ctx, addr)
		if err != nil {
			fmt.Printf("err mvm.GetBridgeUser: %v\n", err)
			return nil, err
		}
		user.MixinUserID = mvmUser.UserID
		user.MixinIdentityNumber = "0"
		user.FullName = mvmUser.FullName
		user.AvatarURL = ""
		user.MvmPublicKey = pubkey

	} else {
		return nil, core.ErrInvalidAuthParams
	}

	existing, err := s.users.GetUserByMixinID(ctx, user.MixinUserID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		fmt.Printf("err users.GetUserByMixinID: %v\n", err)
		return nil, err
	}

	// create
	if errors.Is(err, gorm.ErrRecordNotFound) {
		newUserId, err := s.users.CreateUser(ctx, user.FullName, user.AvatarURL, user.MixinUserID, user.MixinIdentityNumber,
			user.Lang, user.MvmPublicKey)
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

func (s *UserService) ConsumeCreditsByModel(ctx context.Context, userID uint64, model string, tokenCount uint64) error {
	price := decimal.Zero
	switch model {
	case "gpt-3.5-turbo":
		// $0.002 per 1000 tokens
		price = decimal.NewFromFloat(0.000002)
	case "text-davinci-003":
		// $0.02 per 1000 tokens
		price = decimal.NewFromFloat(0.00002)
	case "text-embedding-ada-002":
		// $0.0004 per 1000 tokens
		price = decimal.NewFromFloat(0.0000004)
	default:
		return core.ErrInvalidModel
	}

	credits := price.Mul(decimal.NewFromInt(int64(tokenCount)))
	fmt.Printf("model: %v, price: $%s, token: %d, credits: $%s\n", model, price.StringFixed(8), tokenCount, credits.StringFixed(8))
	return s.ConsumeCredits(ctx, userID, credits)
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
