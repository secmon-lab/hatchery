package hatchery

import (
	"log/slog"

	"github.com/m-mizutani/goerr"
)

// WithStream is an option to add a new stream to the hatchery. The stream can be configured with a source and a destination. If the stream ID is conflicted with existing streams, it returns an error. The source and the destination can be implemented by the user. The multiple streams can be added to the hatchery.
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

// WithLogger is an option to set a logger to the hatchery. The logger is used to log messages from the hatchery. If the logger is nil, it uses the default logger.
func WithLogger(logger *slog.Logger) Option {
	return func(h *Hatchery) error {
		h.logger = logger
		return nil
	}
}
