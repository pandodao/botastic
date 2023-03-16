package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/internal/ddg"
	"github.com/pandodao/botastic/session"
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

func (s *service) ProcessIntentRecognition(ctx context.Context, m *core.Middleware, input string) (*core.MiddlewareProcessResult, error) {
	prompt := `You will analyze the intent of the request.
You will output the analyze result at the beginning of your response as json format: {"intent": Here is your intent analyze result}
The possible intents should be one of following. If you have no confident about the intent, please use "unknown intent":`

	var intents []string
	val, ok := m.Options["intents"]
	if ok {
		_val, ok := val.([]interface{})
		if ok {
			for _, v := range _val {
				str, ok := v.(string)
				if !ok {
					continue
				}
				intents = append(intents, str)
			}
		}
	} else {
		return nil, nil
	}

	ret := &core.MiddlewareProcessResult{
		Name:   m.Name,
		Code:   core.MiddlewareProcessCodeOK,
		Result: fmt.Sprintf("%s\n\n[intents-begin]\n%s\n[intents-end]\n", prompt, strings.Join(intents, "\n")),
	}
	return ret, nil
}

func (s *service) ProcessBotasticSearch(ctx context.Context, m *core.Middleware, input string) (*core.MiddlewareProcessResult, error) {
	limit := 3
	val, ok := m.Options["limit"]
	if ok {
		limit = int(val.(float64))
	}

	app := session.AppFrom(ctx)
	searchResult, err := s.indexz.SearchIndex(ctx, app.UserID, input, limit)
	if err != nil {
		return nil, err
	}

	arr := make([]string, 0)
	for ix, r := range searchResult {
		line := strings.ReplaceAll(strings.ReplaceAll(r.Data, "\n", " "), "\r", "")
		line = strings.TrimSpace(line)
		if line != "" {
			arr = append(arr, fmt.Sprintf("%d: %s\n", ix+1, line))
		}
	}

	ret := &core.MiddlewareProcessResult{
		Name:   m.Name,
		Code:   core.MiddlewareProcessCodeOK,
		Result: fmt.Sprintf("[context-begin]\n%s\n[context-end]\n", strings.Join(arr, "\n")),
	}
	return ret, nil
}

func (s *service) ProcessDuckduckgoSearch(ctx context.Context, m *core.Middleware, input string) (*core.MiddlewareProcessResult, error) {
	limit := 3
	val, ok := m.Options["limit"]
	if ok {
		limit = int(val.(float64))
	}
	r, err := ddg.WebSearch(ctx, input, limit)
	if err != nil {
		return nil, err
	}
	result, err := r.Text()
	if err != nil {
		return nil, err
	}
	ret := &core.MiddlewareProcessResult{
		Name:   m.Name,
		Code:   core.MiddlewareProcessCodeOK,
		Result: result,
	}
	return ret, nil
}

func (s *service) Process(ctx context.Context, m *core.Middleware, input string) (*core.MiddlewareProcessResult, error) {
	switch m.Name {
	case core.MiddlewareBotasticSearch:
		return s.ProcessBotasticSearch(ctx, m, input)
	case core.MiddlewareDuckduckgoSearch:
		return s.ProcessDuckduckgoSearch(ctx, m, input)
	case core.MiddlewareIntentRecognition:
		return s.ProcessIntentRecognition(ctx, m, input)
	}
	return nil, nil
}
