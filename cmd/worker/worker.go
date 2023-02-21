package worker

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/pandodao/botastic/config"
	"github.com/pandodao/botastic/handler/hc"
	"github.com/pandodao/botastic/internal/gpt"
	botServ "github.com/pandodao/botastic/service/bot"
	convServ "github.com/pandodao/botastic/service/conv"
	"github.com/pandodao/botastic/store"
	"github.com/pandodao/botastic/store/app"
	"github.com/pandodao/botastic/store/conv"
	"github.com/pandodao/botastic/worker"
	"github.com/pandodao/botastic/worker/rotater"

	"github.com/fox-one/pkg/logger"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
	"golang.org/x/sync/errgroup"

	"github.com/spf13/cobra"
)

func NewCmdWorker() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "worker [health check port]",
		Short: "run workers",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			ctx := cmd.Context()
			cfg := config.C()

			h := store.MustInit(store.Config{
				Driver: cfg.DB.Driver,
				DSN:    cfg.DB.DSN,
			})

			gptHandler := gpt.New(gpt.Config{
				Keys:    cfg.OpenAPI.Keys,
				Timeout: cfg.OpenAPI.Timeout,
			})

			apps := app.New(h.DB)
			convs := conv.New(h.DB)

			botz := botServ.New(botServ.Config{}, apps)
			convz := convServ.New(convServ.Config{}, apps, convs, botz)

			workers := []worker.Worker{
				// rotater
				rotater.New(rotater.Config{}, gptHandler, convs, convz, botz),
			}

			// run them all
			g, ctx := errgroup.WithContext(ctx)
			for idx := range workers {
				w := workers[idx]
				g.Go(func() error {
					fmt.Println(" workers:", w, ctx)
					return w.Run(ctx)
				})
			}

			// start the health check server
			g.Go(func() error {
				mux := chi.NewMux()
				mux.Use(middleware.Recoverer)
				mux.Use(middleware.StripSlashes)
				mux.Use(cors.AllowAll().Handler)
				mux.Use(logger.WithRequestID)
				mux.Use(middleware.Logger)

				{
					// hc for workers
					mux.Mount("/hc", hc.Handle(cmd.Version))
				}

				// launch server
				port := 8081
				if len(args) > 0 {
					port, err = strconv.Atoi(args[0])
					if err != nil {
						port = 8081
					}
				}

				addr := fmt.Sprintf(":%d", port)
				return http.ListenAndServe(addr, mux)
			})

			if err := g.Wait(); err != nil {
				cmd.PrintErrln("run worker", err)
			}

			return nil
		},
	}

	return cmd
}
