package config

import (
	"context"
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
	return []cli.Flag{
		&cli.TimestampFlag{
			Name:        "start-time",
			Category:    "Time Range",
			Aliases:     []string{"s"},
			Sources:     cli.EnvVars("HATCHERY_START_TIME"),
			Usage:       "Start time to load data",
			Destination: &x.start,
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
			Config: cli.TimestampConfig{
				Layouts: []string{time.RFC3339, time.RFC3339Nano},
			},
		},
		&cli.DurationFlag{
			Name:        "tick",
			Category:    "Time Range",
			Aliases:     []string{"d"},
			Sources:     cli.EnvVars("HATCHERY_TICK"),
			Usage:       "Tick interval to load data",
			Destination: &x.tick,
			Value:       time.Second,
		},
	}
}

func (x *Range) Validate() error {
	if !x.start.IsZero() && !x.end.IsZero() && x.tick > 0 {
		return nil
	}
	if x.start.Before(x.end) {
		return goerr.New("start-time is after end-time")
	}

	if !x.start.IsZero() || !x.end.IsZero() || x.tick > 0 {
		return goerr.New("Some of start-time, end-time, and tick are not set")
	}

	return nil
}

func (x *Range) IsEnable(ctx context.Context) bool {
	return !x.start.IsZero() && !x.end.IsZero() && x.tick > 0
}

func (x *Range) Generate(yield func(time.Time) bool) {
	if x.start.IsZero() || x.end.IsZero() || x.tick == 0 {
		panic("Range.Generate is called without setting start-time, end-time, and tick")
	}

	for t := x.start; t.Before(x.end); t = t.Add(x.tick) {
		if !yield(t.Add(x.tick)) {
			return
		}
	}
}
