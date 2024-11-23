package hatchery

import (
	"log/slog"
	"time"

	"github.com/m-mizutani/goerr"
	"github.com/secmon-lab/hatchery/pkg/config"
	"github.com/secmon-lab/hatchery/pkg/logging"
	"github.com/secmon-lab/hatchery/pkg/timestamp"
	"github.com/urfave/cli/v2"
)

func (h *Hatchery) CLI(argv []string) error {

	var (
		streamIDs cli.StringSlice
		tags      cli.StringSlice
		forAll    bool
		baseTime  cli.Timestamp

		startTime cli.Timestamp
		endTime   cli.Timestamp
		tick      time.Duration

		cfgLogging config.Logging
	)

	flags := []cli.Flag{
		&cli.StringSliceFlag{
			Name:        "stream-id",
			Aliases:     []string{"i"},
			EnvVars:     []string{"HATCHERY_STREAM_ID"},
			Usage:       "Target stream ID, multiple IDs can be specified",
			Destination: &streamIDs,
		},
		&cli.StringSliceFlag{
			Name:        "stream-tag",
			Aliases:     []string{"t"},
			EnvVars:     []string{"HATCHERY_STREAM_TAG"},
			Usage:       "Tag for the stream, multiple tags can be specified",
			Destination: &tags,
		},
		&cli.BoolFlag{
			Name:        "stream-all",
			Aliases:     []string{"a"},
			EnvVars:     []string{"HATCHERY_STREAM_ALL"},
			Usage:       "Run all streams",
			Destination: &forAll,
		},

		&cli.TimestampFlag{
			Name:        "base-time",
			Aliases:     []string{"b"},
			EnvVars:     []string{"HATCHERY_BASE_TIME"},
			Usage:       "Base time to load data. Default is current time",
			Destination: &baseTime,
			Layout:      time.RFC3339,
		},

		&cli.TimestampFlag{
			Name:        "start-time",
			Category:    "Time Range",
			Aliases:     []string{"s"},
			EnvVars:     []string{"HATCHERY_START_TIME"},
			Usage:       "Start time to load data",
			Destination: &startTime,
			Layout:      time.RFC3339,
		},
		&cli.TimestampFlag{
			Name:        "end-time",
			Category:    "Time Range",
			Aliases:     []string{"e"},
			EnvVars:     []string{"HATCHERY_END_TIME"},
			Usage:       "End time to load data",
			Destination: &endTime,
			Layout:      time.RFC3339,
		},
		&cli.DurationFlag{
			Name:        "tick",
			Category:    "Time Range",
			Aliases:     []string{"d"},
			EnvVars:     []string{"HATCHERY_TICK"},
			Usage:       "Tick duration to load data",
			Value:       time.Minute,
			Destination: &tick,
		},
	}

	flags = append(flags, cfgLogging.Flags()...)

	app := &cli.App{
		Name:  "hatchery",
		Usage: "A tool to load log data from various sources for security",
		Flags: flags,

		Action: func(c *cli.Context) error {
			selectors := []Selector{}
			if forAll {
				selectors = append(selectors, SelectAll())
			}
			if len(tags.Value()) > 0 {
				selectors = append(selectors, SelectByTag(tags.Value()...))
			}
			if len(streamIDs.Value()) > 0 {
				selectors = append(selectors, SelectByID(streamIDs.Value()...))
			}

			var logger *slog.Logger
			if h.loggerIsDefault {
				newLogger, closer, err := cfgLogging.Build()
				if err != nil {
					return err
				}
				defer closer()
				logger = newLogger
				logger.Info("Logger is initialized", "config", cfgLogging)
			} else {
				logger = h.logger
				logger.Info("Logger is used from option")
			}

			ctx := logging.InjectCtx(c.Context, logger)

			if t := baseTime.Value(); t != nil && !t.IsZero() {
				ctx = timestamp.InjectCtx(ctx, *t)
			}

			return h.Run(ctx, selectors...)
		},
	}

	if err := app.Run(argv); err != nil {
		return goerr.Wrap(err, "failed to run CLI app")
	}

	return nil
}
