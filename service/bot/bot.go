package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/store"
	"github.com/patrickmn/go-cache"
)

func New(
	cfg Config,
	apps core.AppStore,
	bots core.BotStore,
	models core.ModelStore,
	middlewarez core.MiddlewareService,
) *service {
	botCache := cache.New(time.Minute*5, time.Minute*5)

	conversationMap := make(map[string]*core.Conversation)

	return &service{
		cfg:             cfg,
		apps:            apps,
		bots:            bots,
		models:          models,
		middlewarez:     middlewarez,
		botCache:        botCache,
		conversationMap: conversationMap,
	}
}

type (
	Config struct {
	}

	service struct {
		cfg             Config
		apps            core.AppStore
		bots            core.BotStore
		models          core.ModelStore
		middlewarez     core.MiddlewareService
		botCache        *cache.Cache
		conversationMap map[string]*core.Conversation
	}
)

func (s *service) ReplaceStore(bots core.BotStore) core.BotService {
	return New(s.cfg, s.apps, bots, s.models, s.middlewarez)
}

func (s *service) GetBot(ctx context.Context, id uint64) (*core.Bot, error) {
	key := fmt.Sprintf("bot-%d", id)
	if bot, found := s.botCache.Get(key); found {
		return bot.(*core.Bot), nil
	}

	bot, err := s.bots.GetBot(ctx, id)
	if err != nil {
		return nil, err
	}

	s.botCache.Set(key, bot, cache.DefaultExpiration)
	return bot, nil
}

func (s *service) GetPublicBots(ctx context.Context) ([]*core.Bot, error) {
	key := "public-bots"
	if bots, found := s.botCache.Get(key); found {
		return bots.([]*core.Bot), nil
	}

	bots, err := s.bots.GetPublicBots(ctx)
	if err != nil {
		return nil, err
	}

	s.botCache.Set(key, bots, cache.DefaultExpiration)

	return bots, nil
}

func (s *service) GetBotsByUserID(ctx context.Context, userID uint64) ([]*core.Bot, error) {
	key := fmt.Sprintf("user-bots-%d", userID)
	if bots, found := s.botCache.Get(key); found {
		return bots.([]*core.Bot), nil
	}

	bots, err := s.bots.GetBotsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	s.botCache.Set(key, bots, cache.DefaultExpiration)

	return bots, nil
}

func (s *service) CreateBot(ctx context.Context,
	id uint64,
	name, model, prompt string,
	temperature float32,
	max_turn_count, context_turn_count int,
	middlewares core.MiddlewareConfig,
	public bool,
) (*core.Bot, error) {

	// check model if exists
	if _, err := s.models.GetModel(ctx, model); err != nil {
		if store.IsNotFoundErr(err) {
			return nil, core.ErrInvalidModel
		}
		fmt.Printf("models.GetModel err: %v\n", err)
		return nil, err
	}

	botID, err := s.bots.CreateBot(ctx, id, name, model, prompt, temperature, max_turn_count, context_turn_count, middlewares, public)
	if err != nil {
		fmt.Printf("bots.CreateBot err: %v\n", err)
		return nil, err
	}

	bot, err := s.bots.GetBot(ctx, botID)
	if err != nil {
		fmt.Printf("bots.GetBot err: %v\n", err)
		return nil, err
	}

	key := fmt.Sprintf("bot-%d", botID)

	s.botCache.Set(key, bot, cache.DefaultExpiration)
	s.botCache.Delete(fmt.Sprintf("user-bots-%d", id))

	return bot, nil
}

func (s *service) UpdateBot(ctx context.Context, id uint64, name, model, prompt string, temperature float32, maxTurnCount, contextTurnCount int, middlewares core.MiddlewareConfig, public bool) error {
	bot, err := s.bots.GetBot(ctx, id)
	if err != nil {
		fmt.Printf("bots.GetBot err: %v\n", err)
		return err
	}

	err = s.bots.UpdateBot(ctx, id, name, model, prompt, temperature, maxTurnCount, contextTurnCount, middlewares, public)
	if err != nil {
		fmt.Printf("bots.UpdateBot err: %v\n", err)
		return err
	}

	s.botCache.Delete(fmt.Sprintf("bot-%d", id))
	s.botCache.Delete(fmt.Sprintf("user-bots-%d", bot.UserID))
	return nil
}

func (s *service) DeleteBot(ctx context.Context, id uint64) error {
	bot, err := s.bots.GetBot(ctx, id)
	if err != nil {
		fmt.Printf("bots.GetBot err: %v\n", err)
		return err
	}

	if err := s.bots.DeleteBot(ctx, id); err != nil {
		return err
	}

	s.botCache.Delete(fmt.Sprintf("bot-%d", bot.ID))
	s.botCache.Delete(fmt.Sprintf("user-bots-%d", bot.UserID))
	return nil
}
