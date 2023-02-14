package property

import (
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/store"
	"github.com/pandodao/botastic/store/property/dao"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func init() {
	cfg := gen.Config{
		OutPath: "store/property/dao",
	}
	store.RegistGenerate(
		cfg,
		func(g *gen.Generator) {
			g.ApplyInterface(func(core.PropertyStore) {}, core.Property{})
		},
	)
}

func New(db *gorm.DB) core.PropertyStore {
	dao.SetDefault(db)
	s := &storeImpl{}
	if v, ok := interface{}(dao.Property).(core.PropertyStore); ok {
		s.PropertyStore = v
	}
	return s
}

type storeImpl struct {
	core.PropertyStore
}
