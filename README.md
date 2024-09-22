# hatchery
A tool to gather data + logs from SaaS and save them to object storage

![overview](https://github.com/m-mizutani/hatchery/assets/605953/0d065e1e-1b40-493b-a9c5-8215f2e1691e)

## Motivation

Many SaaS services offer APIs for accessing data and logs, but managing them can be challenging due to various reasons:

- Audit logs are often set to expire after a few months.
- The built-in log search console provided by the service is not user-friendly and lacks centralized functionality for searching and analysis.

As a result, security administrators are required to gather logs from multiple services and store them in object storage for long-term retention and analysis. However, this process is complicated by the fact that each service has its own APIs and data formats, making it difficult to implement and maintain a tool to gather logs.

`hatchery` is a solution designed to address these challenges by collecting data and logs from SaaS services and storing them in object storage. This facilitates log retention and prepares the data for analysis by security administrators.

For those interested in importing logs from Cloud Storage to BigQuery, please refer to [swarm](https://github.com/m-mizutani/swarm).

## Getting Started

### Prerequisites

- Go 1.22 or later
- Scheduled binary execution service (e.g., cron, AWS Lambda, Google Cloud Run)

### Build your binary

```go
package main

import (
	"os"

	"github.com/secmon-as-code/hatchery"
	"github.com/secmon-as-code/hatchery/destination/gcs"
	"github.com/secmon-as-code/hatchery/pkg/types/secret"
	"github.com/secmon-as-code/hatchery/source/slack"
)

func main() {
	streams := []hatchery.Stream{
		{
			// StreamID
			ID: "slack-to-gcs",
			// Source: Slack Audit API
			Src: slack.New(secret.NewString(os.Getenv("SLACK_TOKEN"))),
			// Destination: Google Cloud Storage, bucket name is "mizutani-test"
			Dst: gcs.New("mizutani-test"),
		},
	}

	// You can run CLI with args such as `go run main.go -s slack-to-gcs`
	if err := hatchery.New(streams).CLI(os.Args); err != nil {
		panic(err)
	}
}
```

## License

Apache License 2.0
