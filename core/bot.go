package core

import (
	"context"
	"text/template"
	"time"
)

type (
	Bot struct {
		ID        uint64             `yaml:"id" json:"id"`
		Name      string             `yaml:"name" json:"name"`
		Cmd       string             `yaml:"cmd" json:"cmd"`
		Prompt    string             `yaml:"prompt" json:"-"`
		ModelID   string             `yaml:"model_id" json:"-"`
		PromptTpl *template.Template `yaml:"-" json:"prompt_tpl"`
	}

	BotConversation struct {
		ID           string         `yaml:"id" json:"id"`
		Bot          *Bot           `yaml:"bot" json:"bot"`
		App          *App           `yaml:"app" json:"app"`
		UserIdentity string         `yaml:"user_identity" json:"user_identity"`
		Lang         string         `yaml:"lang" json:"lang"`
		History      []*BotConvTurn `yaml:"history" json:"history"`
		ExpiredAt    time.Time      `yaml:"expired_at" json:"expired_at"`
	}

	BotConvTurn struct {
		ID           uint64 `yaml:"id" json:"id"`
		BotID        uint64 `yaml:"bot_id" json:"bot_id"`
		AppID        uint64 `yaml:"app_id" json:"app_id"`
		UserIdentity string `yaml:"user_identity" json:"user_identity"`
		Request      string `yaml:"request" json:"request"`
		Response     string `yaml:"response" json:"response"`
		Status       int    `yaml:"status" json:"status"`
		CreatedAt    int64  `yaml:"created_at" json:"created_at"`
		UpdatedAt    int64  `yaml:"updated_at" json:"updated_at"`
	}

	BotService interface {
		CreateConversation(ctx context.Context, botID, appID uint64, userIdentity, lang string) (*BotConversation, error)
		ClearExpiredConversations(ctx context.Context) error
		DeleteConversation(ctx context.Context, convID string) error
		GetConversation(ctx context.Context, convID string) (*BotConversation, error)
		PostToConversation(ctx context.Context, conv *BotConversation, input string) (*BotConvTurn, error)
	}
)

func (c *BotConversation) IsExpired() bool {
	return c.ExpiredAt.Before(time.Now())
}
