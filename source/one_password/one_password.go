package one_password

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/m-mizutani/goerr"
	"github.com/secmon-lab/hatchery"
	"github.com/secmon-lab/hatchery/pkg/interfaces"
	"github.com/secmon-lab/hatchery/pkg/logging"
	"github.com/secmon-lab/hatchery/pkg/metadata"
	"github.com/secmon-lab/hatchery/pkg/timestamp"
	"github.com/secmon-lab/hatchery/pkg/types"
	"github.com/secmon-lab/hatchery/pkg/types/secret"
)

const (
	// 1Password API endpoint for Business Plan
	// See https://developer.1password.com/docs/events-api/reference/
	APIEndpoint = "https://events.1password.com/api/v1/auditevents"

	// Time format for 1Password API
	// 2023-03-15T16:32:50-03:00
	timeFormat = "2006-01-02T15:04:05-07:00"
)

type config struct {
	APIToken   secret.String
	MaxPages   int
	Limit      int
	Duration   time.Duration
	httpClient interfaces.HTTPClient
}

type Option func(*config)

// WithMaxPages sets the maximum number of pages to load. If 0, it loads all pages. Default is 0.
func WithMaxPages(n int) Option {
	return func(x *config) {
		x.MaxPages = n
	}
}

// WithLimit sets the number of logs to load per page. Default is 100.
func WithLimit(n int) Option {
	return func(x *config) {
		x.Limit = n
	}
}

// WithDuration sets the duration of logs to load. Default is 10 minutes.
func WithDuration(d time.Duration) Option {
	return func(x *config) {
		x.Duration = d
	}
}

// WithHTTPClient sets the HTTP client to send requests. Default is http.DefaultClient. This option is mainly for testing.
func WithHTTPClient(httpClient interfaces.HTTPClient) Option {
	return func(x *config) {
		x.httpClient = httpClient
	}
}

// New creates a source to load audit logs from 1Password API.
func New(apiToken secret.String, opts ...Option) hatchery.Source {
	x := &config{
		APIToken:   apiToken,
		MaxPages:   0,
		Limit:      100,
		Duration:   time.Minute * 10,
		httpClient: http.DefaultClient,
	}

	for _, opt := range opts {
		opt(x)
	}

	return func(ctx context.Context, p *hatchery.Pipe) error {
		var nextCursor string
		now := timestamp.FromCtx(ctx)

		logger := logging.FromCtx(ctx).With("source", "one_password")
		logger.Info("New source (1Password)", "config", x, "base_time", now)
		ctx = logging.InjectCtx(ctx, logger)

		slug, err := metadata.RandomSlug()
		if err != nil {
			return goerr.Wrap(err, "failed to generate random slug")
		}

		for seq := 0; x.MaxPages == 0 || seq < x.MaxPages; seq++ {
			cursor, err := x.crawl(ctx, p, now, seq, nextCursor, slug)
			if err != nil {
				return goerr.Wrap(err, "failed to crawl 1Password logs").With("seq", seq).With("cursor", nextCursor)
			}
			if cursor == nil {
				break
			}
			nextCursor = *cursor
		}

		return nil
	}
}

func (x *config) crawl(ctx context.Context, p *hatchery.Pipe, end time.Time, seq int, cursor, slug string) (*string, error) {
	startTime := end.Add(-x.Duration)
	var body []byte
	if cursor != "" {
		raw, err := json.Marshal(apiResponseWithCursor{Cursor: cursor})
		if err != nil {
			return nil, goerr.Wrap(err, "failed to marshal API request")
		}
		body = raw
	} else {
		raw, err := json.Marshal(apiRequest{
			Limit:     x.Limit,
			StartTime: startTime.Format(timeFormat),
			EndTime:   end.Format(timeFormat),
		})
		if err != nil {
			return nil, goerr.Wrap(err, "failed to marshal API request")
		}
		body = raw
	}
	reader := bytes.NewReader(body)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, APIEndpoint, reader)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to create HTTP request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+x.APIToken.Unsafe())

	httpResp, err := x.httpClient.Do(httpReq)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to send HTTP request")
	}

	if httpResp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(httpResp.Body)
		return nil, goerr.New("unexpected status code").With("status", httpResp.Status).With("body", string(data))
	}

	body, err = io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to read response body")
	}

	var resp apiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, goerr.Wrap(err, "failed to unmarshal response body")
	}

	md := metadata.New(
		metadata.WithTimestamp(end),
		metadata.WithSeq(seq),
		metadata.WithFormat(types.FmtJSON),
		metadata.WithSlug(slug),
	)

	if err := p.Spout(ctx, bytes.NewReader(body), md); err != nil {
		return nil, goerr.Wrap(err, "failed to spout 1Password logs")
	}

	if resp.HasMore {
		return &resp.Cursor, nil
	}

	return nil, nil
}

type apiRequest struct {
	Limit     int    `json:"limit"`
	StartTime string `json:"start_time,omitempty"`
	EndTime   string `json:"end_time,omitempty"`
}

type apiResponseWithCursor struct {
	Cursor string `json:"cursor"`
}

type apiResponse struct {
	Cursor  string `json:"cursor"`
	HasMore bool   `json:"has_more"`
}
