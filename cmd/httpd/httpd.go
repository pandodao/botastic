package httpd

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/fox-one/pkg/logger"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/pandodao/botastic/config"
	"github.com/pandodao/botastic/handler"
	"github.com/pandodao/botastic/handler/hc"
	"github.com/pandodao/botastic/internal/chanhub"
	"github.com/pandodao/botastic/internal/gpt"
	"github.com/pandodao/botastic/internal/tiktoken"
	appServ "github.com/pandodao/botastic/service/app"
	botServ "github.com/pandodao/botastic/service/bot"
	convServ "github.com/pandodao/botastic/service/conv"
	indexServ "github.com/pandodao/botastic/service/index"
	middlewareServ "github.com/pandodao/botastic/service/middleware"
	orderServ "github.com/pandodao/botastic/service/order"
	userServ "github.com/pandodao/botastic/service/user"
	"github.com/pandodao/botastic/session"
	"github.com/pandodao/botastic/store"
	"github.com/pandodao/botastic/store/app"
	"github.com/pandodao/botastic/store/bot"
	"github.com/pandodao/botastic/store/conv"
	"github.com/pandodao/botastic/store/index"
	"github.com/pandodao/botastic/store/model"
	"github.com/pandodao/botastic/store/order"
	"github.com/pandodao/botastic/store/user"
	"github.com/pandodao/botastic/worker"
	"github.com/pandodao/botastic/worker/ordersyncer"
	"github.com/pandodao/botastic/worker/rotater"
	"github.com/pandodao/lemon-checkout-go"
	"github.com/pandodao/mixpay-go"
	"github.com/pandodao/twitter-login-go"
	"github.com/rs/cors"
	"golang.org/x/sync/errgroup"

	"github.com/drone/signal"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewCmdHttpd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "httpd [port]",
		Short: "start the httpd daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			ctx := cmd.Context()
			s := session.From(ctx)
			s.WithJWTSecret([]byte(config.C().Auth.JwtSecret))

			cfg := config.C()

			client, err := s.GetClient()
			if err != nil {
				return err
			}

			twitterClient := twitter.New(cfg.Twitter.ApiKey, cfg.Twitter.ApiSecret)

			h := store.MustInit(store.Config{
				Driver: cfg.DB.Driver,
				DSN:    cfg.DB.DSN,
			})
			gptHandler := gpt.New(gpt.Config{
				Keys:    cfg.OpenAI.Keys,
				Timeout: cfg.OpenAI.Timeout,
			})

			mixpayClient := mixpay.New()

			lemonClient := lemon.New(cfg.Lemon.Key)

			apps := app.New(h)
			convs := conv.New(h)
			users := user.New(h)
			bots := bot.New(h)
			orders := order.New(h)

			indexes, err := index.Init(ctx, cfg.IndexStore)
			if err != nil {
				return err
			}
			if err := indexes.Init(ctx); err != nil {
				return err
			}

			models := model.New(h)
			tiktokenHandler, err := tiktoken.Init()
			if err != nil {
				return err
			}

			userz := userServ.New(userServ.Config{
				InitUserCredits: cfg.Sys.InitUserCredits,
			}, client, twitterClient, users)
			indexz := indexServ.NewService(ctx, gptHandler, indexes, userz, models, tiktokenHandler)
			appz := appServ.New(appServ.Config{
				SecretKey: cfg.Sys.SecretKey,
			}, apps, indexz)

			botz := botServ.New(botServ.Config{}, apps, bots, models)
			convz := convServ.New(convServ.Config{}, convs, botz, apps)
			orderz := orderServ.New(orderServ.Config{
				PayeeId:           cfg.Mixpay.PayeeId,
				QuoteAssetId:      cfg.Mixpay.QuoteAssetId,
				SettlementAssetId: cfg.Mixpay.SettlementAssetId,
				CallbackUrl:       cfg.Mixpay.CallbackUrl,
				ReturnTo:          cfg.Mixpay.ReturnTo,
				FailedReturnTo:    cfg.Mixpay.FailedReturnTo,
			}, orders, userz, mixpayClient, lemonClient)
			hub := chanhub.New()
			// var userz core.UserService

			middlewarez := middlewareServ.New()
			rotaterWorker := rotater.New(rotater.Config{}, gptHandler, convs, apps, models, convz, botz, middlewarez, userz, hub, tiktokenHandler)

			middlewarez.Register(
				middlewareServ.InitBotastic(rotaterWorker, bots),
				middlewareServ.InitBotasticSearch(apps, indexz),
				middlewareServ.InitDuckduckgoSearch(),
				middlewareServ.InitFetch(),
				middlewareServ.InitIntentRecognition(),
			)

			// httpd's workers
			workers := []worker.Worker{
				// rotater
				rotaterWorker,

				ordersyncer.New(ordersyncer.Config{
					Interval:       cfg.OrderSyncer.Interval,
					CheckInterval:  cfg.OrderSyncer.CheckInterval,
					CancelInterval: cfg.OrderSyncer.CancelInterval,
				}, orders, orderz),
			}

			g, ctx := errgroup.WithContext(ctx)
			for idx := range workers {
				w := workers[idx]
				g.Go(func() error {
					return w.Run(ctx)
				})
			}

			g.Go(func() error {
				mux := chi.NewMux()
				mux.Use(middleware.Recoverer)
				mux.Use(middleware.StripSlashes)
				mux.Use(cors.AllowAll().Handler)
				mux.Use(logger.WithRequestID)
				mux.Use(middleware.Logger)
				mux.Use(middleware.NewCompressor(5).Handler)

				{
					mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
						w.Write([]byte("hello world"))
					})
				}

				// hc
				{
					mux.Mount("/hc", hc.Handle(cmd.Version))
				}

				{
					svr := handler.New(handler.Config{
						ClientID:           client.ClientID,
						TrustDomains:       cfg.Auth.TrustDomains,
						Lemon:              cfg.Lemon,
						Variants:           cfg.TopupVariants,
						TwitterCallbackUrl: cfg.Twitter.CallbackUrl,
						AppPerUserLimit:    cfg.Sys.AppPerUserLimit,
						BotPerUserLimit:    cfg.Sys.BotPerUserLimit,
					}, s, twitterClient, apps, indexes, users, convs, models, appz, botz, indexz, userz, convz, orderz, hub)

					// api v1
					restHandler := svr.HandleRest()
					mux.Mount("/api", restHandler)
				}

				port := 8080
				if len(args) > 0 {
					port, err = strconv.Atoi(args[0])
					if err != nil {
						port = 8080
					}
				}

				// launch server
				if err != nil {
					panic(err)
				}
				addr := fmt.Sprintf(":%d", port)

				svr := &http.Server{
					Addr:    addr,
					Handler: mux,
				}

				done := make(chan struct{}, 1)
				ctx = signal.WithContextFunc(ctx, func() {
					logrus.Debug("shutdown server...")

					// create context with timeout
					ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
					defer cancel()

					if err := svr.Shutdown(ctx); err != nil {
						logrus.WithError(err).Error("graceful shutdown server failed")
					}

					close(done)
				})

				logrus.Infoln("serve at", addr)
				if err := svr.ListenAndServe(); err != http.ErrServerClosed {
					logrus.WithError(err).Fatal("server aborted")
				}

				<-done
				return nil
			})

			if err := g.Wait(); err != nil {
				cmd.PrintErrln("run httpd & worker", err)
			}

			return nil
		},
	}

	return cmd
}
