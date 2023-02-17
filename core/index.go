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
		Similarity float64    `gorm:"-" json:"similarity"`
	}

	IndexStore interface {
		// INSERT INTO @@table
		// 	("app_id", "data", "vectors", "object_id", "index_name", "category", "properties", "created_at", "updated_at")
		// VALUES
		// 	(@idx.AppID, @idx.Data, @idx.Vectors, @idx.ObjectID, @idx.IndexName, @idx.Category, @idx.Properties, NOW(), NOW())
		// ON CONFLICT ("app_id", "object_id") DO
		// 	UPDATE SET "data" = @idx.Data, "vectors" = @idx.Vectors, "index_name" = @idx.IndexName, "category" = @idx.Category, "properties" = @idx.Properties, "updated_at" = NOW()
		UpsertIndex(ctx context.Context, idx *Index) error

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
