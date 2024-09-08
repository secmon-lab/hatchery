package hatchery

import "context"

type StreamID string

// Stream is a pipeline of data processing.
type Stream struct {
	id  StreamID
	src Source
	dst Destination
}

// ID returns the ID of the stream.
func (x *Stream) ID() StreamID {
	return x.id
}

// Run executes the stream, which invokes Source.Load and saves data via Destination.
func (x *Stream) Run(ctx context.Context) error {
	return x.src(ctx, NewPipe(x.dst))
}
