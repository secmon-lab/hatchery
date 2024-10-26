package main

import (
	"context"
	"net/http"
	"os"

	"github.com/secmon-lab/hatchery"
	"github.com/secmon-lab/hatchery/destination/gcs"
	"github.com/secmon-lab/hatchery/pkg/metadata"
)

func main() {
	getAuditLogs := func(ctx context.Context, p *hatchery.Pipe) error {
		resp, err := http.Get("https://example.com/api/audit")
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if err := p.Spout(ctx, resp.Body, metadata.New()); err != nil {
			return err
		}

		return nil
	}

	streams := hatchery.Streams{
		hatchery.NewStream(
			// Source: Some audit API
			getAuditLogs,
			// Destination: Google Cloud Storage, bucket name is "mizutani-test"
			gcs.New("mizutani-test"),
			// With ID
			hatchery.WithID("my-own-service"),
		),
	}

	// You can run CLI with args such as `go run main.go -s my-own-service`
	if err := hatchery.New(streams).CLI(os.Args); err != nil {
		panic(err)
	}
}
