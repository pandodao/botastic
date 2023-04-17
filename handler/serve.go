package handler

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/pandodao/botastic/config"
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/handler/app"
	"github.com/pandodao/botastic/handler/auth"
	"github.com/pandodao/botastic/handler/bot"
	"github.com/pandodao/botastic/handler/conv"
	indexHandler "github.com/pandodao/botastic/handler/index"
	"github.com/pandodao/botastic/handler/model"
	"github.com/pandodao/botastic/handler/order"
	"github.com/pandodao/botastic/handler/render"
	"github.com/pandodao/botastic/handler/user"
	"github.com/pandodao/botastic/internal/chanhub"
	"github.com/pandodao/botastic/session"
	"github.com/pandodao/twitter-login-go"
)

func New(cfg Config, s *session.Session,
	twitterClient *twitter.Client,
	apps core.AppStore,
	indexs core.IndexStore,
	users core.UserStore,
	convs core.ConversationStore,
	models core.ModelStore,
	appz core.AppService,
	botz core.BotService,
	indexz core.IndexService,
	userz core.UserService,
	convz core.ConversationService,
	orderz core.OrderService,
	hub *chanhub.Hub,
) Server {
	return Server{
		cfg:           cfg,
		apps:          apps,
		indexes:       indexs,
		users:         users,
		appz:          appz,
		models:        models,
		indexz:        indexz,
		botz:          botz,
		convz:         convz,
		userz:         userz,
		convs:         convs,
		orderz:        orderz,
		hub:           hub,
		session:       s,
		twitterClient: twitterClient,
	}
}

type (
	Config struct {
		Mode               config.SystemMode
		ClientID           string
		TrustDomains       []string
		Lemon              config.Lemonsqueezy
		Variants           []config.TopupVariant
		TwitterCallbackUrl string
		AppPerUserLimit    int
		BotPerUserLimit    int
	}

	Server struct {
		cfg Config

		session       *session.Session
		twitterClient *twitter.Client

		apps    core.AppStore
		indexes core.IndexStore
		users   core.UserStore
		convs   core.ConversationStore
		models  core.ModelStore

		botz   core.BotService
		appz   core.AppService
		indexz core.IndexService
		convz  core.ConversationService
		userz  core.UserService
		orderz core.OrderService

		hub *chanhub.Hub
	}
)

func (s Server) HandleRest() http.Handler {
	r := chi.NewRouter()
	r.Use(render.WrapResponse(true))
	r.Use(auth.HandleAppAuthentication(s.session, s.appz))
	r.Use(auth.HandleAuthentication(s.session, s.users, s.cfg.Mode))

	userCreditRequiredHandler := auth.UserCreditRequired(s.users, s.cfg.Mode)

	r.Route("/indexes", func(r chi.Router) {
		r.With(auth.HandleAppSecretRequired(), userCreditRequiredHandler).Post("/", indexHandler.CreateIndex(s.indexz))
		r.With(auth.HandleAppSecretRequired()).Post("/reset", indexHandler.ResetIndexes(s.indexz))
		r.With(userCreditRequiredHandler).Get("/search", indexHandler.Search(s.apps, s.indexz))
		r.With(auth.HandleAppSecretRequired()).Delete("/{objectID}", indexHandler.Delete(s.apps, s.indexes))
	})

	r.Route("/models", func(r chi.Router) {
		r.Get("/", model.GetModels(s.models))
		r.Post("/", model.GetModels(s.models))
	})

	r.Route("/conversations", func(r chi.Router) {
		r.Post("/", conv.CreateConversation(s.botz, s.convz))
		r.With(userCreditRequiredHandler).Post("/oneway", conv.CreateOnewayConversation(s.convz, s.convs, s.hub))
		r.Get("/{conversationID}", conv.GetConversation(s.botz, s.convz))
		r.With(userCreditRequiredHandler).Post("/{conversationID}", conv.PostToConversation(s.botz, s.convz))
		r.Delete("/{conversationID}", conv.DeleteConversation(s.botz, s.convz))
		r.Put("/{conversationID}", conv.UpdateConversation())
		r.Get("/{conversationID}/{turnID}", conv.GetConversationTurn(s.botz, s.convs, s.hub))
	})

	r.Route("/bots", func(r chi.Router) {
		r.Get("/public", bot.GetPublicBots(s.botz))
		r.With(auth.LoginRequired()).Get("/{botID}", bot.GetBot(s.botz))
		r.With(auth.LoginRequired()).Post("/", bot.CreateBot(s.botz, s.models, s.cfg.BotPerUserLimit))
		r.With(auth.LoginRequired()).Put("/{botID}", bot.UpdateBot(s.botz))
		r.With(auth.LoginRequired()).Get("/", bot.GetMyBots(s.botz))
		r.With(auth.LoginRequired()).Delete("/{botID}", bot.DeleteBot(s.botz))
	})

	r.With(auth.LoginRequired()).Route("/users", func(r chi.Router) {
		r.Get("/{userID}", user.GetUser(s.users))
	})

	r.With(auth.LoginRequired()).Route("/me", func(r chi.Router) {
		r.Get("/", user.GetMe(s.users))
	})

	r.With(auth.LoginRequired()).Route("/apps", func(r chi.Router) {
		r.Get("/{appID}", app.GetApp(s.appz))
		r.Post("/", app.CreateApp(s.appz, s.cfg.AppPerUserLimit))
		r.Get("/", app.GetMyApps(s.appz))
		r.Put("/{appID}", app.UpdateApp(s.appz))
		r.Delete("/{appID}", app.DeleteApp(s.appz))
	})

	if s.cfg.Mode == config.SystemModeCloud {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", auth.Login(s.session, s.userz, s.cfg.ClientID, s.cfg.TrustDomains))
			r.Get("/twitter/url", auth.GetTwitterURL(s.twitterClient, s.cfg.TwitterCallbackUrl))
		})

		r.With(auth.LoginRequired()).Route("/orders", func(r chi.Router) {
			r.Post("/", order.CreateOrder(s.orderz, s.cfg.Lemon, s.cfg.Variants))
			r.Get("/variants", order.GetVariants(s.cfg.Variants))
		})
		r.Route("/payments", func(r chi.Router) {
			r.Post("/mixpay", order.HandleMixpayCallback(s.orderz))
			r.Post("/lemon", order.HandleLemonCallback(s.orderz))
		})
	}

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		render.Error(w, http.StatusNotFound, fmt.Errorf("not found"))
	})

	return r
}
