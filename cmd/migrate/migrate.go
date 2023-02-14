package migrate

import (
	"botastic/config"
	"botastic/store/db"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
)

func NewCmdMigrate() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "migrate",
		Aliases: []string{"setdb"},
		Short:   "migrate database tables",
		Run: func(cmd *cobra.Command, args []string) {

			conn, err := sqlx.Connect(config.C().DB.Driver, config.C().DB.Datasource)
			if err != nil {
				log.Fatalln("connect to database failed", err)
			}
			defer conn.Close()

			if err := db.Migrate(conn.DB); err != nil {
				cmd.PrintErrln("migrate tables failed: ", err)
				return
			}
			cmd.Println("migrate done")
		},
	}

	return cmd
}
