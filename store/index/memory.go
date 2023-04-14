package index

import (
	"context"
	"sort"
	"time"

	"github.com/pandodao/botastic/core"
)

type memoryStore struct {
	m map[string]map[string]*core.Index
}

func initMemoryStore() *memoryStore {
	return &memoryStore{
		m: make(map[string]map[string]*core.Index),
	}
}

func (s *memoryStore) Init(ctx context.Context) error {
	return nil
}

func (s *memoryStore) Upsert(ctx context.Context, appID string, idx []*core.Index) error {
	indexMap, ok := s.m[appID]
	if !ok {
		indexMap = make(map[string]*core.Index)
		s.m[appID] = indexMap
	}

	now := time.Now()
	for _, i := range idx {
		i = s.cloneIndex(i)
		i.CreatedAt = now.Unix()
		indexMap[i.ObjectID] = s.cloneIndex(i)
	}

	return nil
}

func (s *memoryStore) Search(ctx context.Context, appId string, vectors []float32, n int) ([]*core.Index, error) {
	indexMap, ok := s.m[appId]
	if !ok {
		return []*core.Index{}, nil
	}

	similarities := make([]*core.Index, 0, len(indexMap))

	for _, ix := range indexMap {
		ix = s.cloneIndex(ix)
		ix.Score = cosineSimilarity(ix.Vectors, vectors)
		similarities = append(similarities, ix)
	}

	sort.Slice(similarities, func(i, j int) bool {
		return similarities[i].Score > similarities[j].Score
	})

	if len(similarities) > n {
		return similarities[:n], nil
	}

	return similarities, nil
}

func (s *memoryStore) Reset(ctx context.Context, appId string) error {
	delete(s.m, appId)
	return nil
}

func (s *memoryStore) Delete(ctx context.Context, appID string, items []*core.Index) error {
	indexMap, ok := s.m[appID]
	if !ok {
		return nil
	}

	for _, item := range items {
		delete(indexMap, item.ObjectID)
	}
	return nil
}

func (s *memoryStore) cloneIndex(ix *core.Index) *core.Index {
	newVectors := make([]float32, len(ix.Vectors))
	copy(newVectors, ix.Vectors)

	nix := &core.Index{}
	*nix = *ix
	nix.Vectors = newVectors
	return nix
}
