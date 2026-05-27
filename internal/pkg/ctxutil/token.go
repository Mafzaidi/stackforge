package ctxutil

import (
	"context"

	"github.com/mafzaidi/stackforge/internal/domain/service"
)

type contextKey string

const (
	tokenKey contextKey = "auth_token"
	userKey  contextKey = "authorizer_user"
)

// WithToken returns a new context with the JWT token stored.
func WithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenKey, token)
}

// TokenFromContext extracts the JWT token from context.
func TokenFromContext(ctx context.Context) string {
	token, ok := ctx.Value(tokenKey).(string)
	if !ok {
		return ""
	}
	return token
}

// WithUser returns a new context with the AuthorizerUser stored.
func WithUser(ctx context.Context, user *service.AuthorizerUser) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// UserFromContext extracts the AuthorizerUser from context.
func UserFromContext(ctx context.Context) *service.AuthorizerUser {
	user, ok := ctx.Value(userKey).(*service.AuthorizerUser)
	if !ok {
		return nil
	}
	return user
}
