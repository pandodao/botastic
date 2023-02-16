package core

import (
	"context"
	"time"
)

type (
	Index struct {
		ID         uint64     `json:"id"`
		AppID      uint64     `json:"app_id"`
		Data       string     `json:"data"`
		Vectors    []float64  `gorm:"type:numeric[]" json:"vectors"`
		ObjectID   string     `json:"object_id"`
		IndexName  string     `json:"index_name"`
		Category   string     `json:"category"`
		Properties string     `json:"properties"`
		CreatedAt  *time.Time `db:"created_at" json:"created_at"`
		UpdatedAt  *time.Time `db:"updated_at" json:"updated_at"`
		DeletedAt  *time.Time `db:"deleted_at" json:"-"`
	}

	IndexStore interface {
		// INSERT INTO @@table
		// 	("app_id", "data", "vectors", "object_id", "index_name", "category", "properties", "created_at", "updated_at")
		// VALUES
		// 	(@appID, @data, @vectors, @objectID, @indexName, @category, @properties, NOW(), NOW())
		// ON CONFLICT ("text") DO NOTHING
		// RETURNING "id"
		CreateIndex(ctx context.Context, appID uint64, data string, vectors []float64,
			objectID string, indexName string, category string, properties string) (uint64, error)

		// SELECT * FROM @@table WHERE
		// 	"deleted_at" IS NULL
		GetIndexes(ctx context.Context) ([]*Index, error)
	}

	IndexService interface {
		CreateIndex(ctx context.Context, data string,
			objectID string, indexName string, category string, properties string) error
		SearchIndex(ctx context.Context, indexName, data string, limit int) ([]*Index, error)
	}
)

