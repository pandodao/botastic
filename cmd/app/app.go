package app

import (
	"github.com/pandodao/botastic/config"
	appServ "github.com/pandodao/botastic/service/app"
	"github.com/pandodao/botastic/store"
	"github.com/pandodao/botastic/store/app"
	"github.com/spf13/cobra"
)

func NewCmdApp() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "app commands",
	}

	cmd.AddCommand(NewCmdAppCreate())
	cmd.AddCommand(NewCmdAppList())
	return cmd
}

func NewCmdAppList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list apps",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			cfg := config.C()
			h := store.MustInit(store.Config{
				Driver: cfg.DB.Driver,
				DSN:    cfg.DB.DSN,
			})
			apps := app.New(h.DB)

			items, _ := apps.GetApps(ctx)
			for _, item := range items {
				err := item.Decrypt(cfg.Sys.SecretKey)
				if err != nil {
					cmd.PrintErrf("decrypt app %d:%s failed: %v'n", item.ID, item.AppID, err)
					continue
				}
				cmd.Printf("id: %d, app_id: %s, app_secret: %s", item.ID, item.AppID, item.AppSecret)
			}
		},
	}

	return cmd
}

func NewCmdAppCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create new app key/secret",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			cfg := config.C()
			h := store.MustInit(store.Config{
				Driver: cfg.DB.Driver,
				DSN:    cfg.DB.DSN,
			})
			apps := app.New(h.DB)
			appz := appServ.New(appServ.Config{
				SecretKey: cfg.Sys.SecretKey,
			}, apps)

			app, err := appz.CreateApp(ctx)
			if err != nil {
				cmd.PrintErrf("create app failed: %v'n", err)
				return
			}

			cmd.Printf("app_id: %s, app_secret: %s", app.AppID, app.AppSecret)
		},
	}

	return cmd
}
