package bot

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/pandodao/botastic/handler/render"
)

func GetBot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		botIDStr := chi.URLParam(r, "botID")
		botID, _ := strconv.ParseUint(botIDStr, 10, 64)

		if botID <= 0 {
			render.Error(w, http.StatusBadRequest, nil)
			return
		}

		// @TODO get bot by botID

		render.JSON(w, nil)
	}
}

func GetBots() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// @TODO get bots
		render.JSON(w, []interface{}{})
	}
}
