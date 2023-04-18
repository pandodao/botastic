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
	"github.com/pandodao/botastic/core"
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
			ctx := cmd.Context()
			s := session.From(ctx)
			cfg := config.C()

			h := store.MustInit(store.Config{
				Driver: cfg.DB.Driver,
				DSN:    cfg.DB.DSN,
			})

			var (
				users             core.UserStore
				mixpayClient      *mixpay.Client
				lemonClient       *lemon.Client
				twitterClient     *twitter.Client
				userz             core.UserService
				mixinClientID     string
				orders            core.OrderStore
				orderz            core.OrderService
				ordersyncerWorker *ordersyncer.Worker
				appPerUserLimit   int
				botPerUserLimit   int
			)

			if cfg.Mode == config.ModeSaaS {
				s.WithJWTSecret([]byte(config.C().SaaS.Auth.JwtSecret))
				client, err := s.GetClient()
				if err != nil {
					return err
				}
				appPerUserLimit = cfg.SaaS.AppPerUserLimit
				botPerUserLimit = cfg.SaaS.BotPerUserLimit
				mixinClientID = client.ClientID
				users = user.New(h)
				mixpayClient = mixpay.New()
				lemonClient = lemon.New(cfg.SaaS.Lemon.Key)
				twitterClient = twitter.New(cfg.SaaS.Twitter.ApiKey, cfg.SaaS.Twitter.ApiSecret)
				userz := userServ.New(userServ.Config{
					InitUserCredits: cfg.SaaS.InitUserCredits,
				}, client, twitterClient, users)

				orders = order.New(h)
				orderz := orderServ.New(orderServ.Config{
					PayeeId:           cfg.SaaS.Mixpay.PayeeId,
					QuoteAssetId:      cfg.SaaS.Mixpay.QuoteAssetId,
					SettlementAssetId: cfg.SaaS.Mixpay.SettlementAssetId,
					CallbackUrl:       cfg.SaaS.Mixpay.CallbackUrl,
					ReturnTo:          cfg.SaaS.Mixpay.ReturnTo,
					FailedReturnTo:    cfg.SaaS.Mixpay.FailedReturnTo,
				}, orders, userz, mixpayClient, lemonClient)
				ordersyncerWorker = ordersyncer.New(ordersyncer.Config{
					Interval:       cfg.SaaS.OrderSyncer.Interval,
					CheckInterval:  cfg.SaaS.OrderSyncer.CheckInterval,
					CancelInterval: cfg.SaaS.OrderSyncer.CancelInterval,
				}, orders, orderz)
			} else {
				users = user.NewLocalModeStore(&core.User{})
			}

			gptHandler := gpt.New(gpt.Config{
				Keys:    cfg.OpenAI.Keys,
				Timeout: cfg.OpenAI.Timeout,
			})

			apps := app.New(h)
			convs := conv.New(h)
			bots := bot.New(h)

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

			indexService := indexServ.NewService(ctx, gptHandler, indexes, userz, models, tiktokenHandler)
			appz := appServ.New(appServ.Config{
				SecretKey: cfg.Sys.SecretKey,
			}, apps, indexService)

			middlewarez := middlewareServ.New(middlewareServ.Config{}, apps, indexService)
			botz := botServ.New(botServ.Config{}, apps, bots, models, middlewarez)
			convz := convServ.New(convServ.Config{}, convs, botz, apps)
			hub := chanhub.New()
			// var userz core.UserService

			// httpd's workers
			workers := []worker.Worker{
				// rotater
				rotater.New(rotater.Config{}, gptHandler, convs, apps, models, convz, botz, middlewarez, userz, hub, tiktokenHandler),
			}
			if ordersyncerWorker != nil {
				workers = append(workers, ordersyncerWorker)
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
					mux.Get("/hc", hc.Handle(cmd.Version).ServeHTTP)
				}

				{
					svr := handler.New(handler.Config{
						Mode:               cfg.Mode,
						ClientID:           mixinClientID,
						TrustDomains:       cfg.SaaS.Auth.TrustDomains,
						Lemon:              cfg.SaaS.Lemon,
						Variants:           cfg.SaaS.TopupVariants,
						TwitterCallbackUrl: cfg.SaaS.Twitter.CallbackUrl,
						AppPerUserLimit:    appPerUserLimit,
						BotPerUserLimit:    botPerUserLimit,
					}, s, twitterClient, apps, indexes, users, convs, models, appz, botz, indexService, userz, convz, orderz, hub)

					// api v1
					restHandler := svr.HandleRest()
					mux.Mount("/api", restHandler)
				}

				fmt.Println("All routes:")
				chi.Walk(mux, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
					fmt.Printf("[%s]: %s \n", method, route)
					return nil
				})

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
