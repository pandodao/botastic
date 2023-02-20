package migrate

import (
	"fmt"
	"strconv"

	"github.com/pandodao/botastic/config"
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/internal/milvus"
	"github.com/pandodao/botastic/store"
	"github.com/spf13/cobra"
)

func NewCmdMigrate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "migrate database tables",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cfg := config.C()
			h := store.MustInit(store.Config{
				Driver: cfg.DB.Driver,
				DSN:    cfg.DB.DSN,
			})
			cmd.SetContext(store.NewContext(cmd.Context(), h))
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "up",
		Short: "Migrate the DB to the most recent version available",
		RunE: func(cmd *cobra.Command, args []string) error {
			return store.WithContext(cmd.Context()).MigrationUp()
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "up-to VERSION",
		Short: "Migrate the DB to a specific VERSION",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("up-to requires a version argument")
			}
			version, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid version: %w", err)
			}
			return store.WithContext(cmd.Context()).MigrationUpTo(version)
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "down",
		Short: "Roll back the version by 1",
		RunE: func(cmd *cobra.Command, args []string) error {
			return store.WithContext(cmd.Context()).MigrationUp()
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "down-to VERSION",
		Short: "Roll back to a specific VERSION",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("up-to requires a version argument")
			}
			version, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid version: %w", err)
			}
			return store.WithContext(cmd.Context()).MigrationDownTo(version)
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "redo",
		Short: "Re-run the latest migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return store.WithContext(cmd.Context()).MigrationRedo()
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Dump the migration status for the current DB",
		RunE: func(cmd *cobra.Command, args []string) error {
			return store.WithContext(cmd.Context()).MigrationStatus()
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "create NAME",
		Short: "Creates new migration file with the current timestamp",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("create requires a name argument")
			}
			return store.WithContext(cmd.Context()).MigrationCreate(args[0])
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "milvus",
		Short: "Update milvus collection",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg := config.C()
			client, err := milvus.Init(ctx, cfg.Milvus.Address)
			if err != nil {
				return err
			}

			index := core.Index{}
			return client.CreateCollectionIfNotExist(ctx, index.Schema(), 2)
		},
	})

	return cmd
}
