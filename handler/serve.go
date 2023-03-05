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
	appz core.AppService,
	botz core.BotService,
	indexService core.IndexService,
	convz core.ConversationService,
	indexs core.IndexStore,
) Server {
	return Server{
		cfg:          cfg,
		apps:         apps,
		appz:         appz,
		indexService: indexService,
		indexes:      indexs,
		botz:         botz,
		convz:        convz,
		session:      s,
	}
}

type (
	Config struct {
	}

	Server struct {
		cfg Config

		session      *session.Session
		apps         core.AppStore
		botz         core.BotService
		appz         core.AppService
		indexService core.IndexService
		indexes      core.IndexStore
		convz        core.ConversationService
	}
)

func (s Server) HandleRest() http.Handler {
	r := chi.NewRouter()
	r.Use(render.WrapResponse(true))
	r.Use(auth.HandleAuthentication(s.session, s.appz))

	r.Route("/indices", func(r chi.Router) {
		r.Post("/", indexHandler.CreateIndex(s.indexService))
		r.Get("/search", indexHandler.Search(s.apps, s.indexService))
		r.Delete("/{objectID}", indexHandler.Delete(s.apps, s.indexes))
	})

	r.Route("/bots", func(r chi.Router) {
		r.Get("/", bot.GetBots())
		r.Get("/{botID}", bot.GetBot())
	})

	r.Route("/conversations", func(r chi.Router) {
		r.Post("/", conv.CreateConversation(s.botz, s.convz))
		r.Get("/{conversationID}", conv.GetConversation(s.botz, s.convz))
		r.Post("/{conversationID}", conv.PostToConversation(s.botz, s.convz))
		r.Delete("/{conversationID}", conv.DeleteConversation(s.botz, s.convz))
		r.Put("/{conversationID}", conv.UpdateConversation())
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		render.Error(w, http.StatusNotFound, fmt.Errorf("not found"))
	})

	return r
}
