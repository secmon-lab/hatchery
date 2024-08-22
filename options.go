package hatchery

import "github.com/m-mizutani/goerr"

func WithPipeline(id PipelineID, src Source, dst Destination) Option {
	return func(h *Hatchery) error {
		if _, ok := h.pipelines[id]; ok {
			return goerr.Wrap(ErrPipelineConflicted).With("id", id).Unstack()
		}

		p := &Pipeline{
			src: src,
			dst: dst,
		}
		h.pipelines[p.ID()] = p

		return nil
	}
}
