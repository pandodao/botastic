package core

import (
	"context"

	"github.com/shopspring/decimal"
)

const (
	ModelProviderOpenAI = "openai"
)

const (
	ModelOpenAIGPT4               = "openai:gpt-4"
	ModelOpenAIGPT3Dot5Turbo      = "openai:gpt-3.5-turbo"
	ModelOpenAIGPT3TextDavinci003 = "openai:text-davinci-003"
	ModelOpenAIAdaEmbeddingV2     = "openai:text-embedding-ada-002"
)

type (
	Model struct {
		Provider           string          `yaml:"provider"`
		ProviderModel      string          `yaml:"provider_model"`
		MaxToken           int             `yaml:"max_token"`
		PromptPriceUSD     decimal.Decimal `yaml:"prompt_price_usd"`
		CompletionPriceUSD decimal.Decimal `yaml:"completion_price_usd"`
		PriceUSD           decimal.Decimal `yaml:"price_usd"`

		Props struct {
			IsOpenAIChatModel       bool `yaml:"is_openai_chat_model"`
			IsOpenAICompletionModel bool `yaml:"is_openai_completion_model"`
			IsOpenAIEmbeddingModel  bool `yaml:"is_openai_embedding_model"`
		}
	}

	ModelStore interface {
		GetModel(ctx context.Context, name string) (*Model, error)
	}
)

func (m *Model) CalculateTokenCost(promptCount, completionCount int64) decimal.Decimal {
	pc := decimal.NewFromInt(promptCount)
	cc := decimal.NewFromInt(completionCount)

	if m.PriceUSD.IsPositive() {
		return m.PriceUSD.Mul(pc.Add(cc))
	}
	if m.PromptPriceUSD.IsPositive() && m.CompletionPriceUSD.IsPositive() {
		return m.PromptPriceUSD.Mul(pc).Add(m.CompletionPriceUSD.Mul(cc))
	}
	return decimal.Zero
}
