package middleware

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/pandodao/botastic/core"
)

type Middleware interface {
	Name() string
	ValidateOptions(opts map[string]any) (any, error)
	Process(ctx context.Context, opts any, input string) (string, error)
}

type generalOptions struct {
	Required bool          `json:"required"`
	Timeout  time.Duration `json:"timeout"`
}

func New(
	cfg Config,
	apps core.AppStore,
	indexz core.IndexService,
) *service {

	middlewareMap := map[string]Middleware{}
	for _, m := range []Middleware{
		&botasticSearch{
			apps:   apps,
			indexz: indexz,
		},
		&duckDuckGoSearch{},
		&intentRecognition{},
		&fetch{},
	} {
		middlewareMap[m.Name()] = m
	}

	return &service{
		cfg:           cfg,
		apps:          apps,
		indexz:        indexz,
		middlewareMap: middlewareMap,
	}
}

type (
	Config struct {
	}

	service struct {
		cfg    Config
		apps   core.AppStore
		indexz core.IndexService

		middlewareMap map[string]Middleware
	}
)

func (s *service) ProcessByConfig(ctx context.Context, mc core.MiddlewareConfig, input string) core.MiddlewareResults {
	var results []*core.MiddlewareProcessResult
	for _, m := range mc.Items {
		result := s.Process(ctx, m, input)
		results = append(results, result)
		if result.Required {
			break
		}
	}
	return results
}

func (s *service) Process(ctx context.Context, m *core.Middleware, input string) *core.MiddlewareProcessResult {
	gopts, err := parseGeneralOptions(ctx, m.Options)
	if err != nil {
		return &core.MiddlewareProcessResult{
			Name:     m.Name,
			Code:     core.MiddlewareProcessCodeInvalidOptions,
			Err:      err,
			Required: true,
		}
	}

	middleware := s.middlewareMap[m.Name]
	if middleware == nil {
		return &core.MiddlewareProcessResult{
			Name:     m.Name,
			Code:     core.MiddlewareProcessCodeUnknown,
			Err:      errors.New("middleware not found"),
			Required: gopts.Required,
		}
	}

	opts, err := middleware.ValidateOptions(m.Options)
	if err != nil {
		return &core.MiddlewareProcessResult{
			Name:     m.Name,
			Code:     core.MiddlewareProcessCodeInvalidOptions,
			Err:      err,
			Required: gopts.Required,
		}
	}

	if gopts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, gopts.Timeout)
		defer cancel()
	}

	result, err := middleware.Process(ctx, opts, input)
	if err != nil {
		code := core.MiddlewareProcessCodeInternalError
		if errors.Is(err, context.DeadlineExceeded) {
			code = core.MiddlewareProcessCodeTimeout
		}

		return &core.MiddlewareProcessResult{
			Name:     m.Name,
			Code:     code,
			Err:      err,
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
