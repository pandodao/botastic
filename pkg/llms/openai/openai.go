package openai

import (
	"context"
	"fmt"

	"github.com/pandodao/botastic/config"
	"github.com/pandodao/botastic/pkg/llms/api"
	"github.com/pkoukk/tiktoken-go"
	"github.com/sashabaranov/go-openai"
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

func (h *Handler) EmbeddingModels() []api.EmbeddingLLM {
	ms := make([]api.EmbeddingLLM, 0, len(h.cfg.EmbeddingModels))
	for _, em := range h.cfg.EmbeddingModels {
		var embeddingModel openai.EmbeddingModel
		_ = embeddingModel.UnmarshalText([]byte(em))
		if embeddingModel == openai.Unknown {
			continue
		}
		ms = append(ms, &HandlerWithModel{
			model:          em,
			embeddingModel: embeddingModel,
			Handler:        h,
		})
	}

	return ms
}

type HandlerWithModel struct {
	*Handler
	model string

	embeddingModel openai.EmbeddingModel
}

func (h *HandlerWithModel) Name() string {
	return h.model
}

func (h *HandlerWithModel) Chat(ctx context.Context, req api.ChatRequest) (*api.ChatResponse, error) {
	chatReq := openai.ChatCompletionRequest{
		Model:       h.model,
		Temperature: req.Temperature,
		Messages:    getMessagesFromRequest(req),
	}

	resp, err := h.client.CreateChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, err
	}
	return &api.ChatResponse{
		Response: resp.Choices[0].Message.Content,
		Usage: api.Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}, nil
}

func (h *HandlerWithModel) CalChatRequestTokens(ctx context.Context, req api.ChatRequest) (int, error) {
	tkm, err := tiktoken.EncodingForModel(h.model)
	if err != nil {
		return 0, err
	}

	var tokensPerMessage int
	var tokensPerName int
	switch h.model {
	case "gpt-3.5-turbo-0301", "gpt-3.5-turbo":
		tokensPerMessage = 4
		tokensPerName = -1
	case "gpt-4-0314", "gpt-4":
		tokensPerMessage = 3
		tokensPerName = 1
	default:
		tokensPerMessage = 3
		tokensPerName = 1
	}

	numTokens := 0
	for _, message := range getMessagesFromRequest(req) {
		numTokens += tokensPerMessage
		numTokens += len(tkm.Encode(message.Content, nil, nil))
		numTokens += len(tkm.Encode(message.Role, nil, nil))
		numTokens += len(tkm.Encode(message.Name, nil, nil))
		if message.Name != "" {
			numTokens += tokensPerName
		}
	}

	return numTokens + 3, nil
}

func (h *HandlerWithModel) CreateEmbedding(ctx context.Context, req api.CreateEmbeddingRequest) (*api.CreateEmbeddingResponse, error) {
	resp, err := h.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: req.Input,
		Model: h.embeddingModel,
	})
	if err != nil {
		return nil, err
	}

	embeddings := make([]api.Embedding, len(resp.Data))
	for i, d := range resp.Data {
		embeddings[i] = api.Embedding{
			Embedding: d.Embedding,
			Index:     d.Index,
		}
	}

	return &api.CreateEmbeddingResponse{
		Data: embeddings,
		Usage: api.Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}, nil
}

func (h *HandlerWithModel) MaxRequestTokens() int {
	switch h.model {
	// chat models
	case "gpt-3.5-turbo-0301", "gpt-3.5-turbo":
		return 4096
	case "gpt-4-0314", "gpt-4":
		return 8192
	case "gpt-4-32k-0314", "gpt-4-32k":
		return 32768

	// embedding models
	case "text-embedding-ada-002":
		return 8191
	}
	return 0
}

func (h *HandlerWithModel) CalEmbeddingRequestTokens(req api.CreateEmbeddingRequest) (int, error) {
	tkm, err := tiktoken.EncodingForModel(h.model)
	if err != nil {
		return 0, fmt.Errorf("model %s not supported", h.model)
	}

	numTokens := 0
	for _, text := range req.Input {
		numTokens += len(tkm.Encode(text, nil, nil))
	}

	return numTokens, nil
}

func getMessagesFromRequest(req api.ChatRequest) []openai.ChatCompletionMessage {
	messages := make([]openai.ChatCompletionMessage, 0, len(req.History)+2)
	if req.Prompt != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: req.Prompt,
		})
	}
	for i := 0; i < len(req.History); i += 2 {
		role := openai.ChatMessageRoleUser
		if i%2 == 1 {
			role = openai.ChatMessageRoleAssistant
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: req.History[i],
		})
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: req.Request,
	})

	if req.BoundaryPrompt != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: req.BoundaryPrompt,
		})
	}

	return messages
}
