package api

import (
	"time"

	"github.com/google/uuid"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=TurnErrorCode -linecomment -trimprefix=TurnErrorCode
//go:generate go run golang.org/x/tools/cmd/stringer -type=TurnStatus -linecomment --trimprefix TurnStatus

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
	Opts     map[string]any `json:"opts,omitempty"`
	Name     string         `json:"name"`
	Code     int            `json:"code"`
	Result   string         `json:"result,omitempty"`
	Err      string         `json:"err,omitempty"`
	Required bool           `json:"-"`
}

type MiddlewareResults []*MiddlewareResult

type Turn struct {
	ID                uint              `json:"id"`
	ConvID            uuid.UUID         `json:"conv_id"`
	BotID             uint              `json:"bot_id"`
	Request           string            `json:"request"`
	Response          string            `json:"response"`
	PromptTokens      int               `json:"prompt_tokens"`
	CompletionTokens  int               `json:"completion_tokens"`
	TotalTokens       int               `json:"total_tokens"`
	Status            TurnStatus        `json:"status"`
	MiddlewareResults MiddlewareResults `json:"middleware_results"`
	ErrorCode         TurnErrorCode     `json:"error_code,omitempty"`
	ErrorMessage      string            `json:"error_message,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

type CreateTurnRequest struct {
	Content string `json:"content" binding:"required"`
}

type CreateTurnResponse Turn

type CreateTurnOnewayRequest struct {
	ConversationID uuid.UUID `json:"conv_id"`
	BotID          uint      `json:"bot_id"`
	UserIdentity   string    `json:"user_identity"`
	CreateTurnRequest
}

type CreateTurnOnewayResponse Turn

type GetTurnRequest struct {
	BlockUntilProcessed bool          `form:"block_until_processed"`
	Timeout             time.Duration `form:"timeout"`
}

type GetTurnResponse Turn

type Bot struct {
	ID               uint             `json:"id"`
	Name             string           `json:"name"`
	ChatModel        string           `json:"chat_model"`
	Prompt           string           `json:"prompt"`
	BoundaryPrompt   string           `json:"boundary_prompt"`
	ContextTurnCount int              `json:"context_turn_count"`
	Temperature      float32          `json:"temperature"`
	Middleware       MiddlewareConfig `json:"middleware"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

type Middleware struct {
	Name    string         `json:"name"`
	Options map[string]any `json:"options,omitempty"`
}

type MiddlewareConfig struct {
	Items []*Middleware `json:"items,omitempty"`
}

type CreateBotRequest struct {
	Name             string           `json:"name" binding:"required"`
	ChatModel        string           `json:"chat_model" binding:"required"`
	Prompt           string           `json:"prompt"`
	BoundaryPrompt   string           `json:"boundary_prompt"`
	Temperature      float32          `json:"temperature" binding:"required"`
	ContextTurnCount int              `json:"context_turn_count" binding:"required"`
	Middlewares      MiddlewareConfig `json:"middlewares"`
}

type CreateBotResponse Bot

type GetBotResponse Bot

type GetBotsResponse []Bot

type UpdateBotRequest CreateBotRequest

type ListModelsResponse struct {
	ChatModels      []string `json:"chat_models"`
	EmbeddingModels []string `json:"embedding_models"`
}
