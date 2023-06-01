package vector

import (
	"context"

	"github.com/pandodao/botastic/config"
)

func Init(ctx context.Context, cfg config.VectorStorageConfig) (Storage, error) {
	switch cfg.Driver {
	case config.VectorStorageRedis:
		return initRedisStore(ctx, cfg.Redis)
	default:
		return nil, nil
	}
}

type Vector struct {
	IndexID uint      `json:"index_id"`
	Data    []float32 `json:"data"`
	Score   float64   `json:"-"`
}

type Storage interface {
	Upsert(ctx context.Context, groupKey string, vs []*Vector) error
	Search(ctx context.Context, groupKey string, vd []float32, n int) ([]*Vector, error)
	Reset(ctx context.Context, groupKey string) error
	Delete(ctx context.Context, groupKey string, indexIDs []uint) error
}
