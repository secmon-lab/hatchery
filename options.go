package hatchery

import (
	"log/slog"
)

// WithStream is an option to add a new stream to the hatchery. The stream can be configured with a source and a destination. If the stream ID is conflicted with existing streams, it returns an error. The source and the destination can be implemented by the user. The multiple streams can be added to the hatchery.
func WithStream(id StreamID, src Source, dst Destination) Option {
	return func(h *Hatchery) {
		h.streams = append(h.streams, &Stream{
			id:  id,
			src: src,
			dst: dst,
		})
	}
}

// WithLogger is an option to set a logger to the hatchery. The logger is used to log messages from the hatchery. If the logger is nil, it uses the default logger.
func WithLogger(logger *slog.Logger) Option {
	return func(h *Hatchery) {
		h.logger = logger
	}
}
