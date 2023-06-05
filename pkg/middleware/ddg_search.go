package middleware

import (
	"context"
	"strconv"

	"github.com/pandodao/botastic/api"
	"github.com/pandodao/botastic/internal/ddg"
	"github.com/pandodao/botastic/models"
)

type DDGSearch struct{}

func NewDDGSearch() *DDGSearch {
	return &DDGSearch{}
}

func (m *DDGSearch) Desc() *api.MiddlewareDesc {
	return &api.MiddlewareDesc{
		Name: "ddg_search",
		Desc: "search by given query using duckduckgo search engine",
		Options: []*api.MiddlewareDescOption{
			{
				Name:         "limit",
				Desc:         "limit number of results",
				DefaultValue: "3",
				ParseValueFunc: func(v string) (any, error) {
					return strconv.Atoi(v)
				},
			},
		},
	}
}

func (m *DDGSearch) Process(ctx context.Context, opts map[string]*api.MiddlewareDescOption, turn *models.Turn) (string, map[string]any, error) {
	limit := opts["limit"].Value.(int)
	r, err := ddg.Search(ctx, turn.Request, limit)
	if err != nil {
		return "", nil, err
	}

	result, err := r.Text()
	return result, nil, err
}
