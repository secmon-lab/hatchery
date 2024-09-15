package hatchery

import (
	"context"
	"io"

	"github.com/m-mizutani/goerr"
	"github.com/secmon-as-code/hatchery/pkg/metadata"
)

type Pipe struct {
	dst Destination
}

func NewPipe(dst Destination) *Pipe {
	return &Pipe{dst: dst}
}

func (p *Pipe) Spout(ctx context.Context, src io.Reader, md metadata.MetaData) error {
	w, err := p.dst(ctx, md)
	if err != nil {
		return err
	}

	defer func() {
		err = w.Close()
		if err != nil {
			err = goerr.Wrap(err, "failed to close destination")
		}
	}()

	if _, err = io.Copy(w, src); err != nil {
		return goerr.Wrap(err, "failed to copy data")
	}

	return err
}
