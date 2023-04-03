package core

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	ConvTurnStatusInit = iota
	ConvTurnStatusPending
	ConvTurnStatusCompleted
	ConvTurnStatusError
)

type BotOverride struct {
	Temperature *float32 `json:"temperature,omitempty"`
}

func (b *BotOverride) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("type assertion to []byte failed:", value))
	}

	return json.Unmarshal(data, b)
}

func (b BotOverride) Value() (driver.Value, error) {
	data, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(data), nil
}

type (
	Conversation struct {
		ID           string      `json:"id"`
		Bot          *Bot        `json:"bot"`
		App          *App        `json:"app"`
		UserIdentity string      `json:"user_identity"`
		Lang         string      `json:"lang"`
		History      []*ConvTurn `json:"history"`
		ExpiredAt    time.Time   `json:"expired_at"`
	}

	ConvTurn struct {
		ID               uint64      `json:"id"`
		ConversationID   string      `json:"conversation_id"`
		BotID            uint64      `json:"bot_id"`
		AppID            uint64      `json:"app_id"`
		UserID           uint64      `json:"user_id"`
		UserIdentity     string      `json:"user_identity"`
		Request          string      `json:"request"`
		Response         string      `json:"response"`
		PromptTokens     int         `json:"prompt_tokens"`
		CompletionTokens int         `json:"completion_tokens"`
		TotalTokens      int         `json:"total_tokens"`
		Status           int         `json:"status"`
		BotOverride      BotOverride `gorm:"type:jsonb"  json:"bot_override"`
		CreatedAt        *time.Time  `json:"created_at"`
		UpdatedAt        *time.Time  `json:"updated_at"`
	}

	ConversationStore interface {
		// INSERT INTO "conv_turns"
		// 	(
		//	"conversation_id", "bot_id", "app_id", "user_id",
		//  "user_identity",
		//  "request", "response", "status", "bot_override",
		//  "created_at", "updated_at"
		//   )
		// VALUES
		// 	(
		//   @convID, @botID, @appID, @userID,
		//   @uid,
		//   @request, '', 0, @bo,
		//   NOW(), NOW()
		//  )
		// RETURNING "id"
		CreateConvTurn(ctx context.Context, convID string, botID, appID, userID uint64, uid, request string, bo BotOverride) (uint64, error)

		// SELECT
		// 	"id",
		//	"conversation_id", "bot_id", "app_id", "user_id",
		//  "user_identity",
		//  "request", "response", "status",
		//  "prompt_tokens", "completion_tokens", "total_tokens", "bot_override",
		//  "created_at", "updated_at"
		// FROM "conv_turns" WHERE
		//  "id" IN (@ids)
		GetConvTurns(ctx context.Context, ids []uint64) ([]*ConvTurn, error)

		// SELECT
		// 	"id",
		//	"conversation_id", "bot_id", "app_id", "user_id",
		//  "user_identity",
		//  "request", "response", "status",
		//  "prompt_tokens", "completion_tokens", "total_tokens", "bot_override",
		//  "created_at", "updated_at"
		// FROM "conv_turns" WHERE
		//  "id" = @id
		GetConvTurn(ctx context.Context, id uint64) (*ConvTurn, error)

		// SELECT
		// 	"id",
		//	"conversation_id", "bot_id", "app_id", "user_id",
		//  "user_identity",
		//  "request", "response", "status",
		//  "prompt_tokens", "completion_tokens", "total_tokens", "bot_override",
		//  "created_at", "updated_at"
		// FROM "conv_turns"
		// {{where}}
		// "status" IN (@status)
		//    {{if len(excludeIDs)>0}}
		//      AND "id" NOT IN (@excludeIDs)
		//    {{end}}
		// {{end}}
		GetConvTurnsByStatus(ctx context.Context, excludeIDs []uint64, status []int) ([]*ConvTurn, error)

		// UPDATE "conv_turns"
		// 	{{set}}
		// 		"response"=@response,
		//    "prompt_tokens"=@promptTokens,
		//    "completion_tokens"=@completionTokens,
		// 		"total_tokens"=@totalTokens,
		// 		"status"=@status,
		// 		"updated_at"=NOW()
		// 	{{end}}
		// WHERE
		// 	"id"=@id
		UpdateConvTurn(ctx context.Context, id uint64, response string, promptTokens, completionTokens, totalTokens int64, status int) error
	}

	ConversationService interface {
		CreateConversation(ctx context.Context, botID, appID uint64, userIdentity, lang string) (*Conversation, error)
		ClearExpiredConversations(ctx context.Context) error
		DeleteConversation(ctx context.Context, convID string) error
		GetConversation(ctx context.Context, convID string) (*Conversation, error)
		PostToConversation(ctx context.Context, conv *Conversation, input string, bo BotOverride) (*ConvTurn, error)
		ReplaceStore(store ConversationStore) ConversationService
	}
)

func (c *Conversation) IsExpired() bool {
	return c.ExpiredAt.Before(time.Now())
}

func (c *Conversation) HistoryToText() string {
	lines := make([]string, 0)
	history := c.History
	if len(history) > c.Bot.ContextTurnCount {
		history = history[len(history)-c.Bot.ContextTurnCount:]
	}
	for _, turn := range history {
		lines = append(lines, fmt.Sprintf("Q: %s", turn.Request))
		if turn.Response != "" {
			lines = append(lines, fmt.Sprintf("A: %s", turn.Response))
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func (c *Conversation) LangHint() string {
	langHint := "If no language is explicitly specified, please respond in %s."
	lang := "Chinese"
	switch c.Lang {
	case "en":
		lang = "English"
	case "ja":
		lang = "Japanese"
	case "zh":
		lang = "Chinese"
	}

	return fmt.Sprintf(langHint, lang)
}

func (c *Conversation) GetKey() string {
	return fmt.Sprintf("%d:%s", c.App.ID, c.ID)
}

func (c *Conversation) GenerateUserText(text string) string {
	if text != "" {
		return fmt.Sprintf("Q: %s", text)
	}
	return ""
}

func (t ConvTurn) IsProcessed() bool {
	return t.Status == ConvTurnStatusCompleted || t.Status == ConvTurnStatusError
}
