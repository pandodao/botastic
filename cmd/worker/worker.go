package worker

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"go-boilerplate/config"
	"go-boilerplate/handler/hc"
	asz "go-boilerplate/service/asset"
	snapsz "go-boilerplate/service/snapshot"
	"go-boilerplate/session"
	"go-boilerplate/store/asset"
	"go-boilerplate/store/property"
	"go-boilerplate/store/snapshot"
	"go-boilerplate/worker"
	"go-boilerplate/worker/messenger"
	"go-boilerplate/worker/syncer"
	"go-boilerplate/worker/timer"

	"github.com/fox-one/pkg/logger"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/jmoiron/sqlx"
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
			s := session.From(ctx)

			conn, err := sqlx.Connect(config.C().DB.Driver, config.C().DB.Datasource)
			if err != nil {
				log.Fatalln("connect to database failed", err)
			}
			conn.SetMaxIdleConns(20)
			conn.SetConnMaxLifetime(time.Hour)
			defer conn.Close()

			propertys := property.New(conn)

			keystore, err := s.GetKeystore()
			if err != nil {
				return err
			}

			client, err := s.GetClient()
			if err != nil {
				return err
			}

			assets := asset.New(conn)
			snapshots := snapshot.New(conn)
			assetz := asz.New(client, assets)
			snapshotz := snapsz.New(client)

			workers := []worker.Worker{
				// syncer
				syncer.New(syncer.Config{
					ClientID: keystore.ClientID,
				}, propertys, snapshots, snapshotz),

				// messenger
				messenger.New(client),

				// timer
				timer.New(timer.Config{}, propertys, assetz),
			}

			// run them all
			g, ctx := errgroup.WithContext(ctx)
			for idx := range workers {
				w := workers[idx]
				g.Go(func() error {
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
