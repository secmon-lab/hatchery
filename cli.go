package hatchery

import (
	"github.com/m-mizutani/goerr"
	"github.com/urfave/cli/v2"
)

func (h *Hatchery) CLI(argv []string) error {

	var (
		streamIDs cli.StringSlice
		tags      cli.StringSlice
		forAll    bool
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

			return h.Run(c.Context, selectors...)
		},
	}

	if err := app.Run(argv); err != nil {
		return goerr.Wrap(err, "failed to run CLI app")
	}

	return nil
}
