package hatchery

import (
	"github.com/m-mizutani/goerr"
	"github.com/urfave/cli/v2"
)

func (h *Hatchery) CLI(argv []string) error {

	var (
		targets cli.StringSlice
	)

	app := &cli.App{
		Name:  "hatchery",
		Usage: "A tool to load log data from various sources for security",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:        "target",
				Aliases:     []string{"t"},
				Usage:       "Target pipeline ID",
				Destination: &targets,
			},
		},
		Action: func(c *cli.Context) error {
			for _, target := range targets.Value() {
				if _, ok := h.pipelines[StreamID(target)]; !ok {
					return goerr.Wrap(ErrStreamNotFound).With("id", target).Unstack()
				}
			}

			return nil
		},
	}

	if err := app.Run(argv); err != nil {
		return goerr.Wrap(err, "failed to run CLI app")
	}

	return nil
}
