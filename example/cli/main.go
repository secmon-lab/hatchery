package main

import (
	"os"

	"github.com/secmon-as-code/hatchery"
	"github.com/secmon-as-code/hatchery/destination/gcs"
	"github.com/secmon-as-code/hatchery/pkg/types"
	"github.com/secmon-as-code/hatchery/source/slack"
)

func main() {
	streams := []hatchery.Stream{
		{
			// StreamID
			ID: "slack-to-gcs",
			// Source: Slack Audit API
			Src: slack.New(types.NewSecretString(os.Getenv("SLACK_TOKEN"))),
			// Destination: Google Cloud Storage, bucket name is "mizutani-test"
			Dst: gcs.New("mizutani-test"),
		},
	}

	// You can run CLI with args such as `go run main.go -s slack-to-gcs`
	if err := hatchery.New(streams).CLI(os.Args); err != nil {
		panic(err)
	}
}
