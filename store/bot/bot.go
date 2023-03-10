package bot

import (
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/store"
	"github.com/pandodao/botastic/store/bot/dao"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func init() {
	cfg := gen.Config{
		OutPath: "store/bot/dao",
	}
	store.RegistGenerate(
		cfg,
		func(g *gen.Generator) {
			g.ApplyInterface(func(core.BotStore) {}, core.Bot{})
		},
	)
}

func New(db *gorm.DB) core.BotStore {
	dao.SetDefault(db)
	s := &storeImpl{}
	v, ok := interface{}(dao.Bot).(core.BotStore)
	if !ok {
		panic("dao.Bot is not core.BotStore")
	}
	s.BotStore = v
	return s
}

type storeImpl struct {
	core.BotStore
}
