package hatchery

import (
	"context"
	"log/slog"
	"sync"

	"github.com/m-mizutani/goerr"
)

// Hatchery is a main manager of this tool.
type Hatchery struct {
	streams []*Stream
	logger  *slog.Logger
}

type Option func(*Hatchery)

// New creates a new Hatchery instance.
func New(opts ...Option) *Hatchery {
	h := &Hatchery{
		logger: slog.Default(),
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *Hatchery) Run(ctx context.Context, streamIDs []string) error {
	var streams []*Stream

	for _, id := range streamIDs {
		for _, stream := range h.streams {
			if stream.ID() == StreamID(id) {
				streams = append(streams, stream)
			}
		}
	}

	if len(streams) == 0 {
		return goerr.Wrap(ErrStreamNotFound).With("ids", streamIDs)
	}

	var wg sync.WaitGroup
	var errCh = make(chan error, len(streamIDs))

	for _, s := range streams {
		wg.Add(1)
		go func(stream *Stream) {
			defer wg.Done()

			if err := stream.Run(ctx); err != nil {
				errCh <- goerr.Wrap(err, "pipeline failed").With("id", stream.ID())
			}
		}(s)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		return err
	}

	return nil
}
