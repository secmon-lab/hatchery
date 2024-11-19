package hatchery

import (
	"log/slog"

	"github.com/m-mizutani/goerr"
	"github.com/secmon-lab/hatchery/pkg/logging"
	"github.com/urfave/cli/v2"
)

func (h *Hatchery) CLI(argv []string) error {

	var (
		streamIDs cli.StringSlice
		tags      cli.StringSlice
		forAll    bool

		logFormat string
		logLevel  string
		logOut    string
	)

	app := &cli.App{
		Name:  "hatchery",
		Usage: "A tool to load log data from various sources for security",
		Flags: []cli.Flag{
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

			&cli.StringFlag{
				Name:        "log-format",
				Aliases:     []string{"f"},
				EnvVars:     []string{"HATCHERY_LOG_FORMAT"},
				Usage:       "Log format (json, text)",
				Value:       "json",
				Destination: &logFormat,
			},
			&cli.StringFlag{
				Name:        "log-level",
				Aliases:     []string{"l"},
				EnvVars:     []string{"HATCHERY_LOG_LEVEL"},
				Usage:       "Log level (debug, info, warn, error)",
				Value:       "info",
				Destination: &logLevel,
			},
			&cli.StringFlag{
				Name:        "log-out",
				Aliases:     []string{"o"},
				EnvVars:     []string{"HATCHERY_LOG_OUT"},
				Usage:       "Log output (stdout or stderr)",
				Value:       "stdout",
				Destination: &logOut,
			},
		},
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
				newLogger, err := buildLogger(logLevel, logFormat, logOut)
				if err != nil {
					return err
				}
				logger = newLogger
				logger.Info("Logger is initialized", "level", logLevel, "format", logFormat, "output", logOut)
			} else {
				logger = h.logger
				logger.Info("Logger is used from option")
			}

			ctx := logging.InjectCtx(c.Context, logger)

			return h.Run(ctx, selectors...)
		},
	}

	if err := app.Run(argv); err != nil {
		return goerr.Wrap(err, "failed to run CLI app")
	}

	return nil
}
