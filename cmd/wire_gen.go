// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package cmd

import (
	"github.com/pandodao/botastic/config"
	"github.com/pandodao/botastic/internal/httpd"
	"github.com/pandodao/botastic/internal/llms"
	"github.com/pandodao/botastic/internal/starter"
	"github.com/pandodao/botastic/pkg/chanhub"
	"github.com/pandodao/botastic/pkg/middleware"
	"github.com/pandodao/botastic/state"
	"github.com/pandodao/botastic/storage"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Injectors from wire.go:

func provideHttpdStarter(cfgFile2 string) (starter.Starter, error) {
	configConfig, err := config.Init(cfgFile2)
	if err != nil {
		return nil, err
	}
	httpdConfig := configConfig.Httpd
	dbConfig := configConfig.DB
	handler, err := storage.Init(dbConfig)
	if err != nil {
		return nil, err
	}
	llMsConfig := configConfig.LLMs
	llmsHandler := llms.New(llMsConfig)
	hub := chanhub.New()
	stateConfig := configConfig.State
	logConfig := configConfig.Log
	logger, err := provideLogger(logConfig)
	if err != nil {
		return nil, err
	}
	stateHandler := state.New(stateConfig, logger, handler, llmsHandler, hub)
	fetch := middleware.NewFetch()
	v := provideMiddlewares(fetch)
	middlewareHandler := middleware.New(v...)
	httpdHandler := httpd.NewHandler(handler, llmsHandler, hub, stateHandler, logger, middlewareHandler)
	server := httpd.New(httpdConfig, httpdHandler, logger)
	v2 := provideStarters(server, stateHandler)
	starterStarter := starter.Multi(v2...)
	return starterStarter, nil
}

// wire.go:

func provideStarters(s1 *httpd.Server, s2 *state.Handler) []starter.Starter {
	return []starter.Starter{s1, s2}
}

func provideLogger(cfg config.LogConfig) (*zap.Logger, error) {
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}
	zapCfg := zap.NewProductionConfig()
	zapCfg.Level = zap.NewAtomicLevelAt(level)
	return zapCfg.Build()
}

func provideMiddlewares(m1 *middleware.Fetch) []middleware.Middleware {
	return []middleware.Middleware{m1}
}