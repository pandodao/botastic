package user

import (
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/store"
	"github.com/pandodao/botastic/store/user/dao"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func init() {
	cfg := gen.Config{
		OutPath: "store/user/dao",
	}
	store.RegistGenerate(
		cfg,
		func(g *gen.Generator) {
			g.ApplyInterface(func(core.UserStore) {}, core.User{})
		},
	)
}

func New(db *gorm.DB) core.UserStore {
	dao.SetDefault(db)
	s := &storeImpl{}
	v, ok := interface{}(dao.User).(core.UserStore)
	if !ok {
		panic("dao.User is not core.UserStore")
	}
	s.UserStore = v
	return s
}

type storeImpl struct {
	core.UserStore
}
