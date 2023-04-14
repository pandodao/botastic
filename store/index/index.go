package index

import (
	"context"
	"math"

	"github.com/pandodao/botastic/config"
	"github.com/pandodao/botastic/core"
)

var (
	_ core.IndexStore = (*memoryStore)(nil)
	_ core.IndexStore = (*milvusStore)(nil)
	_ core.IndexStore = (*redisStore)(nil)
)

func Init(ctx context.Context, cfg config.IndexStoreConfig) (core.IndexStore, error) {
	switch cfg.Driver {
	case config.IndexStoreMilvus:
		return initMilvusStore(ctx, cfg.Milvus)
	case config.IndexStoreRedis:
		return initRedisStore(ctx, cfg.Redis)
	default:
		return initMemoryStore(), nil
	}
}

func cosineSimilarity(a, b []float32) float64 {
	dotProduct := float64(0)
	aMagnitude := float64(0)
	bMagnitude := float64(0)

	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i]) * float64(b[i])
		aMagnitude += math.Pow(float64(a[i]), 2)
		bMagnitude += math.Pow(float64(b[i]), 2)
	}

	return dotProduct / (math.Sqrt(aMagnitude) * math.Sqrt(bMagnitude))
}
