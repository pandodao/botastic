package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/pandodao/botastic/api"
	"gorm.io/gorm"
)

type Bot struct {
	gorm.Model
	ChatModel        string `gorm:"type:varchar(128)"`
	Name             string `gorm:"type:varchar(128)"`
	Prompt           string `gorm:"type:text"`
	BoundaryPrompt   string `gorm:"type:text"`
	ContextTurnCount int
	Temperature      float32
	Middleware       MiddlewareConfig `gorm:"type:json"`
}

func (b Bot) API() api.Bot {
	return api.Bot{
		ID:               b.ID,
		Name:             b.Name,
		ChatModel:        b.ChatModel,
		Prompt:           b.Prompt,
		BoundaryPrompt:   b.BoundaryPrompt,
		ContextTurnCount: b.ContextTurnCount,
		Temperature:      b.Temperature,
		Middleware:       api.MiddlewareConfig(b.Middleware),
		CreatedAt:        b.CreatedAt,
		UpdatedAt:        b.UpdatedAt,
	}
}

type MiddlewareConfig api.MiddlewareConfig

func (a MiddlewareConfig) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *MiddlewareConfig) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, a)
}
