package middleware

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/pandodao/botastic/core"
)

type fetch struct{}

func (m *fetch) Name() string {
	return core.MiddlewareFetch
}

type fetchOptions struct {
	URL string
}

func (m *fetch) ValidateOptions(opts map[string]any) (any, error) {
	options := &fetchOptions{}

	if val, ok := opts["url"]; ok {
		v, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("url must be a string")
		}

		_, err := url.Parse(v)
		if err != nil {
			return nil, fmt.Errorf("url is not valid")
		}

		options.URL = v
	}

	return options, nil
}

func (m *fetch) Process(ctx context.Context, opts any, input string) (string, error) {
	options := opts.(*fetchOptions)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, options.URL, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("fetch failed with status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
