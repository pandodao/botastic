package user

import (
	"net/http"
	"strconv"

	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/handler/render"
	"github.com/pandodao/botastic/session"

	"github.com/go-chi/chi"
)

func GetMe(users core.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, found := session.UserFrom(ctx)
		if !found {
			render.Error(w, http.StatusUnauthorized, core.ErrUnauthorized)
			return
		}

		render.JSON(w, user)
	}
}

func GetUser(users core.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, _ := strconv.ParseUint(chi.URLParam(r, "userID"), 10, 64)
		if userID <= 0 {
			render.Error(w, http.StatusBadRequest, core.ErrInvalidUserID)
			return
		}

		user, err := users.GetUser(ctx, userID)
		if err != nil {
			render.Error(w, http.StatusNotFound, core.ErrNoRecord)
			return
		}

		render.JSON(w, user)
	}
}
