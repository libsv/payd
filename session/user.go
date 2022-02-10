package session

import (
	"context"

	"github.com/libsv/payd"
)

type userKey struct{}

// WithUser store user in request context.
func WithUser(ctx context.Context, user *payd.User) context.Context {
	return context.WithValue(ctx, userKey{}, user)
}

// MustUserFromContext return user from request context. Fail on error.
func MustUserFromContext(ctx context.Context) *payd.User {
	return ctx.Value(userKey{}).(*payd.User)
}
