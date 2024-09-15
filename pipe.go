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

	if _, err = io.Copy(w, src); err != nil {
		_ = w.Close()
		return goerr.Wrap(err, "failed to copy data")
	}

	if err = w.Close(); err != nil {
		return goerr.Wrap(err, "failed to close destination")
	}

	return err
}
