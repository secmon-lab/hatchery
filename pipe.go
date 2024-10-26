package hatchery

import (
	"context"
	"io"

	"github.com/m-mizutani/goerr"
	"github.com/secmon-lab/hatchery/pkg/metadata"
)

// Pipe is a struct that contains a destination. It is middle layer between source and destination and source function receives the Pipe object as the argument.
// This has a Spout method that outputs the data from the source to the destination.
//
// Example:
//
//	func someSource(ctx context.Context, p *hatchery.Pipe) error {
//	  var r io.ReadCloser
//	  //
//	  // Load data from somewhere	to r
//	  //
//	  defer r.Close()
//	  return p.Spout(ctx, r, metadata.MetaData{})
//	}
type Pipe struct {
	dst Destination
}

// NewPipe creates a new Pipe object with the destination. It is for testing.
func NewPipe(dst Destination) *Pipe {
	return &Pipe{dst: dst}
}

// Spout outputs the data from the source to the destination.
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
