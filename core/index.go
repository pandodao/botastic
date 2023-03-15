package core

import (
	"context"
	"strings"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

type (
	Index struct {
		ID         string    `json:"-"`
		AppID      string    `json:"-"`
		Data       string    `json:"data"`
		DataToken  int64     `json:"-"`
		Vectors    []float32 `json:"-"`
		ObjectID   string    `json:"object_id"`
		Category   string    `json:"category"`
		Properties string    `json:"properties"`
		CreatedAt  int64     `db:"created_at" json:"created_at"`
		Score      float32   `json:"score"`
	}

	IndexStore interface {
		CreateIndices(ctx context.Context, idx []*Index) error
		Search(ctx context.Context, appId string, vectors []float32, n int) ([]*Index, error)
		DeleteByPks(ctx context.Context, items []*Index) error
	}

	IndexService interface {
		CreateIndices(ctx context.Context, userID uint64, items []*Index) error
		SearchIndex(ctx context.Context, userID uint64, data string, limit int) ([]*Index, error)
	}
)

func (i Index) CollectionName() string {
	return "indices"
}

func (i Index) PartitionName() string {
	return strings.ReplaceAll(i.AppID, "-", "_")
}

// func (i Index) PartitionName() string {
// 	return fmt.Sprintf("%d_%s", i.AppID, i.IndexName)
// }

func (i Index) Schema() *entity.Schema {
	return &entity.Schema{
		CollectionName: i.CollectionName(),
		AutoID:         true,
		Fields: []*entity.Field{
			{
				Name:       "id",
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: true,
				TypeParams: map[string]string{"max_length": "64"},
			},
			{
				Name:       "app_id",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "32"},
			},
			{
				Name:       "object_id",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "32"},
			},
			{
				Name:       "data",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "2048"},
			},
			{
				Name:     "data_token",
				DataType: entity.FieldTypeInt64,
			},
			{
				Name:       "vectors",
				DataType:   entity.FieldTypeFloatVector,
				TypeParams: map[string]string{"dim": "1536"},
			},
			{
				Name:       "category",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "32"},
			},
			{
				Name:       "properties",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "1024"},
			},
			{
				Name:     "created_at",
				DataType: entity.FieldTypeInt64,
			},
		},
	}
}
