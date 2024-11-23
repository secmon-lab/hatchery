package config

import (
	"time"

	"github.com/m-mizutani/goerr"
	"github.com/urfave/cli/v3"
)

type Range struct {
	start time.Time
	end   time.Time
	tick  time.Duration
}

func (x *Range) Flags() []cli.Flag {
	now := time.Now()
	return []cli.Flag{
		&cli.TimestampFlag{
			Name:        "start-time",
			Category:    "Time Range",
			Aliases:     []string{"s"},
			Sources:     cli.EnvVars("HATCHERY_START_TIME"),
			Usage:       "Start time to load data",
			Destination: &x.start,
			Value:       now,
			Config: cli.TimestampConfig{
				Layouts: []string{time.RFC3339, time.RFC3339Nano},
			},
		},
		&cli.TimestampFlag{
			Name:        "end-time",
			Category:    "Time Range",
			Aliases:     []string{"e"},
			Sources:     cli.EnvVars("HATCHERY_END_TIME"),
			Usage:       "End time to load data",
			Destination: &x.end,
			Value:       now,
			Config: cli.TimestampConfig{
				Layouts: []string{time.RFC3339, time.RFC3339Nano},
			},
		},
		&cli.DurationFlag{
			Name:        "tick",
			Category:    "Time Range",
			Aliases:     []string{"d"},
			Sources:     cli.EnvVars("HATCHERY_TICK"),
			Usage:       "Tick interval to load data. If not set, do not loop",
			Destination: &x.tick,
		},
	}
}

func (x *Range) Validate() error {
	if x.start.After(x.end) {
		return goerr.New("start-time is after end-time")
	}

	return nil
}

func (x *Range) Generate(yield func(time.Time) bool) {
	// If tick is not set, return value only once.
	if x.tick == 0 {
		yield(x.start)
		return
	}

	println(x.start.After(x.end))
	for t := x.start; !t.After(x.end); t = t.Add(x.tick) {
		if !yield(t) {
			return
		}
	}
}
