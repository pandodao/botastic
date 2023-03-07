package core

import (
	"context"
	"time"
)

type (
	User struct {
		ID                  uint64 `json:"id"`
		MixinUserID         string `json:"mixin_user_id"`
		MixinIdentityNumber string `json:"mixin_identity_number"`
		FullName            string `json:"full_name"`
		AvatarURL           string `json:"avatar_url"`
		MvmPublicKey        string `json:"mvm_public_key"`
		Lang                string `json:"-"`

		CreatedAt *time.Time `json:"created_at"`
		UpdatedAt *time.Time `json:"-"`
		DeletedAt *time.Time `json:"-"`
	}

	UserStore interface {

		// SELECT
		//   "id", "mixin_user_id", "mixin_identity_number", "full_name", "avatar_url",
		//   "mvm_public_key",
		//   "lang",
		//   "created_at", "updated_at"
		// FROM @@table WHERE
		// 	 "id"=@id AND "deleted_at" IS NULL
		// LIMIT 1
		GetUser(ctx context.Context, id uint64) (*User, error)

		// SELECT
		//   "id", "mixin_user_id", "mixin_identity_number", "full_name", "avatar_url",
		//   "mvm_public_key",
		//   "lang",
		//   "created_at", "updated_at"
		// FROM @@table WHERE
		// 	 "mixin_user_id"=@mixinUserID AND "deleted_at" IS NULL
		// LIMIT 1
		GetUserByMixinID(ctx context.Context, mixinUserID string) (*User, error)

		// INSERT INTO @@table
		// (
		//	 "full_name", "avatar_url",
		//	 "mixin_user_id", "mixin_identity_number",
		//	 "lang", "mvm_public_key",
		//	 "created_at", "updated_at"
		// )
		// VALUES
		// (
		//   @fullName, @avatarURL,
		//   @mixinUserID, @mixinIdentityNumber,
		//   @lang, @mvmPublicKey,
		//   NOW(), NOW()
		// )
		// RETURNING "id"
		CreateUser(ctx context.Context, fullName, avatarURL, mixinUserID, mixinIdentityNumber, lang, mvmPublicKey string) (uint64, error)

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
	}

	UserService interface {
		LoginWithMixin(ctx context.Context, token, pubkey, lang string) (*User, error)
	}
)

func (u *User) IsMessengerUser() bool {
	return u.MixinIdentityNumber != "0" && len(u.MixinIdentityNumber) < 10
}
