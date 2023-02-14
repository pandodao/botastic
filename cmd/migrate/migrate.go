package migrate

import (
	"github.com/spf13/cobra"
)

func NewCmdMigrate() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "migrate",
		Aliases: []string{"setdb"},
		Short:   "migrate database tables",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO
			cmd.Println("migrate done")
		},
	}

	return cmd
}
