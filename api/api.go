package api

import (
	"time"

	"github.com/google/uuid"
)

type Response struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func NewErrorResponse(code int, message string) *Response {
	return &Response{
		Code:    code,
		Message: message,
	}
}

func NewSuccessResponse(data any) *Response {
	return &Response{
		Data: data,
	}
}

type Conv struct {
	ID        uuid.UUID `json:"id"`
	BotID     uint      `json:"bot_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateConvRequest struct {
	BotID        uint   `json:"bot_id" binding:"required"`
	UserIdentity string `json:"user_identity"`
}

type CreateConvResponse Conv

type UpdateConvRequest CreateConvRequest

type GetConvResponse Conv

type MiddlewareResult struct {
	Middleware
	Code       MiddlewareErrorCode `json:"code"`
	Err        string              `json:"err,omitempty"`
	RenderData map[string]any      `json:"render_data,omitempty"`
}

type TurnError struct {
	Code TurnErrorCode `json:"code"`
	Msg  string        `json:"msg"`
}

type Turn struct {
	ID                uint                `json:"id"`
	ConversationID    uuid.UUID           `json:"conversation_id"`
	BotID             uint                `json:"bot_id"`
	Request           string              `json:"request"`
	Response          string              `json:"response"`
	PromptTokens      int                 `json:"prompt_tokens"`
	CompletionTokens  int                 `json:"completion_tokens"`
	TotalTokens       int                 `json:"total_tokens"`
	Status            TurnStatus          `json:"status"`
	MiddlewareResults []*MiddlewareResult `json:"middleware_results,omitempty"`
	Error             *TurnError          `json:"error,omitempty"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`
}

type CreateTurnRequest struct {
	Content string `json:"content" binding:"required"`
}

type CreateTurnResponse Turn

type CreateTurnOnewayRequest struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	BotID          uint      `json:"bot_id"`
	UserIdentity   string    `json:"user_identity"`
	CreateTurnRequest
	GetTurnRequest
}

type CreateTurnOnewayResponse Turn

type GetTurnRequest struct {
	BlockUntilProcessed bool `form:"block_until_processed" json:"block_until_processed"`
	TimeoutSeconds      int  `form:"timeout_seconds" json:"timeout_seconds"`
}

type GetTurnResponse Turn

type Bot struct {
	ID               uint              `json:"id"`
	Name             string            `json:"name"`
	ChatModel        string            `json:"chat_model"`
	Prompt           string            `json:"prompt"`
	BoundaryPrompt   string            `json:"boundary_prompt"`
	ContextTurnCount int               `json:"context_turn_count"`
	Temperature      float32           `json:"temperature"`
	TimeoutSeconds   int               `json:"timeout_seconds"`
	Middlewares      *MiddlewareConfig `json:"middlewares,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

type Middleware struct {
	ID      string            `json:"id" binding:"required"`
	Name    string            `json:"name" binding:"required"`
	Options map[string]string `json:"options,omitempty"`
}

type MiddlewareConfig struct {
	Items []*Middleware `json:"items,omitempty"`
}

type CreateBotRequest struct {
	Name             string            `json:"name" binding:"required"`
	ChatModel        string            `json:"chat_model" binding:"required"`
	Prompt           string            `json:"prompt"`
	BoundaryPrompt   string            `json:"boundary_prompt"`
	Temperature      float32           `json:"temperature" binding:"required"`
	ContextTurnCount int               `json:"context_turn_count" binding:"required"`
	Middlewares      *MiddlewareConfig `json:"middlewares"`
}

type CreateBotResponse Bot

type GetBotResponse Bot

type GetBotsResponse []Bot

type UpdateBotRequest CreateBotRequest

type ListModelsResponse struct {
	ChatModels      []string `json:"chat_models"`
	EmbeddingModels []string `json:"embedding_models"`
}

type MiddlewareDescOption struct {
	Name         string `json:"name"`
	Desc         string `json:"desc"`
	DefaultValue string `json:"default_value,omitempty"`
	Required     bool   `json:"required,omitempty"`

	Value          any                       `json:"-"`
	ParseValueFunc func(string) (any, error) `json:"-"`
}

type MiddlewareDesc struct {
	Name    string                  `json:"name"`
	Desc    string                  `json:"desc"`
	Options []*MiddlewareDescOption `json:"options"`
}

type ListMiddlewaresResponse struct {
	GeneralOptions []*MiddlewareDescOption `json:"general_options"`
	Middlewares    []*MiddlewareDesc       `json:"middlewares"`
}

type UpsertIndexesRequest struct {
	GroupKey       string `json:"group_key" binding:"required"`
	EmbeddingModel string `json:"embedding_model" binding:"required"`
	ForceRebuild   bool   `json:"force_rebuild"`
	Items          []struct {
		ID         uint           `json:"id"`
		Data       string         `json:"data" binding:"required"`
		Properties map[string]any `json:"properties"`
	} `json:"items" binding:"required"`
}

type UpsertIndexesResponse struct {
	Items []*Index `json:"items"`
}

type Index struct {
	ID         uint           `json:"id"`
	GroupKey   string         `json:"group_key"`
	Data       string         `json:"data"`
	Properties map[string]any `json:"properties"`
	Score      float64        `json:"score,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

type SearchIndexesRequest struct {
	GroupKey       string `form:"group_key" binding:"required"`
	EmbeddingModel string `form:"embedding_model" binding:"required"`
	Keyword        string `form:"keyword" binding:"required"`
	Limit          int    `form:"limit"`
}

type SearchIndexesResponse struct {
	Items []*Index `json:"items"`
}
