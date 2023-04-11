package conv

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/session"
	"github.com/pandodao/botastic/store"
)

func New(
	cfg Config,
	convs core.ConversationStore,
	botz core.BotService,
	apps core.AppStore,
) *service {
	return &service{
		cfg:             cfg,
		convs:           convs,
		botz:            botz,
		apps:            apps,
		conversationMap: make(map[string]*core.Conversation),
	}
}

type (
	Config struct {
	}

	service struct {
		cfg             Config
		convs           core.ConversationStore
		botz            core.BotService
		apps            core.AppStore
		conversationMap map[string]*core.Conversation
	}
)

func (s *service) ReplaceStore(convs core.ConversationStore) core.ConversationService {
	return New(s.cfg, convs, s.botz, s.apps)
}

func (s *service) CreateConversation(ctx context.Context, botID, appID uint64, userIdentity, lang string) (*core.Conversation, error) {
	app := session.AppFrom(ctx)

	bot, err := s.botz.GetBot(ctx, botID)
	if err != nil {
		return nil, err
	}

	if !bot.Public && app.UserID != bot.UserID {
		return nil, core.ErrBotNotFound
	}
	if lang == "" {
		lang = "en"
	}

	now := time.Now()
	conv := &core.Conversation{
		ID:           uuid.New().String(),
		AppID:        app.ID,
		BotID:        bot.ID,
		UserIdentity: userIdentity,
		Lang:         lang,
		CreatedAt:    now,
		UpdatedAt:    now,
		Bot:          bot,
		App:          app,
	}

	if err := s.convs.CreateConversation(ctx, conv); err != nil {
		return nil, err
	}

	s.conversationMap[conv.ID] = conv
	return conv, nil
}

func (s *service) GetConversation(ctx context.Context, convID string) (*core.Conversation, error) {
	conv, ok := s.conversationMap[convID]
	if !ok {
		// load from db
		var err error
		conv, err = s.convs.GetConversation(ctx, convID)
		if err != nil {
			if store.IsNotFoundErr(err) {
				return nil, core.ErrConvNotFound
			}
			return nil, core.ErrInternalServer
		}

		app := session.AppFrom(ctx)
		if app != nil {
			if conv.AppID != app.ID {
				return nil, core.ErrConvNotFound
			}
			conv.App = app
		} else {
			app, err := s.apps.GetApp(ctx, conv.AppID)
			if err != nil {
				if store.IsNotFoundErr(err) {
					return nil, core.ErrAppNotFound
				}
				return nil, core.ErrInternalServer
			}
			conv.App = app
		}

		bot, err := s.botz.GetBot(ctx, conv.BotID)
		if err != nil {
			if store.IsNotFoundErr(err) {
				return nil, core.ErrBotNotFound
			}
			return nil, core.ErrInternalServer
		}
		conv.Bot = bot

		// load history
		turns, err := s.convs.GetConvTurnsByConversationID(ctx, conv.ID, bot.MaxTurnCount)
		if err != nil {
			return nil, core.ErrInternalServer
		}

		conv.History = turns
	}

	ids := []uint64{}
	for ix := len(conv.History) - 1; ix >= 0; ix-- {
		turn := conv.History[ix]
		if !turn.IsProcessed() {
			ids = append(ids, turn.ID)
		}
	}

	if len(ids) != 0 {
		turns, _ := s.convs.GetConvTurns(ctx, ids)
		if len(turns) != 0 {
			turnMap := make(map[uint64]*core.ConvTurn)
			for _, turn := range turns {
				turnMap[turn.ID] = turn
			}

			for ix := len(conv.History) - 1; ix >= 0; ix-- {
				turn := conv.History[ix]
				existed, ok := turnMap[turn.ID]
				if ok && existed.Status != turn.Status {
					*turn = *existed
				}
			}
		}
	}

	return conv, nil
}

func (s *service) PostToConversation(ctx context.Context, conv *core.Conversation, input string, bo core.BotOverride) (*core.ConvTurn, error) {
	turnID, err := s.convs.CreateConvTurn(ctx, conv.ID, conv.Bot.ID, conv.App.ID, conv.App.UserID, conv.UserIdentity, input, bo)
	if err != nil {
		return nil, err
	}

	turn, err := s.convs.GetConvTurn(ctx, turnID)
	if err != nil {
		return nil, err
	}

	bot, err := s.botz.GetBot(ctx, conv.Bot.ID)
	if err != nil {
		return nil, err
	}

	conv.History = append(conv.History, turn)

	if len(conv.History) > bot.MaxTurnCount {
		// reduce to MaxTurnCount
		conv.History = conv.History[len(conv.History)-bot.MaxTurnCount:]
	}

	return turn, nil
}

func (s *service) DeleteConversation(ctx context.Context, convID string) error {
	delete(s.conversationMap, convID)
	return nil
}

func (s *service) getDefaultConversation(app *core.App, bot *core.Bot, uid, lang string) *core.Conversation {
	if lang == "" {
		lang = "en"
	}
	return &core.Conversation{
		ID:           uuid.New().String(),
		App:          app,
		Bot:          bot,
		UserIdentity: uid,
		Lang:         lang,
		History:      []*core.ConvTurn{},
	}
}
