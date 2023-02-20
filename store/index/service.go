package index

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/internal/gpt"
	"github.com/pandodao/botastic/internal/milvus"
	gogpt "github.com/sashabaranov/go-gpt3"
)

func NewService(ctx context.Context, gptHandler *gpt.Handler, indexes core.IndexStore) core.IndexService {
	return &serviceImpl{
		gptHandler: gptHandler,
		indexes:    indexes,
	}
}

type serviceImpl struct {
	gptHandler   *gpt.Handler
	milvusClient *milvus.Client
	indexes      core.IndexStore
}

func (s *serviceImpl) SearchIndex(ctx context.Context, indexName, keywords string, limit int) ([]*core.Index, error) {
	if limit <= 0 {
		return nil, errors.New("limit should be greater than 0")
	}

	// resp, err := s.gptHandler.CreateEmbeddings(ctx, gogpt.EmbeddingRequest{
	// 	Input: []string{keywords},
	// 	Model: gogpt.AdaEmbeddingV2,
	// })
	// if err != nil {
	// 	return nil, err
	// }
	// if len(resp.Data) == 0 {
	// 	return nil, fmt.Errorf("no embedding data")
	// }

	return nil, nil
}

func (s *serviceImpl) CreateIndices(ctx context.Context, items []*core.Index) error {
	input := make([]string, 0, len(items))
	for _, item := range items {
		input = append(input, item.Data)
	}

	if err := s.indexes.DeleteByPks(ctx, items); err != nil {
		return err
	}

	resp, err := s.gptHandler.CreateEmbeddings(ctx, gogpt.EmbeddingRequest{
		Input: input,
		Model: gogpt.AdaEmbeddingV2,
	})
	if err != nil {
		return err
	}
	if len(resp.Data) == 0 {
		return fmt.Errorf("no embedding data")
	}

	for i, embedding := range resp.Data {
		vs := make([]float32, 0, len(embedding.Embedding))
		for _, v := range embedding.Embedding {
			vs = append(vs, float32(v))
		}
		items[i].Vectors = vs
	}

	// create index in milvus
	return s.indexes.CreateIndices(ctx, items)
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
