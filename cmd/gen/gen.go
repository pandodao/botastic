package gen

import (
	"github.com/pandodao/botastic/config"
	"github.com/pandodao/botastic/store"
	_ "github.com/pandodao/botastic/store/app"
	_ "github.com/pandodao/botastic/store/conv"
	_ "github.com/pandodao/botastic/store/index"
	_ "github.com/pandodao/botastic/store/property"
	"github.com/spf13/cobra"
)

func NewCmdGen() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen",
		Short: "generate database operation code",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.C()
			h := store.MustInit(store.Config{
				Driver: cfg.DB.Driver,
				DSN:    cfg.DB.DSN,
			})
			h.Generate()
		},
	}

	return cmd
}
