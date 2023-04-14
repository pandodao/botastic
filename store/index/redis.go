package index

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/pandodao/botastic/config"
	"github.com/pandodao/botastic/core"
	"github.com/redis/go-redis/v9"
)

type redisStore struct {
	cfg    *config.IndexStoreRedisConfig
	client *redis.Client
}

func initRedisStore(ctx context.Context, cfg *config.IndexStoreRedisConfig) (*redisStore, error) {
	if cfg.KeyPrefix == "" {
		cfg.KeyPrefix = "app"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &redisStore{
		cfg:    cfg,
		client: rdb,
	}, nil
}

func (s *redisStore) Init(ctx context.Context) error {
	return nil
}

func (s *redisStore) Upsert(ctx context.Context, appID string, idx []*core.Index) error {
	pipeline := s.client.Pipeline()
	now := time.Now()
	for _, ix := range idx {
		ix.CreatedAt = now.Unix()
		data, _ := json.Marshal(ix)
		pipeline.HSet(ctx, s.getkey(appID), ix.ObjectID, data)
	}
	_, err := pipeline.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *redisStore) Reset(ctx context.Context, appId string) error {
	return s.client.Del(ctx, s.getkey(appId)).Err()
}

func (s *redisStore) Delete(ctx context.Context, appID string, items []*core.Index) error {
	ids := make([]string, len(items))
	for _, ix := range items {
		ids = append(ids, ix.ObjectID)
	}

	return s.client.HDel(ctx, s.getkey(appID), ids...).Err()
}

func (s *redisStore) Search(ctx context.Context, appID string, vectors []float32, n int) ([]*core.Index, error) {
	var items []*core.Index
	cursor := uint64(0)
	for {
		keysValues, nextCursor, err := s.client.HScan(ctx, s.getkey(appID), cursor, "*", 1000).Result()
		if err != nil {
			return nil, err
		}
		cursor = nextCursor
		for i := 0; i < len(keysValues); i += 2 {
			value := keysValues[i+1]
			var ix core.Index
			err = json.Unmarshal([]byte(value), &ix)
			if err != nil {
				continue
			}
			ix.Score = cosineSimilarity(vectors, ix.Vectors)
			items = append(items, &ix)
		}
		if cursor == 0 {
			break
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Score > items[j].Score
	})

	if len(items) > n {
		return items[:n], nil
	}

	return items, nil
}

func (s *redisStore) getkey(appID string) string {
	return s.cfg.KeyPrefix + ":" + appID
}
