package app

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/handler/render"
	"github.com/pandodao/botastic/session"

	"github.com/fox-one/pkg/httputil/param"
	"github.com/go-chi/chi"
)

type CreateOrUpdateAppPayload struct {
	Name string `json:"name"`
}

func GetApp(appz core.AppService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, found := session.UserFrom(ctx)
		if !found {
			render.Error(w, http.StatusUnauthorized, core.ErrUnauthorized)
			return
		}

		appID := chi.URLParam(r, "appID")

		// validate uuid
		if _, err := uuid.Parse(appID); err != nil {
			render.Error(w, http.StatusBadRequest, core.ErrAppNotFound)
			return
		}

		app, err := appz.GetAppByAppID(ctx, appID)
		if err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		if app.UserID != user.ID {
			render.Error(w, http.StatusNotFound, core.ErrAppNotFound)
			return
		}

		render.JSON(w, app)
	}
}

func CreateApp(appz core.AppService, appPerUserLimit int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user, found := session.UserFrom(ctx)
		if !found {
			render.Error(w, http.StatusUnauthorized, core.ErrUnauthorized)
			return
		}

		body := &CreateOrUpdateAppPayload{}
		if err := param.Binding(r, body); err != nil {
			render.Error(w, http.StatusBadRequest, err)
			return
		}

		body.Name = strings.TrimSpace(body.Name)

		appArr, _ := appz.GetAppsByUser(ctx, user.ID)
		if len(appArr) >= 10 {
			render.Error(w, http.StatusBadRequest, core.ErrAppLimitReached)
			return
		}

		if appPerUserLimit != 0 && len(appArr) >= appPerUserLimit {
			render.Error(w, http.StatusBadRequest, core.ErrAppLimitReached)
			return
		}

		if len(body.Name) > 128 || len(body.Name) == 0 {
			render.Error(w, http.StatusBadRequest, nil)
			return
		}

		app, err := appz.CreateApp(ctx, user.ID, body.Name)
		if err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		render.JSON(w, app)
	}
}

func UpdateApp(appz core.AppService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user, found := session.UserFrom(ctx)
		if !found {
			render.Error(w, http.StatusUnauthorized, core.ErrUnauthorized)
			return
		}

		body := &CreateOrUpdateAppPayload{}
		if err := param.Binding(r, body); err != nil {
			render.Error(w, http.StatusBadRequest, err)
			return
		}

		body.Name = strings.TrimSpace(body.Name)

		if len(body.Name) > 128 || len(body.Name) == 0 {
			render.Error(w, http.StatusBadRequest, nil)
			return
		}

		appID := chi.URLParam(r, "appID")

		// validate uuid
		if _, err := uuid.Parse(appID); err != nil {
			render.Error(w, http.StatusNotFound, core.ErrAppNotFound)
			return
		}

		app, err := appz.GetAppByAppID(ctx, appID)
		if err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		if app.UserID != user.ID {
			render.Error(w, http.StatusNotFound, core.ErrAppNotFound)
			return
		}

		if err := appz.UpdateApp(ctx, app.ID, body.Name); err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		app.Name = body.Name

		render.JSON(w, app)
	}
}

func GetMyApps(appz core.AppService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, found := session.UserFrom(ctx)
		if !found {
			render.Error(w, http.StatusUnauthorized, core.ErrUnauthorized)
		}

		appArr, err := appz.GetAppsByUser(ctx, user.ID)
		if err != nil {
			render.JSON(w, []core.App{})
			return
		}

		render.JSON(w, appArr)
	}
}

func DeleteApp(appz core.AppService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, found := session.UserFrom(ctx)
		if !found {
			render.Error(w, http.StatusUnauthorized, core.ErrUnauthorized)
		}

		appID := chi.URLParam(r, "appID")

		// validate uuid
		if _, err := uuid.Parse(appID); err != nil {
			render.Error(w, http.StatusBadRequest, core.ErrAppNotFound)
			return
		}

		app, err := appz.GetAppByAppID(ctx, appID)
		if err != nil {
			render.Error(w, http.StatusNotFound, core.ErrAppNotFound)
			return
		}

		if app.UserID != user.ID {
			render.Error(w, http.StatusNotFound, core.ErrAppNotFound)
			return
		}

		if err := appz.DeleteApp(ctx, app.ID); err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		render.JSON(w, app)
	}
}
