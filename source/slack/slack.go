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
	"github.com/secmon-as-code/hatchery"
	"github.com/secmon-as-code/hatchery/pkg/interfaces"
	"github.com/secmon-as-code/hatchery/pkg/metadata"
	"github.com/secmon-as-code/hatchery/pkg/timestamp"
	"github.com/secmon-as-code/hatchery/pkg/types"
)

// Slack is a source to load audit logs from Slack API.
type Client struct {
	// accessToken is a secret value for Slack API.
	accessToken types.SecretString

	// maxPages is the maximum number of pages to read. If it's nil, it reads logs until there are no more logs.
	maxPages int

	// Limit is the number of logs to read in a single request. If it's nil, it reads logs with the limit of 100 logs.
	limit int

	// Duration is the duration to read logs. If it's nil, it reads logs for the last 10 minutes.
	duration time.Duration

	// httpClient is a HTTP client to send requests to Slack API.
	httpClient interfaces.HTTPClient
}

func New(accessToken types.SecretString, options ...Option) *Client {
	c := &Client{
		accessToken: accessToken,
		httpClient:  http.DefaultClient,
		duration:    10 * time.Minute,
		limit:       100,
		maxPages:    0,
	}

	for _, opt := range options {
		opt(c)
	}

	return c
}

type Option func(*Client)

// WithMaxPages sets the maximum number of pages to read. Default is 0, which means it reads logs until there are no more logs.
func WithMaxPages(maxPages int) Option {
	return func(c *Client) {
		c.maxPages = maxPages
	}
}

// WithLimit sets the number of logs to read in a single request.
func WithLimit(limit int) Option {
	return func(c *Client) {
		c.limit = limit
	}
}

// WithDuration sets the duration to read logs. Default is 10 minutes.
func WithDuration(duration time.Duration) Option {
	return func(c *Client) {
		c.duration = duration
	}
}

// WithHTTPClient sets a HTTP client to send requests to Slack API.
func WithHTTPClient(httpClient interfaces.HTTPClient) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// Load reads audit logs from Slack API and write them to the destination. It reads logs for the duration specified by Duration. If Duration is nil, it reads logs for the last 10 minutes. It reads logs for the maximum number of pages specified by maxPages. If maxPages is nil, it reads logs until there are no more logs. It reads logs with the limit specified by Limit. If Limit is nil, it reads logs with the limit of 100 logs.
func (x *Client) Load(ctx context.Context, p *hatchery.Pipe) error {
	var nextCursor string
	now := timestamp.FromCtx(ctx)

	for seq := 0; x.maxPages == 0 || seq < x.maxPages; seq++ {
		cursor, err := x.crawl(ctx, now, seq, nextCursor, p)
		if err != nil {
			return goerr.Wrap(err, "failed to crawl slack logs").With("seq", seq).With("cursor", nextCursor).With("config", *x)
		}
		if cursor == nil {
			break
		}
		nextCursor = *cursor
	}

	return nil
}

const (
	// Slack API endpoint for Business Plan
	baseURL = "https://api.slack.com/audit/v1/logs"
)

func (x *Client) crawl(ctx context.Context, end time.Time, seq int, cursor string, p *hatchery.Pipe) (*string, error) {
	startTime := end.Add(-x.duration)
	qv := url.Values{}
	qv.Add("limit", fmt.Sprintf("%d", x.limit))
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
	httpReq.Header.Set("Authorization", "Bearer "+x.accessToken.UnsafeString())

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
