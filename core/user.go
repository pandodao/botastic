package core

import (
	"context"
	"time"

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
		Lang                string          `json:"-"`
		Credits             decimal.Decimal `json:"credits"`

		CreatedAt *time.Time `json:"created_at"`
		UpdatedAt *time.Time `json:"-"`
		DeletedAt *time.Time `json:"-"`
	}

	UserStore interface {

		// SELECT
		//   "id", "mixin_user_id", "mixin_identity_number", "full_name", "avatar_url",
		//   "mvm_public_key",
		//   "lang", "credits",
		//   "created_at", "updated_at"
		// FROM @@table WHERE
		// 	 "id"=@id AND "deleted_at" IS NULL
		// LIMIT 1
		GetUser(ctx context.Context, id uint64) (*User, error)

		// SELECT
		//   "id", "mixin_user_id", "mixin_identity_number", "full_name", "avatar_url",
		//   "mvm_public_key",
		//   "lang", "credits",
		//   "created_at", "updated_at"
		// FROM @@table WHERE
		// 	 "mixin_user_id"=@mixinUserID AND "deleted_at" IS NULL
		// LIMIT 1
		GetUserByMixinID(ctx context.Context, mixinUserID string) (*User, error)

		// INSERT INTO @@table
		// (
		//	 "full_name", "avatar_url",
		//	 "mixin_user_id", "mixin_identity_number",
		//   "lang", "credits",
		//   "mvm_public_key",
		//	 "created_at", "updated_at"
		// )
		// VALUES
		// (
		//   @fullName, @avatarURL,
		//   @mixinUserID, @mixinIdentityNumber,
		//   @lang, @credits,
		//   @mvmPublicKey,
		//   NOW(), NOW()
		// )
		// RETURNING "id"
		CreateUser(ctx context.Context, fullName, avatarURL, mixinUserID, mixinIdentityNumber, lang, mvmPublicKey string, credits decimal.Decimal) (uint64, error)

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
		LoginWithMixin(ctx context.Context, token, pubkey, lang string) (*User, error)
		Topup(ctx context.Context, user *User, amount decimal.Decimal) error
		ConsumeCredits(ctx context.Context, userID uint64, amount decimal.Decimal) error
		// ConsumeCreditsByModel(ctx context.Context, userID uint64, model string, amount uint64) error
		ConsumeCreditsByModel(ctx context.Context, userID uint64, model string, promptTokenCount, completionTokenCount int64) error
		ReplaceStore(store UserStore) UserService
	}
)

func (u *User) IsMessengerUser() bool {
	return u.MixinIdentityNumber != "0" && len(u.MixinIdentityNumber) < 10
}
