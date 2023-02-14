package property

import (
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/store"
	"github.com/pandodao/botastic/store/property/dao"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func init() {
	store.RegistGenerate(
		"store/property/dao",
		func(g *gen.Generator) {
			g.ApplyInterface(func(core.PropertyStore) {}, core.Property{})
		},
	)
}

func New(db *gorm.DB) core.PropertyStore {
	dao.SetDefault(db)
	return &storeImpl{
		PropertyStore: dao.Property,
	}
}

type storeImpl struct {
	core.PropertyStore
}
