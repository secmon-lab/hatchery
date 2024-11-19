package hatchery

import (
	"context"
	"log/slog"
	"sync"

	"github.com/m-mizutani/goerr"
	"github.com/secmon-lab/hatchery/pkg/logging"
)

// Hatchery is a main manager of this tool.
type Hatchery struct {
	streams         Streams
	logger          *slog.Logger
	loggerIsDefault bool
}

type Option func(*Hatchery)

// New creates a new Hatchery instance.
func New(streams []*Stream, opts ...Option) *Hatchery {
	h := &Hatchery{
		logger:          logging.Default(),
		loggerIsDefault: true,
	}

	h.streams = streams

	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *Hatchery) Run(ctx context.Context, selectors ...Selector) error {
	targets := map[string]*Stream{}

	if err := h.streams.Validate(); err != nil {
		return goerr.Wrap(err, "failed to validate streams")
	}

	for _, stream := range h.streams {
		for _, selector := range selectors {
			if selector(stream) {
				targets[stream.id] = stream
			}
		}
	}

	if len(targets) == 0 {
		return goerr.Wrap(ErrNoStreamFound)
	}

	var wg sync.WaitGroup
	var errCh = make(chan error, len(targets))

	for _, s := range targets {
		wg.Add(1)
		go func(stream *Stream) {
			defer wg.Done()

			if err := stream.Run(ctx); err != nil {
				errCh <- goerr.Wrap(err, "pipeline failed").With("id", stream.id)
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

type Selector func(*Stream) bool

func SelectByTag(tags ...string) Selector {
	return func(s *Stream) bool {
		for _, tag := range tags {
			for _, sTag := range s.tags {
				if tag == sTag {
					return true
				}
			}
		}
		return false
	}
}

func SelectByID(ids ...string) Selector {
	return func(s *Stream) bool {
		for _, id := range ids {
			if s.id == id {
				return true
			}
		}
		return false
	}
}

func SelectAll() Selector {
	return func(s *Stream) bool {
		return true
	}
}
