package source

import (
	"context"
	"net/http"
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
