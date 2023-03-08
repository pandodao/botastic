package session

import (
	"context"

	"github.com/pandodao/botastic/core"
)

type (
	contextKey struct{}
	appKey     struct{}
	userKey    struct{}
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

func AppFrom(ctx context.Context) *core.App {
	u, ok := ctx.Value(appKey{}).(*core.App)
	if ok && u != nil {
		return u
	}
	return nil
}

func WithUser(ctx context.Context, user *core.User) context.Context {
	return context.WithValue(ctx, userKey{}, user)
}

func UserFrom(ctx context.Context) (*core.User, bool) {
	u, ok := ctx.Value(userKey{}).(*core.User)
	return u, ok
}
