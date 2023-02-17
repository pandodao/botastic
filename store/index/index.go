package index

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/internal/gpt"
	"github.com/pandodao/botastic/session"
	"github.com/pandodao/botastic/store"
	"github.com/pandodao/botastic/store/index/dao"
	gogpt "github.com/sashabaranov/go-gpt3"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func init() {
	cfg := gen.Config{
		OutPath: "store/index/dao",
	}
	store.RegistGenerate(
		cfg,
		func(g *gen.Generator) {
			g.ApplyInterface(func(core.IndexStore) {}, core.Index{})
		},
	)
}

func New(db *gorm.DB) core.IndexStore {
	v, ok := interface{}(dao.Index).(core.IndexStore)
	if !ok {
		panic("dao.Index is not core.IndexStore")
	}
	dao.SetDefault(db)
	return &storeImpl{IndexStore: v}
}

func NewService(ctx context.Context, gptHandler *gpt.Handler, indexes core.IndexStore) (core.IndexService, error) {
	is, err := indexes.GetIndexes(ctx)
	if err != nil {
		return nil, err
	}

	si := &serviceImpl{
		indexData:  map[string]map[string]*core.Index{},
		gptHandler: gptHandler,
		indexes:    indexes,
	}

	for _, i := range is {
		si.indexData[fmt.Sprintf("%d:%s", i.AppID, i.IndexName)][i.ObjectID] = i
	}
	return si, nil
}

type storeImpl struct {
	core.IndexStore
}

type serviceImpl struct {
	indexData  map[string]map[string]*core.Index
	gptHandler *gpt.Handler
	indexes    core.IndexStore
}

func (s *serviceImpl) SearchIndex(ctx context.Context, indexName, keywords string, limit int) ([]*core.Index, error) {
	if limit <= 0 {
		return nil, errors.New("limit should be greater than 0")
	}

	app := session.AppFrom(ctx)
	if app == nil {
		return nil, fmt.Errorf("app is nil")
	}

	key := fmt.Sprintf("%d:%s", app.ID, indexName)
	if indexName != "" {
		if s.indexData[key] == nil {
			return []*core.Index{}, nil
		}
	}

	resp, err := s.gptHandler.CreateEmbeddings(ctx, gogpt.EmbeddingRequest{
		Input: []string{keywords},
		Model: gogpt.AdaEmbeddingV2,
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data")
	}

	indexMap := map[float64]*core.Index{}
	cosines := make([]float64, 0)
	if indexName != "" {
		for _, d := range s.indexData[key] {
			cosine, err := Cosine(d.Vectors, resp.Data[0].Embedding)
			if err != nil {
				continue
			}
			cosines = append(cosines, cosine)
			indexMap[cosine] = d
		}
	} else {
		for key, appIndexData := range s.indexData {
			if !strings.HasPrefix(key, fmt.Sprintf("%d:", app.ID)) {
				continue
			}
			for _, d := range appIndexData {
				cosine, err := Cosine(d.Vectors, resp.Data[0].Embedding)
				if err != nil {
					continue
				}
				cosines = append(cosines, cosine)
				indexMap[cosine] = d
			}
		}
	}

	sort.Slice(cosines, func(i, j int) bool {
		return cosines[i] > cosines[j]
	})

	result := make([]*core.Index, 0, limit)
	for i := 0; i < limit && i < len(cosines); i++ {
		idx := indexMap[cosines[i]]
		idx.Similarity = cosines[i]
		result[i] = idx
	}

	return result, nil
}

func (s *serviceImpl) CreateIndex(ctx context.Context, indexName, objectID, category, properties, data string) error {
	app := session.AppFrom(ctx)
	if app == nil {
		return fmt.Errorf("app is nil")
	}
	key := fmt.Sprintf("%d:%s", app.ID, indexName)
	appIndexData := s.indexData[key]

	resp, err := s.gptHandler.CreateEmbeddings(ctx, gogpt.EmbeddingRequest{
		Input: []string{data},
		Model: gogpt.AdaEmbeddingV2,
	})
	if err != nil {
		return err
	}
	if len(resp.Data) == 0 {
		return fmt.Errorf("no embedding data")
	}

	respResult := resp.Data[0]
	idx := &core.Index{
		AppID:      app.ID,
		Data:       data,
		Vectors:    respResult.Embedding,
		ObjectID:   objectID,
		IndexName:  indexName,
		Category:   category,
		Properties: properties,
	}
	err = s.indexes.UpsertIndex(ctx, idx)
	if err != nil {
		return err
	}

	if appIndexData == nil {
		appIndexData = make(map[string]*core.Index)
		s.indexData[key] = appIndexData
	}
	appIndexData[objectID] = idx

	return nil
}

func Cosine(a []float64, b []float64) (cosine float64, err error) {
	count := 0
	length_a := len(a)
	length_b := len(b)
	if length_a > length_b {
		count = length_a
	} else {
		count = length_b
	}
	sumA := 0.0
	s1 := 0.0
	s2 := 0.0
	for k := 0; k < count; k++ {
		if k >= length_a {
			s2 += math.Pow(b[k], 2)
			continue
		}
		if k >= length_b {
			s1 += math.Pow(a[k], 2)
			continue
		}
		sumA += a[k] * b[k]
		s1 += math.Pow(a[k], 2)
		s2 += math.Pow(b[k], 2)
	}
	if s1 == 0 || s2 == 0 {
		return 0.0, errors.New("Vectors should not be null (all zeros)")
	}
	return sumA / (math.Sqrt(s1) * math.Sqrt(s2)), nil
}
