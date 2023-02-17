package bot

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pandodao/botastic/core"
)

func (s *service) CreateConversation(ctx context.Context, botID, appID uint64, userIdentity, lang string) (*core.BotConversation, error) {
	app, err := s.apps.GetApp(ctx, appID)
	if err != nil {
		return nil, err
	}

	bot, err := s.GetBot(ctx, botID)
	if err != nil {
		return nil, err
	}

	conv := s.getDefaultConversation(app, bot, userIdentity, lang)
	s.conversationMap[conv.ID] = conv

	return conv, nil
}

func (s *service) GetConversation(ctx context.Context, convID string) (*core.BotConversation, error) {
	conv, ok := s.conversationMap[convID]
	if !ok {
		return nil, fmt.Errorf("conversation not found")
	}

	return conv, nil
}

func (s *service) PostToConversation(ctx context.Context, conv *core.BotConversation, input string) (*core.BotConvTurn, error) {
	// @TODO: add to turn and return the turn
	return nil, nil
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

func (s *service) getDefaultConversation(app *core.App, bot *core.Bot, uid, lang string) *core.BotConversation {
	return &core.BotConversation{
		ID:           uuid.New().String(),
		App:          app,
		Bot:          bot,
		UserIdentity: uid,
		Lang:         lang,
		History:      []*core.BotConvTurn{},
	}
}
