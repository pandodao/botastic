package core

import (
	"context"
	"text/template"
)

type (
	Bot struct {
		ID        uint64             `yaml:"id" json:"id"`
		Name      string             `yaml:"name" json:"name"`
		Cmd       string             `yaml:"cmd" json:"cmd"`
		Prompt    string             `yaml:"prompt" json:"-"`
		ModelID   string             `yaml:"model_id" json:"-"`
		PromptTpl *template.Template `yaml:"-" json:"-"`
	}

	BotService interface {
		GetBot(ctx context.Context, id uint64) (*Bot, error)
	}
)
