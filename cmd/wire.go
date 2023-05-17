//go:build wireinject

package cmd

import (
	"github.com/google/wire"
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

func provideHttpdStarter(cfgFile string) (starter.Starter, error) {
	panic(wire.Build(
		provideLogger,
		wire.NewSet(
			config.Init,
			wire.FieldsOf(new(*config.Config), "Log", "Httpd", "DB", "LLMs", "State"),
		),
		wire.NewSet(storage.Init),
		wire.NewSet(llms.New),
		wire.NewSet(chanhub.New),
		wire.NewSet(
			middleware.NewFetch,
			provideMiddlewares,
			middleware.New,
			wire.Bind(new(httpd.MiddlewareHandler), new(*middleware.Handler)),
		),
		wire.NewSet(
			httpd.New,
			httpd.NewHandler,
		),
		wire.NewSet(
			state.New,
			wire.Bind(new(httpd.TurnTransmitter), new(*state.Handler)),
		),
		wire.NewSet(
			provideStarters,
			starter.Multi,
		),
	))
}

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
