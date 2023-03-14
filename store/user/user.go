package user

import (
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/store"
	"github.com/pandodao/botastic/store/user/dao"
	"gorm.io/gen"
)

func init() {
	store.RegistGenerate(
		gen.Config{
			OutPath: "store/user/dao",
		},
		func(g *gen.Generator) {
			g.ApplyInterface(func(core.UserStore) {}, core.User{})
		},
	)
}

func New(h *store.Handler) core.UserStore {
	var q *dao.Query
	if !dao.Q.Available() {
		dao.SetDefault(h.DB)
		q = dao.Q
	} else {
		q = dao.Use(h.DB)
	}

	v, ok := interface{}(q.User).(core.UserStore)
	if !ok {
		panic("dao.User is not core.UserStore")
	}

	return &storeImpl{
		UserStore: v,
	}
}

type storeImpl struct {
	core.UserStore
}
