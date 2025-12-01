// nolint:revive
package util

import (
	"context"
	"errors"
	"time"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const (
	// CTXKeyUser is the key for the user object in the context.
	CTXKeyUser contextKey = "user"
	// CTXKeyAccessToken is the key for the access token in the context.
	CTXKeyAccessToken contextKey = "access_token"
	// CTXKeyCacheControl is the key for cache control directives in the context.
	CTXKeyCacheControl contextKey = "cache_control"
	// CTXKeyRequestID is the key for the request ID in the context.
	CTXKeyRequestID contextKey = "request_id"
	// CTXKeyDisableLogger is the key for the disable logger flag in the context.
	CTXKeyDisableLogger contextKey = "disable_logger"
)

//nolint:containedctx
type detachedContext struct {
	parent context.Context
}

// Deadline returns the zero time and false, indicating no deadline.
func (c detachedContext) Deadline() (time.Time, bool) { return time.Time{}, false }

// Done returns nil, indicating the context is never canceled.
func (c detachedContext) Done() <-chan struct{} { return nil }

// Err returns nil, indicating no error.
func (c detachedContext) Err() error { return nil }

// Value returns the value associated with the key from the parent context.
func (c detachedContext) Value(key interface{}) interface{} { return c.parent.Value(key) }

// DetachContext creates a new context that wraps the parent but ignores cancellation signals (Deadline, Done, Err).
// It retains the values from the parent context.
//
// This is useful for passing context information to goroutines that should outlive the request context,
// but use this sparingly.
//
// Parameters:
//   - ctx: The parent context to detach from.
//
// Returns:
//   - context.Context: A new context that is not cancelable but carries the parent's values.
func DetachContext(ctx context.Context) context.Context {
	return detachedContext{ctx}
}

// RequestIDFromContext retrieves the request ID from the context.
//
// Parameters:
//   - ctx: The context to retrieve the ID from.
//
// Returns:
//   - string: The request ID.
//   - error: An error if the ID is missing or not a string.
func RequestIDFromContext(ctx context.Context) (string, error) {
	val := ctx.Value(CTXKeyRequestID)
	if val == nil {
		return "", errors.New("no request id present in context")
	}

	id, ok := val.(string)
	if !ok {
		return "", errors.New("request id in context is not a string")
	}

	return id, nil
}

// ShouldDisableLogger determines if logging should be disabled for the given context.
//
// It checks for the presence of the CTXKeyDisableLogger key in the context.
//
// Parameters:
//   - ctx: The context to check.
//
// Returns:
//   - bool: True if logging should be disabled, false otherwise.
func ShouldDisableLogger(ctx context.Context) bool {
	s := ctx.Value(CTXKeyDisableLogger)
	if s == nil {
		return false
	}

	shouldDisable, ok := s.(bool)
	if !ok {
		return false
	}

	return shouldDisable
}

// DisableLogger sets a flag in the context to disable or enable logging.
//
// Parameters:
//   - ctx: The parent context.
//   - shouldDisable: True to disable logging, false to enable.
//
// Returns:
//   - context.Context: A new context with the updated flag.
func DisableLogger(ctx context.Context, shouldDisable bool) context.Context {
	return context.WithValue(ctx, CTXKeyDisableLogger, shouldDisable)
}
