package api

import (
	"context"
	"time"
)

type ChatRequest struct {
	Model          string
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
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}

type EmbeddingLLM interface {
	CreateEmbedding(ctx context.Context, req CreateEmbeddingRequest) (*CreateEmbeddingResponse, error)
}
