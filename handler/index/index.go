package index

import (
	"net/http"

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

		err := indexes.CreateIndex(r.Context(), body.Data, body.ObjectID, body.IndexName, body.Category, body.Properties)
		if err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		render.JSON(w, nil)
	}
}

func Search(indexes core.IndexService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		indexName := chi.URLParam(r, "indexName")
		if indexName == "" {
			render.Error(w, http.StatusBadRequest, nil)
			return
		}
		searchData := chi.URLParam(r, "data")
		if searchData == "" {
			render.Error(w, http.StatusBadRequest, nil)
			return
		}

		result, err := indexes.SearchIndex(r.Context(), indexName, searchData, 1)
		if err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		render.JSON(w, result)
	}
}
