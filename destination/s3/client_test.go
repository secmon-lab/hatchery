package s3_test

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/m-mizutani/gt"
	"github.com/secmon-lab/hatchery/destination/s3"
	"github.com/secmon-lab/hatchery/pkg/metadata"

	"github.com/aws/aws-sdk-go-v2/config"
	aws_s3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestIntegration(t *testing.T) {
	bucketName, ok := os.LookupEnv("TEST_S3_BUCKET_NAME")
	if !ok {
		t.Skip("TEST_S3_BUCKET_NAME is not set")
	}
	dst := s3.New("ap-northeast-1", bucketName)

	ctx := context.Background()
	ts := time.Now()
	md := metadata.New(
		metadata.WithTimestamp(ts),
	)
	w, err := dst(ctx, md)
	gt.NoError(t, err)

	gt.R1(w.Write([]byte("Hello, world"))).NoError(t)
	gt.NoError(t, w.Close()).Must()

	awsOpts := []func(*config.LoadOptions) error{
		config.WithRegion("ap-northeast-1"),
	}

	cfg := gt.R1(config.LoadDefaultConfig(ctx, awsOpts...)).NoError(t)

	client := aws_s3.NewFromConfig(cfg)

	expectedKey := ts.Format("2006/01/02/15/20060102T150405_0000.log")
	out, err := client.GetObject(ctx, &aws_s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &expectedKey,
	})
	gt.NoError(t, err).Must()

	buf := gt.R1(io.ReadAll(out.Body)).NoError(t)
	gt.Equal(t, string(buf), "Hello, world")
}
