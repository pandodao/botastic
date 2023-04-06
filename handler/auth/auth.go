package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/fox-one/pkg/httputil/param"
	"github.com/golang-jwt/jwt"
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/handler/render"
	"github.com/pandodao/botastic/session"
	"github.com/pandodao/botastic/util"
	"github.com/pandodao/passport-go/auth"
	"gorm.io/gorm"
)

type LoginPayload struct {
	Method        string `json:"method"`
	MixinToken    string `json:"mixin_token"`
	Signature     string `json:"signature"`
	SignedMessage string `json:"signed_message"`
	Lang          string `json:"lang"`
}

func Login(s *session.Session, userz core.UserService, clientID string, domains []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body := &LoginPayload{}
		if err := param.Binding(r, body); err != nil {
			render.Error(w, http.StatusBadRequest, err)
			return
		}

		authorizer := auth.New([]string{
			clientID,
		}, domains)

		switch body.Method {
		case "mixin_token":
			{
				authUser, err := authorizer.Authorize(ctx, &auth.AuthorizationParams{
					Method:     auth.AuthMethodMixinToken,
					MixinToken: body.MixinToken,
				})
				if err != nil {
					render.Error(w, http.StatusUnauthorized, err)
					return
				}
				user, token, err := s.LoginWithMixin(ctx, userz, authUser, body.Lang)
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
				authUser, err := authorizer.Authorize(ctx, &auth.AuthorizationParams{
					Method:           auth.AuthMethodMvm,
					MvmSignature:     body.Signature,
					MvmSignedMessage: body.SignedMessage,
				})
				if err != nil {
					render.Error(w, http.StatusUnauthorized, err)
					return
				}
				user, token, err := s.LoginWithMixin(ctx, userz, authUser, body.Lang)
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

			if appSecret == "" || app.SecureAppSecret != appSecret {
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

			app.SecureAppSecret = app.AppSecret
			app.AppSecret = ""
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

func LoginRequired() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if _, found := session.UserFrom(ctx); !found {
				render.Error(w, http.StatusUnauthorized, core.ErrUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func UserCreditRequired(users core.UserStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			var user *core.User
			user, found := session.UserFrom(ctx)
			app := session.AppFrom(ctx)
			if !found && app != nil {
				var err error
				user, err = users.GetUser(ctx, app.UserID)
				if err != nil {
					fmt.Printf("users.GetUser err: %v, appID: %d\n", err, app.ID)
					render.Error(w, http.StatusUnauthorized, core.ErrUnauthorized)
					return
				}
				r = r.WithContext(session.WithUser(ctx, user))
			}

			if user == nil || user.ID == 0 {
				render.Error(w, http.StatusUnauthorized, core.ErrUnauthorized)
				return
			}

			if user.Credits.LessThan(util.OneSatoshi) {
				render.Error(w, http.StatusPaymentRequired, core.ErrInsufficientCredit)
				return
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
