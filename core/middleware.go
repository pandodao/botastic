package core

import (
	"context"
	"encoding/json"
	"errors"
)

const (
	MiddlewareBotasticSearch    = "botastic-search"
	MiddlewareDuckduckgoSearch  = "duckduckgo-search"
	MiddlewareIntentRecognition = "intent-recognition"
)

const MiddlewareProcessCodeUnknown = -1

const (
	MiddlewareProcessCodeOK = iota
)

type (
	Middleware struct {
		ID      uint64            `yaml:"id" json:"id"`
		Name    string            `yaml:"name" json:"name"`
		Options MiddlewareOptions `yaml:"options" json:"options"`
	}

	MiddlewareOptions map[string]interface{}

	MiddlewareConfig struct {
		Items []*Middleware `yaml:"items" json:"items"`
	}

	MiddlewareProcessResult struct {
		Name   string `json:"name"`
		Code   uint64 `json:"code"`
		Result string `json:"result"`
	}

	MiddlewareService interface {
		GetMiddlewareByName(ctx context.Context, name string) (*Middleware, error)
		Process(ctx context.Context, m *Middleware, input string) (*MiddlewareProcessResult, error)
	}
)

func (a MiddlewareConfig) Value() ([]byte, error) {
	return json.Marshal(a)
}

func (a *MiddlewareConfig) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &a)
}

func (a MiddlewareOptions) Value() ([]byte, error) {
	return json.Marshal(a)
}

func (a *MiddlewareOptions) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &a)
}
