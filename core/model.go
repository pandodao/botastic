package core

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

const (
	ModelProviderOpenAI = "openai"
	ModelProviderCustom = "custom"

	ModelFunctionChat      = "chat"
	ModelFunctionEmbedding = "embedding"
)

type CustomConfig struct {
	Request *struct {
		URL     string            `json:"url"`
		Method  string            `json:"method"`
		Headers map[string]string `json:"headers"`
		Data    map[string]any    `json:"data"`
	} `json:"request,omitempty"`
	Response *struct {
		Path string `json:"path"`
	} `json:"response,omitempty"`
}

func (c *CustomConfig) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("type assertion to []byte failed:", value))
	}

	return json.Unmarshal(data, c)
}

func (b CustomConfig) Value() (driver.Value, error) {
	return json.Marshal(b)
}

type (
	Model struct {
		ID                 uint64          `json:"id"`
		Provider           string          `json:"provider"`
		ProviderModel      string          `json:"provider_model"`
		MaxToken           int             `json:"max_token"`
		PromptPriceUSD     decimal.Decimal `json:"prompt_price_usd"`
		CompletionPriceUSD decimal.Decimal `json:"completion_price_usd"`
		PriceUSD           decimal.Decimal `json:"price_usd"`
		CustomConfig       CustomConfig    `gorm:"type:jsonb;" json:"custom_config,omitempty"`
		Function           string          `json:"function"`

		CreatedAt time.Time  `json:"-"`
		DeletedAt *time.Time `json:"-"`
	}

	ModelStore interface {

		// SELECT *
		// FROM @@table WHERE
		// 	"deleted_at" IS NULL AND CONCAT(provider, ':', provider_model) = @name;
		GetModel(ctx context.Context, name string) (*Model, error)

		// SELECT *
		// FROM @@table WHERE
		// 	"deleted_at" IS NULL
		//  {{if f !=""}}
		//      AND function=@f
		//  {{end}}
		GetModelsByFunction(ctx context.Context, f string) ([]*Model, error)

		// INSERT INTO @@table
		// 	("provider", "provider_model", "max_token", "prompt_price_usd", "completion_price_usd", "price_usd", "custom_config", "function", "created_at")
		// VALUES
		// 	(@model.Provider, @model.ProviderModel, @model.MaxToken, @model.PromptPriceUSD, @model.CompletionPriceUSD, @model.PriceUSD, @model.CustomConfig, @model.Function, NOW())
		CreateModel(ctx context.Context, model *Model) error
	}
)

func (m Model) Name() string {
	return fmt.Sprintf("%s:%s", m.Provider, m.ProviderModel)
}

func (m Model) CalculateTokenCost(promptCount, completionCount int) decimal.Decimal {
	pc := decimal.NewFromInt(int64(promptCount))
	cc := decimal.NewFromInt(int64(completionCount))

	if m.PriceUSD.IsPositive() {
		return m.PriceUSD.Mul(pc.Add(cc))
	}
	if m.PromptPriceUSD.IsPositive() && m.CompletionPriceUSD.IsPositive() {
		return m.PromptPriceUSD.Mul(pc).Add(m.CompletionPriceUSD.Mul(cc))
	}
	return decimal.Zero
}

func (m Model) IsOpenAIChatModel() bool {
	if m.Provider != ModelProviderOpenAI {
		return false
	}

	switch m.ProviderModel {
	case "gpt-4", "gpt-4-32k", "gpt-3.5-turbo":
		return true
	}

	return false
}

func (m Model) IsOpenAICompletionModel() bool {
	if m.Provider != ModelProviderOpenAI {
		return false
	}
	switch m.ProviderModel {
	case "text-davinci-003":
		return true
	}

	return false
}
