package llms

import (
	"fmt"

	"github.com/pandodao/botastic/config"
	"github.com/pandodao/botastic/internal/llms/api"
	"github.com/pandodao/botastic/internal/llms/openai"
)

type chatLLM struct {
	api.ChatLLM
	model string
}

type embeddingLLM struct {
	api.EmbeddingLLM
	model string
}

type Hanlder struct {
	chatMap      map[string]*chatLLM
	embeddingMap map[string]*embeddingLLM

	chatModles      []string
	embeddingModels []string
}

func New(cfg config.LLMsConfig) *Hanlder {
	h := &Hanlder{
		chatMap:      make(map[string]*chatLLM),
		embeddingMap: make(map[string]*embeddingLLM),
	}
	for _, name := range cfg.Enabled {
		item := cfg.Items[name]
		var r any
		switch item.Provider {
		case "openai":
			r = openai.Init(item.OpenAI)
		}

		if v, ok := r.(api.ChatLLM); ok {
			for _, cm := range item.ChatModels {
				key := fmt.Sprintf("%s:%s", name, cm)
				h.chatModles = append(h.chatModles, key)
				h.chatMap[key] = &chatLLM{
					ChatLLM: v,
					model:   cm,
				}
			}
		}
		if v, ok := r.(api.EmbeddingLLM); ok {
			for _, em := range item.EmbeddingModels {
				key := fmt.Sprintf("%s:%s", name, em)
				h.embeddingModels = append(h.embeddingModels, key)
				h.embeddingMap[key] = &embeddingLLM{
					EmbeddingLLM: v,
					model:        em,
				}
			}
		}
	}

	return h
}

func (h *Hanlder) ChatModels() []string {
	return h.chatModles
}

func (h *Hanlder) EmbeddingModels() []string {
	return h.embeddingModels
}
