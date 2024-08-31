package source

import (
	"context"
	"time"
)

// InjectHTTPClient injects HTTP client to context. It's used to inject mock client for testing.
func InjectHTTPClient(ctx context.Context, c HTTPClient) context.Context {
	return context.WithValue(ctx, ctxHTTPClientKey{}, c)
}

type timeFunc func() time.Time

type ctxTimeFuncKey struct{}

func timeFuncFromCtx(ctx context.Context) timeFunc {
	if f, ok := ctx.Value(ctxTimeFuncKey{}).(timeFunc); ok {
		return f
	}
	return time.Now
}

// InjectTimeFunc injects time function to context. It's used to inject mock time function for testing.
func InjectTimeFunc(ctx context.Context, f timeFunc) context.Context {
	return context.WithValue(ctx, ctxTimeFuncKey{}, f)
}
