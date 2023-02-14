package root

import (
	"github.com/pandodao/botastic/cmd/httpd"
	"github.com/pandodao/botastic/cmd/migrate"
	"github.com/pandodao/botastic/config"
	"github.com/spf13/cobra"
)

func NewCmdRoot(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "botastic <command> <subcommand> [flags]",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       version,
	}

	// load config
	config.MustInit("./config.yaml")

	cmd.AddCommand(httpd.NewCmdHttpd())
	cmd.AddCommand(migrate.NewCmdMigrate())

	return cmd
}
