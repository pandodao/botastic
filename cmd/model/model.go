package model

import (
	"encoding/json"
	"fmt"

	"github.com/pandodao/botastic/config"
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/store"
	"github.com/pandodao/botastic/store/model"
	"github.com/spf13/cobra"
)

func NewCmdModel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "model",
		Short: "model commands",
	}

	cmd.AddCommand(NewCmdModelCreate())
	cmd.AddCommand(NewCmdModelList())
	return cmd
}

func NewCmdModelList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list models",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			cfg := config.C()
			h := store.MustInit(store.Config{
				Driver: cfg.DB.Driver,
				DSN:    cfg.DB.DSN,
			})
			models := model.New(h)

			ms, err := models.GetModelsByFunction(ctx, "")
			if err != nil {
				cmd.PrintErr(err.Error())
				return
			}
			for _, item := range ms {
				cmd.Printf("%+v\n", item)
			}
		},
	}

	return cmd
}

func NewCmdModelCreate() *cobra.Command {
	var data string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create custom model",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg := config.C()
			h := store.MustInit(store.Config{
				Driver: cfg.DB.Driver,
				DSN:    cfg.DB.DSN,
			})
			models := model.New(h)

			m := &core.Model{}
			if err := json.Unmarshal([]byte(data), m); err != nil {
				return fmt.Errorf("invalid model data: %w", err)
			}
			if m.ProviderModel == "" {
				return fmt.Errorf("provider model is empty")
			}

			if m.CustomConfig.Request.URL == "" || m.CustomConfig.Request.Method == "" || len(m.CustomConfig.Request.Data) == 0 {
				return fmt.Errorf("request of custom config is empty")
			}

			switch m.Function {
			case core.ModelFunctionChat, core.ModelFunctionEmbedding:
			default:
				return fmt.Errorf("invalid model function: %s", m.Function)
			}

			m.Provider = core.ModelProviderCustom
			if err := models.CreateModel(ctx, m); err != nil {
				return fmt.Errorf("create model error: %w", err)
			}

			mm, err := models.GetModel(ctx, m.Name())
			if err != nil {
				return err
			}

			data, _ := json.Marshal(mm)
			cmd.Printf("%s\n", string(data))
			return nil
		},
	}

	cmd.Flags().StringVar(&data, "data", "", "model data in JSON format")
	return cmd
}
