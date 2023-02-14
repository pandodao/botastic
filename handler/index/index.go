package index

import (
	"net/http"

	"github.com/fox-one/pkg/httputil/param"
	"github.com/go-chi/chi"
	"github.com/pandodao/botastic/handler/render"
)

type CreateIndexPayload struct {
	ObjectID   string `json:"object_id"`
	IndexName  string `json:"index_name"`
	Category   string `json:"category"`
	Data       string `json:"data"`
	Properties string `json:"properties"`
}

func CreateIndex() http.HandlerFunc {
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

		// @TODO create index

		render.JSON(w, nil)
	}
}

func Search() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		indexName := chi.URLParam(r, "indexName")
		if indexName == "" {
			render.Error(w, http.StatusBadRequest, nil)
		}

		// @TODO search index
		render.JSON(w, []interface{}{})
	}
}
