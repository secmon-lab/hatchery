package main

import (
	"os"

	"github.com/secmon-lab/hatchery"
	"github.com/secmon-lab/hatchery/destination/gcs"
	"github.com/secmon-lab/hatchery/pkg/types/secret"
	"github.com/secmon-lab/hatchery/source/slack"
)

func main() {
	streams := []*hatchery.Stream{
		hatchery.NewStream(
			// Source: Slack Audit API
			slack.New(secret.NewString(os.Getenv("SLACK_TOKEN"))),
			// Destination: Google Cloud Storage, bucket name is "mizutani-test"
			gcs.New("mizutani-test"),

			// With ID
			hatchery.WithID("slack-to-gcs"),
		),
	}

	// You can run CLI with args such as `go run main.go -s slack-to-gcs`
	if err := hatchery.New(streams).CLI(os.Args); err != nil {
		panic(err)
	}
}
