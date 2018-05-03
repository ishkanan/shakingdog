package auth

import (
	"context"
	"strings"
	
	"golang.org/x/oauth2"
)

// unexported key prevents type collisions
type key int

const (
	tokenKey key = iota
	usernameKey
	groupsKey
	authErrorKey
)

// WithToken returns a new Context with the token attached
func WithToken(ctx context.Context, val *oauth2.Token) context.Context {
	return context.WithValue(ctx, tokenKey, val)
}

// TokenFromContext extracts an oauth2.Token from the context if it's present
// Returns nil if it's not present.
func TokenFromContext(ctx context.Context) *oauth2.Token {
	if val, ok := ctx.Value(tokenKey).(*oauth2.Token); ok {
		return val
	}
	return nil
}

// WithUsername adds the username to the context for handlers to use
func WithUsername(ctx context.Context, val string) context.Context {
	return context.WithValue(ctx, usernameKey, val)
}

// UsernameFromContext extracts the username from the context if it's present
// Returns "" if it's not present.
func UsernameFromContext(ctx context.Context) string {
	if val, ok := ctx.Value(usernameKey).(string); ok {
		return val
	}
	return ""
}

// WithError attaches auth errors to the context so a handler can inspect them
func WithError(ctx context.Context, val error) context.Context {
	return context.WithValue(ctx, authErrorKey, val)
}

// ErrorFromContext extracts any auth errors from the context if it's present
// Returns nil if it's not present.
func ErrorFromContext(ctx context.Context) error {
	if val, ok := ctx.Value(authErrorKey).(error); ok {
		return val
	}
	return nil
}

// WithGroups attaches a list of user groups to the context so the app can use
// them for access control
func WithGroups(ctx context.Context, val []string) context.Context {
	return context.WithValue(ctx, groupsKey, val)
}

// GroupsFromContext extracts Okta Groups from the context if it's present
// Returns nil if it's not present.
func GroupsFromContext(ctx context.Context) []string {
	if groups, ok := ctx.Value(groupsKey).([]string); ok {
		// convert to lower case for easier comparison
		for index, group := range groups {
			groups[index] = strings.ToLower(group)
		}
		return groups
	}
	return nil
}
