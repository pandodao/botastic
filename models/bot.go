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
	Name             string `gorm:"type:varchar(128);unique"`
	Prompt           string `gorm:"type:text"`
	BoundaryPrompt   string `gorm:"type:text"`
	ContextTurnCount int
	Temperature      float32
	TimeoutSeconds   int
	Middlewares      *MiddlewareConfig `gorm:"type:json"`
}

func (b Bot) API() api.Bot {
	r := api.Bot{
		ID:               b.ID,
		Name:             b.Name,
		ChatModel:        b.ChatModel,
		Prompt:           b.Prompt,
		BoundaryPrompt:   b.BoundaryPrompt,
		ContextTurnCount: b.ContextTurnCount,
		Temperature:      b.Temperature,
		TimeoutSeconds:   b.TimeoutSeconds,
		CreatedAt:        b.CreatedAt,
		UpdatedAt:        b.UpdatedAt,
	}
	if b.Middlewares != nil {
		v := api.MiddlewareConfig(*b.Middlewares)
		r.Middlewares = &v
	}

	return r
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
