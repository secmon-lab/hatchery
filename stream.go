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
func (p *Stream) ID() StreamID {
	return p.id
}

// Run executes the stream, which invokes Source.Load and saves data via Destination.
func (p *Stream) Run(ctx context.Context) error {
	return p.src.Load(ctx, p.dst)
}
