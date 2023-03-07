package core

import (
	"context"
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

const (
	MaxTurnCount = 8
)

type (
	Conversation struct {
		ID           string      `yaml:"id" json:"id"`
		Bot          *Bot        `yaml:"bot" json:"bot"`
		App          *App        `yaml:"app" json:"app"`
		UserIdentity string      `yaml:"user_identity" json:"user_identity"`
		Lang         string      `yaml:"lang" json:"lang"`
		History      []*ConvTurn `yaml:"history" json:"history"`
		ExpiredAt    time.Time   `yaml:"expired_at" json:"expired_at"`
	}

	ConvTurn struct {
		ID             uint64     `yaml:"id" json:"id"`
		ConversationID string     `yaml:"conversation_id" json:"conversation_id"`
		BotID          uint64     `yaml:"bot_id" json:"bot_id"`
		AppID          uint64     `yaml:"app_id" json:"app_id"`
		UserIdentity   string     `yaml:"user_identity" json:"user_identity"`
		Request        string     `yaml:"request" json:"request"`
		RequestToken   int        `yaml:"request_token" json:"request_token"`
		Response       string     `yaml:"response" json:"response"`
		ResponseToken  int        `yaml:"response_token" json:"response_token"`
		Status         int        `yaml:"status" json:"status"`
		CreatedAt      *time.Time `yaml:"created_at" json:"created_at"`
		UpdatedAt      *time.Time `yaml:"updated_at" json:"updated_at"`
	}

	ConversationStore interface {
		// INSERT INTO "conv_turns"
		// 	(
		//	"conversation_id", "bot_id", "app_id", "user_identity",
		//  "request", "request_token", "response", "status",
		//  "created_at", "updated_at"
		//   )
		// VALUES
		// 	(
		//   @convID, @botID, @appID, @uid,
		//   @request, @reqToken, '', 0,
		//   NOW(), NOW()
		//  )
		// RETURNING "id"
		CreateConvTurn(ctx context.Context, convID string, botID, appID uint64, uid, request string, reqToken int) (uint64, error)

		// SELECT
		//	"id", "conversation_id", "bot_id", "app_id", "user_identity",
		//  "request", "response", "status",
		//  "created_at", "updated_at"
		// FROM "conv_turns" WHERE
		//  "id" IN (@ids)
		GetConvTurns(ctx context.Context, ids []uint64) ([]*ConvTurn, error)

		// SELECT
		//	"id", "conversation_id", "bot_id", "app_id", "user_identity",
		//  "request", "response", "status",
		//  "created_at", "updated_at"
		// FROM "conv_turns" WHERE
		//  "id" = @id AND conversation_id = @conversationID
		GetConvTurn(ctx context.Context, conversationID string, id uint64) (*ConvTurn, error)

		// SELECT
		//	"id", "conversation_id", "bot_id", "app_id", "user_identity",
		//  "request", "response", "status",
		//  "created_at", "updated_at"
		// FROM "conv_turns" WHERE
		//  "status"=@status
		GetConvTurnsByStatus(ctx context.Context, status int) ([]*ConvTurn, error)

		// UPDATE "conv_turns"
		// 	{{set}}
		// 		"response"=@response,
		// 		"response_token"=@responseToken,
		// 		"status"=@status,
		// 		"updated_at"=NOW()
		// 	{{end}}
		// WHERE
		// 	"id"=@id
		UpdateConvTurn(ctx context.Context, id uint64, response string, responseToken int, status int) error
	}

	ConversationService interface {
		CreateConversation(ctx context.Context, botID, appID uint64, userIdentity, lang string) (*Conversation, error)
		ClearExpiredConversations(ctx context.Context) error
		DeleteConversation(ctx context.Context, convID string) error
		GetConversation(ctx context.Context, convID string) (*Conversation, error)
		PostToConversation(ctx context.Context, conv *Conversation, input string) (*ConvTurn, error)
	}
)

func (c *Conversation) IsExpired() bool {
	return c.ExpiredAt.Before(time.Now())
}

func (c *Conversation) HistoryToText() string {
	lines := make([]string, 0)
	for _, turn := range c.History {
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
