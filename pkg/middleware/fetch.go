package middleware

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

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

func (m *Fetch) Process(ctx context.Context, opts map[string]*api.MiddlewareDescOption, turn *models.Turn) (string, map[string]any, error) {
	u := opts["url"].Value.(string)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}

	if resp.StatusCode/100 != 2 {
		return "", nil, fmt.Errorf("failed to fetch url: %s, status code: %d, body: %s", u, resp.StatusCode, string(body))
	}

	return string(body), nil, nil
}
