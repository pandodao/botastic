package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/pandodao/botastic/api"
)

type Index struct {
	ID         uint `gorm:"primarykey"`
	GroupKey   string
	Data       string
	Vector     Vector
	Properties IndexProperties
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (i Index) TableName() string {
	return "indexes"
}

func (i Index) API() *api.Index {
	return &api.Index{
		ID:         i.ID,
		GroupKey:   i.GroupKey,
		Data:       i.Data,
		Properties: i.Properties,
		CreatedAt:  i.CreatedAt,
		UpdatedAt:  i.UpdatedAt,
	}
}

type IndexProperties map[string]any

func (p *IndexProperties) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("type assertion to []byte failed:", value))
	}

	return json.Unmarshal(data, p)
}

func (p IndexProperties) Value() (driver.Value, error) {
	return json.Marshal(p)
}

type Vector []float32

func (v *Vector) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("type assertion to []byte failed:", value))
	}

	return json.Unmarshal(data, v)
}

func (v Vector) Value() (driver.Value, error) {
	return json.Marshal(v)
}
