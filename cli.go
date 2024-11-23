package hatchery

import (
	"context"
	"log/slog"

	"github.com/m-mizutani/goerr"
	"github.com/secmon-lab/hatchery/pkg/config"
	"github.com/secmon-lab/hatchery/pkg/logging"
	"github.com/secmon-lab/hatchery/pkg/timestamp"
	"github.com/urfave/cli/v3"
)

func (h *Hatchery) CLI(argv []string) error {

	var (
		streamIDs []string
		tags      []string
		forAll    bool

		cfgRange   config.Range
		cfgLogging config.Logging
	)

	flags := []cli.Flag{
		&cli.StringSliceFlag{
			Name:        "stream-id",
			Aliases:     []string{"i"},
			Sources:     cli.EnvVars("HATCHERY_STREAM_ID"),
			Usage:       "Target stream ID, multiple IDs can be specified",
			Destination: &streamIDs,
		},
		&cli.StringSliceFlag{
			Name:        "stream-tags",
			Aliases:     []string{"t"},
			Sources:     cli.EnvVars("HATCHERY_STREAM_TAGS"),
			Usage:       "Tag for the stream, multiple tags can be specified",
			Destination: &tags,
		},
		&cli.BoolFlag{
			Name:        "stream-all",
			Aliases:     []string{"a"},
			Sources:     cli.EnvVars("HATCHERY_STREAM_ALL"),
			Usage:       "Run all streams",
			Destination: &forAll,
		},
	}

	flags = append(flags, cfgLogging.Flags()...)
	flags = append(flags, cfgRange.Flags()...)

	app := &cli.Command{
		Name:  "hatchery",
		Usage: "A tool to load log data from various sources for security",
		Flags: flags,

		Before: func(ctx context.Context, _ *cli.Command) (context.Context, error) {
			var logger *slog.Logger
			if h.loggerIsDefault {
				newLogger, closer, err := cfgLogging.Build()
				if err != nil {
					return nil, err
				}
				defer closer()
				logger = newLogger
				logger.Info("Logger is initialized", "config", cfgLogging)
			} else {
				logger = h.logger
				logger.Info("Logger is used from option")
			}
			logging.SetDefault(logger)

			return logging.InjectCtx(ctx, logger), nil
		},

		Action: func(ctx context.Context, cmd *cli.Command) error {
			selectors := []Selector{}
			if forAll {
				selectors = append(selectors, SelectAll())
			}
			if len(tags) > 0 {
				selectors = append(selectors, SelectByTag(tags...))
			}
			if len(streamIDs) > 0 {
				selectors = append(selectors, SelectByID(streamIDs...))
			}

			if err := cfgRange.Validate(); err != nil {
				return err
			}

			for t := range cfgRange.Generate {
				ctx = timestamp.InjectCtx(ctx, t)
				logging.FromCtx(ctx).Info("Start to load data", "time", t)

				if err := h.Run(ctx, selectors...); err != nil {
					return goerr.Wrap(err, "failed to run Hatchery")
				}
			}
			return nil
		},
	}

	if err := app.Run(context.Background(), argv); err != nil {
		return goerr.Wrap(err, err.Error())
	}

	return nil
}
