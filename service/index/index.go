package index

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/internal/gpt"
	"github.com/pandodao/botastic/internal/milvus"
	"github.com/pandodao/botastic/session"
	"github.com/pandodao/tokenizer-go"
	gogpt "github.com/sashabaranov/go-gpt3"
)

func NewService(ctx context.Context, gptHandler *gpt.Handler, indexes core.IndexStore, userz core.UserService) core.IndexService {
	return &serviceImpl{
		gptHandler:                gptHandler,
		indexes:                   indexes,
		userz:                     userz,
		createEmbeddingsLimitChan: make(chan struct{}, 20),
	}
}

type serviceImpl struct {
	gptHandler                *gpt.Handler
	milvusClient              *milvus.Client
	indexes                   core.IndexStore
	userz                     core.UserService
	createEmbeddingsLimitChan chan struct{}
}

func (s *serviceImpl) createEmbeddingsWithLimit(ctx context.Context, req gogpt.EmbeddingRequest, userID uint64) (gogpt.EmbeddingResponse, error) {
	s.createEmbeddingsLimitChan <- struct{}{}
	defer func() {
		<-s.createEmbeddingsLimitChan
	}()

	resp, err := s.gptHandler.CreateEmbeddings(ctx, req)
	if err == nil {
		if err := s.userz.ConsumeCreditsByModel(ctx, userID, "text-embedding-ada-002", uint64(resp.Usage.TotalTokens)); err != nil {
			log.Printf("ConsumeCredits error: %v\n", err)
		}
	}

	return resp, err
}

func (s *serviceImpl) ResetIndexes(ctx context.Context, appID string) error {
	return s.indexes.Reset(ctx, appID)
}

func (s *serviceImpl) SearchIndex(ctx context.Context, userID uint64, query string, limit int) ([]*core.Index, error) {
	if limit <= 0 {
		return nil, errors.New("limit should be greater than 0")
	}

	app := session.AppFrom(ctx)

	resp, err := s.createEmbeddingsWithLimit(ctx, gogpt.EmbeddingRequest{
		Input: []string{query},
		Model: gogpt.AdaEmbeddingV2,
	}, userID)

	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data")
	}
	vs := make([]float32, 0, len(resp.Data[0].Embedding))
	for _, v := range resp.Data[0].Embedding {
		vs = append(vs, float32(v))
	}

	return s.indexes.Search(ctx, app.AppID, vs, limit)
}

func (s *serviceImpl) CreateIndexes(ctx context.Context, userID uint64, items []*core.Index) error {
	input := make([]string, 0, len(items))
	var totalToken uint64
	for _, item := range items {
		token := tokenizer.MustCalToken(item.Data)
		totalToken += uint64(token)
		item.DataToken = int64(token)
		input = append(input, item.Data)
	}

	// @TODO calculate token limit according to the model
	if totalToken > 8191 {
		return core.ErrTokenExceedLimit
	}

	// @TODO should not pending here
	resp, err := s.createEmbeddingsWithLimit(ctx, gogpt.EmbeddingRequest{
		Input: input,
		Model: gogpt.AdaEmbeddingV2,
	}, userID)

	if err != nil {
		return fmt.Errorf("CreateEmbeddings: %w", err)
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
	if err := s.indexes.CreateIndexes(ctx, items); err != nil {
		return fmt.Errorf("indexes.CreateIndexes: %w", err)
	}

	return nil
}
