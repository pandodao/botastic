package core

import (
	"bytes"
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"text/template"
	"time"

	gogpt "github.com/sashabaranov/go-openai"
)

const (
	MiddlewareBotasticSearch    = "botastic-search"
	MiddlewareDuckduckgoSearch  = "duckduckgo-search"
	MiddlewareIntentRecognition = "intent-recognition"
	MiddlewareFetch             = "fetch"
	MiddlewareBotastic          = "botastic"
)

const (
	MiddlewareProcessCodeOK = iota
	MiddlewareProcessCodeUnknown
	MiddlewareProcessCodeInvalidOptions
	MiddlewareProcessCodeInternalError
	MiddlewareProcessCodeTimeout
)

type (
	Middleware struct {
		Name    string         `json:"name"`
		Options map[string]any `json:"options,omitempty"`
	}

	MiddlewareConfig struct {
		Items []*Middleware `json:"items"`
	}

	MiddlewareProcessResult struct {
		Opts     map[string]any `json:"opts,omitempty"`
		Name     string         `json:"name"`
		Code     int            `json:"code"`
		Result   string         `json:"result,omitempty"`
		Err      error          `json:"err,omitempty"`
		Required bool           `json:"required,omitempty"`
	}

	MiddlewareDesc interface {
		Name() string
		ValidateOptions(opts map[string]any) (any, error)
		Process(ctx context.Context, opts any, turn *ConvTurn) (string, error)
	}

	MiddlewareService interface {
		ProcessByConfig(ctx context.Context, m MiddlewareConfig, turn *ConvTurn) MiddlewareResults
		Process(ctx context.Context, m *Middleware, turn *ConvTurn) *MiddlewareProcessResult
		Register(ms ...MiddlewareDesc)
	}
)

func (m Middleware) GetName() string {
	return m.Name
}

func (a MiddlewareConfig) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *MiddlewareConfig) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, a)
}

type (
	Bot struct {
		ID               uint64           `json:"id"`
		Name             string           `json:"name"`
		UserID           uint64           `json:"user_id"`
		Prompt           string           `json:"prompt"`
		BoundaryPrompt   string           `json:"boundary_prompt"`
		Model            string           `json:"model"`
		MaxTurnCount     int              `json:"max_turn_count"`
		ContextTurnCount int              `json:"context_turn_count"`
		Temperature      float32          `json:"temperature"`
		MiddlewareJson   MiddlewareConfig `gorm:"type:jsonb" json:"middlewares"`
		Public           bool             `json:"public"`

		CreatedAt time.Time  `json:"created_at"`
		UpdatedAt time.Time  `json:"updated_at"`
		DeletedAt *time.Time `json:"deleted_at"`

		PromptTpl         *template.Template `gorm:"-" json:"-"`
		BoundaryPromptTpl *template.Template `gorm:"-" json:"-"`
	}

	BotStore interface {
		// SELECT *
		// FROM @@table WHERE
		// 	"id"=@id AND "deleted_at" IS NULL
		// LIMIT 1
		GetBot(ctx context.Context, id uint64) (*Bot, error)

		// SELECT *
		// FROM @@table WHERE
		// 	"user_id"=@userID AND "deleted_at" IS NULL
		GetBotsByUserID(ctx context.Context, userID uint64) ([]*Bot, error)

		// SELECT *
		// FROM @@table WHERE
		// 	"public"='t' AND "deleted_at" IS NULL
		GetPublicBots(ctx context.Context) ([]*Bot, error)

		// INSERT INTO @@table
		// 	("user_id", "name", "model", "prompt", "boundary_prompt", "temperature",
		// 	 "max_turn_count", "context_turn_count",
		//   "middleware_json", "public",
		//   "created_at", "updated_at")
		// VALUES
		// 	(@bot.UserID, @bot.Name, @bot.Model, @bot.Prompt, @bot.BoundaryPrompt, @bot.Temperature,
		//   @bot.MaxTurnCount, @bot.ContextTurnCount,
		//   @bot.MiddlewareJson, @bot.Public,
		//   NOW(), NOW())
		// RETURNING "id"
		CreateBot(ctx context.Context, bot *Bot) (uint64, error)

		// UPDATE @@table
		// 	{{set}}
		// 		"name"=@bot.Name,
		// 		"model"=@bot.Model,
		// 		"prompt"=@bot.Prompt,
		// 		"boundary_prompt"=@bot.BoundaryPrompt,
		// 		"temperature"=@bot.Temperature,
		// 		"max_turn_count"=@bot.MaxTurnCount,
		// 		"context_turn_count"=@bot.ContextTurnCount,
		// 		"middleware_json"=@bot.MiddlewareJson,
		//    "public"=@bot.Public,
		// 		"updated_at"=NOW()
		// 	{{end}}
		// WHERE
		// 	"id"=@bot.ID AND "deleted_at" is NULL
		UpdateBot(ctx context.Context, bot *Bot) error

		// UPDATE @@table
		// 	{{set}}
		// 		"deleted_at"=NOW()
		// 	{{end}}
		// WHERE
		// 	"id"=@id AND "deleted_at" is NULL
		DeleteBot(ctx context.Context, id uint64) error
	}

	BotService interface {
		GetBot(ctx context.Context, id uint64) (*Bot, error)
		GetPublicBots(ctx context.Context) ([]*Bot, error)
		GetBotsByUserID(ctx context.Context, userID uint64) ([]*Bot, error)
		CreateBot(ctx context.Context, b *Bot) error
		UpdateBot(ctx context.Context, b *Bot) error
		DeleteBot(ctx context.Context, id uint64) error
		ReplaceStore(store BotStore) BotService
	}
)

func (t *Bot) GetRequestContent(conv *Conversation, question string, additionData map[string]any) string {
	var buf bytes.Buffer
	data := map[string]interface{}{
		"LangHint": conv.LangHint(),
	}
	for k, v := range additionData {
		data[k] = v
	}

	var result string
	if t.Prompt != "" {
		if t.PromptTpl == nil {
			t.PromptTpl = template.Must(template.New(fmt.Sprintf("%d-prompt-tmpl", t.ID)).Parse(t.Prompt))
		}
		t.PromptTpl.Execute(&buf, data)

		str := buf.String()
		result = strings.TrimSpace(str) + "\n"
	}

	result += conv.HistoryToText()

	if t.BoundaryPrompt != "" {
		if t.BoundaryPromptTpl == nil {
			t.BoundaryPromptTpl = template.Must(template.New(fmt.Sprintf("%d-boundary-prompt-tmpl", t.ID)).Parse(t.BoundaryPrompt))
		}
		t.BoundaryPromptTpl.Execute(&buf, data)

		str := buf.String()
		result += "\n" + strings.TrimSpace(str)
	}

	return result
}

func (t *Bot) GetChatMessages(conv *Conversation, additionData map[string]any) []gogpt.ChatCompletionMessage {
	var buf bytes.Buffer
	data := map[string]interface{}{
		"LangHint": conv.LangHint(),
	}

	for k, v := range additionData {
		data[k] = v
	}

	result := []gogpt.ChatCompletionMessage{}
	if t.Prompt != "" {
		if t.PromptTpl == nil {
			t.PromptTpl = template.Must(template.New(fmt.Sprintf("%d-prompt-tmpl", t.ID)).Parse(t.Prompt))
		}

		t.PromptTpl.Execute(&buf, data)
		str := buf.String()
		result = append(result, gogpt.ChatCompletionMessage{
			Role:    "system",
			Content: str,
		})
	}

	history := conv.History
	if len(history) > t.ContextTurnCount {
		history = history[len(history)-conv.Bot.ContextTurnCount:]
	}

	for _, turn := range history {
		result = append(result, gogpt.ChatCompletionMessage{
			Role:    "user",
			Content: turn.Request,
		})
		if turn.Response != "" {
			result = append(result, gogpt.ChatCompletionMessage{
				Role:    "assistant",
				Content: turn.Response,
			})
		}
	}

	if t.BoundaryPrompt != "" {
		if t.BoundaryPromptTpl == nil {
			t.BoundaryPromptTpl = template.Must(template.New(fmt.Sprintf("%d-boundary-prompt-tmpl", t.ID)).Parse(t.BoundaryPrompt))
		}

		t.BoundaryPromptTpl.Execute(&buf, data)
		str := buf.String()
		result = append(result, gogpt.ChatCompletionMessage{
			Role:    "system",
			Content: str,
		})

	}

	return result
}
