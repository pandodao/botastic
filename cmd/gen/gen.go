package gen

import (
	"github.com/pandodao/botastic/config"
	"github.com/pandodao/botastic/store"
	_ "github.com/pandodao/botastic/store/app"
	_ "github.com/pandodao/botastic/store/property"
	"github.com/spf13/cobra"
)

func NewCmdGen() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen",
		Short: "generate database operation code",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cfg := config.C()
			h := store.MustInit(store.Config{
				Driver: cfg.DB.Driver,
				DSN:    cfg.DB.DSN,
			})
			cmd.SetContext(store.NewContext(cmd.Context(), h))
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "dao",
		Short: "Generate database operation code",
		Run: func(cmd *cobra.Command, args []string) {
			store.WithContext(cmd.Context()).GenerateDAOs()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "model",
		Short: "Generate database models code",
		Run: func(cmd *cobra.Command, args []string) {
			store.WithContext(cmd.Context()).GenerateModels()
		},
	})

	return cmd
}
