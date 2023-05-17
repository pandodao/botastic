package middleware

import (
	"context"
	"fmt"

	"github.com/pandodao/botastic/api"
	"github.com/pandodao/botastic/models"
)

type Middleware interface {
	Desc() *api.MiddlewareDesc
	Parse(map[string]string) (map[string]*api.MiddlewareDescOption, error)
	Process(context.Context, map[string]*api.MiddlewareDescOption, *models.Turn) (string, error)
}

type Handler struct {
	ms []Middleware
	mm map[string]Middleware
}

func New(ms ...Middleware) *Handler {
	h := &Handler{
		ms: ms,
		mm: map[string]Middleware{},
	}

	for _, m := range h.ms {
		h.mm[m.Desc().Name] = m
	}
	return h
}

func (h *Handler) Middlewares() []*api.MiddlewareDesc {
	rs := make([]*api.MiddlewareDesc, len(h.ms))
	for i, m := range h.ms {
		rs[i] = m.Desc()
	}

	return rs
}

func (h *Handler) GeneralOptions() []*api.MiddlewareDescOption {
	return []*api.MiddlewareDescOption{
		{
			Name:         "timeout_seconds",
			Desc:         "middleware execution timeout in seconds",
			DefaultValue: "10",
		},
		{
			Name:         "terminate_if_error",
			Desc:         "terminate the whole flow if error occurs",
			DefaultValue: "false",
		},
	}
}

func (h *Handler) ValidateConfig(mc *api.MiddlewareConfig) error {
	for _, item := range mc.Items {
		m, ok := h.mm[item.Name]
		if !ok {
			return fmt.Errorf("unknown middleware: %s", item.Name)
		}

		_, err := m.Parse(item.Options)
		if err != nil {
			return fmt.Errorf("failed to parse middleware options: %s, err: %w", item.Name, err)
		}
	}

	return nil
}
