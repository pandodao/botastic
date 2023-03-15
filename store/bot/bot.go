package bot

import (
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/store"
	"github.com/pandodao/botastic/store/bot/dao"
	"gorm.io/gen"
)

func init() {
	store.RegistGenerate(
		gen.Config{
			OutPath: "store/bot/dao",
		},
		func(g *gen.Generator) {
			g.ApplyInterface(func(core.BotStore) {}, core.Bot{})
		},
	)
}

func New(h *store.Handler) core.BotStore {
	var q *dao.Query
	if !dao.Q.Available() {
		dao.SetDefault(h.DB)
		q = dao.Q
	} else {
		q = dao.Use(h.DB)
	}

	v, ok := interface{}(q.Bot).(core.BotStore)
	if !ok {
		panic("dao.Bot is not core.BotStore")
	}

	return &storeImpl{
		BotStore: v,
	}
}

type storeImpl struct {
	core.BotStore
}
