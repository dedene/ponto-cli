package ctx

import (
	"context"
	"time"
)

// Context keys for passing data through context.
type contextKey string

const (
	profileKey contextKey = "profile"
	timeoutKey contextKey = "timeout"
	noRetryKey contextKey = "noRetry"
)

// WithProfile adds the profile to the context.
func WithProfile(ctx context.Context, profile string) context.Context {
	return context.WithValue(ctx, profileKey, profile)
}

// ProfileFrom retrieves the profile from the context.
func ProfileFrom(ctx context.Context) string {
	if v, ok := ctx.Value(profileKey).(string); ok {
		return v
	}

	return "default"
}

// WithTimeout adds the timeout to the context.
func WithTimeout(ctx context.Context, timeout time.Duration) context.Context {
	return context.WithValue(ctx, timeoutKey, timeout)
}

// TimeoutFrom retrieves the timeout from the context.
func TimeoutFrom(ctx context.Context) time.Duration {
	if v, ok := ctx.Value(timeoutKey).(time.Duration); ok {
		return v
	}

	return 30 * time.Second
}

// WithNoRetry adds the noRetry flag to the context.
func WithNoRetry(ctx context.Context, noRetry bool) context.Context {
	return context.WithValue(ctx, noRetryKey, noRetry)
}

// NoRetryFrom retrieves the noRetry flag from the context.
func NoRetryFrom(ctx context.Context) bool {
	if v, ok := ctx.Value(noRetryKey).(bool); ok {
		return v
	}

	return false
}
