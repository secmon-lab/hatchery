package twilio

import (
	"context"
	"net/http"
	"net/url"

	"github.com/m-mizutani/goerr"
	"github.com/secmon-lab/hatchery"
	"github.com/secmon-lab/hatchery/pkg/interfaces"
	"github.com/secmon-lab/hatchery/pkg/metadata"
	"github.com/secmon-lab/hatchery/pkg/types/secret"
)

const (
	defaultURL = "https://monitor.twilio.com/v1/Events"
)

type config struct {
	baseURL    string
	httpClient interfaces.HTTPClient
}

func WithBaseURL(url string) Option {
	return func(c *config) {
		c.baseURL = url
	}
}

func WithHTTPClient(client interfaces.HTTPClient) Option {
	return func(c *config) {
		c.httpClient = client
	}
}

func New(sid string, token secret.String, options ...Option) hatchery.Source {
	c := &config{
		baseURL:    defaultURL,
		httpClient: http.DefaultClient,
	}

	for _, opt := range options {
		opt(c)
	}

	// To be updated
	return func(ctx context.Context, p *hatchery.Pipe) error {
		reqURL, err := url.Parse(c.baseURL)
		if err != nil {
			return goerr.Wrap(err, "failed to parse URL").With("url", c.baseURL)
		}
		req, err := http.NewRequestWithContext(ctx, "GET", reqURL.String(), nil)
		if err != nil {
			return goerr.Wrap(err, "failed to create request")
		}

		req.SetBasicAuth(sid, token.Unsafe())

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return goerr.Wrap(err, "failed to send request")
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return goerr.New("unexpected status code").With("code", resp.StatusCode)
		}

		if err := p.Spout(ctx, resp.Body, metadata.New()); err != nil {
			return goerr.Wrap(err, "failed to spout data")
		}
		return nil
	}
}

type Option func(*config)
