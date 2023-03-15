package handler

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/handler/app"
	"github.com/pandodao/botastic/handler/auth"
	"github.com/pandodao/botastic/handler/bot"
	"github.com/pandodao/botastic/handler/conv"
	indexHandler "github.com/pandodao/botastic/handler/index"
	"github.com/pandodao/botastic/handler/render"
	"github.com/pandodao/botastic/handler/user"
	"github.com/pandodao/botastic/internal/chanhub"
	"github.com/pandodao/botastic/session"
)

func New(cfg Config, s *session.Session,
	apps core.AppStore,
	indexs core.IndexStore,
	users core.UserStore,
	convs core.ConversationStore,
	appz core.AppService,
	botz core.BotService,
	indexService core.IndexService,
	userz core.UserService,
	convz core.ConversationService,
	hub *chanhub.Hub,
) Server {
	return Server{
		cfg:          cfg,
		apps:         apps,
		indexes:      indexs,
		users:        users,
		appz:         appz,
		indexService: indexService,
		botz:         botz,
		convz:        convz,
		userz:        userz,
		session:      s,
		convs:        convs,
		hub:          hub,
	}
}

type (
	Config struct {
	}

	Server struct {
		cfg Config

		session *session.Session
		apps    core.AppStore
		indexes core.IndexStore
		users   core.UserStore
		convs   core.ConversationStore

		botz         core.BotService
		appz         core.AppService
		indexService core.IndexService
		convz        core.ConversationService
		userz        core.UserService

		hub *chanhub.Hub
	}
)

func (s Server) HandleRest() http.Handler {
	r := chi.NewRouter()
	r.Use(render.WrapResponse(true))
	r.Use(auth.HandleAppAuthentication(s.session, s.appz))
	r.Use(auth.HandleAuthentication(s.session, s.users))

	r.Route("/indices", func(r chi.Router) {
		r.Post("/", indexHandler.CreateIndex(s.indexService))
		r.Get("/search", indexHandler.Search(s.apps, s.indexService))
		r.Delete("/{objectID}", indexHandler.Delete(s.apps, s.indexes))
	})

	r.Route("/conversations", func(r chi.Router) {
		r.Post("/", conv.CreateConversation(s.botz, s.convz))
		r.Post("/oneway", conv.CreateOnewayConversation(s.convz, s.convs, s.hub))
		r.Get("/{conversationID}", conv.GetConversation(s.botz, s.convz))
		r.Post("/{conversationID}", conv.PostToConversation(s.botz, s.convz))
		r.Delete("/{conversationID}", conv.DeleteConversation(s.botz, s.convz))
		r.Put("/{conversationID}", conv.UpdateConversation())
		r.Get("/{conversationID}/{turnID}", conv.GetConversationTurn(s.botz, s.convs, s.hub))
	})

	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", auth.Login(s.session, s.userz))
	})

	r.Route("/bots", func(r chi.Router) {
		r.Get("/public", bot.GetPublicBots(s.botz))
		r.With(s.LoginRequired()).Get("/{botID}", bot.GetBot(s.botz))
		r.With(s.LoginRequired()).Post("/", bot.CreateBot(s.botz))
		r.With(s.LoginRequired()).Put("/{botID}", bot.UpdateBot(s.botz))
		r.With(s.LoginRequired()).Get("/", bot.GetMyBots(s.botz))
	})

	r.With(s.LoginRequired()).Route("/users", func(r chi.Router) {
		r.Get("/{userID}", user.GetUser(s.users))
	})

	r.With(s.LoginRequired()).Route("/me", func(r chi.Router) {
		r.Get("/", user.GetMe(s.users))
	})

	r.With(s.LoginRequired()).Route("/apps", func(r chi.Router) {
		r.Get("/{appID}", app.GetApp(s.appz))
		r.Post("/", app.CreateApp(s.appz))
		r.Get("/", app.GetMyApps(s.appz))
		r.Delete("/{appID}", app.DeleteApp(s.appz))
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		render.Error(w, http.StatusNotFound, fmt.Errorf("not found"))
	})

	return r
}

func (s Server) LoginRequired() func(http.Handler) http.Handler {
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
