package conv

import (
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/store"
	"github.com/pandodao/botastic/store/conv/dao"

	"gorm.io/gen"
)

func init() {
	store.RegistGenerate(
		gen.Config{
			OutPath: "store/conv/dao",
		},
		func(g *gen.Generator) {
			g.ApplyInterface(func(core.ConversationStore) {}, core.ConvTurn{})
		},
	)
}

func New(h *store.Handler) core.ConversationStore {
	var q *dao.Query
	if !dao.Q.Available() {
		dao.SetDefault(h.DB)
		q = dao.Q
	} else {
		q = dao.Use(h.DB)
	}

	v, ok := interface{}(q.ConvTurn).(core.ConversationStore)
	if !ok {
		panic("dao.Conv is not core.ConversationStore")
	}

	return &storeImpl{
		ConversationStore: v,
	}
}

type storeImpl struct {
	core.ConversationStore
}
