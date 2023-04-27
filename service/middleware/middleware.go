package middleware

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/pandodao/botastic/core"
)

type generalOptions struct {
	Required bool          `json:"required"`
	Timeout  time.Duration `json:"timeout"`
}

func New() *service {
	return &service{
		middlewareMap: map[string]core.MiddlewareDesc{},
	}
}

type (
	service struct {
		middlewareMap map[string]core.MiddlewareDesc
	}
)

func (s *service) Register(ms ...core.MiddlewareDesc) {
	for _, m := range ms {
		s.middlewareMap[m.Name()] = m
	}
}

func (s *service) ProcessByConfig(ctx context.Context, mc core.MiddlewareConfig, turn *core.ConvTurn) core.MiddlewareResults {
	var results []*core.MiddlewareProcessResult
	for _, m := range mc.Items {
		result := s.Process(ctx, m, turn)
		results = append(results, result)
		if result.Required {
			break
		}
	}
	return results
}

func (s *service) Process(ctx context.Context, m *core.Middleware, turn *core.ConvTurn) *core.MiddlewareProcessResult {
	gopts, err := parseGeneralOptions(ctx, m.Options)
	if err != nil {
		return &core.MiddlewareProcessResult{
			Name:     m.Name,
			Code:     core.MiddlewareProcessCodeInvalidOptions,
			Err:      err.Error(),
			Required: true,
		}
	}

	middleware := s.middlewareMap[m.Name]
	if middleware == nil {
		return &core.MiddlewareProcessResult{
			Name:     m.Name,
			Code:     core.MiddlewareProcessCodeUnknown,
			Err:      "middleware not found",
			Required: gopts.Required,
		}
	}

	opts, err := middleware.ValidateOptions(m.Options)
	if err != nil {
		return &core.MiddlewareProcessResult{
			Name:     m.Name,
			Code:     core.MiddlewareProcessCodeInvalidOptions,
			Err:      err.Error(),
			Required: gopts.Required,
		}
	}

	if gopts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, gopts.Timeout)
		defer cancel()
	}

	result, err := middleware.Process(ctx, opts, turn)
	if err != nil {
		code := core.MiddlewareProcessCodeInternalError
		if errors.Is(err, context.DeadlineExceeded) {
			code = core.MiddlewareProcessCodeTimeout
		}

		return &core.MiddlewareProcessResult{
			Name:     m.Name,
			Code:     code,
			Err:      err.Error(),
			Required: gopts.Required,
		}
	}

	return &core.MiddlewareProcessResult{
		Opts:   m.Options,
		Name:   m.Name,
		Code:   core.MiddlewareProcessCodeOK,
		Result: result,
	}
}

func parseGeneralOptions(ctx context.Context, opts map[string]any) (*generalOptions, error) {
	generalOptions := &generalOptions{}

	if val, ok := opts["required"]; ok {
		b, ok := val.(bool)
		if !ok {
			return nil, fmt.Errorf("required should be bool: %v", val)
		}
		generalOptions.Required = b
	}

	if val, ok := opts["timeout"]; ok {
		s, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("timeout should be string: %v", val)
		}
		d, err := time.ParseDuration(s)
		if err != nil {
			return nil, fmt.Errorf("timeout should be valid duration: %v", val)
		}
		generalOptions.Timeout = d
	}

	return generalOptions, nil
}
