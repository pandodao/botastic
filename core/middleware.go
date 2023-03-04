package core

import (
	"context"
)

const (
	MiddlewareBotasticSearch   = "botastic-search"
	MiddlewareDuckduckgoSearch = "duckduckgo-search"
)

type (
	Middleware struct {
		ID      uint64                 `yaml:"id" json:"id"`
		Name    string                 `yaml:"name" json:"name"`
		Options map[string]interface{} `yaml:"options" json:"options"`
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
