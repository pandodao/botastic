package handler

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/handler/auth"
	"github.com/pandodao/botastic/handler/bot"
	"github.com/pandodao/botastic/handler/conv"
	indexHandler "github.com/pandodao/botastic/handler/index"
	"github.com/pandodao/botastic/handler/render"
	"github.com/pandodao/botastic/session"
)

func New(cfg Config, s *session.Session,
	apps core.AppStore,
	appz core.AppService, bots core.BotService, indexes core.IndexService) Server {
	return Server{
		cfg:     cfg,
		apps:    apps,
		appz:    appz,
		indexes: indexes,
		session: s,
	}
}

type (
	Config struct {
	}

	Server struct {
		cfg Config

		session *session.Session
		apps    core.AppStore
		botz    core.BotService
		appz    core.AppService
		indexes core.IndexService
	}
)

func (s Server) HandleRest() http.Handler {
	r := chi.NewRouter()
	r.Use(render.WrapResponse(true))
	r.Use(auth.HandleAuthentication(s.session, s.appz))

	r.Route("/indexes", func(r chi.Router) {
		r.Post("/{indexName}", indexHandler.CreateIndex(s.indexes))
		r.Get("/{indexName}/search/{data}", indexHandler.Search(s.indexes))
	})

	r.Route("/bots", func(r chi.Router) {
		r.Get("/", bot.GetBots())
		r.Get("/{botID}", bot.GetBot())
	})

	r.Route("/conversations", func(r chi.Router) {
		r.Post("/", conv.CreateConversation(s.botz))
		r.Get("/{conversationID}", conv.GetConversation(s.botz))
		r.Post("/{conversationID}", conv.PostToConversation(s.botz))
		r.Delete("/{conversationID}", conv.DeleteConversation(s.botz))
		r.Put("/{conversationID}", conv.UpdateConversation())
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		render.Error(w, http.StatusNotFound, fmt.Errorf("not found"))
	})

	return r
}
