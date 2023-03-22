package order

import (
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/store"
	"github.com/pandodao/botastic/store/order/dao"
	"gorm.io/gen"
)

func init() {
	store.RegistGenerate(
		gen.Config{
			OutPath: "store/order/dao",
		},
		func(g *gen.Generator) {
			g.ApplyInterface(func(core.OrderStore) {}, core.Order{})
		},
	)
}

func New(h *store.Handler) core.OrderStore {
	var q *dao.Query
	if !dao.Q.Available() {
		dao.SetDefault(h.DB)
		q = dao.Q
	} else {
		q = dao.Use(h.DB)
	}

	v, ok := interface{}(q.Order).(core.OrderStore)
	if !ok {
		panic("dao.Order is not core.OrderStore")
	}

	return &storeImpl{
		OrderStore: v,
	}
}

type storeImpl struct {
	core.OrderStore
}
