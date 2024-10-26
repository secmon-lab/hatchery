package gcs_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/m-mizutani/gt"
	"github.com/secmon-lab/hatchery/destination/gcs"
	"github.com/secmon-lab/hatchery/pkg/metadata"
)

func TestIntegration(t *testing.T) {
	var bucketName string
	if v, ok := os.LookupEnv("TEST_GCS_BUCKET_NAME"); !ok {
		t.Skip("TEST_GCS_BUCKET_NAME is not set")
	} else {
		bucketName = v
	}

	ctx := context.Background()
	md := metadata.New(
		metadata.WithTimestamp(time.Now()),
	)

	w, err := gcs.New(bucketName, gcs.WithGzip(true))(ctx, md)
	gt.NoError(t, err)
	n := gt.R1(w.Write([]byte("hello, world\n"))).NoError(t)
	gt.N(t, n).Greater(0)
	gt.NoError(t, w.Close())
}
