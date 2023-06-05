package vector

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"

	"github.com/pandodao/botastic/config"
	"github.com/pandodao/botastic/internal/utils"
	"github.com/redis/go-redis/v9"
)

type redisStore struct {
	cfg    *config.VectorStorageRedisConfig
	client *redis.Client
}

func initRedisStore(ctx context.Context, cfg *config.VectorStorageRedisConfig) (*redisStore, error) {
	if cfg.KeyPrefix == "" {
		cfg.KeyPrefix = "botastic:vector"
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

func (s *redisStore) Upsert(ctx context.Context, groupKey string, vs []*Vector) error {
	pipeline := s.client.Pipeline()
	for _, v := range vs {
		data, _ := json.Marshal(v)
		pipeline.HSet(ctx, s.getkey(groupKey), v.IndexID, data)
	}
	if _, err := pipeline.Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s *redisStore) Reset(ctx context.Context, groupKey string) error {
	return s.client.Del(ctx, s.getkey(groupKey)).Err()
}

func (s *redisStore) Delete(ctx context.Context, groupKey string, indexIDs []uint) error {
	ids := make([]string, len(indexIDs))
	for _, id := range indexIDs {
		ids = append(ids, strconv.Itoa(int(id)))
	}
	return s.client.HDel(ctx, s.getkey(groupKey), ids...).Err()
}

func (s *redisStore) Search(ctx context.Context, groupKey string, vd []float32, n int) ([]*Vector, error) {
	var items []*Vector
	cursor := uint64(0)
	for {
		keysValues, nextCursor, err := s.client.HScan(ctx, s.getkey(groupKey), cursor, "*", 1000).Result()
		if err != nil {
			return nil, err
		}
		cursor = nextCursor
		for i := 0; i < len(keysValues); i += 2 {
			value := keysValues[i+1]
			var v Vector
			err = json.Unmarshal([]byte(value), &v)
			if err != nil {
				continue
			}
			v.Score = utils.CosineSimilarity(vd, v.Data)
			items = append(items, &v)
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

func (s *redisStore) getkey(groupKey string) string {
	return s.cfg.KeyPrefix + ":" + groupKey
}
