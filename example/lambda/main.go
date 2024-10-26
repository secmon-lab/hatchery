package main

import (
	"context"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/secmon-lab/hatchery"
	"github.com/secmon-lab/hatchery/destination/s3"
	"github.com/secmon-lab/hatchery/pkg/types/secret"
	"github.com/secmon-lab/hatchery/source/slack"
)

// HandleRequest receives SNS event and run hatchery
func HandleRequest(ctx context.Context, snsEvent events.SNSEvent) error {
	streams := []*hatchery.Stream{
		hatchery.NewStream(
			// Source: Slack Audit API
			slack.New(secret.NewString(os.Getenv("SLACK_TOKEN"))),
			// Destination: S3, bucket name is "mizutani-test"
			s3.New("ap-northeast-1", "mizutani-test"),
			// With ID
			hatchery.WithID("slack-to-s3"),
		),
	}

	// In this example, SNS message has comma separated stream IDs, e.g., "slack-to-s3,some-other-stream"
	for _, record := range snsEvent.Records {
		targets := strings.Split(record.SNS.Message, ",")
		if err := hatchery.New(streams).Run(ctx, hatchery.SelectByID(targets...)); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
