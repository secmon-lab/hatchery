package hatchery

import (
	"context"
	"sync"

	"github.com/m-mizutani/goerr"
)

// Hatchery is a main manager of this tool.
type Hatchery struct {
	pipelines map[PipelineID]*Pipeline
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

func (h *Hatchery) Run(ctx context.Context, pipelineIDs []string) error {
	for _, id := range pipelineIDs {
		if _, ok := h.pipelines[PipelineID(id)]; !ok {
			return ErrPipelineNotFound
		}
	}

	var wg sync.WaitGroup
	var errCh = make(chan error, len(pipelineIDs))

	for _, pID := range pipelineIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			if err := h.pipelines[PipelineID(id)].Run(ctx); err != nil {
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
