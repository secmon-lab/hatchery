package hatchery

import (
	"context"

	"github.com/m-mizutani/goerr"
)

type StreamID string

type Streams []*Stream

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
	id  StreamID
	src Source
	dst Destination
}

func NewStream(id StreamID, src Source, dst Destination) *Stream {
	return &Stream{id: id, src: src, dst: dst}
}

// Run executes the stream, which invokes Source.Load and saves data via Destination.
func (x *Stream) Run(ctx context.Context) error {
	return x.src(ctx, NewPipe(x.dst))
}

// Validate checks the stream is valid or not.
func (x *Stream) Validate() error {
	if x.id == "" {
		return goerr.Wrap(ErrInvalidStream, "ID is not defined")
	}
	if x.src == nil {
		return goerr.Wrap(ErrInvalidStream, "source is not defined").With("id", x.id)
	}
	if x.dst == nil {
		return goerr.Wrap(ErrInvalidStream, "destination is not defined").With("id", x.id)
	}
	return nil
}
