package middleware

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/pandodao/botastic/api"
	"github.com/pandodao/botastic/internal/vector"
	"github.com/pandodao/botastic/models"
	"github.com/pandodao/botastic/pkg/llms"
)

type BotasticSearch struct {
	vih   *vector.IndexHandler
	llmsh *llms.Handler
}

func NewBotasticSearch(vih *vector.IndexHandler, llmsh *llms.Handler) *BotasticSearch {
	return &BotasticSearch{
		vih:   vih,
		llmsh: llmsh,
	}
}

func (m *BotasticSearch) Desc() *api.MiddlewareDesc {
	return &api.MiddlewareDesc{
		Name: "botastic-search",
		Desc: "botastic-search is a middleware that searches for a given query in botastic vector store",
		Options: []*api.MiddlewareDescOption{
			{
				Name:         "limit",
				Desc:         "limit is the number of results to return",
				DefaultValue: "3",
				ParseValueFunc: func(v string) (any, error) {
					return strconv.Atoi(v)
				},
			},
			{
				Name: "embedding_model",
				Desc: "the embedding model to use",
				ParseValueFunc: func(v string) (any, error) {
					_, err := m.llmsh.GetEmbeddingModel(v)
					if err != nil {
						return nil, fmt.Errorf("embedding model %s not found", v)
					}
					return v, nil
				},
			},
			{
				Name:         "group_key",
				Desc:         "the group key to search in",
				DefaultValue: "",
				ParseValueFunc: func(v string) (any, error) {
					v = strings.TrimSpace(v)
					if v == "" {
						return nil, fmt.Errorf("group key cannot be empty")
					}
					return v, nil
				},
			},
		},
	}
}

func (m *BotasticSearch) Process(ctx context.Context, opts map[string]*api.MiddlewareDescOption, turn *models.Turn) (string, map[string]any, error) {
	limit := opts["limit"].Value.(int)
	groupKey := opts["group_key"].Value.(string)
	embeddingModel := opts["embedding_model"].Value.(string)

	indexes, err := m.vih.SearchIndexes(ctx, embeddingModel, turn.Request, groupKey, limit)
	if err != nil {
		return "", nil, err
	}

	result := ""
	for i, index := range indexes {
		result += fmt.Sprintf("%d. %s", i, index.Data)
		if i != len(indexes)-1 {
			result += "\n"
		}
	}

	return result, map[string]any{
		"indexes": indexes,
	}, nil
}
