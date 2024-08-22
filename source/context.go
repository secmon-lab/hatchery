package source

import (
	"context"
	"net/http"
	"time"
)

// HTTPClient is an interface for HTTP client. It's used to inject mock client for testing.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type ctxHTTPClientKey struct{}

func httpClientFromCtx(ctx context.Context) HTTPClient {
	if c, ok := ctx.Value(ctxHTTPClientKey{}).(HTTPClient); ok {
		return c
	}
	return http.DefaultClient
}

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
