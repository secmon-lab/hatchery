package hatchery

import "github.com/m-mizutani/goerr"

func WithStream(id StreamID, src Source, dst Destination) Option {
	return func(h *Hatchery) error {
		if _, ok := h.pipelines[id]; ok {
			return goerr.Wrap(ErrStreamConflicted).With("id", id).Unstack()
		}

		p := &Stream{
			src: src,
			dst: dst,
		}
		h.pipelines[p.ID()] = p

		return nil
	}
}
