package index

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/fox-one/pkg/httputil/param"
	"github.com/go-chi/chi"
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/handler/render"
)

type CreateIndexPayload struct {
	ObjectID   string `json:"object_id"`
	IndexName  string `json:"index_name"`
	Category   string `json:"category"`
	Data       string `json:"data"`
	Properties string `json:"properties"`
}

func CreateIndex(indexes core.IndexService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		indexName := chi.URLParam(r, "indexName")
		if indexName == "" {
			render.Error(w, http.StatusBadRequest, nil)
		}

		body := &CreateIndexPayload{}
		if err := param.Binding(r, body); err != nil {
			render.Error(w, http.StatusBadRequest, err)
			return
		}
		if body.Category != "plain-text" {
			render.Error(w, http.StatusBadRequest, fmt.Errorf("category not supported"))
			return
		}

		err := indexes.CreateIndex(r.Context(), body.Data, body.ObjectID, body.IndexName, body.Category, body.Properties)
		if err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		render.JSON(w, nil)
	}
}

func Search(apps core.AppStore, indexes core.IndexService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		indexName := chi.URLParam(r, "indexName")
		keywords := chi.URLParam(r, "keywords")
		if keywords == "" {
			render.Error(w, http.StatusBadRequest, fmt.Errorf("keywords is required"))
			return
		}

		limit := 1
		limitStr := chi.URLParam(r, "n")
		if limitStr != "" {
			n, err := strconv.Atoi(limitStr)
			if err == nil {
				limit = n
			}
		}

		result, err := indexes.SearchIndex(ctx, indexName, keywords, limit)
		if err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		render.JSON(w, result)
	}
}
