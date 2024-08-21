package hatchery

import (
	"github.com/m-mizutani/goerr"
	"github.com/urfave/cli/v2"
)

func (h *Hatchery) CLI(argv []string) error {
	app := &cli.App{}

	if err := app.Run(argv); err != nil {
		return goerr.Wrap(err, "failed to run CLI app")
	}

	return nil
}
