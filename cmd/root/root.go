package root

import (
	"fmt"
	"net/url"
	"os"

	"botastic/cmd/echo"
	"botastic/cmd/httpd"
	"botastic/cmd/migrate"
	"botastic/cmd/worker"
	"botastic/cmdutil"
	"botastic/config"
	"botastic/session"

	"github.com/fox-one/mixin-sdk-go"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewCmdRoot(version string) *cobra.Command {
	var opt struct {
		host         string
		KeystoreFile string
		accessToken  string
		Pin          string
	}

	cmd := &cobra.Command{
		Use:           "go-boilerplate <command> <subcommand> [flags]",
		Short:         "gb",
		Long:          `A boilerplate for go programe.`,
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

				if opt.Pin != "" {
					pin = opt.Pin
				}

				s.WithKeystore(keystore)

				if pin != "" {
					s.WithPin(pin)
				}
			}

			if opt.accessToken != "" {
				s.WithAccessToken(opt.accessToken)
			}

			if cmd.Flags().Changed("host") {
				u, err := url.Parse(opt.host)
				if err != nil {
					return err
				}

				if u.Scheme == "" {
					u.Scheme = "https"
				}

				mixin.UseApiHost(u.String())
			}

			return nil
		},
	}

	// load config
	config.InitConfig()

	cmd.PersistentFlags().StringVar(&opt.host, "host", mixin.DefaultApiHost, "custom api host")
	cmd.PersistentFlags().StringVarP(&opt.KeystoreFile, "file", "f", "", "keystore file path")
	cmd.PersistentFlags().StringVar(&opt.accessToken, "token", "", "custom access token")
	cmd.PersistentFlags().StringVar(&opt.Pin, "pin", "", "raw pin")

	cmd.AddCommand(httpd.NewCmdHttpd())
	cmd.AddCommand(echo.NewCmdEcho())
	cmd.AddCommand(migrate.NewCmdMigrate())
	cmd.AddCommand(worker.NewCmdWorker())

	return cmd
}
