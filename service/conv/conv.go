package conv

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pandodao/botastic/core"
)

func New(
	cfg Config,
	apps core.AppStore,
	convs core.ConversationStore,

	botz core.BotService,
) *service {

	conversationMap := make(map[string]*core.Conversation)

	return &service{
		cfg:             cfg,
		apps:            apps,
		convs:           convs,
		botz:            botz,
		conversationMap: conversationMap,
	}
}

type (
	Config struct {
	}

	service struct {
		cfg             Config
		apps            core.AppStore
		convs           core.ConversationStore
		botz            core.BotService
		conversationMap map[string]*core.Conversation
	}
)

func (s *service) CreateConversation(ctx context.Context, botID, appID uint64, userIdentity, lang string) (*core.Conversation, error) {
	app, err := s.apps.GetApp(ctx, appID)
	if err != nil {
		return nil, err
	}

	bot, err := s.botz.GetBot(ctx, botID)
	if err != nil {
		return nil, err
	}

	conv := s.getDefaultConversation(app, bot, userIdentity, lang)
	s.conversationMap[conv.ID] = conv

	return conv, nil
}

func (s *service) GetConversation(ctx context.Context, convID string) (*core.Conversation, error) {
	conv, ok := s.conversationMap[convID]
	if !ok {
		return nil, fmt.Errorf("conversation not found")
	}

	return conv, nil
}

func (s *service) PostToConversation(ctx context.Context, conv *core.Conversation, input string) (*core.ConvTurn, error) {
	// @TODO: add to turn and return the turn
	turnID, err := s.convs.CreateConvTurn(ctx, conv.Bot.ID, conv.App.ID, conv.UserIdentity, input)
	if err != nil {
		return nil, err
	}
	turns, err := s.convs.GetConvTurns(ctx, []uint64{turnID})
	if err != nil {
		return nil, err
	}

	turn := turns[0]
	conv.History = append(conv.History, turn)

	return turn, nil
}

func (s *service) ClearExpiredConversations(ctx context.Context) error {
	for key, conv := range s.conversationMap {
		if conv.IsExpired() {
			delete(s.conversationMap, key)
		}
	}
	return nil
}

func (s *service) DeleteConversation(ctx context.Context, convID string) error {
	delete(s.conversationMap, convID)
	return nil
}

func (s *service) getDefaultConversation(app *core.App, bot *core.Bot, uid, lang string) *core.Conversation {
	return &core.Conversation{
		ID:           uuid.New().String(),
		App:          app,
		Bot:          bot,
		UserIdentity: uid,
		Lang:         lang,
		History:      []*core.ConvTurn{},
		ExpiredAt:    time.Now().Add(10 * time.Minute),
	}
}
