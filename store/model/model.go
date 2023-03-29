package model

import (
	"context"
	_ "embed"
	"errors"

	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/store"
	"github.com/pandodao/botastic/store/model/dao"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func init() {
	store.RegistGenerate(
		gen.Config{
			OutPath: "store/model/dao",
		},
		func(g *gen.Generator) {
			g.ApplyInterface(func(core.ModelStore) {}, core.Model{})
		},
	)
}

func New(h *store.Handler) core.ModelStore {
	var q *dao.Query
	if !dao.Q.Available() {
		dao.SetDefault(h.DB)
		q = dao.Q
	} else {
		q = dao.Use(h.DB)
	}

	v, ok := interface{}(q.Model).(core.ModelStore)
	if !ok {
		panic("dao.Model is not core.ModelStore")
	}

	return &storeImpl{
		ModelStore: v,
	}
}

type storeImpl struct {
	core.ModelStore
}

func (s *storeImpl) GetModel(ctx context.Context, name string) (*core.Model, error) {
	m, err := s.ModelStore.GetModel(ctx, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrInvalidModel
		}
		return nil, err
	}

	if m.Provider == core.ModelProviderOpenAI {
		switch m.ProviderModel {
		case "gpt-4", "gpt-4-32k", "gpt-4-0314", "gpt-4-32k-0314", "gpt-3.5-turbo", "gpt-3.5-turbo-0301":
			m.Props.IsOpenAIChatModel = true
		case "text-davinci-003":
			m.Props.IsOpenAICompletionModel = true
		case "text-embedding-ada-002":
			m.Props.IsOpenAIEmbeddingModel = true
		}
	}

	return m, nil
}
