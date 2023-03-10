package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/pandodao/botastic/core"
	"github.com/patrickmn/go-cache"

	"gopkg.in/yaml.v2"
)

func New(
	cfg Config,
	apps core.AppStore,
	bots core.BotStore,
	middlewarez core.MiddlewareService,
) *service {
	// botMap, err := LoadBots()
	// if err != nil {
	// 	panic(err)
	// }
	botCache := cache.New(time.Minute*5, time.Minute*5)

	conversationMap := make(map[string]*core.Conversation)

	return &service{
		cfg:             cfg,
		apps:            apps,
		bots:            bots,
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
		middlewarez     core.MiddlewareService
		botCache        *cache.Cache
		conversationMap map[string]*core.Conversation
	}
)

func LoadBots() (map[uint64]*core.Bot, error) {
	avatarMap := make(map[uint64]*core.Bot)
	base := "./bot_data"
	filenames := make([]string, 0)
	items, _ := os.ReadDir(base)
	for _, item := range items {
		if !item.IsDir() {
			if !strings.HasSuffix(item.Name(), ".yaml") {
				continue
			}
			filenames = append(filenames, item.Name())
		}
	}

	// read yaml file into s.avatarMap
	for _, filename := range filenames {
		// read yaml file
		data, err := os.ReadFile(path.Join(base, filename))
		if err != nil {
			continue
		}
		output := &core.Bot{}
		err = yaml.Unmarshal(data, output)
		if err != nil {
			fmt.Printf("Error unmarshaling yaml content of file %s: %v\n", filename, err)
			continue
		}

		output.PromptTpl = template.
			Must(template.New(fmt.Sprintf("%d-prompt-tmpl", output.ID)).Parse(output.Prompt))

		avatarMap[output.ID] = output
	}
	return avatarMap, nil
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

	if err := bot.DecodeMiddlewares(); err != nil {
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
	for _, bot := range bots {
		if err := bot.DecodeMiddlewares(); err != nil {
			continue
		}
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
	for _, bot := range bots {
		if err := bot.DecodeMiddlewares(); err != nil {
			continue
		}
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

	bytes, err := json.Marshal(middlewares)
	if err != nil {
		fmt.Printf("json.Marshal err: %v\n", err)
		return nil, err
	}

	jsonb := core.JSONB{}
	if err := jsonb.Scan(bytes); err != nil {
		fmt.Printf("jsonb.Scan err: %v\n", err)
		return nil, err
	}

	botID, err := s.bots.CreateBot(ctx, id, name, model, prompt, temperature, max_turn_count, context_turn_count, jsonb, public)
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

	return bot, nil
}

func (s *service) UpdateBot(ctx context.Context, id uint64, name, model, prompt string, temperature float32, maxTurnCount, contextTurnCount int, middlewares core.MiddlewareConfig, public bool) error {
	bytes, err := json.Marshal(middlewares)
	if err != nil {
		fmt.Printf("json.Marshal err: %v\n", err)
		return err
	}

	jsonb := core.JSONB{}
	if err := jsonb.Scan(bytes); err != nil {
		fmt.Printf("jsonb.Scan err: %v\n", err)
		return err
	}

	err = s.bots.UpdateBot(ctx, id, name, model, prompt, temperature, maxTurnCount, contextTurnCount, jsonb, public)
	if err != nil {
		fmt.Printf("bots.UpdateBot err: %v\n", err)
		return err
	}

	key := fmt.Sprintf("bot-%d", id)

	s.botCache.Delete(key)

	return nil
}
