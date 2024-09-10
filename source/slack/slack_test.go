package slack_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/m-mizutani/gt"
	"github.com/secmon-as-code/hatchery"
	"github.com/secmon-as-code/hatchery/pkg/metadata"
	"github.com/secmon-as-code/hatchery/pkg/mock"
	"github.com/secmon-as-code/hatchery/pkg/timestamp"
	"github.com/secmon-as-code/hatchery/pkg/types"
	"github.com/secmon-as-code/hatchery/source/slack"
)

type writeCloseBuffer struct {
	bytes.Buffer
	closed bool
}

func (w *writeCloseBuffer) Close() error {
	w.closed = true
	return nil
}

func loadOrSkip(t *testing.T, key string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		t.Skipf("missing %s", key)
	}
	return v
}

func TestSlackIntegration(t *testing.T) {
	var bufList []*writeCloseBuffer
	var called int
	now := time.Now()
	dst := func(ctx context.Context, md metadata.MetaData) (io.WriteCloser, error) {
		called++
		gt.Equal(t, md.Timestamp(), now)
		buf := &writeCloseBuffer{}
		bufList = append(bufList, buf)
		return buf, nil
	}

	ctx := context.Background()
	ctx = timestamp.InjectCtx(ctx, now)

	var (
		maxPages = 2
		limit    = 10
		duration = time.Hour * 24
	)

	src := slack.New(
		types.NewSecretString(loadOrSkip(t, "TEST_SLACK_ACCESS_TOKEN")),
		slack.WithMaxPages(maxPages),
		slack.WithLimit(limit),
		slack.WithDuration(duration),
	)

	gt.NoError(t, src(ctx, hatchery.NewPipe(dst)))
	gt.A(t, bufList).Longer(0)

	for _, buf := range bufList {
		var resp struct {
			Entries []struct {
				ID string `json:"id"`
			} `json:"entries"`
			ResponseMetadata struct {
				NextCursor string `json:"next_cursor"`
			} `json:"response_metadata"`
		}

		gt.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
		gt.S(t, resp.ResponseMetadata.NextCursor).IsNotEmpty()
		gt.A(t, resp.Entries).Longer(0)
		gt.S(t, resp.Entries[0].ID).IsNotEmpty()
		gt.Equal(t, called, 1)
	}
}

//go:embed testdata/resp1.json
var resp1 []byte

//go:embed testdata/resp2.json
var resp2 []byte

func TestSlackCrawler(t *testing.T) {

	now := time.Now()
	ctx := context.Background()
	ctx = timestamp.InjectCtx(ctx, now)

	var (
		maxPages = 2
		limit    = 10
		duration = time.Hour * 24
	)

	var httpCount int
	httpMock := &mock.HTTPClientMock{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			httpCount++
			switch httpCount {
			case 1:
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewReader(resp1)),
				}, nil
			case 2:
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewReader(resp2)),
				}, nil
			default:
				return nil, io.EOF
			}
		},
	}

	var bufList []*writeCloseBuffer
	dstMock := func(ctx context.Context, md metadata.MetaData) (io.WriteCloser, error) {
		buf := &writeCloseBuffer{}
		bufList = append(bufList, buf)
		return buf, nil
	}

	src := slack.New(
		types.NewSecretString("dummy"),
		slack.WithMaxPages(maxPages),
		slack.WithLimit(limit),
		slack.WithDuration(duration),
		slack.WithHTTPClient(httpMock),
	)

	gt.NoError(t, src(ctx, hatchery.NewPipe(dstMock)))

	gt.A(t, httpMock.DoCalls()).Length(2).
		At(0, func(t testing.TB, v struct{ Req *http.Request }) {
			gt.S(t, v.Req.URL.Path).Equal("/audit/v1/logs")
			gt.S(t, v.Req.URL.Query().Get("limit")).Equal("10")
			gt.S(t, v.Req.URL.Query().Get("cursor")).Equal("")
		}).
		At(1, func(t testing.TB, v struct{ Req *http.Request }) {
			gt.S(t, v.Req.URL.Path).Equal("/audit/v1/logs")
			gt.S(t, v.Req.URL.Query().Get("limit")).Equal("10")
			gt.S(t, v.Req.URL.Query().Get("cursor")).NotEqual("")
		})

	gt.A(t, bufList).Length(2).
		At(0, func(t testing.TB, buf *writeCloseBuffer) {
			gt.True(t, buf.closed)
			gt.Equal(t, buf.Bytes(), resp1)
		}).
		At(1, func(t testing.TB, buf *writeCloseBuffer) {
			gt.True(t, buf.closed)
			gt.Equal(t, buf.Bytes(), resp2)
		})
}

/*
func TestIntegration(t *testing.T) {
	prefix := "slack-2/"
	req := &config.SlackImpl{
		AccessToken: utils.LoadEnv(t, "TEST_SLACK_ACCESS_TOKEN"),
		Bucket:      utils.LoadEnv(t, "TEST_BUCKET"),
		Prefix:      &prefix,
		Duration: &pkl.Duration{
			Value: 1,
			Unit:  pkl.Hour,
		},
		Limit: 1000,
	}

	ctx := context.Background()
	csClient := gt.R1(cs.New(ctx)).NoError(t)
	clients := infra.New(infra.WithCloudStorage(csClient))

	gt.NoError(t, slack.Exec(ctx, clients, req)).Must()
}
*/
