package model

import (
	"net/http"

	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/handler/render"
)

func GetModels(models core.ModelStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ms, err := models.GetModels(r.Context())
		if err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		render.JSON(w, ms)
	}
}
