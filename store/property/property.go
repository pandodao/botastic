package property

import (
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/store"
	"github.com/pandodao/botastic/store/property/dao"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func init() {
	store.RegistGenerateDAO(
		gen.Config{
			OutPath: "store/property/dao",
		},
		func(g *gen.Generator) {
			g.ApplyInterface(func(core.PropertyStore) {}, core.Property{})
		},
	)

	store.RegistGenerateModel(
		gen.Config{
			ModelPkgPath:     "./core",
			FieldWithTypeTag: true,
		},
		func(g *gen.Generator) {
			g.GenerateModel("properties")
		},
	)
}

func New(db *gorm.DB) core.PropertyStore {
	dao.SetDefault(db)
	s := &storeImpl{}
	v, ok := interface{}(dao.Property).(core.PropertyStore)
	if !ok {
		panic("dao.Property is not core.PropertyStore")
	}
	s.PropertyStore = v
	return s
}

type storeImpl struct {
	core.PropertyStore
}
