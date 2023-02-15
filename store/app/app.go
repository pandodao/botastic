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
	if v, ok := interface{}(dao.App).(core.AppStore); ok {
		s.AppStore = v
	}
	return s
}

type storeImpl struct {
	core.AppStore
}
