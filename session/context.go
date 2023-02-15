package session

import (
	"context"

	"github.com/pandodao/botastic/core"
)

type (
	contextKey struct{}
	appKey     struct{}
)

func With(ctx context.Context, s *Session) context.Context {
	return context.WithValue(ctx, contextKey{}, s)
}

func From(ctx context.Context) *Session {
	return ctx.Value(contextKey{}).(*Session)
}

func WithApp(ctx context.Context, app *core.App) context.Context {
	return context.WithValue(ctx, appKey{}, app)
}

func AppFrom(ctx context.Context) (*core.App, bool) {
	u, ok := ctx.Value(appKey{}).(*core.App)
	return u, ok
}
