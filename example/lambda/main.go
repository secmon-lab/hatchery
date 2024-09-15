package main

import (
	"context"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/secmon-as-code/hatchery"
	"github.com/secmon-as-code/hatchery/destination/s3"
	"github.com/secmon-as-code/hatchery/pkg/types"
	"github.com/secmon-as-code/hatchery/source/slack"
)

// HandleRequest receives SNS event and run hatchery
func HandleRequest(ctx context.Context, snsEvent events.SNSEvent) error {
	streams := []hatchery.Stream{
		{
			// StreamID
			ID: "slack-to-s3",
			// Source: Slack Audit API
			Src: slack.New(types.NewSecretString(os.Getenv("SLACK_TOKEN"))),
			// Destination: S3, bucket name is "mizutani-test"
			Dst: s3.New("ap-northeast-1", "mizutani-test"),
		},
	}

	// In this example, SNS message has comma separated stream IDs, e.g., "slack-to-s3,some-other-stream"
	for _, record := range snsEvent.Records {
		targets := strings.Split(record.SNS.Message, ",")
		if err := hatchery.New(streams).Run(ctx, targets); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
