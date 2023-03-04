package middleware

import (
	"context"
	"strings"

	"github.com/pandodao/botastic/core"
)

func New(
	cfg Config,
	indexz core.IndexService,
) *service {
	middlewareMap := make(map[string]*core.Middleware)
	middlewareMap["botastic-search"] = &core.Middleware{
		ID:   1,
		Name: "botastic-search",
	}
	middlewareMap["duckduckgo-search"] = &core.Middleware{
		ID:   2,
		Name: "duckduckgo-search",
	}
	return &service{
		cfg:    cfg,
		indexz: indexz,

		middlewareMap: middlewareMap,
	}
}

type (
	Config struct {
	}

	service struct {
		cfg    Config
		indexz core.IndexService

		middlewareMap map[string]*core.Middleware
	}
)

func (s *service) GetMiddlewareByName(ctx context.Context, name string) (*core.Middleware, error) {
	middleware, found := s.middlewareMap[name]
	if !found {
		return nil, core.ErrMiddlewareNotFound
	}
	return middleware, nil
}

func (s *service) ProcessBotasticSearch(ctx context.Context, m *core.Middleware, input string) (*core.MiddlewareProcessResult, error) {
	searchResult, err := s.indexz.SearchIndex(ctx, input, 10)
	if err != nil {
		return nil, err
	}

	arr := make([]string, len(searchResult))
	for _, r := range searchResult {
		arr = append(arr, r.Data)
	}

	ret := &core.MiddlewareProcessResult{
		Name:   m.Name,
		Code:   200,
		Result: strings.Join(arr, "\n"),
	}
	return ret, nil
}

func (s *service) ProcessDuckduckgoSearch(ctx context.Context, m *core.Middleware, input string) (*core.MiddlewareProcessResult, error) {
	ret := &core.MiddlewareProcessResult{
		Name:   m.Name,
		Code:   200,
		Result: "duckduckgo-search doesn't work yet!",
	}
	return ret, nil
}

func (s *service) Process(ctx context.Context, m *core.Middleware, input string) (*core.MiddlewareProcessResult, error) {
	switch m.Name {
	case core.MiddlewareBotasticSearch:
		return s.ProcessBotasticSearch(ctx, m, input)
	case core.MiddlewareDuckduckgoSearch:
		return s.ProcessDuckduckgoSearch(ctx, m, input)
	}
	return nil, nil
}
