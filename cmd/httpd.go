package cmd

import (
	"github.com/spf13/cobra"
)

// httpdCmd represents the httpd command
var httpdCmd = &cobra.Command{
	Use:   "httpd",
	Short: "Start the HTTP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		httpdStarter, err := provideHttpdStarter(ctx, cfgFile)
		if err != nil {
			return err
		}
		return httpdStarter.Start(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(httpdCmd)
}
