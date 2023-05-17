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
	Error             *TurnError        `gorm:"type:json"`
}

func (t Turn) API() api.Turn {
	r := api.Turn{
		ID:                t.ID,
		ConversationID:    t.ConvID,
		BotID:             t.BotID,
		Request:           t.Request,
		Response:          t.Response,
		PromptTokens:      t.PromptTokens,
		CompletionTokens:  t.CompletionTokens,
		TotalTokens:       t.TotalTokens,
		Status:            t.Status,
		MiddlewareResults: []*api.MiddlewareResult(t.MiddlewareResults),
		CreatedAt:         t.CreatedAt,
		UpdatedAt:         t.UpdatedAt,
	}
	if t.Error != nil {
		v := api.TurnError(*t.Error)
		r.Error = &v
	}
	return r
}

func (t Turn) IsProcessed() bool {
	return t.Status == api.TurnStatusSuccess || t.Status == api.TurnStatusFailed
}

type MiddlewareResults []*api.MiddlewareResult

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

type TurnError api.TurnError

func (e *TurnError) Error() string {
	if e.Msg != "" {
		return e.Msg
	}

	return e.Code.String()
}

func NewTurnError(code api.TurnErrorCode, msg ...string) *TurnError {
	e := &TurnError{
		Code: code,
		Msg:  code.String(),
	}
	if len(msg) > 0 {
		e.Msg = msg[0]
	}

	return e
}

func (te *TurnError) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("type assertion to []byte failed:", value))
	}
	if len(data) == 0 {
		return nil
	}

	return json.Unmarshal(data, te)
}

func (te TurnError) Value() (driver.Value, error) {
	return json.Marshal(te)
}
