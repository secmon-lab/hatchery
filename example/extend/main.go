package main

import (
	"context"
	"net/http"
	"os"

	"github.com/secmon-as-code/hatchery"
	"github.com/secmon-as-code/hatchery/destination/gcs"
	"github.com/secmon-as-code/hatchery/pkg/metadata"
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

	streams := []hatchery.Stream{
		{
			// StreamID
			ID: "my-own-service",
			// Source: Some audit API
			Src: getAuditLogs,
			// Destination: Google Cloud Storage, bucket name is "mizutani-test"
			Dst: gcs.New("mizutani-test"),
		},
	}

	// You can run CLI with args such as `go run main.go -s my-own-service`
	if err := hatchery.New(streams).CLI(os.Args); err != nil {
		panic(err)
	}
}
