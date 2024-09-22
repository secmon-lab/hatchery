package timestamp

import (
	"context"
	"time"
)

type ctxTimeKey struct{}

// InjectCtx injects timestamp to context. It's used to inject mock timestamp for testing.
func InjectCtx(ctx context.Context, ts time.Time) context.Context {
	return context.WithValue(ctx, ctxTimeKey{}, ts)
}

// FromCtx returns timestamp from context. If timestamp is not found, it returns current time.
func FromCtx(ctx context.Context) time.Time {
	if ts, ok := ctx.Value(ctxTimeKey{}).(time.Time); ok {
		return ts
	}
	return time.Now()
}
