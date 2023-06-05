package api

import (
	"context"
	"errors"
)

var (
	ErrModelNotFound        = errors.New("model not found")
	ErrTooManyRequestTokens = errors.New("too many request tokens")
)

type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

type ChatRequest struct {
	Temperature    float32
	Prompt         string
	BoundaryPrompt string
	History        []string
	Request        string
}

type ChatResponse struct {
	Response string
	Usage    Usage
}

type CreateEmbeddingRequest struct {
	Input []string
}

type Embedding struct {
	Embedding []float32
	Index     int
}

type CreateEmbeddingResponse struct {
	Data  []Embedding
	Usage Usage
}

type ChatLLM interface {
	Name() string
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	MaxRequestTokens() int // 0 means unlimited
}

type EmbeddingLLM interface {
	Name() string
	CreateEmbedding(ctx context.Context, req CreateEmbeddingRequest) (*CreateEmbeddingResponse, error)
	MaxRequestTokens() int // 0 means unlimited
}
