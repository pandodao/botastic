package api

import (
	"context"
	"time"
)

type ChatRequest struct {
	Temperature    float32
	Prompt         string
	BoundaryPrompt string
	History        []string
	Request        string
}

type ChatResponse struct {
	Duration         time.Duration
	Response         string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

type CreateEmbeddingRequest struct{}

type CreateEmbeddingResponse struct{}

type ChatLLM interface {
	Name() string
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}

type EmbeddingLLM interface {
	Name() string
	CreateEmbedding(ctx context.Context, req CreateEmbeddingRequest) (*CreateEmbeddingResponse, error)
}
