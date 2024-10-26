package main

import (
	"os"

	"github.com/secmon-lab/hatchery"
	"github.com/secmon-lab/hatchery/destination/gcs"
	"github.com/secmon-lab/hatchery/destination/s3"
	"github.com/secmon-lab/hatchery/pkg/types/secret"
	"github.com/secmon-lab/hatchery/source/one_password"
	"github.com/secmon-lab/hatchery/source/slack"
)

func main() {
	streams := []*hatchery.Stream{
		hatchery.NewStream(
			// Source: Slack Audit API
			slack.New(secret.NewString(os.Getenv("SLACK_TOKEN"))),
			// Destination: Google Cloud Storage
			gcs.New("mizutani-test"),

			// Identifiers
			hatchery.WithID("slack-to-gcs"),
			hatchery.WithTags("hourly"),
		),

		hatchery.NewStream(
			// Source: 1Password
			one_password.New(secret.NewString(os.Getenv("ONE_PASSWORD_TOKEN"))),
			// Destination: Amazon S3
			s3.New("ap-northeast1", "mizutani-test"),

			// Identifiers
			hatchery.WithID("1pw-to-s3"),
			hatchery.WithTags("daily"),
		),
	}

	// You can run CLI with args such as `go run main.go -s slack-to-gcs`
	if err := hatchery.New(streams).CLI(os.Args); err != nil {
		panic(err)
	}
}
