package source_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"testing"
	"time"

	"github.com/m-mizutani/gt"
	"github.com/secmon-as-code/hatchery/mock"
	"github.com/secmon-as-code/hatchery/pkg/metadata"
	"github.com/secmon-as-code/hatchery/pkg/types"
	"github.com/secmon-as-code/hatchery/source"
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
	dst := &mock.DestinationMock{
		NewWriterFunc: func(ctx context.Context, md metadata.MetaData) (io.WriteCloser, error) {
			buf := &writeCloseBuffer{}
			bufList = append(bufList, buf)
			return buf, nil
		},
	}

	now := time.Now()
	ctx := context.Background()
	ctx = source.InjectTimeFunc(ctx, func() time.Time {
		return now
	})

	var (
		maxPages = 2
		limit    = 10
		duration = time.Hour * 24
	)

	slack := &source.Slack{
		AccessToken: types.NewSecretString(loadOrSkip(t, "TEST_SLACK_ACCESS_TOKEN")),
		MaxPages:    &maxPages,
		Limit:       &limit,
		Duration:    &duration,
	}

	gt.NoError(t, slack.Load(ctx, dst))
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
	}
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
