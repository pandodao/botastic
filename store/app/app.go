package app

import (
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/store"
	"github.com/pandodao/botastic/store/app/dao"
	"gorm.io/gen"
)

func init() {
	store.RegistGenerate(
		gen.Config{
			OutPath: "store/app/dao",
		},
		func(g *gen.Generator) {
			g.ApplyInterface(func(core.AppStore) {}, core.App{})
		},
	)
}

func New(h *store.Handler) core.AppStore {
	var q *dao.Query
	if !dao.Q.Available() {
		dao.SetDefault(h.DB)
		q = dao.Q
	} else {
		q = dao.Use(h.DB)
	}

	v, ok := interface{}(q.App).(core.AppStore)
	if !ok {
		panic("dao.App is not core.AppStore")
	}

	return &storeImpl{
		AppStore: v,
	}
}

type storeImpl struct {
	core.AppStore
}
