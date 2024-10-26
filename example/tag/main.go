package main

import (
	"os"

	"github.com/secmon-lab/hatchery"
	"github.com/secmon-lab/hatchery/destination/gcs"
	"github.com/secmon-lab/hatchery/pkg/types/secret"
	"github.com/secmon-lab/hatchery/source/one_password"
)

func main() {
	streams := hatchery.Streams{
		hatchery.NewStream(
			// Source: 1Password audit log
			one_password.New(secret.NewString("1password")),
			// Destination: Google Cloud Storage, bucket name is "mizutani-test"
			gcs.New("mizutani-test"),
			// Tags
			hatchery.WithTags("hourly"),
		),
	}

	// You can run CLI with args such as `go run main.go -t hourly`
	if err := hatchery.New(streams).CLI(os.Args); err != nil {
		panic(err)
	}
}
