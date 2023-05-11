package llms

import (
	"fmt"

	"github.com/pandodao/botastic/config"
	"github.com/pandodao/botastic/internal/llms/api"
	"github.com/pandodao/botastic/internal/llms/openai"
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
		case "openai":
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

func (h *Handler) GetChatModel(key string) (api.ChatLLM, bool) {
	v, ok := h.chatMap[key]
	return v, ok
}

func (h *Handler) GetEmbeddingModel(key string) (api.EmbeddingLLM, bool) {
	v, ok := h.embeddingMap[key]
	return v, ok
}

func (h *Handler) ChatModels() []string {
	return h.chatModles
}

func (h *Handler) EmbeddingModels() []string {
	return h.embeddingModels
}
