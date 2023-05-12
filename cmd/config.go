package cmd

import (
	"fmt"

	"github.com/pandodao/botastic/config"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Display config info",
	RunE: func(cmd *cobra.Command, args []string) error {
		showExample, _ := cmd.Flags().GetBool("example")
		showDefault, _ := cmd.Flags().GetBool("default")
		var cfg *config.Config
		switch {
		case showExample:
			cfg = config.ExampleConfig()
		case showDefault:
			cfg = config.DefaultConfig()
		default:
			var err error
			cfg, err = config.Init(cfgFile)
			if err != nil {
				return err
			}
		}

		fmt.Print(cfg)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.Flags().BoolP("example", "e", false, "Display example config")
	configCmd.Flags().BoolP("default", "d", false, "Display default config")
}
