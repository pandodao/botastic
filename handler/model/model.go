package model

import (
	"net/http"

	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/handler/render"
)

func GetModels(models core.ModelStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ms, err := models.GetModelsByFunction(r.Context(), r.URL.Query().Get("function"))
		if err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		for _, m := range ms {
			// do not expose custom config, it may contain sensitive information
			m.CustomConfig = core.CustomConfig{}
		}

		render.JSON(w, ms)
	}
}
