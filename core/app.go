package core

import (
	"context"
	"time"
)

type (
	App struct {
		ID        uint64 `json:"id"`
		AppID     string `json:"app_id"`
		AppSecret string `json:"app_secret"`

		CreatedAt *time.Time `db:"created_at" json:"created_at"`
		UpdatedAt *time.Time `db:"updated_at" json:"updated_at"`
		DeletedAt *time.Time `db:"deleted_at" json:"-"`
	}

	AppStore interface {
		// SELECT
		// 	"id", "app_id", "app_secret", "created_at", "updated_at"
		// FROM @@table WHERE
		// 	"id"=@id AND "deleted_at" IS NULL
		// LIMIT 1
		GetApp(ctx context.Context, id uint64) (*App, error)

		// SELECT
		// 	"id", "app_id", "app_secret", "created_at", "updated_at"
		// FROM @@table WHERE
		// 	"app_id"=@appID AND "deleted_at" IS NULL
		// LIMIT 1
		GetAppByAppID(ctx context.Context, appID string) (*App, error)

		// INSERT INTO @@table
		// 	("app_id", "app_secret", "created_at", "updated_at")
		// VALUES
		// 	(@appID, @appSecret, NOW(), NOW())
		// ON CONFLICT ("app_id") DO NOTHING
		// RETURNING "id"
		CreateApp(ctx context.Context, appID, appSecret string) (uint64, error)

		// UPDATE @@table
		// 	{{set}}
		// 		"app_secret"=@appSecret,
		// 		"updated_at"=NOW()
		// 	{{end}}
		// WHERE
		// 	"id"=@id
		UpdateAppSecret(ctx context.Context, id uint64, appSecret string) error
	}
)
