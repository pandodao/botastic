package auth

import (
	"errors"
	"net/http"

	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/handler/render"
	"github.com/pandodao/botastic/session"
	"gorm.io/gorm"
)

func HandleAuthentication(s *session.Session, appz core.AppService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appID, appSecret := getAuthInfo(r)
			if appID == "" || appSecret == "" {
				render.Error(w, http.StatusUnauthorized, errors.New("missing app id or secret"))
				return
			}

			app, err := appz.GetAppByAppID(ctx, appID)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				render.Error(w, http.StatusInternalServerError, err)
				return
			}
			if app == nil {
				render.Error(w, http.StatusUnauthorized, errors.New("app not found"))
				return
			}

			if app.AppSecret != appSecret {
				render.Error(w, http.StatusUnauthorized, errors.New("invalid app id or secret"))
				return
			}

			next.ServeHTTP(w, r.WithContext(
				session.WithApp(r.Context(), app),
			))
		}

		return http.HandlerFunc(fn)
	}
}

func getAuthInfo(r *http.Request) (string, string) {
	appID := r.Header.Get("X-BOTASTIC-APPID")
	appSecret := r.Header.Get("X-BOTASTIC-SECRET")
	return appID, appSecret
}
