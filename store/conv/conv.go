package conv

import (
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/store"
	"github.com/pandodao/botastic/store/conv/dao"

	"gorm.io/gen"
	"gorm.io/gorm"
)

func init() {
	cfg := gen.Config{
		OutPath: "store/conv/dao",
	}
	store.RegistGenerate(
		cfg,
		func(g *gen.Generator) {
			g.ApplyInterface(func(core.ConversationStore) {}, core.ConvTurn{})
		},
	)
}

func New(db *gorm.DB) core.ConversationStore {
	dao.SetDefault(db)
	s := &storeImpl{}
	v, ok := interface{}(dao.ConvTurn).(core.ConversationStore)
	if !ok {
		panic("dao.Conv is not core.ConversationStore")
	}
	s.ConversationStore = v
	return s
}

type storeImpl struct {
	core.ConversationStore
}
