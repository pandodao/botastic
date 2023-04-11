package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/session"
)

type botasticSearch struct {
	apps   core.AppStore
	indexz core.IndexService
}

type botasticSearchOptions struct {
	Limit int
	AppID string
}

func (m *botasticSearch) Name() string {
	return core.MiddlewareBotasticSearch
}

func (m *botasticSearch) ValidateOptions(opts map[string]any) (any, error) {
	options := &botasticSearchOptions{
		Limit: 3,
	}

	if val, ok := opts["limit"]; ok {
		v, ok := val.(float64)
		if !ok {
			return nil, fmt.Errorf("limit is not a number: %v", val)
		}

		if v <= 0 || float64(int(v)) != v {
			return nil, fmt.Errorf("limit is not a positive integer: %v", v)
		}
		options.Limit = int(v)
	}

	if val, ok := opts["app_id"]; ok {
		appID, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("app_id must be a string: %v", val)
		}

		options.AppID = appID
	}

	return options, nil
}

func (m *botasticSearch) Process(ctx context.Context, opts any, input string) (string, error) {
	options := opts.(*botasticSearchOptions)
	app := session.AppFrom(ctx)
	if options.AppID != "" {
		optionsApp, err := m.apps.GetAppByAppID(ctx, options.AppID)
		if err != nil {
			return "", fmt.Errorf("error when getting app by app_id: %s", optionsApp.AppID)
		}
		if optionsApp.UserID != app.UserID {
			return "", fmt.Errorf("app_id not found: %s", options.AppID)
		}
		app = optionsApp
		ctx = session.WithApp(ctx, app)
	}

	searchResult, err := m.indexz.SearchIndex(ctx, app.UserID, input, options.Limit)
	if err != nil {
		return "", err
	}

	arr := make([]string, 0)
	for ix, r := range searchResult {
		line := strings.ReplaceAll(strings.ReplaceAll(r.Data, "\n", " "), "\r", "")
		line = strings.TrimSpace(line)
		if line != "" {
			arr = append(arr, fmt.Sprintf("%d: %s\n", ix+1, line))
		}
	}

	return strings.Join(arr, "\n"), nil
}
