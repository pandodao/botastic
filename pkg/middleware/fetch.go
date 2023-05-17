package middleware

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pandodao/botastic/api"
	"github.com/pandodao/botastic/models"
)

type Fetch struct{}

func NewFetch() *Fetch {
	return &Fetch{}
}

func (m *Fetch) Desc() *api.MiddlewareDesc {
	return &api.MiddlewareDesc{
		Name: "fetch",
		Desc: "fetch middleware will send a GET request to the specified URL and return the response body as string",
		Options: []*api.MiddlewareDescOption{
			{
				Name:     "url",
				Desc:     "URL to fetch",
				Required: true,
				ParseValueFunc: func(v string) (any, error) {
					_, err := url.Parse(v)
					return v, err
				},
			},
		},
	}
}

func (m *Fetch) Parse(opts map[string]string) (map[string]*api.MiddlewareDescOption, error) {
	result := map[string]*api.MiddlewareDescOption{}
	desc := m.Desc()
	for _, opt := range m.Desc().Options {
		opts[opt.Name] = strings.TrimSpace(opts[opt.Name])
		if opt.Required && opts[opt.Name] == "" {
			return nil, fmt.Errorf("missing required option: %s, middleware: %s", opt.Name, desc.Name)
		}
		if opts[opt.Name] == "" {
			opts[opt.Name] = opt.DefaultValue
		}

		if opt.ParseValueFunc != nil {
			v, err := opt.ParseValueFunc(opts[opt.Name])
			if err != nil {
				return nil, fmt.Errorf("failed to parse option: %s, middleware: %s, err: %w", opt.Name, desc.Name, err)
			}
			opt.Value = v
		}
		result[opt.Name] = opt
	}

	return result, nil
}

func (m *Fetch) Process(ctx context.Context, opts map[string]*api.MiddlewareDescOption, turn *models.Turn) (string, error) {
	u := opts["url"].Value.(string)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("failed to fetch url: %s, status code: %d, body: %s", u, resp.StatusCode, string(body))
	}

	return string(body), nil
}
