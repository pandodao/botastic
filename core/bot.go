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

	gogpt "github.com/sashabaranov/go-gpt3"
)

type JSONB json.RawMessage

// implement sql.Scanner interface, Scan value into Jsonb
func (j *JSONB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("type assertion to []byte failed:", value))
	}

	result := json.RawMessage{}
	err := json.Unmarshal(bytes, &result)
	if err != nil {
		return err
	}
	*j = JSONB(result)
	return err
}

// implement driver.Valuer interface, Value return json value
func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}

type (
	Bot struct {
		ID               uint64  `json:"id"`
		Name             string  `json:"name"`
		UserID           uint64  `json:"user_id"`
		Prompt           string  `json:"prompt"`
		Model            string  `json:"model"`
		MaxTurnCount     int     `json:"max_turn_count"`
		ContextTurnCount int     `json:"context_turn_count"`
		Temperature      float32 `json:"temperature"`
		MiddlewareJson   JSONB   `gorm:"type:jsonb" json:"-"`
		Public           bool    `json:"public"`

		CreatedAt *time.Time `json:"created_at"`
		UpdatedAt *time.Time `json:"updated_at"`
		DeletedAt *time.Time `json:"deleted_at"`

		PromptTpl   *template.Template `gorm:"-" json:"-"`
		Middlewares MiddlewareConfig   `gorm:"-" json:"middlewares"`
	}

	BotStore interface {
		// SELECT "id",
		// 	 "user_id", "name", "model", "prompt", "temperature",
		// 	 "max_turn_count", "context_turn_count",
		//   "middleware_json", "public",
		//   "created_at", "updated_at"
		// FROM @@table WHERE
		// 	"id"=@id AND "deleted_at" IS NULL
		// LIMIT 1
		GetBot(ctx context.Context, id uint64) (*Bot, error)

		// SELECT "id",
		// 	 "user_id", "name", "model", "prompt", "temperature",
		// 	 "max_turn_count", "context_turn_count",
		//   "middleware_json", "public",
		//   "created_at", "updated_at"
		// FROM @@table WHERE
		// 	"user_id"=@userID AND "deleted_at" IS NULL
		GetBotsByUserID(ctx context.Context, userID uint64) ([]*Bot, error)

		// SELECT "id",
		// 	 "user_id", "name", "model", "prompt", "temperature",
		// 	 "max_turn_count", "context_turn_count",
		//   "middleware_json", "public",
		//   "created_at", "updated_at"
		// FROM @@table WHERE
		// 	"public"='t' AND "deleted_at" IS NULL
		GetPublicBots(ctx context.Context) ([]*Bot, error)

		// INSERT INTO @@table
		// 	("user_id", "name", "model", "prompt", "temperature",
		// 	 "max_turn_count", "context_turn_count",
		//   "middleware_json", "public",
		//   "created_at", "updated_at")
		// VALUES
		// 	(@userID, @name, @model, @prompt, @temperature,
		//   @maxTurnCount, @contextTurnCount,
		//   @middlewareJson, @public,
		//   NOW(), NOW())
		// RETURNING "id"
		CreateBot(ctx context.Context, userID uint64,
			name, model, prompt string,
			temperature float32,
			maxTurnCount,
			contextTurnCount int,
			middlewareJson JSONB, public bool,
		) (uint64, error)

		// UPDATE @@table
		// 	{{set}}
		// 		"name"=@name,
		// 		"model"=@model,
		// 		"prompt"=@prompt,
		// 		"temperature"=@temperature,
		// 		"max_turn_count"=@maxTurnCount,
		// 		"context_turn_count"=@contextTurnCount,
		// 		"middleware_json"=@middlewareJson,
		//    "public"=@public,
		// 		"updated_at"=NOW()
		// 	{{end}}
		// WHERE
		// 	"id"=@id AND "deleted_at" is NULL
		UpdateBot(ctx context.Context, id uint64,
			name, model, prompt string,
			temperature float32,
			maxTurnCount,
			contextTurnCount int,
			middlewareJson JSONB,
			public bool,
		) error

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
		CreateBot(ctx context.Context, userID uint64, name, model, prompt string, temperature float32, maxTurnCount, contextTurnCount int, middlewares MiddlewareConfig, public bool) (*Bot, error)
		UpdateBot(ctx context.Context, id uint64, name, model, prompt string, temperature float32, maxTurnCount, contextTurnCount int, middlewares MiddlewareConfig, public bool) error
	}
)

func (t *Bot) DecodeMiddlewares() error {
	if t.MiddlewareJson == nil {
		return nil
	}

	val, err := t.MiddlewareJson.Value()
	if err != nil {
		return err
	}

	return json.Unmarshal(val.([]byte), &t.Middlewares)
}

func (t *Bot) GetPrompt(conv *Conversation, question string) string {
	var buf bytes.Buffer
	data := map[string]interface{}{
		"LangHint": conv.LangHint(),
		"History":  conv.HistoryToText(),
	}
	if t.PromptTpl == nil {
		t.PromptTpl = template.Must(template.New(fmt.Sprintf("%d-prompt-tmpl", t.ID)).Parse(t.Prompt))
	}
	t.PromptTpl.Execute(&buf, data)

	str := buf.String()

	return strings.TrimSpace(str) + "\n"
}

func (t *Bot) GetChatMessages(conv *Conversation, additionData map[string]interface{}) []gogpt.ChatCompletionMessage {
	var buf bytes.Buffer
	data := map[string]interface{}{
		"LangHint": conv.LangHint(),
	}

	for k, v := range additionData {
		data[k] = v
	}

	if t.PromptTpl == nil {
		t.PromptTpl = template.Must(template.New(fmt.Sprintf("%d-prompt-tmpl", t.ID)).Parse(t.Prompt))
	}

	t.PromptTpl.Execute(&buf, data)

	str := buf.String()

	result := []gogpt.ChatCompletionMessage{
		{
			Role:    "system",
			Content: str,
		},
	}

	history := conv.History
	if len(history) > conv.Bot.ContextTurnCount {
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

	return result
}
