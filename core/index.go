package core

import (
	"context"
)

type (
	Index struct {
		AppID      string    `json:"app_id,omitempty"`
		Data       string    `json:"data"`
		DataToken  int64     `json:"data_token,omitempty"`
		Vectors    []float32 `json:"vectors,omitempty"`
		ObjectID   string    `json:"object_id"`
		Category   string    `json:"category"`
		Properties string    `json:"properties"`
		CreatedAt  int64     `json:"created_at"`
		Score      float64   `json:"score"`
	}

	IndexStore interface {
		Init(ctx context.Context) error
		Upsert(ctx context.Context, appID string, idx []*Index) error
		Search(ctx context.Context, appId string, vectors []float32, n int) ([]*Index, error)
		Reset(ctx context.Context, appId string) error
		Delete(ctx context.Context, appID string, items []*Index) error
	}

	IndexService interface {
		CreateIndexes(ctx context.Context, userID uint64, appId string, items []*Index) error
		SearchIndex(ctx context.Context, userID uint64, data string, limit int) ([]*Index, error)
		ResetIndexes(ctx context.Context, appID string) error
	}
)

func (ix *Index) Mask() {
	ix.AppID = ""
	ix.DataToken = 0
	ix.Vectors = nil
}
