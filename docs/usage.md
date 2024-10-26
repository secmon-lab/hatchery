# How to use hatchery

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
