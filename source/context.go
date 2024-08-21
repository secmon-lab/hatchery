package source

import (
	"context"
	"net/http"
	"time"
)

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

func InjectTimeFunc(ctx context.Context, f timeFunc) context.Context {
	return context.WithValue(ctx, ctxTimeFuncKey{}, f)
}
