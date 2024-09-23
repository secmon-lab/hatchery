# hatchery

A code-based audit log collector for SaaS services

![overview](https://github.com/m-mizutani/hatchery/assets/605953/0d065e1e-1b40-493b-a9c5-8215f2e1691e)

## Motivation

Many SaaS services offer APIs for accessing data and logs, but managing them can be challenging due to various reasons:

- Audit logs are often set to expire after a few months.
- The built-in log search console provided by the service is not user-friendly and lacks centralized functionality for searching and analysis.

As a result, security administrators are required to gather logs from multiple services and store them in object storage for long-term retention and analysis. However, this process is complicated by the fact that each service has its own APIs and data formats, making it difficult to implement and maintain a tool to gather logs.

`hatchery` is a solution designed to address these challenges by collecting data and logs from SaaS services and storing them in object storage. This facilitates log retention and prepares the data for analysis by security administrators.

## Documentation

- About Hatchery
  - [Overview](docs/README.md)
  - [How to Use hatchery](docs/usage.md)
  - [How to Develop Hatchery Extension](docs/extension.md)
- Source
  - [Slack](source/slack)
  - [1Password](source/1password)
  - [Falcon Data Replicator](source/falcon_data_replicator)
  - [Twilio](source/twilio)
- Destination
  - [Google Cloud Storage](destination/gcs)
  - [Amazon S3](destination/s3)

## Getting Started

### Prerequisites

- Go 1.22 or later
- Scheduled binary execution service (e.g., cron, AWS Lambda, Google Cloud Run)

### Build your binary

Write your own main.go. For example, the following code collects logs from Slack and stores them in Google Cloud Storage.

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
	streams := []*hatchery.Stream{
		hatchery.NewStream(
			// Source: Slack Audit API
			slack.New(secret.NewString(os.Getenv("SLACK_TOKEN"))),
			// Destination: Google Cloud Storage, bucket name is "mizutani-test"
			gcs.New("mizutani-test"),

			// Option: WithID sets the stream ID
			hatchery.WithID("slack-to-gcs"),
		),
	}

	// You can run CLI with args such as `go run main.go -i slack-to-gcs`
	if err := hatchery.New(streams).CLI(os.Args); err != nil {
		panic(err)
	}
}
```

Build your binary.

```sh
$ go build -o myhatchery main.go
```

### Run your binary

Run your binary.

```sh
$ env SLACK_TOKEN=your-slack-token ./myhatchery -s slack-to-gcs
```

It will collect logs from Slack and store them in Google Cloud Storage.


## License

Apache License 2.0
