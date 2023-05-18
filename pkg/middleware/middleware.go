package middleware

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pandodao/botastic/api"
	"github.com/pandodao/botastic/models"
)

const (
	generalOptionTimeoutSeconds   = "timeout_seconds"
	generalOptionTerminateIfError = "terminate_if_error"
)

type Middleware interface {
	Desc() *api.MiddlewareDesc
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
			Name:         generalOptionTerminateIfError,
			Desc:         "terminate the whole flow if error occurs",
			DefaultValue: "true",
			ParseValueFunc: func(v string) (any, error) {
				return strconv.ParseBool(v)
			},
		},
		{
			Name:         generalOptionTimeoutSeconds,
			Desc:         "middleware execution timeout in seconds",
			DefaultValue: "10",
			ParseValueFunc: func(v string) (any, error) {
				return strconv.Atoi(v)
			},
		},
	}
}

func (h *Handler) Process(ctx context.Context, mc api.MiddlewareConfig, turn *models.Turn) ([]*api.MiddlewareResult, bool) {
	rs := []*api.MiddlewareResult{}
	for _, item := range mc.Items {
		r := &api.MiddlewareResult{
			Middleware: *item,
		}
		generalOptions, options, err := h.validateMiddleware(item)
		if err != nil {
			r.Code = api.MiddlewareErrorCodeConfigInvalid
			r.Err = err.Error()
		}
		rs = append(rs, r)

		terminateIfError, ok := generalOptions[generalOptionTerminateIfError].Value.(bool)
		if !ok {
			terminateIfError = true
		}
		if terminateIfError && r.Code != 0 {
			return rs, false
		}

		result, err := func() (string, error) {
			timeoutSeconds := generalOptions[generalOptionTimeoutSeconds].Value.(int)
			ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
			defer cancel()

			return h.mm[item.Name].Process(ctx, options, turn)
		}()
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				r.Code = api.MiddlewareErrorCodeTimeout
			} else {
				r.Code = api.MiddlewareErrorCodeProcessFailed
			}
			r.Err = err.Error()
			if terminateIfError {
				return rs, false
			}
		} else {
			r.Result = result
			r.RenderName = "MIDDLEWARE_RESULT_" + item.ID
		}
	}
	return rs, true
}

func (h *Handler) ValidateConfig(mc *api.MiddlewareConfig) error {
	idMap := map[string]bool{}
	for _, item := range mc.Items {
		if item.Options == nil {
			item.Options = map[string]string{}
		}

		if _, _, err := h.validateMiddleware(item); err != nil {
			return err
		}

		if idMap[item.ID] {
			return fmt.Errorf("duplicate middleware id: %s", item.ID)
		}
		idMap[item.ID] = true
	}

	return nil
}

func (h *Handler) validateMiddleware(item *api.Middleware) (map[string]*api.MiddlewareDescOption, map[string]*api.MiddlewareDescOption, error) {
	m, ok := h.mm[item.Name]
	if !ok {
		return nil, nil, fmt.Errorf("unknown middleware: %s", item.Name)
	}

	generalOptions, err := parseOptions(item.Name, h.GeneralOptions(), item.Options)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse general options, middleware: %s, err: %w", item.Name, err)
	}

	middlewareOptions, err := parseOptions(item.Name, m.Desc().Options, item.Options)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse options, middleware: %s, err: %w", item.Name, err)
	}

	return generalOptions, middlewareOptions, nil
}

func parseOptions(name string, descOptions []*api.MiddlewareDescOption, opts map[string]string) (map[string]*api.MiddlewareDescOption, error) {
	result := map[string]*api.MiddlewareDescOption{}
	for _, opt := range descOptions {
		opts[opt.Name] = strings.TrimSpace(opts[opt.Name])
		if opt.Required && opts[opt.Name] == "" {
			return nil, fmt.Errorf("missing required option: %s, middleware: %s", opt.Name, name)
		}
		if opts[opt.Name] == "" {
			opts[opt.Name] = opt.DefaultValue
		}

		if opt.ParseValueFunc != nil {
			v, err := opt.ParseValueFunc(opts[opt.Name])
			if err != nil {
				return nil, fmt.Errorf("failed to parse option: %s, middleware: %s, err: %w", opt.Name, name, err)
			}
			opt.Value = v
		}
		result[opt.Name] = opt
	}

	return result, nil
}
