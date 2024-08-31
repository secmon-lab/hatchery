package hatchery

import (
	"context"
	"sync"

	"github.com/m-mizutani/goerr"
)

// Hatchery is a main manager of this tool.
type Hatchery struct {
	pipelines map[StreamID]*Stream
}

type Option func(*Hatchery) error

// New creates a new Hatchery instance.
func New(opts ...Option) (*Hatchery, error) {
	h := &Hatchery{}
	for _, opt := range opts {
		if err := opt(h); err != nil {
			return nil, err
		}
	}
	return h, nil
}

func (h *Hatchery) Run(ctx context.Context, streamIDs []string) error {
	for _, id := range streamIDs {
		if _, ok := h.pipelines[StreamID(id)]; !ok {
			return ErrStreamNotFound
		}
	}

	var wg sync.WaitGroup
	var errCh = make(chan error, len(streamIDs))

	for _, pID := range streamIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			if err := h.pipelines[StreamID(id)].Run(ctx); err != nil {
				errCh <- goerr.Wrap(err, "pipeline failed").With("id", id)
			}
		}(pID)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		return err
	}

	return nil
}
