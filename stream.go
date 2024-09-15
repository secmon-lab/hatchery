package hatchery

import (
	"context"

	"github.com/m-mizutani/goerr"
)

type StreamID string

type Streams []Stream

func (s Streams) Validate() error {
	for _, stream := range s {
		if err := stream.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Stream is a pipeline of data processing.
type Stream struct {
	ID  StreamID
	Src Source
	Dst Destination
}

// Run executes the stream, which invokes Source.Load and saves data via Destination.
func (x *Stream) Run(ctx context.Context) error {
	return x.Src(ctx, NewPipe(x.Dst))
}

// Validate checks the stream is valid or not.
func (x *Stream) Validate() error {
	if x.ID == "" {
		return goerr.Wrap(ErrInvalidStream, "ID is not defined")
	}
	if x.Src == nil {
		return goerr.Wrap(ErrInvalidStream, "source is not defined").With("id", x.ID)
	}
	if x.Dst == nil {
		return goerr.Wrap(ErrInvalidStream, "destination is not defined").With("id", x.ID)
	}
	return nil
}
