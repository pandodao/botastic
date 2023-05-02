package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/pandodao/botastic/api"
	"gorm.io/gorm"
)

type Turn struct {
	gorm.Model
	ConvID            uuid.UUID `gorm:"index"`
	BotID             uint      `gorm:"index"`
	Request           string    `gorm:"type:text"`
	Response          string    `gorm:"type:text"`
	PromptTokens      int
	CompletionTokens  int
	TotalTokens       int
	Status            api.TurnStatus    `gorm:"index"`
	MiddlewareResults MiddlewareResults `gorm:"type:json"`
	ErrorCode         int
	ErrorMessage      string `gorm:"type:text"`
}

func (t Turn) API() api.Turn {
	return api.Turn{
		ID:                t.ID,
		ConvID:            t.ConvID,
		BotID:             t.BotID,
		Request:           t.Request,
		Response:          t.Response,
		PromptTokens:      t.PromptTokens,
		CompletionTokens:  t.CompletionTokens,
		TotalTokens:       t.TotalTokens,
		Status:            t.Status,
		MiddlewareResults: api.MiddlewareResults(t.MiddlewareResults),
		ErrorCode:         t.ErrorCode,
		ErrorMessage:      t.ErrorMessage,
		CreatedAt:         t.CreatedAt,
		UpdatedAt:         t.UpdatedAt,
	}
}

type MiddlewareResults api.MiddlewareResults

func (mr *MiddlewareResults) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("type assertion to []byte failed:", value))
	}

	return json.Unmarshal(data, mr)
}

func (mr MiddlewareResults) Value() (driver.Value, error) {
	return json.Marshal(mr)
}
