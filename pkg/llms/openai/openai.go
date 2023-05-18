package openai

import (
	"context"
	"time"

	"github.com/pandodao/botastic/config"
	"github.com/pandodao/botastic/pkg/llms/api"
	"github.com/sashabaranov/go-openai"
	gogpt "github.com/sashabaranov/go-openai"
)

type Handler struct {
	cfg    *config.OpenAIConfig
	client *openai.Client
}

func Init(cfg *config.OpenAIConfig) *Handler {
	return &Handler{
		cfg:    cfg,
		client: openai.NewClient(cfg.Key),
	}
}

func (h *Handler) ChatModels() []api.ChatLLM {
	ms := make([]api.ChatLLM, 0, len(h.cfg.ChatModels))
	for _, cm := range h.cfg.ChatModels {
		ms = append(ms, &HandlerWithModel{
			model:   cm,
			Handler: h,
		})
	}
	return ms
}

type HandlerWithModel struct {
	*Handler
	model string
}

func (h *HandlerWithModel) Name() string {
	return h.model
}

func (h *HandlerWithModel) Chat(ctx context.Context, req api.ChatRequest) (*api.ChatResponse, error) {
	start := time.Now()
	chatReq := openai.ChatCompletionRequest{
		Model:       h.model,
		Temperature: req.Temperature,
	}
	if req.Prompt != "" {
		chatReq.Messages = append(chatReq.Messages, gogpt.ChatCompletionMessage{
			Role:    "system",
			Content: req.Prompt,
		})
	}
	for i := 0; i < len(req.History); i += 2 {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		chatReq.Messages = append(chatReq.Messages, gogpt.ChatCompletionMessage{
			Role:    role,
			Content: req.History[i],
		})
	}

	chatReq.Messages = append(chatReq.Messages, gogpt.ChatCompletionMessage{
		Role:    "user",
		Content: req.Request,
	})

	if req.BoundaryPrompt != "" {
		chatReq.Messages = append(chatReq.Messages, gogpt.ChatCompletionMessage{
			Role:    "system",
			Content: req.BoundaryPrompt,
		})
	}

	resp, err := h.client.CreateChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, err
	}
	return &api.ChatResponse{
		Duration:         time.Since(start),
		Response:         resp.Choices[0].Message.Content,
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
	}, nil
}

func (h *Handler) CreateEmbedding(ctx context.Context, req api.CreateEmbeddingRequest) (*api.CreateEmbeddingResponse, error) {
	return nil, nil
}
