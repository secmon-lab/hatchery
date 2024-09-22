package hatchery

import (
	"github.com/m-mizutani/goerr"
	"github.com/urfave/cli/v2"
)

func (h *Hatchery) CLI(argv []string) error {

	var (
		streamIDs cli.StringSlice
	)

	app := &cli.App{
		Name:  "hatchery",
		Usage: "A tool to load log data from various sources for security",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:        "stream-id",
				Aliases:     []string{"s"},
				Usage:       "Target stream ID",
				Destination: &streamIDs,
			},
		},
		Action: func(c *cli.Context) error {
			return h.Run(c.Context, streamIDs.Value())
		},
	}

	if err := app.Run(argv); err != nil {
		return goerr.Wrap(err, "failed to run CLI app")
	}

	return nil
}
