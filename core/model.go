package core

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

const (
	ModelProviderOpenAI = "openai"
)

type (
	Model struct {
		ID                 uint64          `json:"id"`
		Provider           string          `json:"provider"`
		ProviderModel      string          `json:"provider_model"`
		MaxToken           int             `json:"max_token"`
		PromptPriceUSD     decimal.Decimal `json:"prompt_price_usd"`
		CompletionPriceUSD decimal.Decimal `json:"completion_price_usd"`
		PriceUSD           decimal.Decimal `json:"price_usd"`
		CreatedAt          time.Time       `json:"-"`
		DeletedAt          *time.Time      `json:"-"`

		Props struct {
			IsOpenAIChatModel       bool `yaml:"is_openai_chat_model"`
			IsOpenAICompletionModel bool `yaml:"is_openai_completion_model"`
			IsOpenAIEmbeddingModel  bool `yaml:"is_openai_embedding_model"`
		} `gorm:"-" json:"-"`
	}

	ModelStore interface {

		// SELECT *
		// FROM @@table WHERE
		// 	"deleted_at" IS NULL AND CONCAT(provider, ':', provider_model) = @name;
		GetModel(ctx context.Context, name string) (*Model, error)

		// SELECT *
		// FROM @@table WHERE
		// 	"deleted_at" IS NULL
		GetModels(ctx context.Context) ([]*Model, error)
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
