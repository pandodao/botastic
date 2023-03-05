package core

import (
	"bytes"
	"context"
	"strings"
	"text/template"

	gogpt "github.com/sashabaranov/go-gpt3"
)

type (
	Bot struct {
		ID          uint64             `yaml:"id" json:"id"`
		Name        string             `yaml:"name" json:"name"`
		Prompt      string             `yaml:"prompt" json:"-"`
		Model       string             `yaml:"model" json:"-"`
		Middlewares []*Middleware      `yaml:"middlewares" json:"-"`
		Temperature float32            `yaml:"temperature" json:"-"`
		PromptTpl   *template.Template `yaml:"-" json:"-"`
	}

	BotService interface {
		GetBot(ctx context.Context, id uint64) (*Bot, error)
	}
)

func (t *Bot) GetPrompt(conv *Conversation, question string) string {
	var buf bytes.Buffer
	data := map[string]interface{}{
		"LangHint":    conv.LangHint(),
		"History":     conv.HistoryToText(),
		"WithHistory": true,
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

	t.PromptTpl.Execute(&buf, data)

	str := buf.String()

	result := []gogpt.ChatCompletionMessage{
		{
			Role:    "system",
			Content: str,
		},
	}
	for _, turn := range conv.History {
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
