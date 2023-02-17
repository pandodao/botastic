package bot

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/pandodao/botastic/core"

	"gopkg.in/yaml.v2"
)

func New(
	cfg Config,
	apps core.AppStore,
) *service {
	botMap, err := LoadBots()
	if err != nil {
		panic(err)
	}

	conversationMap := make(map[string]*core.BotConversation)

	return &service{
		cfg:  cfg,
		apps: apps,

		botMap:          botMap,
		conversationMap: conversationMap,
	}
}

type (
	Config struct {
	}

	service struct {
		cfg             Config
		apps            core.AppStore
		botMap          map[uint64]*core.Bot
		conversationMap map[string]*core.BotConversation
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
	bot, found := s.botMap[id]
	if !found {
		return nil, core.ErrBotNotFound
	}
	return bot, nil
}
