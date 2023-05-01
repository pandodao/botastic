package cmd

import (
	"github.com/pandodao/botastic/config"
	"github.com/pandodao/botastic/internal/httpd"
	"github.com/pandodao/botastic/internal/llms"
	"github.com/pandodao/botastic/storage"
	"github.com/spf13/cobra"
)

// httpdCmd represents the httpd command
var httpdCmd = &cobra.Command{
	Use:   "httpd",
	Short: "Start the HTTP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Init(cfgFile)
		if err != nil {
			return err
		}
		sh, err := storage.Init(cfg.DB)
		if err != nil {
			return err
		}
		if err := sh.Migrate(); err != nil {
			return err
		}

		lh := llms.New(cfg.LLMs)

		s := httpd.New(cfg.Httpd, httpd.NewHandler(sh, lh))
		return s.Start()
	},
}

func init() {
	rootCmd.AddCommand(httpdCmd)
}
