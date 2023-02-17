package index

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/fox-one/pkg/httputil/param"
	"github.com/go-chi/chi"
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/handler/render"
	"github.com/pandodao/botastic/session"
	"gorm.io/gorm"
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

		app := session.AppFrom(r.Context())
		if app == nil {
			appIDStr := chi.URLParam(r, "appid")
			if appIDStr == "" {
				render.Error(w, http.StatusBadRequest, fmt.Errorf("app id is required"))
				return
			}
			var err error
			app, err = apps.GetAppByAppID(r.Context(), appIDStr)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					render.Error(w, http.StatusBadRequest, fmt.Errorf("app not found"))
					return
				}
				render.Error(w, http.StatusInternalServerError, err)
				return
			}
			r = r.WithContext(session.WithApp(r.Context(), app))
		}

		result, err := indexes.SearchIndex(r.Context(), indexName, keywords, limit)
		if err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		render.JSON(w, result)
	}
}
