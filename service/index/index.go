package index

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/internal/gpt"
	"github.com/pandodao/botastic/internal/tiktoken"
	"github.com/pandodao/botastic/session"
	gogpt "github.com/sashabaranov/go-openai"
)

func NewService(ctx context.Context, gptHandler *gpt.Handler, indexes core.IndexStore, userz core.UserService, models core.ModelStore, tiktokenHandler *tiktoken.Handler) core.IndexService {
	return &serviceImpl{
		gptHandler:                gptHandler,
		indexes:                   indexes,
		userz:                     userz,
		models:                    models,
		createEmbeddingsLimitChan: make(chan struct{}, 20),
		tiktokenHandler:           tiktokenHandler,
	}
}

type serviceImpl struct {
	gptHandler                *gpt.Handler
	indexes                   core.IndexStore
	userz                     core.UserService
	models                    core.ModelStore
	createEmbeddingsLimitChan chan struct{}
	tiktokenHandler           *tiktoken.Handler
}

func (s *serviceImpl) createEmbeddingsWithLimit(ctx context.Context, req gogpt.EmbeddingRequest, userID uint64, m *core.Model) (gogpt.EmbeddingResponse, error) {
	s.createEmbeddingsLimitChan <- struct{}{}
	defer func() {
		<-s.createEmbeddingsLimitChan
	}()

	resp, err := s.gptHandler.CreateEmbeddings(ctx, req)
	if err == nil {
		if err := s.userz.ConsumeCreditsByModel(ctx, userID, *m, int64(resp.Usage.PromptTokens), int64(resp.Usage.CompletionTokens)); err != nil {
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

	provider, providerModel := core.ModelProviderOpenAI, gogpt.AdaEmbeddingV2.String()
	m, err := s.models.GetModel(ctx, fmt.Sprintf("%s:%s", provider, providerModel))
	if err != nil {
		return nil, fmt.Errorf("models.GetModel: %w", err)
	}

	resp, err := s.createEmbeddingsWithLimit(ctx, gogpt.EmbeddingRequest{
		Input: []string{query},
		Model: gogpt.AdaEmbeddingV2,
	}, userID, m)

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

func (s *serviceImpl) CreateIndexes(ctx context.Context, userID uint64, appID string, items []*core.Index) error {
	input := make([]string, 0, len(items))
	var totalToken int

	provider, providerModel := core.ModelProviderOpenAI, gogpt.AdaEmbeddingV2.String()
	m, err := s.models.GetModel(ctx, fmt.Sprintf("%s:%s", provider, providerModel))
	if err != nil {
		return fmt.Errorf("models.GetModel: %w", err)
	}

	for _, item := range items {
		tokens, err := s.tiktokenHandler.CalToken(provider, providerModel, item.Data)
		if err != nil {
			return err
		}
		totalToken += tokens
		item.DataToken = int64(tokens)
		input = append(input, item.Data)
	}

	if totalToken > m.MaxToken {
		return core.ErrTokenExceedLimit
	}

	resp, err := s.createEmbeddingsWithLimit(ctx, gogpt.EmbeddingRequest{
		Input: input,
		Model: gogpt.AdaEmbeddingV2,
	}, userID, m)

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

	if err := s.indexes.Upsert(ctx, appID, items); err != nil {
		return fmt.Errorf("indexes.CreateIndexes: %w", err)
	}

	return nil
}
