package core

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/pandodao/botastic/util"
)

type (
	App struct {
		ID                 uint64 `json:"id"`
		AppID              string `json:"app_id"`
		AppSecret          string `json:"-"`
		AppSecretEncrypted string `json:"-"`

		CreatedAt *time.Time `db:"created_at" json:"created_at"`
		UpdatedAt *time.Time `db:"updated_at" json:"updated_at"`
		DeletedAt *time.Time `db:"deleted_at" json:"-"`
	}

	AppStore interface {
		// SELECT
		// 	"id", "app_id", "app_secret_encrypted", "created_at", "updated_at"
		// FROM @@table WHERE
		// 	"id"=@id AND "deleted_at" IS NULL
		// LIMIT 1
		GetApp(ctx context.Context, id uint64) (*App, error)

		// SELECT
		// 	"id", "app_id", "app_secret_encrypted", "created_at", "updated_at"
		// FROM @@table WHERE
		// 	"deleted_at" IS NULL
		GetApps(ctx context.Context) ([]*App, error)

		// SELECT
		// 	"id", "app_id", "app_secret_encrypted", "created_at", "updated_at"
		// FROM @@table WHERE
		// 	"app_id"=@appID AND "deleted_at" IS NULL
		// LIMIT 1
		GetAppByAppID(ctx context.Context, appID string) (*App, error)

		// INSERT INTO @@table
		// 	("app_id", "app_secret_encrypted", "created_at", "updated_at")
		// VALUES
		// 	(@appID, @appSecretEncrypted, NOW(), NOW())
		// ON CONFLICT ("app_id") DO NOTHING
		// RETURNING "id"
		CreateApp(ctx context.Context, appID, appSecretEncrypted string) (uint64, error)

		// UPDATE @@table
		// 	{{set}}
		// 		"app_secret_encrypted"=@appSecretEncrypted,
		// 		"updated_at"=NOW()
		// 	{{end}}
		// WHERE
		// 	"id"=@id
		UpdateAppSecret(ctx context.Context, id uint64, appSecretEncrypted string) error
	}

	AppService interface {
		CreateApp(ctx context.Context) (*App, error)
		GetAppByAppID(ctx context.Context, appID string) (*App, error)
	}
)

func (s *App) Encrypt(rawkey string) error {
	if s.AppSecretEncrypted == "" {
		key, err := hex.DecodeString(rawkey)
		if err != nil {
			return err
		}

		encrypted, err := util.Encrypt(key, s.AppSecret)
		if err != nil {
			return err
		}

		s.AppSecretEncrypted = encrypted
	}
	return nil
}

func (s *App) Decrypt(rawkey string) error {
	if s.AppSecretEncrypted != "" {
		key, err := hex.DecodeString(rawkey)
		if err != nil {
			return err
		}

		decrypted, err := util.Decrypt(key, s.AppSecretEncrypted)
		if err != nil {
			return err
		}

		s.AppSecret = decrypted
	}
	return nil
}
