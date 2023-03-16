package index

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/fox-one/pkg/httputil/param"
	"github.com/go-chi/chi"
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/handler/render"
	"github.com/pandodao/botastic/session"
)

type CreateIndexPayload struct {
	ObjectID   string `json:"object_id"`
	Category   string `json:"category"`
	Data       string `json:"data"`
	Properties string `json:"properties"`
}

func ResetIndexes(indexes core.IndexService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		app := session.AppFrom(ctx)
		err := indexes.ResetIndexes(ctx, app.AppID)
		if err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		render.JSON(w, map[string]interface{}{})
	}
}

func CreateIndex(indexes core.IndexService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var data struct {
			Items []*CreateIndexPayload `json:"items"`
		}
		if err := param.Binding(r, &data); err != nil {
			render.Error(w, http.StatusBadRequest, err)
			return
		}
		app := session.AppFrom(r.Context())
		is := make([]*core.Index, len(data.Items))
		for i, item := range data.Items {
			if item.Category != "plain-text" {
				render.Error(w, http.StatusBadRequest, fmt.Errorf("category not supported"))
				return
			}
			is[i] = &core.Index{
				AppID:      app.AppID,
				Data:       item.Data,
				ObjectID:   item.ObjectID,
				Category:   item.Category,
				Properties: item.Properties,
			}
		}

		if len(is) == 0 {
			render.Error(w, http.StatusBadRequest, fmt.Errorf("empty data"))
			return
		}

		err := indexes.CreateIndexes(r.Context(), app.UserID, is)
		if err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		render.JSON(w, map[string]interface{}{})
	}
}

func Delete(apps core.AppStore, indexes core.IndexStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		app := session.AppFrom(r.Context())

		objectID := chi.URLParam(r, "objectID")
		if objectID == "" {
			render.Error(w, http.StatusBadRequest, nil)
			return
		}

		if err := indexes.DeleteByPks(ctx, []*core.Index{
			{
				AppID:    app.AppID,
				ObjectID: objectID,
			},
		}); err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		render.JSON(w, map[string]interface{}{})
	}
}

func Search(apps core.AppStore, indexes core.IndexService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		app := session.AppFrom(ctx)
		query := r.URL.Query().Get("query")
		if query == "" {
			query = r.URL.Query().Get("keywords")
		}
		if query == "" {
			render.Error(w, http.StatusBadRequest, fmt.Errorf("query is required"))
			return
		}

		limit := 1
		limitStr := r.URL.Query().Get("n")
		if limitStr != "" {
			n, err := strconv.Atoi(limitStr)
			if err == nil {
				limit = n
			}
		}

		result, err := indexes.SearchIndex(ctx, app.UserID, query, limit)
		if err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		render.JSON(w, result)
	}
}
