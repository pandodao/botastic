package core

import (
	"context"
	"time"

	"github.com/pandodao/passport-go/auth"
	"github.com/shopspring/decimal"
)

type (
	User struct {
		ID                  uint64          `json:"id"`
		MixinUserID         string          `json:"mixin_user_id"`
		MixinIdentityNumber string          `json:"mixin_identity_number"`
		FullName            string          `json:"full_name"`
		AvatarURL           string          `json:"avatar_url"`
		MvmPublicKey        string          `json:"mvm_public_key"`
		Email               string          `json:"email"`
		TwitterID           string          `json:"twitter_id"`
		TwitterScreenName   string          `json:"twitter_screen_name"`
		Lang                string          `json:"-"`
		Credits             decimal.Decimal `json:"credits"`

		CreatedAt *time.Time `json:"created_at"`
		UpdatedAt *time.Time `json:"-"`
		DeletedAt *time.Time `json:"-"`
	}

	UserStore interface {

		// SELECT
		//   *
		// FROM @@table WHERE
		// 	 "id"=@id AND "deleted_at" IS NULL
		// LIMIT 1
		GetUser(ctx context.Context, id uint64) (*User, error)

		// SELECT
		//   *
		// FROM @@table WHERE
		// 	 "mixin_user_id"=@mixinUserID AND "deleted_at" IS NULL
		// LIMIT 1
		GetUserByMixinID(ctx context.Context, mixinUserID string) (*User, error)

		// SELECT
		//   *
		// FROM @@table WHERE
		// 	 "twitter_id"=@twitterID AND "deleted_at" IS NULL
		// LIMIT 1
		GetUserByTwitterID(ctx context.Context, twitterID string) (*User, error)

		// INSERT INTO @@table
		// (
		//	 "full_name", "avatar_url",
		//	 "mixin_user_id", "mixin_identity_number",
		//   "mvm_public_key",
		//   "email", "twitter_id", "twitter_screen_name",
		//   "lang", "credits",
		//	 "created_at", "updated_at"
		// )
		// VALUES
		// (
		//   @user.FullName, @user.AvatarURL,
		//   @user.MixinUserID, @user.MixinIdentityNumber,
		//   @user.MvmPublicKey,
		//   @user.Email, @user.TwitterID, @user.TwitterScreenName,
		//   @user.Lang, @user.Credits,
		//   NOW(), NOW()
		// )
		// RETURNING "id"
		CreateUser(ctx context.Context, user *User) (uint64, error)

		// UPDATE @@table
		// 	{{set}}
		// 		"full_name"=@fullName,
		// 		"avatar_url"=@avatarURL,
		// 		"lang"=@lang,
		// 		"updated_at"=NOW()
		// 	{{end}}
		// WHERE
		// 	"id"=@id
		UpdateInfo(ctx context.Context, id uint64, fullName, avatarURL, lang string) error

		// UPDATE @@table
		// 	{{set}}
		// 		"credits"=@amount,
		// 		"updated_at"=NOW()
		// 	{{end}}
		// WHERE
		// 	"id"=@id
		UpdateCredits(ctx context.Context, id uint64, amount decimal.Decimal) error
	}

	UserService interface {
		LoginWithMixin(ctx context.Context, authUser *auth.User, lang string) (*User, error)
		LoginWithTwitter(ctx context.Context, oauthToken, oauthVerifier, lang string) (*User, error)
		Topup(ctx context.Context, user *User, amount decimal.Decimal) error
		ConsumeCredits(ctx context.Context, userID uint64, amount decimal.Decimal) error
		ConsumeCreditsByModel(ctx context.Context, userID uint64, model Model, promptTokenCount, completionTokenCount int) error
		ReplaceStore(store UserStore) UserService
	}
)

func (u *User) IsMessengerUser() bool {
	return u.MixinIdentityNumber != "0" && len(u.MixinIdentityNumber) < 10
}
