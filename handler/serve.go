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

func New(cfg Config, s *session.Session, apps core.AppStore) Server {
	return Server{
		cfg:     cfg,
		apps:    apps,
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
	}
)

func (s Server) HandleRest() http.Handler {
	r := chi.NewRouter()
	r.Use(render.WrapResponse(true))
	r.Use(auth.HandleAuthentication(s.session, s.apps))

	r.Route("/indexes", func(r chi.Router) {
		r.Post("/{indexName}", indexHandler.CreateIndex())
		r.Get("/{indexName}/search", indexHandler.Search())
	})

	r.Route("/bots", func(r chi.Router) {
		r.Get("/", bot.GetBots())
		r.Get("/{botID}", bot.GetBot())
	})

	r.Route("/conversations", func(r chi.Router) {
		r.Post("/", conv.CreateConversation())
		r.Get("/{conversationID}", conv.GetConversation())
		r.Post("/{conversationID}", conv.PostToConversation())
		r.Delete("/{conversationID}", conv.DeleteConversation())
		r.Put("/{conversationID}", conv.UpdateConversation())
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		render.Error(w, http.StatusNotFound, fmt.Errorf("not found"))
	})

	return r
}
