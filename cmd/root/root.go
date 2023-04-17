package root

import (
	"fmt"
	"os"

	"github.com/pandodao/botastic/cmd/app"
	"github.com/pandodao/botastic/cmd/gen"
	"github.com/pandodao/botastic/cmd/httpd"
	"github.com/pandodao/botastic/cmd/migrate"
	"github.com/pandodao/botastic/cmd/model"
	"github.com/pandodao/botastic/cmdutil"
	"github.com/pandodao/botastic/config"
	"github.com/pandodao/botastic/session"
	"github.com/spf13/cobra"
)

var opt struct {
	KeystoreFile string
}

func NewCmdRoot(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "botastic <command> <subcommand> [flags]",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			s := session.From(cmd.Context())

			if opt.KeystoreFile != "" {
				data, err := os.ReadFile(opt.KeystoreFile)
				if err != nil {
					return fmt.Errorf("read keystore file %s failed: %w", opt.KeystoreFile, err)
				}
				keystore, pin, err := cmdutil.DecodeKeystore(data)
				if err != nil {
					return fmt.Errorf("decode keystore failed: %w", err)
				}
				s.WithKeystore(keystore)
				if pin != "" {
					s.WithPin(pin)
				}
			}

			return nil
		},
	}

	// load config
	config.C()

	cmd.PersistentFlags().StringVarP(&opt.KeystoreFile, "file", "f", "", "keystore file path")

	cmd.AddCommand(httpd.NewCmdHttpd())
	cmd.AddCommand(migrate.NewCmdMigrate())
	cmd.AddCommand(gen.NewCmdGen())
	cmd.AddCommand(app.NewCmdApp())
	cmd.AddCommand(model.NewCmdModel())

	return cmd
}
