package auth

import (
	"net/http"

	"github.com/pandodao/botastic/session"
)

func HandleAuthentication(s *session.Session) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// ctx := r.Context()

			appID, appSecret := getAuthInfo(r)
			if appID == "" || appSecret == "" {
				next.ServeHTTP(w, r)
				return
			}

			// @TODO check appID and appSecret
			// app, err := foo(ctx, appID, appSecret)
			// if err != nil {
			// 	fmt.Println("auth app", err)
			// 	next.ServeHTTP(w, r)
			// 	return
			// }
			// next.ServeHTTP(w, r.WithContext(
			// 	session.WithUser(ctx, app),
			// ))
		}

		return http.HandlerFunc(fn)
	}
}

func getAuthInfo(r *http.Request) (string, string) {
	appID := r.Header.Get("X-BOTASTIC-APPID")
	appSecret := r.Header.Get("X-BOTASTIC-SECRET")
	return appID, appSecret
}
