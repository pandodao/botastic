package model

import (
	_ "embed"

	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/store"
	"github.com/pandodao/botastic/store/model/dao"
	"gorm.io/gen"
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
