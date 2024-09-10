package main

import (
	"os"

	"github.com/secmon-as-code/hatchery"
	"github.com/secmon-as-code/hatchery/destination/gcs"
	"github.com/secmon-as-code/hatchery/pkg/types"
	"github.com/secmon-as-code/hatchery/source/slack"
)

func main() {
	h := hatchery.New(
		hatchery.WithStream(
			// StreamID
			"slack-to-gcs",

			// Source: Slack Audit API
			slack.New(types.NewSecretString(os.Getenv("SLACK_TOKEN"))),

			// Destination: Google Cloud Storage, bucket name is "mizutani-test"
			gcs.New("mizutani-test"),
		),
	)

	// You can run CLI with args such as `go run main.go -s slack-to-gcs`
	if err := h.CLI(os.Args); err != nil {
		panic(err)
	}
}
