package app

import (
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/store"
	"github.com/pandodao/botastic/store/app/dao"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func init() {
	cfg := gen.Config{
		OutPath: "store/app/dao",
	}
	store.RegistGenerate(
		cfg,
		func(g *gen.Generator) {
			g.ApplyInterface(func(core.AppStore) {}, core.App{})
		},
	)
}

func New(db *gorm.DB) core.AppStore {
	dao.SetDefault(db)
	s := &storeImpl{}
	v, ok := interface{}(dao.App).(core.AppStore)
	if !ok {
		panic("dao.App is not core.AppStore")
	}
	s.AppStore = v
	return s
}

type storeImpl struct {
	core.AppStore
}
