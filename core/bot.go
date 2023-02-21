package core

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"
)

type (
	Bot struct {
		ID        uint64             `yaml:"id" json:"id"`
		Name      string             `yaml:"name" json:"name"`
		Prompt    string             `yaml:"prompt" json:"-"`
		Model     string             `yaml:"model" json:"-"`
		PromptTpl *template.Template `yaml:"-" json:"-"`
	}

	BotService interface {
		GetBot(ctx context.Context, id uint64) (*Bot, error)
	}
)

func (t *Bot) GetPrompt(conv *Conversation, question string) string {
	langHint := "If no language is explicitly specified, please respond in %s."
	lang := "Chinese"
	switch conv.Lang {
	case "en":
		lang = "English"
	case "ja":
		lang = "Japanese"
	case "zh":
		lang = "Chinese"
	}
	langHint = fmt.Sprintf(langHint, lang)

	var buf bytes.Buffer
	data := map[string]interface{}{
		"LangHint": langHint,
		"History":  conv.HistoryToText(),
	}
	t.PromptTpl.Execute(&buf, data)

	str := buf.String()

	return strings.TrimSpace(str) + "\n"
}
