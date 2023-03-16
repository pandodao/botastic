package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fox-one/passport-go/eip4361"
	"github.com/fox-one/pkg/httputil/param"
	"github.com/golang-jwt/jwt"
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/handler/render"
	"github.com/pandodao/botastic/session"
	"gorm.io/gorm"
)

type LoginPayload struct {
	Method        string `json:"method"`
	MixinToken    string `json:"mixin_token"`
	Signature     string `json:"signature"`
	SignedMessage string `json:"signed_message"`
	Lang          string `json:"lang"`
}

func Login(s *session.Session, userz core.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body := &LoginPayload{}
		if err := param.Binding(r, body); err != nil {
			render.Error(w, http.StatusBadRequest, err)
			return
		}

		switch body.Method {
		case "mixin_token":
			{
				mixinToken := body.MixinToken
				user, token, err := s.LoginWithMixin(ctx, userz, mixinToken, "", body.Lang)
				if err != nil {
					render.Error(w, http.StatusUnauthorized, err)
					return
				}
				render.JSON(w, map[string]interface{}{
					"user":         user,
					"access_token": token,
				})
				return
			}
		case "mvm":
			{
				if body.Signature == "" {
					render.Error(w, http.StatusBadRequest, core.ErrBadMvmLoginSignature)
					return
				}

				message, err := eip4361.Parse(body.SignedMessage)
				if err != nil {
					render.Error(w, http.StatusBadRequest, core.ErrBadMvmLoginMessage)
					return
				}

				if err := message.Validate(time.Now()); err != nil {
					render.Error(w, http.StatusBadRequest, core.ErrBadMvmLoginMessage)
					return
				}

				if err := eip4361.Verify(message, body.Signature); err != nil {
					render.Error(w, http.StatusUnauthorized, core.ErrBadMvmLoginSignature)
					return
				}

				pubkey := message.Address

				user, token, err := s.LoginWithMixin(ctx, userz, "", pubkey, body.Lang)
				if err != nil {
					render.Error(w, http.StatusUnauthorized, err)
					return
				}
				render.JSON(w, map[string]interface{}{
					"user":         user,
					"access_token": token,
				})
				return
			}
		default:
			render.JSON(w, nil)
			return
		}
	}
}

func HandleAuthentication(s *session.Session, users core.UserStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			accessToken := getBearerToken(r)
			if accessToken == "" {
				next.ServeHTTP(w, r)
				return
			}

			claims := &session.JwtClaims{}

			tkn, err := jwt.ParseWithClaims(accessToken, claims,
				func(token *jwt.Token) (interface{}, error) {
					return s.JwtSecret, nil
				},
			)

			if err != nil {
				fmt.Println("jwt.ParseWithClaims", err)
				next.ServeHTTP(w, r)
				return
			}
			if !tkn.Valid {
				next.ServeHTTP(w, r)
				return
			}

			user, err := users.GetUser(ctx, claims.UserID)
			if err != nil {
				fmt.Println("users.GetUser", err)
				next.ServeHTTP(w, r)
				return
			}

			next.ServeHTTP(w, r.WithContext(
				session.WithUser(ctx, user),
			))
		}

		return http.HandlerFunc(fn)
	}
}

func HandleAppSecretRequired() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			app := session.AppFrom(ctx)
			appSecret := getAuthAppSecret(r)

			if appSecret == "" || app.AppSecret != appSecret {
				render.Error(w, http.StatusUnauthorized, errors.New("invalid app_secret"))
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func HandleAppAuthentication(s *session.Session, appz core.AppService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appAuthRoutes := []string{
				"/api/indexes",
				"/api/conversations",
			}

			meet := false
			for _, route := range appAuthRoutes {
				if strings.HasPrefix(r.URL.Path, route) {
					meet = true
					break
				}
			}

			if !meet {
				next.ServeHTTP(w, r)
				return
			}

			appID := getAuthAppID(r)
			if appID == "" {
				render.Error(w, http.StatusUnauthorized, errors.New("app_id is required"))
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

			next.ServeHTTP(w, r.WithContext(
				session.WithApp(r.Context(), app),
			))
		}

		return http.HandlerFunc(fn)
	}
}

func getAuthAppID(r *http.Request) string {
	return r.Header.Get("X-BOTASTIC-APPID")
}

func getAuthAppSecret(r *http.Request) string {
	return r.Header.Get("X-BOTASTIC-SECRET")
}

func getBearerToken(r *http.Request) string {
	s := r.Header.Get("Authorization")
	return strings.TrimPrefix(s, "Bearer ")
}
