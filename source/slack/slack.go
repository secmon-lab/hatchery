package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

// Slack is a source to load audit logs from Slack API.
type config struct {
	// AccessToken is a secret value for Slack API.
	AccessToken secret.String

	// MaxPages is the maximum number of pages to read. If it's nil, it reads logs until there are no more logs.
	MaxPages int

	// Limit is the number of logs to read in a single request. If it's nil, it reads logs with the limit of 100 logs.
	Limit int

	// Duration is the duration to read logs. If it's nil, it reads logs for the last 10 minutes.
	Duration time.Duration

	// httpClient is a HTTP client to send requests to Slack API.
	httpClient interfaces.HTTPClient
}

func New(accessToken secret.String, options ...Option) hatchery.Source {
	c := &config{
		AccessToken: accessToken,
		httpClient:  http.DefaultClient,
		Duration:    10 * time.Minute,
		Limit:       100,
		MaxPages:    0,
	}

	for _, opt := range options {
		opt(c)
	}

	return func(ctx context.Context, p *hatchery.Pipe) error {
		var nextCursor string
		now := timestamp.FromCtx(ctx)

		logger := logging.FromCtx(ctx).With("source", "slack")
		logger.Info("New source (Slack)", "config", c)
		ctx = logging.InjectCtx(ctx, logger)

		for seq := 0; c.MaxPages == 0 || seq < c.MaxPages; seq++ {
			cursor, err := c.crawl(ctx, now, seq, nextCursor, p)
			if err != nil {
				return goerr.Wrap(err, "failed to crawl slack logs").With("seq", seq).With("cursor", nextCursor).With("config", *c)
			}
			if cursor == nil {
				break
			}
			nextCursor = *cursor
		}

		return nil
	}
}

type Option func(*config)

// WithMaxPages sets the maximum number of pages to read. Default is 0, which means it reads logs until there are no more logs.
func WithMaxPages(MaxPages int) Option {
	return func(c *config) {
		c.MaxPages = MaxPages
	}
}

// WithLimit sets the number of logs to read in a single request.
func WithLimit(limit int) Option {
	return func(c *config) {
		c.Limit = limit
	}
}

// WithDuration sets the duration to read logs. Default is 10 minutes.
func WithDuration(duration time.Duration) Option {
	return func(c *config) {
		c.Duration = duration
	}
}

// WithHTTPClient sets a HTTP client to send requests to Slack API.
func WithHTTPClient(httpClient interfaces.HTTPClient) Option {
	return func(c *config) {
		c.httpClient = httpClient
	}
}

// Load reads audit logs from Slack API and write them to the destination. It reads logs for the duration specified by Duration. If Duration is nil, it reads logs for the last 10 minutes. It reads logs for the maximum number of pages specified by MaxPages. If MaxPages is nil, it reads logs until there are no more logs. It reads logs with the limit specified by Limit. If Limit is nil, it reads logs with the limit of 100 logs.

const (
	// Slack API endpoint for Business Plan
	baseURL = "https://api.slack.com/audit/v1/logs"
)

func (x *config) crawl(ctx context.Context, end time.Time, seq int, cursor string, p *hatchery.Pipe) (*string, error) {
	startTime := end.Add(-x.Duration)
	qv := url.Values{}
	qv.Add("limit", fmt.Sprintf("%d", x.Limit))
	qv.Add("oldest", fmt.Sprintf("%d", startTime.Unix()))

	if cursor != "" {
		qv.Add("cursor", cursor)
	}

	endpoint, err := url.Parse(baseURL)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to parse URL").With("url", baseURL)
	}
	endpoint.RawQuery = qv.Encode()

	apiURL := endpoint.String()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to create HTTP request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+x.AccessToken.Unsafe())

	httpResp, err := x.httpClient.Do(httpReq)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to send HTTP request")
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(httpResp.Body)
		return nil, goerr.New("unexpected status code").With("status", httpResp.Status).With("body", string(data))
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to read response body")
	}

	var resp struct {
		ResponseMetadata struct {
			NextCursor string `json:"next_cursor"`
		} `json:"response_metadata"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, goerr.Wrap(err, "failed to unmarshal response body")
	}

	md := metadata.New(
		metadata.WithTimestamp(end),
		metadata.WithSeq(seq),
		metadata.WithFormat(types.FmtJSON),
	)
	if err := p.Spout(ctx, bytes.NewReader(body), md); err != nil {
		return nil, goerr.Wrap(err, "failed to write response to destination")
	}

	if resp.ResponseMetadata.NextCursor != "" {
		return &resp.ResponseMetadata.NextCursor, nil
	}

	return nil, nil
}
