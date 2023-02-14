package echo

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewCmdEcho() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "echo <message>",
		Short: "echo a message back",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				fmt.Println("Please provide a message")
				return nil
			}
			fmt.Println(args[0])
			return nil
		},
	}

	return cmd
}
