package source

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
	"github.com/secmon-as-code/hatchery/pkg/metadata"
	"github.com/secmon-as-code/hatchery/pkg/types"
)

// Slack is a source to load audit logs from Slack API.
type Slack struct {
	// AccessToken is a secret value for Slack API.
	AccessToken types.SecretString

	// MaxPages is the maximum number of pages to read. If it's nil, it reads logs until there are no more logs.
	MaxPages *int

	// Limit is the number of logs to read in a single request. If it's nil, it reads logs with the limit of 100 logs.
	Limit *int

	// Duration is the duration to read logs. If it's nil, it reads logs for the last 10 minutes.
	Duration *time.Duration
}

// Load reads audit logs from Slack API and write them to the destination. It reads logs for the duration specified by Duration. If Duration is nil, it reads logs for the last 10 minutes. It reads logs for the maximum number of pages specified by MaxPages. If MaxPages is nil, it reads logs until there are no more logs. It reads logs with the limit specified by Limit. If Limit is nil, it reads logs with the limit of 100 logs.
func (x *Slack) Load(ctx context.Context, dst hatchery.Destination) error {
	var nextCursor string
	now := timeFuncFromCtx(ctx)()

	for seq := 0; x.MaxPages == nil || seq < *x.MaxPages; seq++ {
		cursor, err := x.crawl(ctx, now, nextCursor, dst)
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

func (x *Slack) crawl(ctx context.Context, end time.Time, cursor string, dst hatchery.Destination) (*string, error) {
	d := 10 * time.Minute
	if x.Duration != nil {
		d = *x.Duration
	}

	limit := 100
	if x.Limit != nil {
		limit = *x.Limit
	}

	startTime := end.Add(-d)
	qv := url.Values{}
	qv.Add("limit", fmt.Sprintf("%d", limit))
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
	httpReq.Header.Set("Authorization", "Bearer "+x.AccessToken.UnsafeString())

	client := httpClientFromCtx(ctx)

	httpResp, err := client.Do(httpReq)
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

	var resp apiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, goerr.Wrap(err, "failed to unmarshal response body")
	}

	md := metadata.New(metadata.WithTimestamp(end))
	w, err := dst.NewWriter(ctx, md)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to create object writer")
	}

	n, err := io.Copy(w, bytes.NewReader(body))
	if err != nil {
		return nil, goerr.Wrap(err, "failed to write response to object writer").With("bytes", n)
	}

	if err := w.Close(); err != nil {
		return nil, goerr.Wrap(err, "failed to close dst writer")
	}

	if resp.ResponseMetadata.NextCursor != "" {
		return &resp.ResponseMetadata.NextCursor, nil
	}

	return nil, nil
}

type apiResponse struct {
	ResponseMetadata struct {
		NextCursor string `json:"next_cursor"`
	} `json:"response_metadata"`
}
