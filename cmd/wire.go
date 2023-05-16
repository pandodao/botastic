//go:build wireinject

package cmd

import (
	"github.com/google/wire"
	"github.com/pandodao/botastic/config"
	"github.com/pandodao/botastic/internal/httpd"
	"github.com/pandodao/botastic/internal/llms"
	"github.com/pandodao/botastic/internal/starter"
	"github.com/pandodao/botastic/pkg/chanhub"
	"github.com/pandodao/botastic/state"
	"github.com/pandodao/botastic/storage"
)

func provideHttpdStarter() (starter.Starter, error) {
	panic(wire.Build(
		wire.NewSet(
			wire.Value(cfgFile),
			config.Init,
			wire.FieldsOf(new(*config.Config), "Httpd", "DB", "LLMS", "State"),
		),
		wire.NewSet(storage.Init),
		wire.NewSet(llms.New),
		wire.NewSet(chanhub.New),
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
