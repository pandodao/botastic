package llms

import (
	"fmt"

	"github.com/pandodao/botastic/config"
	"github.com/pandodao/botastic/pkg/llms/api"
	"github.com/pandodao/botastic/pkg/llms/openai"
)

type embeddingLLM struct {
	api.EmbeddingLLM
	model string
}

type Handler struct {
	chatMap      map[string]api.ChatLLM
	embeddingMap map[string]api.EmbeddingLLM

	chatModles      []string
	embeddingModels []string
}

func New(cfg config.LLMsConfig) *Handler {
	h := &Handler{
		chatMap:      make(map[string]api.ChatLLM),
		embeddingMap: make(map[string]api.EmbeddingLLM),
	}
	for _, name := range cfg.Enabled {
		item := cfg.Items[name]
		var r any
		switch item.Provider {
		case config.LLMProviderOpenAI:
			r = openai.Init(item.OpenAI)
		}

		if v, ok := r.(interface{ ChatModels() []api.ChatLLM }); ok {
			for _, m := range v.ChatModels() {
				key := fmt.Sprintf("%s:%s", name, m.Name())
				h.chatModles = append(h.chatModles, key)
				h.chatMap[key] = m
			}
		}

		if v, ok := r.(interface{ EmbeddingModels() []api.EmbeddingLLM }); ok {
			for _, m := range v.EmbeddingModels() {
				key := fmt.Sprintf("%s:%s", name, m.Name())
				h.embeddingModels = append(h.embeddingModels, key)
				h.embeddingMap[key] = m
			}
		}
	}

	return h
}

func (h *Handler) GetChatModel(key string) (api.ChatLLM, error) {
	v, ok := h.chatMap[key]
	if !ok {
		return nil, api.ErrModelNotFound
	}
	return v, nil
}

func (h *Handler) GetEmbeddingModel(key string) (api.EmbeddingLLM, error) {
	v, ok := h.embeddingMap[key]
	if !ok {
		return nil, api.ErrModelNotFound
	}
	return v, nil
}

func (h *Handler) ChatModels() []string {
	return h.chatModles
}

func (h *Handler) EmbeddingModels() []string {
	return h.embeddingModels
}
