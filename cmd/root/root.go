package root

import (
	"fmt"
	"os"

	jsoniter "github.com/json-iterator/go"
	"github.com/pandodao/botastic/cmd/app"
	"github.com/pandodao/botastic/cmd/gen"
	"github.com/pandodao/botastic/cmd/httpd"
	"github.com/pandodao/botastic/cmd/migrate"
	"github.com/pandodao/botastic/cmdutil"
	"github.com/pandodao/botastic/config"
	"github.com/pandodao/botastic/session"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

			v := viper.New()
			v.SetConfigType("yaml")

			if opt.KeystoreFile != "" {
				f, err := os.Open(opt.KeystoreFile)
				if err != nil {
					return fmt.Errorf("open keystore file %s failed: %w", opt.KeystoreFile, err)
				}

				defer f.Close()
				_ = v.ReadConfig(f)
			}

			if values := v.AllSettings(); len(values) > 0 {
				b, _ := jsoniter.Marshal(values)
				keystore, pin, err := cmdutil.DecodeKeystore(b)
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

	return cmd
}
