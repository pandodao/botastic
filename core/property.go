package core

import (
	"context"
	"strconv"
	"time"
)

type (
	Property struct {
		Key       string        `json:"key"`
		Value     PropertyValue `json:"value"`
		UpdatedAt *time.Time    `json:"updated_at"`
	}

	PropertyValue string

	PropertyStore interface {
		Get(ctx context.Context, key string) (PropertyValue, error)
		Set(ctx context.Context, key string, value interface{}) error
	}
)

func (pv PropertyValue) Time() time.Time {
	t, _ := time.Parse(time.RFC3339Nano, string(pv))
	return t
}

func (pv PropertyValue) String() string {
	return string(pv)
}
func (pv PropertyValue) Int64() (int64, error) {
	return strconv.ParseInt(string(pv), 10, 64)
}
